package entities

import (
	"time"
)

// MessageType représente le type de message
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeFile     MessageType = "file"
	MessageTypeLocation MessageType = "location"
	MessageTypeSystem   MessageType = "system"
)

// MessageStatus représente le statut d'un message
type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
)

// Message représente un message dans une conversation
type Message struct {
	ID        uint                   `json:"id" db:"id"`
	RoomID    uint                   `json:"room_id" db:"room_id"`
	UserID    uint                   `json:"user_id" db:"user_id"`
	Type      MessageType            `json:"type" db:"type"`
	Content   string                 `json:"content" db:"content"`
	Status    MessageStatus          `json:"status" db:"status"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time             `json:"deleted_at,omitempty" db:"deleted_at"`
	Room      *Room                  `json:"room,omitempty"`
	User      *User                  `json:"user,omitempty"`
}

// MessageCreateRequest représente la requête de création d'un message
type MessageCreateRequest struct {
	RoomID   uint                   `json:"room_id" validate:"required"`
	Type     MessageType            `json:"type" validate:"required"`
	Content  string                 `json:"content" validate:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// MessageUpdateRequest représente la requête de mise à jour d'un message
type MessageUpdateRequest struct {
	Content  *string                `json:"content,omitempty"`
	Status   *MessageStatus         `json:"status,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
