package entities

import (
	"time"
)

// TrackStatus représente le statut d'une piste
type TrackStatus string

const (
	TrackStatusActive   TrackStatus = "active"
	TrackStatusInactive TrackStatus = "inactive"
	TrackStatusDeleted  TrackStatus = "deleted"
)

// TrackType représente le type de piste
type TrackType string

const (
	TrackTypeAudio   TrackType = "audio"
	TrackTypeVideo   TrackType = "video"
	TrackTypePodcast TrackType = "podcast"
	TrackTypeLive    TrackType = "live"
)

// Track représente une piste audio/vidéo
type Track struct {
	ID           uint        `json:"id" db:"id"`
	UserID       uint        `json:"user_id" db:"user_id"`
	Title        string      `json:"title" db:"title"`
	Description  string      `json:"description" db:"description"`
	Type         TrackType   `json:"type" db:"type"`
	Status       TrackStatus `json:"status" db:"status"`
	Duration     int         `json:"duration" db:"duration"` // en secondes
	FileURL      string      `json:"file_url" db:"file_url"`
	ThumbnailURL string      `json:"thumbnail_url" db:"thumbnail_url"`
	Genre        string      `json:"genre" db:"genre"`
	Tags         []string    `json:"tags" db:"tags"`
	PlayCount    int         `json:"play_count" db:"play_count"`
	LikeCount    int         `json:"like_count" db:"like_count"`
	ShareCount   int         `json:"share_count" db:"share_count"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`
	User         *User       `json:"user,omitempty"`
}

// TrackCreateRequest représente la requête de création d'une piste
type TrackCreateRequest struct {
	Title        string    `json:"title" validate:"required,min=1,max=200"`
	Description  string    `json:"description" validate:"omitempty,max=2000"`
	Type         TrackType `json:"type" validate:"required"`
	Duration     int       `json:"duration" validate:"required,min=1"`
	FileURL      string    `json:"file_url" validate:"required,url"`
	ThumbnailURL string    `json:"thumbnail_url" validate:"omitempty,url"`
	Genre        string    `json:"genre" validate:"omitempty"`
	Tags         []string  `json:"tags,omitempty"`
}

// TrackUpdateRequest représente la requête de mise à jour d'une piste
type TrackUpdateRequest struct {
	Title        *string      `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description  *string      `json:"description,omitempty" validate:"omitempty,max=2000"`
	Type         *TrackType   `json:"type,omitempty"`
	Status       *TrackStatus `json:"status,omitempty"`
	Duration     *int         `json:"duration,omitempty" validate:"omitempty,min=1"`
	FileURL      *string      `json:"file_url,omitempty" validate:"omitempty,url"`
	ThumbnailURL *string      `json:"thumbnail_url,omitempty" validate:"omitempty,url"`
	Genre        *string      `json:"genre,omitempty"`
	Tags         []string     `json:"tags,omitempty"`
}

// TrackFilters représente les filtres pour la recherche de pistes
type TrackFilters struct {
	Type        TrackType   `json:"type,omitempty"`
	Genre       string      `json:"genre,omitempty"`
	Status      TrackStatus `json:"status,omitempty"`
	UserID      *uint       `json:"user_id,omitempty"`
	SearchTerm  string      `json:"search_term,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	MinDuration *int        `json:"min_duration,omitempty"`
	MaxDuration *int        `json:"max_duration,omitempty"`
}
