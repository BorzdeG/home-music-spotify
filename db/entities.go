package db

import "time"

const (
	TablePlaylist = "music-spotify.playlist"
	TableTrack    = "music-spotify.track"

	TablePlaylistTrack = "music-spotify.playlist_track"
)

type PlaylistEntity struct {
	ID   string `gorm:"column:id;primaryKey"`
	Name string `gorm:"column:name"`
}

type TrackEntity struct {
	ID      string `gorm:"column:id;primaryKey"`
	Name    string `gorm:"column:name"`
	AlbumID string `gorm:"column:album_id"`
	Type    string `gorm:"column:type"`

	URI        string `gorm:"column:uri"`
	IsDownload bool   `gorm:"column:is_download"`
}

type PlaylistTrackEntity struct {
	PlaylistID string    `gorm:"column:playlist_id;primaryKey"`
	TrackID    string    `gorm:"column:track_id;primaryKey"`
	AddedAt    time.Time `gorm:"column:added_at"`
	Sync       bool      `gorm:"column:sync"`
}
