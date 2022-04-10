package main

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	"gorm.io/gorm/clause"
	"music-spotify/db"
)

func syncPlaylists() {
	log.Infoln("Playlist sync...")

	playlistPage, err := spotifyClient.CurrentUsersPlaylists(ctx)
	if err != nil {
		log.Panicln(err)
	}

	for page := 1; ; page++ {
		log.Debugf(
			"page: %d, playlists: {Total: %d, Offset: %d, Limit: %d}\n",
			page,
			playlistPage.Total, playlistPage.Offset, playlistPage.Limit,
		)

		for _, playlist := range playlistPage.Playlists {
			syncPlaylist(playlist)
		}

		err := spotifyClient.NextPage(ctx, playlistPage)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			log.Panicln(err)
		}
	}

	log.Infoln("Playlist sync completed.")
}

func syncLikesPlaylist() {
	log.Infoln("sync likes playlist")

	playlistID := spotify.ID(spotifyUser.ID)

	storePlaylist2DB(
		&db.PlaylistEntity{
			ID:   playlistID.String(),
			Name: fmt.Sprintf("Liked Songs for user:%s", playlistID),
		},
	)

	savedTrackPage, err := spotifyClient.CurrentUsersTracks(ctx)
	if err != nil {
		log.Panicln("Failed to get track list for playlist:", playlistID)
	}

	beforeSyncPlaylistTracks(playlistID)
	for page := 1; ; page++ {
		log.Debugf(
			"page: %d, tracks: {Total: %d, Offset: %d, Limit: %d}\n",
			page,
			savedTrackPage.Total, savedTrackPage.Offset, savedTrackPage.Limit,
		)

		for _, savedTrack := range savedTrackPage.Tracks {
			syncPlaylistTrack(playlistID, savedTrack.FullTrack, savedTrack.AddedAt)
		}

		err := spotifyClient.NextPage(ctx, savedTrackPage)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			log.Panicln(err)
		}
	}
	afterSyncPlaylistTracks(playlistID)
}

func syncPlaylist(playlist spotify.SimplePlaylist) {
	playlistID := playlist.ID
	playlistName := playlist.Name

	log.Infoln("sync playlist:", playlistName)
	log.Debugf("playlist tracks: %d\n", playlist.Tracks.Total)

	storePlaylist2DB(
		&db.PlaylistEntity{
			ID:   playlistID.String(),
			Name: playlistName,
		},
	)

	if playlist.Owner.ID == spotifyUser.ID {
		syncPlaylistTracks(playlistID)
	}
}

func storePlaylist2DB(playlistEntity *db.PlaylistEntity) {
	if res := gormDB.Table(db.TablePlaylist).
		Clauses(clause.OnConflict{DoNothing: true}).
		Save(playlistEntity); res.Error != nil {
		log.Errorln("Playlist save error:", res.Error)
	}
}

func syncPlaylistTracks(playlistID spotify.ID) {
	playlistTrackPage, err := spotifyClient.GetPlaylistTracks(ctx, playlistID)
	if err != nil {
		log.Panicln("Failed to get track list for playlist:", playlistID)
	}

	beforeSyncPlaylistTracks(playlistID)
	for page := 1; ; page++ {
		log.Debugf(
			"page: %d, tracks: {Total: %d, Offset: %d, Limit: %d}\n",
			page,
			playlistTrackPage.Total, playlistTrackPage.Offset, playlistTrackPage.Limit,
		)

		for _, playlistTrack := range playlistTrackPage.Tracks {
			syncPlaylistTrack(playlistID, playlistTrack.Track, playlistTrack.AddedAt)
		}

		err := spotifyClient.NextPage(ctx, playlistTrackPage)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			log.Panicln(err)
		}
	}
	afterSyncPlaylistTracks(playlistID)
}
