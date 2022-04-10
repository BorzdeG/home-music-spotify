package main

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	"gorm.io/gorm/clause"
	"music-spotify/db"
)

const (
	// language=PostgreSQL
	sqlClearSyncFlag = `UPDATE "music-spotify"."playlist_track" SET "sync"=FALSE WHERE "playlist_id"=?`
	// language=PostgreSQL
	sqlDropNotSyncPlaylistTracks = `DELETE FROM "music-spotify"."playlist_track" WHERE "playlist_id"=? AND "sync"=FALSE`
)

func beforeSyncPlaylistTracks(playlistID spotify.ID) {
	if res := gormDB.Exec(sqlClearSyncFlag, playlistID.String()); res.Error != nil {
		log.Panicln(res.Error)
	}
}
func afterSyncPlaylistTracks(playlistID spotify.ID) {
	if res := gormDB.Exec(sqlDropNotSyncPlaylistTracks, playlistID.String()); res.Error != nil {
		log.Panicln(res.Error)
	}
}

func syncPlaylistTrack(playlistID spotify.ID, track spotify.FullTrack, playlistAddedAt string) {
	log.Debugf("sync track for playlist `%s`: %s\n", playlistID, track)

	syncTrack(track)

	trackID := track.ID
	addedAt, err := time.Parse(time.RFC3339, playlistAddedAt)
	if err != nil {
		log.Panicln(err)
	}

	playlistTrackEntity := &db.PlaylistTrackEntity{
		PlaylistID: playlistID.String(),
		TrackID:    trackID.String(),
		AddedAt:    addedAt,
		Sync:       true,
	}

	if res := gormDB.Table(db.TablePlaylistTrack).
		Clauses(
			clause.OnConflict{
				Columns: []clause.Column{
					{Name: "playlist_id"},
					{Name: "track_id"},
				},
				DoUpdates: clause.AssignmentColumns(
					[]string{
						"added_at",
						"sync",
					},
				),
			},
		).
		Save(playlistTrackEntity); res.Error != nil {
		log.Panicln(res.Error)
	}
}
