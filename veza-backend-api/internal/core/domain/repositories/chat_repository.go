package repositories

import (
	"context"
	"time"

	"github.com/okinrev/veza-web-app/internal/core/domain/entities"
)

// ChatRepository définit les opérations de persistence pour le système de chat
type ChatRepository interface {
	// ============================================================================
	// GESTION DES ROOMS
	// ============================================================================

	// CreateRoom crée une nouvelle room de chat
	CreateRoom(ctx context.Context, room *entities.ChatRoom) error

	// GetRoomByID récupère une room par ID
	GetRoomByID(ctx context.Context, roomID int64) (*entities.ChatRoom, error)

	// GetRoomBySlug récupère une room par son slug
	GetRoomBySlug(ctx context.Context, slug string) (*entities.ChatRoom, error)

	// UpdateRoom met à jour une room
	UpdateRoom(ctx context.Context, room *entities.ChatRoom) error

	// DeleteRoom supprime une room (soft delete)
	DeleteRoom(ctx context.Context, roomID int64) error

	// ListRooms liste les rooms avec pagination et filtres
	ListRooms(ctx context.Context, filters RoomFilters) ([]*entities.ChatRoom, error)

	// SearchRooms recherche des rooms par nom/description
	SearchRooms(ctx context.Context, query string, limit, offset int) ([]*entities.ChatRoom, error)

	// GetUserRooms récupère les rooms d'un utilisateur
	GetUserRooms(ctx context.Context, userID int64, limit, offset int) ([]*entities.ChatRoom, error)

	// GetPublicRooms récupère les rooms publiques
	GetPublicRooms(ctx context.Context, limit, offset int) ([]*entities.ChatRoom, error)

	// GetRoomStats récupère les statistiques d'une room
	GetRoomStats(ctx context.Context, roomID int64) (*RoomStats, error)

	// ============================================================================
	// GESTION DES MEMBRES
	// ============================================================================

	// JoinRoom ajoute un utilisateur à une room
	JoinRoom(ctx context.Context, roomID, userID int64) error

	// LeaveRoom retire un utilisateur d'une room
	LeaveRoom(ctx context.Context, roomID, userID int64) error

	// KickUser expulse un utilisateur d'une room
	KickUser(ctx context.Context, roomID, userID, moderatorID int64, reason string) error

	// BanUser bannit un utilisateur d'une room
	BanUser(ctx context.Context, roomID, userID, moderatorID int64, reason string, duration *time.Duration) error

	// UnbanUser débannit un utilisateur
	UnbanUser(ctx context.Context, roomID, userID, moderatorID int64) error

	// GetRoomMembers récupère les membres d'une room
	GetRoomMembers(ctx context.Context, roomID int64, limit, offset int) ([]*RoomMember, error)

	// GetRoomModerators récupère les modérateurs d'une room
	GetRoomModerators(ctx context.Context, roomID int64) ([]*RoomMember, error)

	// IsUserInRoom vérifie si un utilisateur est dans une room
	IsUserInRoom(ctx context.Context, roomID, userID int64) (bool, error)

	// IsUserBanned vérifie si un utilisateur est banni d'une room
	IsUserBanned(ctx context.Context, roomID, userID int64) (bool, error)

	// GetUserRoomRole récupère le rôle d'un utilisateur dans une room
	GetUserRoomRole(ctx context.Context, roomID, userID int64) (entities.RoomRole, error)

	// UpdateUserRoomRole met à jour le rôle d'un utilisateur dans une room
	UpdateUserRoomRole(ctx context.Context, roomID, userID int64, role entities.RoomRole) error

	// ============================================================================
	// GESTION DES MESSAGES
	// ============================================================================

	// CreateMessage crée un nouveau message
	CreateMessage(ctx context.Context, message *entities.ChatMessage) error

	// GetMessageByID récupère un message par ID
	GetMessageByID(ctx context.Context, messageID int64) (*entities.ChatMessage, error)

	// UpdateMessage met à jour un message
	UpdateMessage(ctx context.Context, message *entities.ChatMessage) error

	// DeleteMessage supprime un message (soft delete)
	DeleteMessage(ctx context.Context, messageID, userID int64) error

	// GetRoomMessages récupère les messages d'une room avec pagination
	GetRoomMessages(ctx context.Context, roomID int64, limit, offset int) ([]*entities.ChatMessage, error)

	// GetRoomMessagesAfter récupère les messages après un timestamp
	GetRoomMessagesAfter(ctx context.Context, roomID int64, after time.Time, limit int) ([]*entities.ChatMessage, error)

	// GetRoomMessagesBefore récupère les messages avant un timestamp
	GetRoomMessagesBefore(ctx context.Context, roomID int64, before time.Time, limit int) ([]*entities.ChatMessage, error)

	// SearchMessages recherche des messages dans une room
	SearchMessages(ctx context.Context, roomID int64, query string, limit, offset int) ([]*entities.ChatMessage, error)

	// GetUserMessages récupère les messages d'un utilisateur
	GetUserMessages(ctx context.Context, userID int64, limit, offset int) ([]*entities.ChatMessage, error)

	// GetMessageReplies récupère les réponses à un message
	GetMessageReplies(ctx context.Context, messageID int64, limit, offset int) ([]*entities.ChatMessage, error)

	// ============================================================================
	// RÉACTIONS ET INTERACTIONS
	// ============================================================================

	// AddReaction ajoute une réaction à un message
	AddReaction(ctx context.Context, messageID, userID int64, emoji string) error

	// RemoveReaction retire une réaction d'un message
	RemoveReaction(ctx context.Context, messageID, userID int64, emoji string) error

	// GetMessageReactions récupère les réactions d'un message
	GetMessageReactions(ctx context.Context, messageID int64) ([]*MessageReaction, error)

	// PinMessage épingle un message
	PinMessage(ctx context.Context, messageID, userID int64) error

	// UnpinMessage désépingle un message
	UnpinMessage(ctx context.Context, messageID, userID int64) error

	// GetPinnedMessages récupère les messages épinglés d'une room
	GetPinnedMessages(ctx context.Context, roomID int64) ([]*entities.ChatMessage, error)

	// ============================================================================
	// MODÉRATION
	// ============================================================================

	// ReportMessage signale un message
	ReportMessage(ctx context.Context, report *MessageReport) error

	// GetReports récupère les signalements
	GetReports(ctx context.Context, roomID int64, status ReportStatus, limit, offset int) ([]*MessageReport, error)

	// UpdateReportStatus met à jour le statut d'un signalement
	UpdateReportStatus(ctx context.Context, reportID int64, status ReportStatus, moderatorID int64) error

	// GetModerationLogs récupère les logs de modération
	GetModerationLogs(ctx context.Context, roomID int64, limit, offset int) ([]*ModerationLog, error)

	// CreateModerationLog crée un log de modération
	CreateModerationLog(ctx context.Context, log *ModerationLog) error

	// ============================================================================
	// PRÉSENCE ET ACTIVITÉ
	// ============================================================================

	// UpdateUserPresence met à jour la présence d'un utilisateur dans une room
	UpdateUserPresence(ctx context.Context, roomID, userID int64, isTyping bool) error

	// GetTypingUsers récupère les utilisateurs en train de taper
	GetTypingUsers(ctx context.Context, roomID int64) ([]*entities.User, error)

	// GetOnlineUsers récupère les utilisateurs en ligne dans une room
	GetOnlineUsers(ctx context.Context, roomID int64) ([]*entities.User, error)

	// GetRoomActivity récupère l'activité récente d'une room
	GetRoomActivity(ctx context.Context, roomID int64, since time.Time) (*RoomActivity, error)

	// ============================================================================
	// INVITATIONS
	// ============================================================================

	// CreateInvitation crée une invitation à rejoindre une room
	CreateInvitation(ctx context.Context, invitation *RoomInvitation) error

	// GetInvitation récupère une invitation par code
	GetInvitation(ctx context.Context, code string) (*RoomInvitation, error)

	// AcceptInvitation accepte une invitation
	AcceptInvitation(ctx context.Context, code string, userID int64) error

	// DeclineInvitation refuse une invitation
	DeclineInvitation(ctx context.Context, code string, userID int64) error

	// GetUserInvitations récupère les invitations d'un utilisateur
	GetUserInvitations(ctx context.Context, userID int64) ([]*RoomInvitation, error)

	// GetRoomInvitations récupère les invitations d'une room
	GetRoomInvitations(ctx context.Context, roomID int64) ([]*RoomInvitation, error)

	// ============================================================================
	// STATISTIQUES ET ANALYTICS
	// ============================================================================

	// GetChatStats récupère les statistiques générales du chat
	GetChatStats(ctx context.Context) (*ChatStats, error)

	// GetUserChatStats récupère les statistiques de chat d'un utilisateur
	GetUserChatStats(ctx context.Context, userID int64) (*UserChatStats, error)

	// GetRoomUsageStats récupère les statistiques d'utilisation d'une room
	GetRoomUsageStats(ctx context.Context, roomID int64, period time.Duration) (*RoomUsageStats, error)

	// GetPopularRooms récupère les rooms les plus populaires
	GetPopularRooms(ctx context.Context, period time.Duration, limit int) ([]*entities.ChatRoom, error)

	// GetActiveUsers récupère les utilisateurs les plus actifs
	GetActiveUsers(ctx context.Context, roomID int64, period time.Duration, limit int) ([]*entities.User, error)
}

// ============================================================================
// TYPES ET STRUCTURES
// ============================================================================

// RoomFilters filtres pour la recherche de rooms
type RoomFilters struct {
	Type       entities.RoomType    `json:"type,omitempty"`
	Privacy    entities.RoomPrivacy `json:"privacy,omitempty"`
	Status     entities.RoomStatus  `json:"status,omitempty"`
	CreatorID  int64                `json:"creator_id,omitempty"`
	Tags       []string             `json:"tags,omitempty"`
	MinMembers int                  `json:"min_members,omitempty"`
	MaxMembers int                  `json:"max_members,omitempty"`
	CreatedAt  *TimeRange           `json:"created_at,omitempty"`
	Limit      int                  `json:"limit"`
	Offset     int                  `json:"offset"`
	SortBy     string               `json:"sort_by"`    // "created_at", "member_count", "activity"
	SortOrder  string               `json:"sort_order"` // "asc", "desc"
}

// TimeRange période de temps
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// RoomMember membre d'une room avec son rôle
type RoomMember struct {
	ID       int64             `json:"id" db:"id"`
	RoomID   int64             `json:"room_id" db:"room_id"`
	UserID   int64             `json:"user_id" db:"user_id"`
	Role     entities.RoomRole `json:"role" db:"role"`
	JoinedAt time.Time         `json:"joined_at" db:"joined_at"`
	LastSeen time.Time         `json:"last_seen" db:"last_seen"`
	IsOnline bool              `json:"is_online" db:"is_online"`

	// Relations
	User *entities.User     `json:"user,omitempty"`
	Room *entities.ChatRoom `json:"room,omitempty"`
}

// MessageReaction réaction à un message
type MessageReaction struct {
	ID        int64     `json:"id" db:"id"`
	MessageID int64     `json:"message_id" db:"message_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Emoji     string    `json:"emoji" db:"emoji"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Relations
	User    *entities.User        `json:"user,omitempty"`
	Message *entities.ChatMessage `json:"message,omitempty"`
}

// ReportStatus statut d'un signalement
type ReportStatus string

const (
	ReportStatusPending   ReportStatus = "pending"
	ReportStatusReviewed  ReportStatus = "reviewed"
	ReportStatusResolved  ReportStatus = "resolved"
	ReportStatusDismissed ReportStatus = "dismissed"
)

// MessageReport signalement d'un message
type MessageReport struct {
	ID          int64        `json:"id" db:"id"`
	MessageID   int64        `json:"message_id" db:"message_id"`
	ReporterID  int64        `json:"reporter_id" db:"reporter_id"`
	Reason      string       `json:"reason" db:"reason"`
	Description string       `json:"description" db:"description"`
	Status      ReportStatus `json:"status" db:"status"`
	ModeratorID *int64       `json:"moderator_id,omitempty" db:"moderator_id"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`

	// Relations
	Message   *entities.ChatMessage `json:"message,omitempty"`
	Reporter  *entities.User        `json:"reporter,omitempty"`
	Moderator *entities.User        `json:"moderator,omitempty"`
}

// ModerationLog log d'action de modération
type ModerationLog struct {
	ID           int64     `json:"id" db:"id"`
	RoomID       int64     `json:"room_id" db:"room_id"`
	ModeratorID  int64     `json:"moderator_id" db:"moderator_id"`
	TargetUserID *int64    `json:"target_user_id,omitempty" db:"target_user_id"`
	MessageID    *int64    `json:"message_id,omitempty" db:"message_id"`
	Action       string    `json:"action" db:"action"` // "ban", "kick", "delete_message", "mute"
	Reason       string    `json:"reason" db:"reason"`
	Duration     *int64    `json:"duration,omitempty" db:"duration"` // en secondes
	CreatedAt    time.Time `json:"created_at" db:"created_at"`

	// Relations
	Room       *entities.ChatRoom    `json:"room,omitempty"`
	Moderator  *entities.User        `json:"moderator,omitempty"`
	TargetUser *entities.User        `json:"target_user,omitempty"`
	Message    *entities.ChatMessage `json:"message,omitempty"`
}

// RoomInvitation invitation à rejoindre une room
type RoomInvitation struct {
	ID        int64                     `json:"id" db:"id"`
	RoomID    int64                     `json:"room_id" db:"room_id"`
	InviterID int64                     `json:"inviter_id" db:"inviter_id"`
	InviteeID *int64                    `json:"invitee_id,omitempty" db:"invitee_id"`
	Code      string                    `json:"code" db:"code"`
	Uses      int                       `json:"uses" db:"uses"`
	MaxUses   int                       `json:"max_uses" db:"max_uses"`
	ExpiresAt *time.Time                `json:"expires_at,omitempty" db:"expires_at"`
	Status    entities.InvitationStatus `json:"status" db:"status"`
	CreatedAt time.Time                 `json:"created_at" db:"created_at"`
	UpdatedAt time.Time                 `json:"updated_at" db:"updated_at"`

	// Relations
	Room    *entities.ChatRoom `json:"room,omitempty"`
	Inviter *entities.User     `json:"inviter,omitempty"`
	Invitee *entities.User     `json:"invitee,omitempty"`
}

// RoomActivity activité d'une room
type RoomActivity struct {
	RoomID             int64         `json:"room_id"`
	MessageCount       int64         `json:"message_count"`
	ActiveUserCount    int64         `json:"active_user_count"`
	PeakOnlineUsers    int64         `json:"peak_online_users"`
	AverageOnlineUsers float64       `json:"average_online_users"`
	LastActivity       time.Time     `json:"last_activity"`
	Period             time.Duration `json:"period"`
}

// RoomStats statistiques d'une room
type RoomStats struct {
	RoomID        int64     `json:"room_id"`
	MemberCount   int64     `json:"member_count"`
	MessageCount  int64     `json:"message_count"`
	OnlineCount   int64     `json:"online_count"`
	TodayMessages int64     `json:"today_messages"`
	WeekMessages  int64     `json:"week_messages"`
	MonthMessages int64     `json:"month_messages"`
	LastMessage   time.Time `json:"last_message"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ChatStats statistiques générales du chat
type ChatStats struct {
	TotalRooms    int64     `json:"total_rooms"`
	TotalMessages int64     `json:"total_messages"`
	TotalUsers    int64     `json:"total_users"`
	ActiveRooms   int64     `json:"active_rooms"`
	OnlineUsers   int64     `json:"online_users"`
	TodayMessages int64     `json:"today_messages"`
	TodayUsers    int64     `json:"today_users"`
	PopularRooms  int64     `json:"popular_rooms"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserChatStats statistiques de chat d'un utilisateur
type UserChatStats struct {
	UserID            int64     `json:"user_id"`
	TotalMessages     int64     `json:"total_messages"`
	RoomsJoined       int64     `json:"rooms_joined"`
	RoomsCreated      int64     `json:"rooms_created"`
	ReactionsGiven    int64     `json:"reactions_given"`
	ReactionsReceived int64     `json:"reactions_received"`
	AveragePerDay     float64   `json:"average_per_day"`
	MostActiveRoom    int64     `json:"most_active_room"`
	LastMessage       time.Time `json:"last_message"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// RoomUsageStats statistiques d'utilisation d'une room
type RoomUsageStats struct {
	RoomID              int64            `json:"room_id"`
	Period              time.Duration    `json:"period"`
	MessageCount        int64            `json:"message_count"`
	UniqueUsers         int64            `json:"unique_users"`
	PeakConcurrentUsers int64            `json:"peak_concurrent_users"`
	AverageSessionTime  time.Duration    `json:"average_session_time"`
	MostActiveHour      int              `json:"most_active_hour"`
	MessagesByHour      map[int]int64    `json:"messages_by_hour"`
	UserActivity        map[int64]int64  `json:"user_activity"`
	TopEmojis           map[string]int64 `json:"top_emojis"`
	UpdatedAt           time.Time        `json:"updated_at"`
}
