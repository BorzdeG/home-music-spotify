package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	"gorm.io/gorm"
	"music-spotify/core"
	"music-spotify/db"
	"music-spotify/spotify-client"
)

var ctx context.Context
var spotifyClient *spotify.Client
var spotifyUser *spotify.PrivateUser

var gormDB *gorm.DB

func main() {
	core.InitLog()

	gormDB = db.NewGorm()

	spotifyClient, ctx = spotify_client.NewSpotifyClient()

	getCurrentUser()
	syncPlaylists()
	syncLikesPlaylist()
}

func getCurrentUser() {
	user, err := spotifyClient.CurrentUser(ctx)
	if err != nil {
		log.Panicln("Error getting current user:", err)
	}
	log.Debugln("You are logged in as:", user.Email)
	spotifyUser = user
}
