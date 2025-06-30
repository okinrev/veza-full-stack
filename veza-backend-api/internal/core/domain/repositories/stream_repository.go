package repositories

import (
	"context"
	"time"

	"github.com/okinrev/veza-web-app/internal/core/domain/entities"
)

// StreamRepository définit les opérations de persistence pour le système de streaming
type StreamRepository interface {
	// ============================================================================
	// GESTION DES STREAMS
	// ============================================================================

	// CreateStream crée un nouveau stream
	CreateStream(ctx context.Context, stream *entities.Stream) error

	// GetStreamByID récupère un stream par ID
	GetStreamByID(ctx context.Context, streamID int64) (*entities.Stream, error)

	// GetStreamBySlug récupère un stream par son slug
	GetStreamBySlug(ctx context.Context, slug string) (*entities.Stream, error)

	// UpdateStream met à jour un stream
	UpdateStream(ctx context.Context, stream *entities.Stream) error

	// DeleteStream supprime un stream (soft delete)
	DeleteStream(ctx context.Context, streamID int64) error

	// StartStream démarre un stream
	StartStream(ctx context.Context, streamID int64, serverURL string) error

	// StopStream arrête un stream
	StopStream(ctx context.Context, streamID int64) error

	// ListStreams liste les streams avec pagination et filtres
	ListStreams(ctx context.Context, filters StreamFilters) ([]*entities.Stream, error)

	// SearchStreams recherche des streams
	SearchStreams(ctx context.Context, query string, limit, offset int) ([]*entities.Stream, error)

	// GetUserStreams récupère les streams d'un utilisateur
	GetUserStreams(ctx context.Context, userID int64, limit, offset int) ([]*entities.Stream, error)

	// GetLiveStreams récupère les streams en direct
	GetLiveStreams(ctx context.Context, limit, offset int) ([]*entities.Stream, error)

	// GetPopularStreams récupère les streams populaires
	GetPopularStreams(ctx context.Context, period time.Duration, limit int) ([]*entities.Stream, error)

	// GetStreamsByCategory récupère les streams par catégorie
	GetStreamsByCategory(ctx context.Context, category string, limit, offset int) ([]*entities.Stream, error)

	// ============================================================================
	// GESTION DES LISTENERS
	// ============================================================================

	// JoinStream ajoute un listener au stream
	JoinStream(ctx context.Context, streamID, userID int64, clientIP string) error

	// LeaveStream retire un listener du stream
	LeaveStream(ctx context.Context, streamID, userID int64) error

	// GetStreamListeners récupère les listeners d'un stream
	GetStreamListeners(ctx context.Context, streamID int64, limit, offset int) ([]*StreamListener, error)

	// GetListenerCount récupère le nombre de listeners
	GetListenerCount(ctx context.Context, streamID int64) (int64, error)

	// GetCurrentListeners récupère les listeners actuellement connectés
	GetCurrentListeners(ctx context.Context, streamID int64) ([]*StreamListener, error)

	// UpdateListenerPosition met à jour la position du listener
	UpdateListenerPosition(ctx context.Context, streamID, userID int64, position time.Duration) error

	// GetListenerSession récupère une session d'écoute
	GetListenerSession(ctx context.Context, sessionToken string) (*ListenerSession, error)

	// CreateListenerSession crée une session d'écoute
	CreateListenerSession(ctx context.Context, session *ListenerSession) error

	// UpdateListenerSession met à jour une session d'écoute
	UpdateListenerSession(ctx context.Context, sessionToken string, heartbeat time.Time) error

	// EndListenerSession termine une session d'écoute
	EndListenerSession(ctx context.Context, sessionToken string) error

	// ============================================================================
	// GESTION DES PLAYLISTS
	// ============================================================================

	// CreatePlaylist crée une nouvelle playlist
	CreatePlaylist(ctx context.Context, playlist *StreamPlaylist) error

	// GetPlaylistByID récupère une playlist par ID
	GetPlaylistByID(ctx context.Context, playlistID int64) (*StreamPlaylist, error)

	// UpdatePlaylist met à jour une playlist
	UpdatePlaylist(ctx context.Context, playlist *StreamPlaylist) error

	// DeletePlaylist supprime une playlist
	DeletePlaylist(ctx context.Context, playlistID int64) error

	// GetUserPlaylists récupère les playlists d'un utilisateur
	GetUserPlaylists(ctx context.Context, userID int64, limit, offset int) ([]*StreamPlaylist, error)

	// GetStreamPlaylists récupère les playlists d'un stream
	GetStreamPlaylists(ctx context.Context, streamID int64) ([]*StreamPlaylist, error)

	// AddTrackToPlaylist ajoute une track à une playlist
	AddTrackToPlaylist(ctx context.Context, playlistID, trackID int64, order int) error

	// RemoveTrackFromPlaylist retire une track d'une playlist
	RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID int64) error

	// ReorderPlaylistTracks réordonne les tracks d'une playlist
	ReorderPlaylistTracks(ctx context.Context, playlistID int64, trackOrders map[int64]int) error

	// ============================================================================
	// GESTION DES TRACKS
	// ============================================================================

	// CreateTrack crée une nouvelle track
	CreateTrack(ctx context.Context, track *StreamTrack) error

	// GetTrackByID récupère une track par ID
	GetTrackByID(ctx context.Context, trackID int64) (*StreamTrack, error)

	// UpdateTrack met à jour une track
	UpdateTrack(ctx context.Context, track *StreamTrack) error

	// DeleteTrack supprime une track
	DeleteTrack(ctx context.Context, trackID int64) error

	// GetStreamTracks récupère les tracks d'un stream
	GetStreamTracks(ctx context.Context, streamID int64, limit, offset int) ([]*StreamTrack, error)

	// GetCurrentTrack récupère la track actuellement jouée
	GetCurrentTrack(ctx context.Context, streamID int64) (*StreamTrack, error)

	// SetCurrentTrack définit la track actuellement jouée
	SetCurrentTrack(ctx context.Context, streamID, trackID int64, position time.Duration) error

	// GetTrackHistory récupère l'historique des tracks jouées
	GetTrackHistory(ctx context.Context, streamID int64, limit, offset int) ([]*TrackHistory, error)

	// AddTrackToHistory ajoute une track à l'historique
	AddTrackToHistory(ctx context.Context, history *TrackHistory) error

	// ============================================================================
	// QUALITÉ ET TRANSCODING
	// ============================================================================

	// CreateStreamQuality crée une qualité de stream
	CreateStreamQuality(ctx context.Context, quality *StreamQualityConfig) error

	// GetStreamQualities récupère les qualités disponibles pour un stream
	GetStreamQualities(ctx context.Context, streamID int64) ([]*StreamQualityConfig, error)

	// UpdateStreamQuality met à jour une configuration de qualité
	UpdateStreamQuality(ctx context.Context, quality *StreamQualityConfig) error

	// DeleteStreamQuality supprime une configuration de qualité
	DeleteStreamQuality(ctx context.Context, streamID int64, quality entities.StreamQuality) error

	// GetOptimalQuality récupère la qualité optimale pour un listener
	GetOptimalQuality(ctx context.Context, streamID, userID int64, bandwidth int64) (*StreamQualityConfig, error)

	// ============================================================================
	// ENREGISTREMENT ET ARCHIVES
	// ============================================================================

	// CreateRecording crée un enregistrement
	CreateRecording(ctx context.Context, recording *StreamRecording) error

	// GetRecordingByID récupère un enregistrement par ID
	GetRecordingByID(ctx context.Context, recordingID int64) (*StreamRecording, error)

	// UpdateRecording met à jour un enregistrement
	UpdateRecording(ctx context.Context, recording *StreamRecording) error

	// DeleteRecording supprime un enregistrement
	DeleteRecording(ctx context.Context, recordingID int64) error

	// GetStreamRecordings récupère les enregistrements d'un stream
	GetStreamRecordings(ctx context.Context, streamID int64, limit, offset int) ([]*StreamRecording, error)

	// GetUserRecordings récupère les enregistrements d'un utilisateur
	GetUserRecordings(ctx context.Context, userID int64, limit, offset int) ([]*StreamRecording, error)

	// StartRecording démarre l'enregistrement d'un stream
	StartRecording(ctx context.Context, streamID int64, settings RecordingSettings) (*StreamRecording, error)

	// StopRecording arrête l'enregistrement
	StopRecording(ctx context.Context, recordingID int64) error

	// ============================================================================
	// ANALYTICS ET MÉTRIQUES
	// ============================================================================

	// RecordListenerEvent enregistre un événement de listener
	RecordListenerEvent(ctx context.Context, event *ListenerEvent) error

	// GetStreamAnalytics récupère les analytics d'un stream
	GetStreamAnalytics(ctx context.Context, streamID int64, period time.Duration) (*StreamAnalytics, error)

	// GetUserListeningStats récupère les statistiques d'écoute d'un utilisateur
	GetUserListeningStats(ctx context.Context, userID int64) (*UserListeningStats, error)

	// GetTopStreams récupère les streams les plus populaires
	GetTopStreams(ctx context.Context, period time.Duration, limit int) ([]*entities.Stream, error)

	// GetTopGenres récupère les genres les plus écoutés
	GetTopGenres(ctx context.Context, period time.Duration, limit int) ([]*GenreStats, error)

	// GetRealtimeMetrics récupère les métriques en temps réel
	GetRealtimeMetrics(ctx context.Context, streamID int64) (*RealtimeMetrics, error)

	// UpdateStreamMetrics met à jour les métriques d'un stream
	UpdateStreamMetrics(ctx context.Context, streamID int64, metrics StreamMetricsUpdate) error

	// ============================================================================
	// CHAT ET INTERACTIONS
	// ============================================================================

	// EnableStreamChat active le chat pour un stream
	EnableStreamChat(ctx context.Context, streamID int64) error

	// DisableStreamChat désactive le chat pour un stream
	DisableStreamChat(ctx context.Context, streamID int64) error

	// GetStreamChatRoom récupère la room de chat associée au stream
	GetStreamChatRoom(ctx context.Context, streamID int64) (*entities.ChatRoom, error)

	// CreateStreamPoll crée un sondage dans le stream
	CreateStreamPoll(ctx context.Context, poll *StreamPoll) error

	// VoteStreamPoll vote dans un sondage
	VoteStreamPoll(ctx context.Context, pollID, userID int64, optionID int64) error

	// GetStreamPolls récupère les sondages d'un stream
	GetStreamPolls(ctx context.Context, streamID int64, active bool) ([]*StreamPoll, error)

	// ============================================================================
	// RECOMMANDATIONS ET DÉCOUVERTE
	// ============================================================================

	// GetRecommendedStreams récupère des streams recommandés pour un utilisateur
	GetRecommendedStreams(ctx context.Context, userID int64, limit int) ([]*entities.Stream, error)

	// GetSimilarStreams récupère des streams similaires
	GetSimilarStreams(ctx context.Context, streamID int64, limit int) ([]*entities.Stream, error)

	// GetTrendingStreams récupère les streams en tendance
	GetTrendingStreams(ctx context.Context, limit int) ([]*entities.Stream, error)

	// UpdateUserPreferences met à jour les préférences d'écoute d'un utilisateur
	UpdateUserPreferences(ctx context.Context, userID int64, preferences UserStreamingPreferences) error

	// GetUserPreferences récupère les préférences d'un utilisateur
	GetUserPreferences(ctx context.Context, userID int64) (*UserStreamingPreferences, error)

	// ============================================================================
	// MONÉTISATION ET ABONNEMENTS
	// ============================================================================

	// CreateSubscription crée un abonnement à un stream
	CreateSubscription(ctx context.Context, subscription *StreamSubscription) error

	// GetUserSubscriptions récupère les abonnements d'un utilisateur
	GetUserSubscriptions(ctx context.Context, userID int64) ([]*StreamSubscription, error)

	// GetStreamSubscribers récupère les abonnés d'un stream
	GetStreamSubscribers(ctx context.Context, streamID int64, limit, offset int) ([]*StreamSubscription, error)

	// CancelSubscription annule un abonnement
	CancelSubscription(ctx context.Context, subscriptionID int64) error

	// ProcessDonation traite une donation
	ProcessDonation(ctx context.Context, donation *StreamDonation) error

	// GetStreamDonations récupère les donations d'un stream
	GetStreamDonations(ctx context.Context, streamID int64, limit, offset int) ([]*StreamDonation, error)
}

// ============================================================================
// TYPES ET STRUCTURES
// ============================================================================

// StreamFilters filtres pour la recherche de streams
type StreamFilters struct {
	Type         entities.StreamType    `json:"type,omitempty"`
	Status       entities.StreamStatus  `json:"status,omitempty"`
	Privacy      entities.StreamPrivacy `json:"privacy,omitempty"`
	Quality      entities.StreamQuality `json:"quality,omitempty"`
	Genre        string                 `json:"genre,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CreatorID    int64                  `json:"creator_id,omitempty"`
	IsLive       *bool                  `json:"is_live,omitempty"`
	MinListeners int64                  `json:"min_listeners,omitempty"`
	MaxListeners int64                  `json:"max_listeners,omitempty"`
	CreatedAt    *TimeRange             `json:"created_at,omitempty"`
	Limit        int                    `json:"limit"`
	Offset       int                    `json:"offset"`
	SortBy       string                 `json:"sort_by"`    // "created_at", "listener_count", "popularity"
	SortOrder    string                 `json:"sort_order"` // "asc", "desc"
}

// StreamListener représente un listener connecté à un stream
type StreamListener struct {
	ID           int64                  `json:"id" db:"id"`
	StreamID     int64                  `json:"stream_id" db:"stream_id"`
	UserID       int64                  `json:"user_id" db:"user_id"`
	SessionToken string                 `json:"session_token" db:"session_token"`
	ClientIP     string                 `json:"client_ip" db:"client_ip"`
	UserAgent    string                 `json:"user_agent" db:"user_agent"`
	Quality      entities.StreamQuality `json:"quality" db:"quality"`
	Position     time.Duration          `json:"position" db:"position"`
	JoinedAt     time.Time              `json:"joined_at" db:"joined_at"`
	LastSeen     time.Time              `json:"last_seen" db:"last_seen"`
	IsActive     bool                   `json:"is_active" db:"is_active"`

	// Relations
	User   *entities.User   `json:"user,omitempty"`
	Stream *entities.Stream `json:"stream,omitempty"`
}

// ListenerSession session d'écoute d'un utilisateur
type ListenerSession struct {
	ID            int64                  `json:"id" db:"id"`
	StreamID      int64                  `json:"stream_id" db:"stream_id"`
	UserID        int64                  `json:"user_id" db:"user_id"`
	SessionToken  string                 `json:"session_token" db:"session_token"`
	Quality       entities.StreamQuality `json:"quality" db:"quality"`
	Position      time.Duration          `json:"position" db:"position"`
	Volume        float64                `json:"volume" db:"volume"`
	IsMuted       bool                   `json:"is_muted" db:"is_muted"`
	LastHeartbeat time.Time              `json:"last_heartbeat" db:"last_heartbeat"`
	StartedAt     time.Time              `json:"started_at" db:"started_at"`
	EndedAt       *time.Time             `json:"ended_at,omitempty" db:"ended_at"`
	TotalDuration time.Duration          `json:"total_duration" db:"total_duration"`
}

// StreamPlaylist playlist de tracks pour un stream
type StreamPlaylist struct {
	ID          int64         `json:"id" db:"id"`
	StreamID    int64         `json:"stream_id" db:"stream_id"`
	UserID      int64         `json:"user_id" db:"user_id"`
	Name        string        `json:"name" db:"name"`
	Description string        `json:"description" db:"description"`
	IsActive    bool          `json:"is_active" db:"is_active"`
	IsPublic    bool          `json:"is_public" db:"is_public"`
	TrackCount  int64         `json:"track_count" db:"track_count"`
	Duration    time.Duration `json:"duration" db:"duration"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`

	// Relations
	Stream *entities.Stream `json:"stream,omitempty"`
	User   *entities.User   `json:"user,omitempty"`
	Tracks []*StreamTrack   `json:"tracks,omitempty"`
}

// StreamTrack track audio dans un stream
type StreamTrack struct {
	ID         int64                  `json:"id" db:"id"`
	StreamID   int64                  `json:"stream_id" db:"stream_id"`
	Title      string                 `json:"title" db:"title"`
	Artist     string                 `json:"artist" db:"artist"`
	Album      string                 `json:"album" db:"album"`
	Genre      string                 `json:"genre" db:"genre"`
	Duration   time.Duration          `json:"duration" db:"duration"`
	FileURL    string                 `json:"file_url" db:"file_url"`
	CoverURL   string                 `json:"cover_url" db:"cover_url"`
	Metadata   map[string]interface{} `json:"metadata" db:"metadata"`
	PlayCount  int64                  `json:"play_count" db:"play_count"`
	LikeCount  int64                  `json:"like_count" db:"like_count"`
	IsExplicit bool                   `json:"is_explicit" db:"is_explicit"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`

	// Relations
	Stream *entities.Stream `json:"stream,omitempty"`
}

// TrackHistory historique des tracks jouées
type TrackHistory struct {
	ID        int64         `json:"id" db:"id"`
	StreamID  int64         `json:"stream_id" db:"stream_id"`
	TrackID   int64         `json:"track_id" db:"track_id"`
	PlayedAt  time.Time     `json:"played_at" db:"played_at"`
	Duration  time.Duration `json:"duration" db:"duration"`
	Listeners int64         `json:"listeners" db:"listeners"`

	// Relations
	Stream *entities.Stream `json:"stream,omitempty"`
	Track  *StreamTrack     `json:"track,omitempty"`
}

// StreamQualityConfig configuration de qualité pour un stream
type StreamQualityConfig struct {
	ID         int64                  `json:"id" db:"id"`
	StreamID   int64                  `json:"stream_id" db:"stream_id"`
	Quality    entities.StreamQuality `json:"quality" db:"quality"`
	Bitrate    int64                  `json:"bitrate" db:"bitrate"`
	SampleRate int64                  `json:"sample_rate" db:"sample_rate"`
	Codec      string                 `json:"codec" db:"codec"`
	URL        string                 `json:"url" db:"url"`
	IsActive   bool                   `json:"is_active" db:"is_active"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
}

// StreamRecording enregistrement d'un stream
type StreamRecording struct {
	ID          int64                  `json:"id" db:"id"`
	StreamID    int64                  `json:"stream_id" db:"stream_id"`
	Title       string                 `json:"title" db:"title"`
	Description string                 `json:"description" db:"description"`
	FileURL     string                 `json:"file_url" db:"file_url"`
	FileSize    int64                  `json:"file_size" db:"file_size"`
	Duration    time.Duration          `json:"duration" db:"duration"`
	Quality     entities.StreamQuality `json:"quality" db:"quality"`
	Status      RecordingStatus        `json:"status" db:"status"`
	StartedAt   time.Time              `json:"started_at" db:"started_at"`
	EndedAt     *time.Time             `json:"ended_at,omitempty" db:"ended_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`

	// Relations
	Stream *entities.Stream `json:"stream,omitempty"`
}

// RecordingStatus statut d'un enregistrement
type RecordingStatus string

const (
	RecordingStatusRecording  RecordingStatus = "recording"
	RecordingStatusProcessing RecordingStatus = "processing"
	RecordingStatusCompleted  RecordingStatus = "completed"
	RecordingStatusFailed     RecordingStatus = "failed"
	RecordingStatusDeleted    RecordingStatus = "deleted"
)

// RecordingSettings paramètres d'enregistrement
type RecordingSettings struct {
	Quality     entities.StreamQuality `json:"quality"`
	Format      string                 `json:"format"`
	MaxDuration time.Duration          `json:"max_duration"`
	AutoStop    bool                   `json:"auto_stop"`
}

// ListenerEvent événement de listener pour analytics
type ListenerEvent struct {
	ID        int64                  `json:"id" db:"id"`
	StreamID  int64                  `json:"stream_id" db:"stream_id"`
	UserID    int64                  `json:"user_id" db:"user_id"`
	EventType string                 `json:"event_type" db:"event_type"` // "join", "leave", "quality_change", "position_change"
	EventData map[string]interface{} `json:"event_data" db:"event_data"`
	ClientIP  string                 `json:"client_ip" db:"client_ip"`
	UserAgent string                 `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// StreamAnalytics analytics d'un stream
type StreamAnalytics struct {
	StreamID            int64            `json:"stream_id"`
	Period              time.Duration    `json:"period"`
	TotalListeners      int64            `json:"total_listeners"`
	UniqueListeners     int64            `json:"unique_listeners"`
	PeakListeners       int64            `json:"peak_listeners"`
	AverageListeners    float64          `json:"average_listeners"`
	TotalListenTime     time.Duration    `json:"total_listen_time"`
	AverageListenTime   time.Duration    `json:"average_listen_time"`
	DropoffRate         float64          `json:"dropoff_rate"`
	QualityDistribution map[string]int64 `json:"quality_distribution"`
	GeographicData      map[string]int64 `json:"geographic_data"`
	DeviceData          map[string]int64 `json:"device_data"`
	HourlyListeners     map[int]int64    `json:"hourly_listeners"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

// UserListeningStats statistiques d'écoute d'un utilisateur
type UserListeningStats struct {
	UserID           int64                  `json:"user_id"`
	TotalListenTime  time.Duration          `json:"total_listen_time"`
	StreamsListened  int64                  `json:"streams_listened"`
	FavoriteGenres   map[string]int64       `json:"favorite_genres"`
	FavoriteStreams  []int64                `json:"favorite_streams"`
	AverageSession   time.Duration          `json:"average_session"`
	PreferredQuality entities.StreamQuality `json:"preferred_quality"`
	ListeningHours   map[int]int64          `json:"listening_hours"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// GenreStats statistiques par genre
type GenreStats struct {
	Genre         string        `json:"genre"`
	ListenerCount int64         `json:"listener_count"`
	StreamCount   int64         `json:"stream_count"`
	TotalTime     time.Duration `json:"total_time"`
	GrowthRate    float64       `json:"growth_rate"`
}

// RealtimeMetrics métriques en temps réel
type RealtimeMetrics struct {
	StreamID         int64            `json:"stream_id"`
	CurrentListeners int64            `json:"current_listeners"`
	QualityBreakdown map[string]int64 `json:"quality_breakdown"`
	GeographicSpread map[string]int64 `json:"geographic_spread"`
	LastUpdate       time.Time        `json:"last_update"`
}

// StreamMetricsUpdate mise à jour des métriques
type StreamMetricsUpdate struct {
	ListenerCount *int64                  `json:"listener_count,omitempty"`
	Quality       *entities.StreamQuality `json:"quality,omitempty"`
	Bandwidth     *int64                  `json:"bandwidth,omitempty"`
	Bitrate       *int64                  `json:"bitrate,omitempty"`
	DroppedFrames *int64                  `json:"dropped_frames,omitempty"`
	BufferHealth  *float64                `json:"buffer_health,omitempty"`
}

// StreamPoll sondage dans un stream
type StreamPoll struct {
	ID        int64        `json:"id" db:"id"`
	StreamID  int64        `json:"stream_id" db:"stream_id"`
	CreatorID int64        `json:"creator_id" db:"creator_id"`
	Question  string       `json:"question" db:"question"`
	Options   []PollOption `json:"options" db:"options"`
	IsActive  bool         `json:"is_active" db:"is_active"`
	ExpiresAt *time.Time   `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
}

// PollOption option de sondage
type PollOption struct {
	ID    int64  `json:"id"`
	Text  string `json:"text"`
	Votes int64  `json:"votes"`
}

// UserStreamingPreferences préférences d'écoute utilisateur
type UserStreamingPreferences struct {
	UserID               int64                   `json:"user_id"`
	DefaultQuality       entities.StreamQuality  `json:"default_quality"`
	AutoAdjustQuality    bool                    `json:"auto_adjust_quality"`
	PreferredGenres      []string                `json:"preferred_genres"`
	BlockedGenres        []string                `json:"blocked_genres"`
	NotificationSettings NotificationPreferences `json:"notification_settings"`
	PrivacySettings      StreamPrivacySettings   `json:"privacy_settings"`
	UpdatedAt            time.Time               `json:"updated_at"`
}

// NotificationPreferences préférences de notification
type NotificationPreferences struct {
	StreamStart    bool `json:"stream_start"`
	StreamEnd      bool `json:"stream_end"`
	NewFollower    bool `json:"new_follower"`
	FavoriteOnline bool `json:"favorite_online"`
	PollCreated    bool `json:"poll_created"`
}

// StreamPrivacySettings paramètres de confidentialité
type StreamPrivacySettings struct {
	ShowInDirectory bool `json:"show_in_directory"`
	AllowRecording  bool `json:"allow_recording"`
	ShowListeners   bool `json:"show_listeners"`
	AllowChat       bool `json:"allow_chat"`
}

// StreamSubscription abonnement à un stream
type StreamSubscription struct {
	ID        int64              `json:"id" db:"id"`
	StreamID  int64              `json:"stream_id" db:"stream_id"`
	UserID    int64              `json:"user_id" db:"user_id"`
	Type      SubscriptionType   `json:"type" db:"type"`
	Price     int64              `json:"price" db:"price"` // en centimes
	Currency  string             `json:"currency" db:"currency"`
	Status    SubscriptionStatus `json:"status" db:"status"`
	ExpiresAt *time.Time         `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" db:"updated_at"`

	// Relations
	Stream *entities.Stream `json:"stream,omitempty"`
	User   *entities.User   `json:"user,omitempty"`
}

// SubscriptionType type d'abonnement
type SubscriptionType string

const (
	SubscriptionTypeMonthly  SubscriptionType = "monthly"
	SubscriptionTypeYearly   SubscriptionType = "yearly"
	SubscriptionTypeLifetime SubscriptionType = "lifetime"
)

// SubscriptionStatus statut d'abonnement
type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
	SubscriptionStatusPending   SubscriptionStatus = "pending"
)

// StreamDonation donation à un stream
type StreamDonation struct {
	ID          int64          `json:"id" db:"id"`
	StreamID    int64          `json:"stream_id" db:"stream_id"`
	UserID      int64          `json:"user_id" db:"user_id"`
	Amount      int64          `json:"amount" db:"amount"` // en centimes
	Currency    string         `json:"currency" db:"currency"`
	Message     string         `json:"message" db:"message"`
	IsAnonymous bool           `json:"is_anonymous" db:"is_anonymous"`
	Status      DonationStatus `json:"status" db:"status"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`

	// Relations
	Stream *entities.Stream `json:"stream,omitempty"`
	User   *entities.User   `json:"user,omitempty"`
}

// DonationStatus statut d'une donation
type DonationStatus string

const (
	DonationStatusPending   DonationStatus = "pending"
	DonationStatusCompleted DonationStatus = "completed"
	DonationStatusRefunded  DonationStatus = "refunded"
	DonationStatusFailed    DonationStatus = "failed"
)
