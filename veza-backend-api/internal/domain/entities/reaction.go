package entities

import (
	"time"
)

// ReactionType représente le type de réaction
type ReactionType string

const (
	ReactionTypeLike       ReactionType = "like"
	ReactionTypeLove       ReactionType = "love"
	ReactionTypeHaha       ReactionType = "haha"
	ReactionTypeWow        ReactionType = "wow"
	ReactionTypeSad        ReactionType = "sad"
	ReactionTypeAngry      ReactionType = "angry"
	ReactionTypeThumbsUp   ReactionType = "thumbs_up"
	ReactionTypeThumbsDown ReactionType = "thumbs_down"
)

// Reaction représente une réaction sur un message
type Reaction struct {
	ID        uint         `json:"id" db:"id"`
	MessageID uint         `json:"message_id" db:"message_id"`
	UserID    uint         `json:"user_id" db:"user_id"`
	Type      ReactionType `json:"type" db:"type"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time   `json:"deleted_at,omitempty" db:"deleted_at"`
	Message   *Message     `json:"message,omitempty"`
	User      *User        `json:"user,omitempty"`
}

// ReactionCreateRequest représente la requête de création d'une réaction
type ReactionCreateRequest struct {
	MessageID uint         `json:"message_id" validate:"required"`
	Type      ReactionType `json:"type" validate:"required"`
}

// ReactionUpdateRequest représente la requête de mise à jour d'une réaction
type ReactionUpdateRequest struct {
	Type ReactionType `json:"type" validate:"required"`
}
