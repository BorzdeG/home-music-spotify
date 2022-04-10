package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	"gorm.io/gorm/clause"
	"music-spotify/db"
)

func syncTrack(track spotify.FullTrack) {
	log.Debugln("sync track:", track)
	log.Tracef("ExternalURLs: %+v\n", track.ExternalURLs)

	var uri string
	s, exist := track.ExternalURLs["spotify"]
	if exist {
		uri = s
	}
	trackEntity := &db.TrackEntity{
		ID:      track.ID.String(),
		Name:    track.Name,
		AlbumID: track.Album.ID.String(),
		Type:    track.Type,
		URI:     uri,
	}
	log.Tracef("track entity: %+v\n", trackEntity)

	if res := gormDB.Table(db.TableTrack).
		Clauses(
			clause.OnConflict{
				Columns: []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns(
					[]string{
						"name",
						"album_id",
						"type",
						"uri",
					},
				),
			},
		).
		Save(trackEntity); res.Error != nil {
		log.Panicln(res.Error)
	}
}
