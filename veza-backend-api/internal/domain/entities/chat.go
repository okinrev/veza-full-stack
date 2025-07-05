package entities

import (
	"time"
)

// ChatType représente le type de chat
type ChatType string

const (
	ChatTypeDirect  ChatType = "direct"
	ChatTypeGroup   ChatType = "group"
	ChatTypeChannel ChatType = "channel"
	ChatTypeSupport ChatType = "support"
)

// ChatStatus représente le statut d'un chat
type ChatStatus string

const (
	ChatStatusActive   ChatStatus = "active"
	ChatStatusInactive ChatStatus = "inactive"
	ChatStatusArchived ChatStatus = "archived"
)

// Chat représente un chat dans le système
type Chat struct {
	ID          uint       `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	Type        ChatType   `json:"type" db:"type"`
	Status      ChatStatus `json:"status" db:"status"`
	CreatedBy   uint       `json:"created_by" db:"created_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	Users       []User     `json:"users,omitempty"`
	Messages    []Message  `json:"messages,omitempty"`
	Creator     *User      `json:"creator,omitempty"`
}

// ChatCreateRequest représente la requête de création d'un chat
type ChatCreateRequest struct {
	Name        string   `json:"name" validate:"required,min=1,max=100"`
	Description string   `json:"description" validate:"omitempty,max=500"`
	Type        ChatType `json:"type" validate:"required"`
	UserIDs     []uint   `json:"user_ids" validate:"required,min=1"`
}

// ChatUpdateRequest représente la requête de mise à jour d'un chat
type ChatUpdateRequest struct {
	Name        *string     `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string     `json:"description,omitempty" validate:"omitempty,max=500"`
	Status      *ChatStatus `json:"status,omitempty"`
}

// ChatMember représente un membre d'un chat
type ChatMember struct {
	ChatID   uint      `json:"chat_id" db:"chat_id"`
	UserID   uint      `json:"user_id" db:"user_id"`
	Role     string    `json:"role" db:"role"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
	User     *User     `json:"user,omitempty"`
}
