package entities

import (
	"time"
)

// FileType représente le type de fichier
type FileType string

const (
	FileTypeImage    FileType = "image"
	FileTypeVideo    FileType = "video"
	FileTypeAudio    FileType = "audio"
	FileTypeDocument FileType = "document"
	FileTypeArchive  FileType = "archive"
	FileTypeOther    FileType = "other"
)

// FileStatus représente le statut d'un fichier
type FileStatus string

const (
	FileStatusUploading  FileStatus = "uploading"
	FileStatusUploaded   FileStatus = "uploaded"
	FileStatusProcessing FileStatus = "processing"
	FileStatusReady      FileStatus = "ready"
	FileStatusFailed     FileStatus = "failed"
	FileStatusDeleted    FileStatus = "deleted"
)

// File représente un fichier dans le système
type File struct {
	ID           uint                   `json:"id" db:"id"`
	UserID       uint                   `json:"user_id" db:"user_id"`
	Name         string                 `json:"name" db:"name"`
	OriginalName string                 `json:"original_name" db:"original_name"`
	Type         FileType               `json:"type" db:"type"`
	Status       FileStatus             `json:"status" db:"status"`
	Size         int64                  `json:"size" db:"size"`
	MimeType     string                 `json:"mime_type" db:"mime_type"`
	URL          string                 `json:"url" db:"url"`
	ThumbnailURL string                 `json:"thumbnail_url" db:"thumbnail_url"`
	Hash         string                 `json:"hash" db:"hash"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time             `json:"deleted_at,omitempty" db:"deleted_at"`
	User         *User                  `json:"user,omitempty"`
}

// FileCreateRequest représente la requête de création d'un fichier
type FileCreateRequest struct {
	Name         string                 `json:"name" validate:"required"`
	OriginalName string                 `json:"original_name" validate:"required"`
	Type         FileType               `json:"type" validate:"required"`
	Size         int64                  `json:"size" validate:"required,min=1"`
	MimeType     string                 `json:"mime_type" validate:"required"`
	URL          string                 `json:"url" validate:"required,url"`
	ThumbnailURL string                 `json:"thumbnail_url,omitempty" validate:"omitempty,url"`
	Hash         string                 `json:"hash" validate:"required"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// FileUpdateRequest représente la requête de mise à jour d'un fichier
type FileUpdateRequest struct {
	Name         *string                `json:"name,omitempty"`
	Status       *FileStatus            `json:"status,omitempty"`
	ThumbnailURL *string                `json:"thumbnail_url,omitempty" validate:"omitempty,url"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// FileFilters représente les filtres pour la recherche de fichiers
type FileFilters struct {
	Type       FileType   `json:"type,omitempty"`
	Status     FileStatus `json:"status,omitempty"`
	UserID     *uint      `json:"user_id,omitempty"`
	SearchTerm string     `json:"search_term,omitempty"`
	MinSize    *int64     `json:"min_size,omitempty"`
	MaxSize    *int64     `json:"max_size,omitempty"`
}
