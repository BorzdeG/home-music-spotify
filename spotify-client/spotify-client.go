package spotify_client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const (
	tokenFile    = `.spotify_token.json`
	listenPort   = 8888
	callbackPath = `/callback`
)

var oauthScopes = []string{
	spotifyauth.ScopeUserReadEmail,
	spotifyauth.ScopeUserReadPrivate,
	spotifyauth.ScopePlaylistReadPrivate,
	spotifyauth.ScopePlaylistReadCollaborative,
	spotifyauth.ScopeUserLibraryRead,
}

var redirectURI = url.URL{
	Scheme: `http`,
	Host:   fmt.Sprintf("%s:%d", "localhost", listenPort),
	Path:   callbackPath,
}

type SpotifyClientBuilder struct {
	auth       *spotifyauth.Authenticator
	ch         chan *spotify.Client
	oauthState string
	ctx        context.Context
}

func NewSpotifyClient() (*spotify.Client, context.Context) {

	ctx := context.Background()

	builder := SpotifyClientBuilder{
		ctx:        ctx,
		oauthState: os.Getenv("SPOTIFY_AUTH_STATE"),
		auth: spotifyauth.New(
			spotifyauth.WithRedirectURL(redirectURI.String()),
			spotifyauth.WithScopes(oauthScopes...),
		),
		ch: make(chan *spotify.Client),
	}

	// TODO Попробовать пройти повторную oauth аутентификацию в автоматическом режиме, чтобы не сохранять токен
	client := builder.getClientWithToken()
	if client == nil {
		client = builder.getClientWithOauth()
	}

	return client, ctx
}

func (receiver *SpotifyClientBuilder) getClientWithOauth() *spotify.Client {
	http.HandleFunc(callbackPath, receiver.completeAuth)
	http.HandleFunc(
		"/", func(writer http.ResponseWriter, request *http.Request) {
			log.Infoln("Got request for:", request.URL.String())
		},
	)
	go func() {
		err := http.ListenAndServe(redirectURI.Host, nil)
		if err != nil {
			log.Panicln(err)
		}
	}()

	authURL := receiver.auth.AuthURL(receiver.oauthState)
	log.Infoln("Please log in to Spotify by visiting the following page in your browser:", authURL)

	client := <-receiver.ch

	return client
}

func (receiver *SpotifyClientBuilder) completeAuth(writer http.ResponseWriter, request *http.Request) {
	token, err := receiver.auth.Token(request.Context(), receiver.oauthState, request)
	if err != nil {
		http.Error(writer, `Couldn't get token`, http.StatusForbidden)
		log.Panicln(err)
	}
	log.Tracef("token: %+v\n", token)
	receiver.storeToken(token)

	if st := request.FormValue("state"); st != receiver.oauthState {
		http.NotFound(writer, request)
		log.Panicf("State mismatch: %s != %s\n", st, receiver.oauthState)
	}

	client := spotify.New(receiver.auth.Client(request.Context(), token))

	user, err := client.CurrentUser(receiver.ctx)
	if err != nil {
		log.Errorln("Error getting current user:", err)
	}
	log.Infoln("Spotify login completed as:", user.DisplayName)

	receiver.ch <- client
}

func (receiver *SpotifyClientBuilder) getClientWithToken() *spotify.Client {
	token := receiver.readToken()
	if token == nil {
		log.Debugln("Missing or failed to use stored token...")
		return nil
	}

	client := spotify.New(receiver.auth.Client(receiver.ctx, token))

	user, err := client.CurrentUser(receiver.ctx)
	if err != nil {
		log.Traceln("Failed to use saved token:", err)
		return nil
	}
	log.Infoln("You are logged in as:", user.Email)

	return client
}

func (receiver *SpotifyClientBuilder) storeToken(token *oauth2.Token) {
	bytes, err := json.Marshal(token)
	if err != nil {
		log.Panicln("Failed to serialize token:", err)
	}

	err = ioutil.WriteFile(tokenFile, bytes, 0644)
	if err != nil {
		log.Panicln("Failed to save token:", err)
	}
}

func (receiver *SpotifyClientBuilder) readToken() *oauth2.Token {
	_, err := os.Stat(tokenFile)
	if errors.Is(err, os.ErrNotExist) {
		log.Debugln("Saved token missing:", err)
		return nil
	}

	bytes, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		log.Panicln("Failed to read token:", err)
	}

	var token *oauth2.Token
	err = json.Unmarshal(bytes, &token)
	if err != nil {
		log.Panicln("Failed to deserialize token:", err)
	}

	return token
}
