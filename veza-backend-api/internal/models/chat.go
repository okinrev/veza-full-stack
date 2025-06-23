// internal/models/chat.go
package models

import (
	"database/sql"
	"time"
)

// Message represents a chat message (direct or room)
type Message struct {
	ID        int            `db:"id" json:"id"`
	FromUser  int            `db:"from_user" json:"from_user"`
	ToUser    sql.NullInt32  `db:"to_user" json:"to_user,omitempty"`     // For direct messages
	Room      sql.NullString `db:"room" json:"room,omitempty"`           // For room messages
	Content   string         `db:"content" json:"content"`
	IsRead    bool           `db:"is_read" json:"is_read"`
	EditedAt  sql.NullTime   `db:"edited_at" json:"edited_at,omitempty"`
	Timestamp time.Time      `db:"timestamp" json:"timestamp"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
}

// MessageWithUser represents a message with sender information
type MessageWithUser struct {
	Message
	FromUsername string         `db:"from_username" json:"from_username,omitempty"`
	FromAvatar   sql.NullString `db:"from_avatar" json:"from_avatar,omitempty"`
}

// Room represents a chat room
type Room struct {
	ID           int            `db:"id" json:"id"`
	Name         string         `db:"name" json:"name"`
	Description  sql.NullString `db:"description" json:"description,omitempty"`
	IsPrivate    bool           `db:"is_private" json:"is_private"`
	CreatorID    sql.NullInt32  `db:"creator_id" json:"creator_id,omitempty"`
	PasswordHash sql.NullString `db:"password_hash" json:"-"` // Never serialize password
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
}

// RoomWithDetails represents a room with additional information
type RoomWithDetails struct {
	Room
	CreatorName   sql.NullString `db:"creator_name" json:"creator_name,omitempty"`
	MemberCount   int            `db:"member_count" json:"member_count,omitempty"`
	OnlineCount   int            `db:"online_count" json:"online_count,omitempty"`
	LastActivity  sql.NullTime   `db:"last_activity" json:"last_activity,omitempty"`
	LastMessage   sql.NullString `db:"last_message" json:"last_message,omitempty"`
	UnreadCount   int            `db:"unread_count" json:"unread_count,omitempty"`
	IsMember      bool           `db:"is_member" json:"is_member,omitempty"`
	UserRole      sql.NullString `db:"user_role" json:"user_role,omitempty"` // owner, admin, member
}

// RoomMember represents room membership
type RoomMember struct {
	ID       int       `db:"id" json:"id"`
	RoomID   int       `db:"room_id" json:"room_id"`
	UserID   int       `db:"user_id" json:"user_id"`
	Role     string    `db:"role" json:"role"` // owner, admin, member
	JoinedAt time.Time `db:"joined_at" json:"joined_at"`
}

// RoomMemberWithUser represents room membership with user information
type RoomMemberWithUser struct {
	RoomMember
	Username     string         `db:"username" json:"username,omitempty"`
	Avatar       sql.NullString `db:"avatar" json:"avatar,omitempty"`
	LastSeen     sql.NullTime   `db:"last_seen" json:"last_seen,omitempty"`
	IsOnline     bool           `db:"is_online" json:"is_online,omitempty"`
	MessageCount int            `db:"message_count" json:"message_count,omitempty"`
}

// RoomMessage represents a message in a room (extends Message for room-specific features)
type RoomMessage struct {
	ID          int            `db:"id" json:"id"`
	RoomID      int            `db:"room_id" json:"room_id"`
	UserID      sql.NullInt32  `db:"user_id" json:"user_id,omitempty"` // Null for system messages
	Content     string         `db:"content" json:"content"`
	MessageType string         `db:"message_type" json:"message_type"` // message, join, leave, system
	EditedAt    sql.NullTime   `db:"edited_at" json:"edited_at,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
}

// RoomMessageWithUser represents a room message with user information
type RoomMessageWithUser struct {
	RoomMessage
	Username string         `db:"username" json:"username,omitempty"`
	Avatar   sql.NullString `db:"avatar" json:"avatar,omitempty"`
}

// ConversationSummary represents a summary of a direct message conversation
type ConversationSummary struct {
	UserID       int            `db:"user_id" json:"user_id"`
	Username     string         `db:"username" json:"username"`
	FirstName    sql.NullString `db:"first_name" json:"first_name,omitempty"`
	LastName     sql.NullString `db:"last_name" json:"last_name,omitempty"`
	Avatar       sql.NullString `db:"avatar" json:"avatar,omitempty"`
	LastMessage  sql.NullString `db:"last_message" json:"last_message,omitempty"`
	LastActivity sql.NullTime   `db:"last_activity" json:"last_activity,omitempty"`
	UnreadCount  int            `db:"unread_count" json:"unread_count"`
	IsOnline     bool           `db:"is_online" json:"is_online,omitempty"`
	LastSeen     sql.NullTime   `db:"last_seen" json:"last_seen,omitempty"`
}