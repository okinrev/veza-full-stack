package repositories

import (
	"context"
	"time"

	"github.com/okinrev/veza-web-app/internal/core/domain/entities"
)

// UserRepository définit les opérations de persistence pour les utilisateurs
type UserRepository interface {
	// CRUD de base
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id int64) (*entities.User, error)
	GetByUUID(ctx context.Context, uuid string) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id int64) error
	SoftDelete(ctx context.Context, id int64) error

	// Recherche et listing
	List(ctx context.Context, limit, offset int) ([]*entities.User, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.User, error)
	GetByRole(ctx context.Context, role entities.UserRole, limit, offset int) ([]*entities.User, error)
	GetByStatus(ctx context.Context, status entities.UserStatus, limit, offset int) ([]*entities.User, error)

	// Authentification et sécurité
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
	UpdateLoginAttempts(ctx context.Context, userID int64, attempts int, lockedUntil *time.Time) error
	ResetLoginAttempts(ctx context.Context, userID int64) error
	UpdateLastLogin(ctx context.Context, userID int64, ipAddress string) error

	// Vérification d'email et 2FA
	UpdateEmailVerification(ctx context.Context, userID int64, verified bool, token string) error
	UpdateTwoFactor(ctx context.Context, userID int64, enabled bool, secret string) error

	// Gestion des rôles et statuts
	UpdateRole(ctx context.Context, userID int64, role entities.UserRole) error
	UpdateStatus(ctx context.Context, userID int64, status entities.UserStatus) error

	// Présence et activité
	UpdateOnlineStatus(ctx context.Context, userID int64, isOnline bool) error
	GetOnlineUsers(ctx context.Context, limit int) ([]*entities.User, error)
	UpdateLastSeen(ctx context.Context, userID int64, lastSeen time.Time) error

	// Statistiques et analytics
	GetUserStats(ctx context.Context, userID int64) (*UserStats, error)
	GetUserCount(ctx context.Context) (int64, error)
	GetUserCountByRole(ctx context.Context, role entities.UserRole) (int64, error)
	GetUserCountByStatus(ctx context.Context, status entities.UserStatus) (int64, error)
	GetActiveUsersCount(ctx context.Context, since time.Time) (int64, error)

	// Validation et vérification d'unicité
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	EmailExistsExcludingUser(ctx context.Context, email string, userID int64) (bool, error)
	UsernameExistsExcludingUser(ctx context.Context, username string, userID int64) (bool, error)

	// Préférences utilisateur
	UpdatePreferences(ctx context.Context, userID int64, preferences *UserPreferences) error
	GetPreferences(ctx context.Context, userID int64) (*UserPreferences, error)

	// Gestion des sessions
	CreateSession(ctx context.Context, session *UserSession) error
	GetSession(ctx context.Context, sessionToken string) (*UserSession, error)
	UpdateSession(ctx context.Context, sessionToken string, lastActivity time.Time) error
	InvalidateSession(ctx context.Context, sessionToken string) error
	InvalidateAllUserSessions(ctx context.Context, userID int64) error
	GetUserSessions(ctx context.Context, userID int64) ([]*UserSession, error)

	// Audit et logs
	CreateAuditLog(ctx context.Context, log *UserAuditLog) error
	GetUserAuditLogs(ctx context.Context, userID int64, limit, offset int) ([]*UserAuditLog, error)

	// Relations et contacts
	AddContact(ctx context.Context, userID, contactID int64) error
	RemoveContact(ctx context.Context, userID, contactID int64) error
	GetContacts(ctx context.Context, userID int64) ([]*entities.User, error)
	BlockUser(ctx context.Context, userID, blockedUserID int64) error
	UnblockUser(ctx context.Context, userID, blockedUserID int64) error
	GetBlockedUsers(ctx context.Context, userID int64) ([]*entities.User, error)
	IsBlocked(ctx context.Context, userID, otherUserID int64) (bool, error)
}

// UserStats contient les statistiques d'un utilisateur
type UserStats struct {
	UserID          int64     `json:"user_id"`
	MessageCount    int64     `json:"message_count"`
	StreamCount     int64     `json:"stream_count"`
	TotalStreamTime int64     `json:"total_stream_time"` // en secondes
	TotalListenTime int64     `json:"total_listen_time"` // en secondes
	RoomsJoined     int64     `json:"rooms_joined"`
	FriendsCount    int64     `json:"friends_count"`
	FollowersCount  int64     `json:"followers_count"`
	FollowingCount  int64     `json:"following_count"`
	LikesReceived   int64     `json:"likes_received"`
	LikesGiven      int64     `json:"likes_given"`
	AccountAge      int64     `json:"account_age"` // en jours
	LastActive      time.Time `json:"last_active"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UserPreferences contient les préférences d'un utilisateur
type UserPreferences struct {
	UserID               int64                `json:"user_id"`
	Theme                string               `json:"theme"`
	Language             string               `json:"language"`
	Timezone             string               `json:"timezone"`
	NotificationSettings NotificationSettings `json:"notification_settings"`
	PrivacySettings      PrivacySettings      `json:"privacy_settings"`
	AudioSettings        AudioSettings        `json:"audio_settings"`
	ChatSettings         ChatSettings         `json:"chat_settings"`
	UpdatedAt            time.Time            `json:"updated_at"`
}

// NotificationSettings contient les paramètres de notification
type NotificationSettings struct {
	EmailNotifications   bool `json:"email_notifications"`
	PushNotifications    bool `json:"push_notifications"`
	SoundNotifications   bool `json:"sound_notifications"`
	DesktopNotifications bool `json:"desktop_notifications"`

	// Types de notifications
	NewMessages         bool `json:"new_messages"`
	Mentions            bool `json:"mentions"`
	StreamStarted       bool `json:"stream_started"`
	StreamInvites       bool `json:"stream_invites"`
	RoomInvites         bool `json:"room_invites"`
	FriendRequests      bool `json:"friend_requests"`
	SystemAnnouncements bool `json:"system_announcements"`

	// Heures de silence
	QuietHoursEnabled bool   `json:"quiet_hours_enabled"`
	QuietHoursStart   string `json:"quiet_hours_start"` // "22:00"
	QuietHoursEnd     string `json:"quiet_hours_end"`   // "08:00"
}

// PrivacySettings contient les paramètres de confidentialité
type PrivacySettings struct {
	ProfileVisibility   string `json:"profile_visibility"` // "public", "friends", "private"
	OnlineStatusVisible bool   `json:"online_status_visible"`
	LastSeenVisible     bool   `json:"last_seen_visible"`
	AllowDirectMessages string `json:"allow_direct_messages"` // "everyone", "friends", "none"
	AllowRoomInvites    string `json:"allow_room_invites"`    // "everyone", "friends", "none"
	AllowStreamInvites  string `json:"allow_stream_invites"`  // "everyone", "friends", "none"
	ShowInSearchResults bool   `json:"show_in_search_results"`
	ShowEmail           bool   `json:"show_email"`
	ShowRealName        bool   `json:"show_real_name"`
}

// AudioSettings contient les paramètres audio
type AudioSettings struct {
	DefaultQuality     entities.StreamQuality `json:"default_quality"`
	AutoAdjustQuality  bool                   `json:"auto_adjust_quality"`
	Volume             float64                `json:"volume"`
	Muted              bool                   `json:"muted"`
	EnableEqualizer    bool                   `json:"enable_equalizer"`
	EqualizerPreset    string                 `json:"equalizer_preset"`
	EnableSpatialAudio bool                   `json:"enable_spatial_audio"`
	BufferSize         int                    `json:"buffer_size"` // en secondes
}

// ChatSettings contient les paramètres de chat
type ChatSettings struct {
	ShowTimestamps       bool     `json:"show_timestamps"`
	Show24HourTime       bool     `json:"show_24_hour_time"`
	EnableEmojis         bool     `json:"enable_emojis"`
	EnableAnimatedEmojis bool     `json:"enable_animated_emojis"`
	FontSize             int      `json:"font_size"`
	MessageGrouping      bool     `json:"message_grouping"`
	CompactMode          bool     `json:"compact_mode"`
	FilterWords          []string `json:"filter_words"`
	AllowMentions        bool     `json:"allow_mentions"`
	AllowDirectMentions  bool     `json:"allow_direct_mentions"`
}

// UserSession représente une session utilisateur
type UserSession struct {
	ID           int64     `json:"id" db:"id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	SessionToken string    `json:"session_token" db:"session_token"`
	RefreshToken string    `json:"refresh_token" db:"refresh_token"`
	DeviceInfo   string    `json:"device_info" db:"device_info"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	Location     string    `json:"location" db:"location"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	LastActivity time.Time `json:"last_activity" db:"last_activity"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`

	// Relation
	User *entities.User `json:"user,omitempty"`
}

// UserAuditLog représente un log d'audit pour un utilisateur
type UserAuditLog struct {
	ID           int64     `json:"id" db:"id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	Action       string    `json:"action" db:"action"`
	Resource     string    `json:"resource" db:"resource"`
	ResourceID   *int64    `json:"resource_id,omitempty" db:"resource_id"`
	Details      string    `json:"details" db:"details"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	Success      bool      `json:"success" db:"success"`
	ErrorMessage *string   `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`

	// Relation
	User *entities.User `json:"user,omitempty"`
}

// UserContact représente un contact/ami d'un utilisateur
type UserContact struct {
	ID        int64             `json:"id" db:"id"`
	UserID    int64             `json:"user_id" db:"user_id"`
	ContactID int64             `json:"contact_id" db:"contact_id"`
	Status    UserContactStatus `json:"status" db:"status"`
	CreatedAt time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt time.Time         `json:"updated_at" db:"updated_at"`

	// Relations
	User    *entities.User `json:"user,omitempty"`
	Contact *entities.User `json:"contact,omitempty"`
}

// UserContactStatus définit le statut d'un contact
type UserContactStatus string

const (
	ContactStatusPending  UserContactStatus = "pending"
	ContactStatusAccepted UserContactStatus = "accepted"
	ContactStatusBlocked  UserContactStatus = "blocked"
	ContactStatusDeclined UserContactStatus = "declined"
)

// UserBlock représente un utilisateur bloqué
type UserBlock struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	BlockedID int64     `json:"blocked_id" db:"blocked_id"`
	Reason    string    `json:"reason" db:"reason"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Relations
	User    *entities.User `json:"user,omitempty"`
	Blocked *entities.User `json:"blocked,omitempty"`
}
