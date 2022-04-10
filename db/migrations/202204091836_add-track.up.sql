CREATE TABLE IF NOT EXISTS "music-spotify"."track"
(
   "id"          varchar(62)  NOT NULL PRIMARY KEY,
   "name"        varchar(255) NOT NULL,
   "album_id"    varchar(62),
   "type"        varchar(255),
   "uri"         varchar(255) NOT NULL,
   "is_download" boolean      NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS "music-spotify"."playlist_track"
(
   "playlist_id" varchar(62) NOT NULL
      CONSTRAINT "playlist_track-playlist_id-fk"
         REFERENCES "music-spotify"."playlist",
   "track_id"    varchar(62) NOT NULL
      CONSTRAINT "playlist_track-track_id-fk"
         REFERENCES "music-spotify"."track",
   "added_at"    timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP,
   "sync"        boolean     NOT NULL DEFAULT FALSE,
   PRIMARY KEY ("playlist_id", "track_id")
);
