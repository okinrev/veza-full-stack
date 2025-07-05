package entities

import (
	"time"
)

// RoomType représente le type de salle de chat
type RoomType string

const (
	RoomTypeDirect  RoomType = "direct"
	RoomTypeGroup   RoomType = "group"
	RoomTypeChannel RoomType = "channel"
	RoomTypeSupport RoomType = "support"
)

// RoomStatus représente le statut d'une salle
type RoomStatus string

const (
	RoomStatusActive   RoomStatus = "active"
	RoomStatusInactive RoomStatus = "inactive"
	RoomStatusArchived RoomStatus = "archived"
)

// RoomPrivacy représente le niveau de confidentialité d'une salle
type RoomPrivacy string

const (
	RoomPrivacyPublic  RoomPrivacy = "public"
	RoomPrivacyPrivate RoomPrivacy = "private"
	RoomPrivacySecret  RoomPrivacy = "secret"
)

// Room représente une salle de chat
type Room struct {
	ID          uint        `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	Type        RoomType    `json:"type" db:"type"`
	Status      RoomStatus  `json:"status" db:"status"`
	Privacy     RoomPrivacy `json:"privacy" db:"privacy"`
	CreatedBy   uint        `json:"created_by" db:"created_by"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`
	Users       []User      `json:"users,omitempty"`
	Messages    []Message   `json:"messages,omitempty"`
	Creator     *User       `json:"creator,omitempty"`
}

// RoomCreateRequest représente la requête de création d'une salle
type RoomCreateRequest struct {
	Name        string   `json:"name" validate:"required,min=1,max=100"`
	Description string   `json:"description" validate:"omitempty,max=500"`
	Type        RoomType `json:"type" validate:"required"`
	UserIDs     []uint   `json:"user_ids" validate:"required,min=1"`
}

// RoomUpdateRequest représente la requête de mise à jour d'une salle
type RoomUpdateRequest struct {
	Name        *string     `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string     `json:"description,omitempty" validate:"omitempty,max=500"`
	Status      *RoomStatus `json:"status,omitempty"`
}

// RoomMember représente un membre d'une salle
type RoomMember struct {
	RoomID   uint      `json:"room_id" db:"room_id"`
	UserID   uint      `json:"user_id" db:"user_id"`
	Role     string    `json:"role" db:"role"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
	User     *User     `json:"user,omitempty"`
}
