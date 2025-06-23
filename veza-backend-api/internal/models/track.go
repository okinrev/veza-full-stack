// internal/models/track.go
package models

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// Track represents an audio track in the system
type Track struct {
	ID              int            `db:"id" json:"id"`
	Title           string         `db:"title" json:"title"`
	Artist          string         `db:"artist" json:"artist"`
	Filename        string         `db:"filename" json:"filename"`
	DurationSeconds sql.NullInt32  `db:"duration_seconds" json:"duration_seconds,omitempty"`
	Tags            pq.StringArray `db:"tags" json:"tags"`
	IsPublic        bool           `db:"is_public" json:"is_public"`
	UploaderID      int            `db:"uploader_id" json:"uploader_id"`
	CreatedAt       time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at" json:"updated_at"`
}

// TrackWithUploader represents a track with uploader information
type TrackWithUploader struct {
	Track
	UploaderUsername string         `db:"uploader_username" json:"uploader_username,omitempty"`
	UploaderAvatar   sql.NullString `db:"uploader_avatar" json:"uploader_avatar,omitempty"`
}

// TrackResponse represents track data with computed fields for API responses
type TrackResponse struct {
	ID              int            `json:"id"`
	Title           string         `json:"title"`
	Artist          string         `json:"artist"`
	Filename        string         `json:"filename"`
	DurationSeconds sql.NullInt32  `json:"duration_seconds,omitempty"`
	Tags            []string       `json:"tags"`
	IsPublic        bool           `json:"is_public"`
	UploaderID      int            `json:"uploader_id"`
	UploaderName    string         `json:"uploader_name,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	StreamURL       string         `json:"stream_url,omitempty"`
}