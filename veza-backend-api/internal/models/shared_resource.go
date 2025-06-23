// internal/models/shared_resource.go
package models

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// SharedResource represents a shared file/resource in the system
type SharedResource struct {
	ID            int            `db:"id" json:"id"`
	Title         string         `db:"title" json:"title"`
	Description   sql.NullString `db:"description" json:"description,omitempty"`
	Filename      string         `db:"filename" json:"filename"`
	URL           string         `db:"url" json:"url"`
	Type          string         `db:"type" json:"type"` // sample, preset, plugin, template, midi, document
	Tags          pq.StringArray `db:"tags" json:"tags"`
	UploaderID    int            `db:"uploader_id" json:"uploader_id"`
	IsPublic      bool           `db:"is_public" json:"is_public"`
	DownloadCount int            `db:"download_count" json:"download_count"`
	UploadedAt    time.Time      `db:"uploaded_at" json:"uploaded_at"`
	UpdatedAt     time.Time      `db:"updated_at" json:"updated_at"`
}

// SharedResourceWithUploader represents a shared resource with uploader information
type SharedResourceWithUploader struct {
	SharedResource
	UploaderUsername string         `db:"uploader_username" json:"uploader_username,omitempty"`
	UploaderAvatar   sql.NullString `db:"uploader_avatar" json:"uploader_avatar,omitempty"`
}

// SharedResourceResponse represents shared resource data for API responses
type SharedResourceResponse struct {
	ID               int           `json:"id"`
	Title            string        `json:"title"`
	Description      sql.NullString `json:"description,omitempty"`
	Filename         string        `json:"filename"`
	URL              string        `json:"url"`
	Type             string        `json:"type"`
	Tags             []string      `json:"tags"`
	UploaderID       int           `json:"uploader_id"`
	UploaderUsername string        `json:"uploader_username,omitempty"`
	IsPublic         bool          `json:"is_public"`
	DownloadCount    int           `json:"download_count"`
	UploadedAt       time.Time     `json:"uploaded_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	DownloadURL      string        `json:"download_url,omitempty"`
}