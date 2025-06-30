package entities

import (
	"errors"
	"fmt"
	"time"
)

// StreamStatus définit les statuts des streams
type StreamStatus string

const (
	StreamStatusScheduled StreamStatus = "scheduled"
	StreamStatusLive      StreamStatus = "live"
	StreamStatusPaused    StreamStatus = "paused"
	StreamStatusEnded     StreamStatus = "ended"
	StreamStatusRecording StreamStatus = "recording"
	StreamStatusArchived  StreamStatus = "archived"
)

// StreamType définit les types de streams
type StreamType string

const (
	StreamTypeLive     StreamType = "live"
	StreamTypeRecorded StreamType = "recorded"
	StreamTypePodcast  StreamType = "podcast"
	StreamTypeRadio    StreamType = "radio"
	StreamTypeMusic    StreamType = "music"
	StreamTypeTalk     StreamType = "talk"
)

// StreamPrivacy définit les niveaux de confidentialité
type StreamPrivacy string

const (
	StreamPrivacyPublic   StreamPrivacy = "public"
	StreamPrivacyPrivate  StreamPrivacy = "private"
	StreamPrivacyUnlisted StreamPrivacy = "unlisted"
	StreamPrivacyMembers  StreamPrivacy = "members_only"
)

// StreamQuality définit les qualités de streaming
type StreamQuality string

const (
	Quality64kbps   StreamQuality = "64kbps"
	Quality128kbps  StreamQuality = "128kbps"
	Quality256kbps  StreamQuality = "256kbps"
	Quality320kbps  StreamQuality = "320kbps"
	QualityLossless StreamQuality = "lossless"
)

// Stream représente un stream audio
type Stream struct {
	ID          int64  `json:"id" db:"id"`
	UUID        string `json:"uuid" db:"uuid"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`

	// Type et confidentialité
	Type    StreamType    `json:"type" db:"type"`
	Privacy StreamPrivacy `json:"privacy" db:"privacy"`

	// Propriétaire et permissions
	StreamerID   int64    `json:"streamer_id" db:"streamer_id"`
	RequiredRole UserRole `json:"required_role" db:"required_role"`
	IsPrivate    bool     `json:"is_private" db:"is_private"`
	Password     string   `json:"-" db:"password_hash"`

	// État du stream
	Status     StreamStatus  `json:"status" db:"status"`
	Quality    StreamQuality `json:"quality" db:"quality"`
	Bitrate    int           `json:"bitrate" db:"bitrate"`
	SampleRate int           `json:"sample_rate" db:"sample_rate"`

	// URLs et configuration technique
	StreamURL    string `json:"stream_url" db:"stream_url"`
	PlaylistURL  string `json:"playlist_url" db:"playlist_url"`
	ThumbnailURL string `json:"thumbnail_url" db:"thumbnail_url"`
	RecordingURL string `json:"recording_url,omitempty" db:"recording_url"`

	// Statistiques en temps réel
	CurrentListeners int   `json:"current_listeners" db:"current_listeners"`
	MaxListeners     int   `json:"max_listeners" db:"max_listeners"`
	TotalViews       int64 `json:"total_views" db:"total_views"`

	// Durée et timing
	ScheduledAt *time.Time `json:"scheduled_at,omitempty" db:"scheduled_at"`
	StartedAt   *time.Time `json:"started_at,omitempty" db:"started_at"`
	EndedAt     *time.Time `json:"ended_at,omitempty" db:"ended_at"`
	Duration    int64      `json:"duration" db:"duration"` // en secondes

	// Configuration avancée
	EnableChat      bool     `json:"enable_chat" db:"enable_chat"`
	EnableRecording bool     `json:"enable_recording" db:"enable_recording"`
	AutoArchive     bool     `json:"auto_archive" db:"auto_archive"`
	MaxDuration     int64    `json:"max_duration" db:"max_duration"` // en secondes
	Tags            []string `json:"tags" db:"tags"`
	Category        string   `json:"category" db:"category"`

	// Métadonnées
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Relations (chargées séparément)
	Streamer     *User            `json:"streamer,omitempty"`
	CurrentTrack *Track           `json:"current_track,omitempty"`
	Playlist     []*Track         `json:"playlist,omitempty"`
	Analytics    *StreamAnalytics `json:"analytics,omitempty"`
}

// Track représente une piste audio dans un stream
type Track struct {
	ID     int64  `json:"id" db:"id"`
	UUID   string `json:"uuid" db:"uuid"`
	Title  string `json:"title" db:"title"`
	Artist string `json:"artist" db:"artist"`
	Album  string `json:"album" db:"album"`
	Genre  string `json:"genre" db:"genre"`

	// Informations techniques
	Duration   int64  `json:"duration" db:"duration"` // en secondes
	FileSize   int64  `json:"file_size" db:"file_size"`
	Format     string `json:"format" db:"format"`
	Bitrate    int    `json:"bitrate" db:"bitrate"`
	SampleRate int    `json:"sample_rate" db:"sample_rate"`

	// URLs
	FileURL     string `json:"file_url" db:"file_url"`
	CoverURL    string `json:"cover_url" db:"cover_url"`
	WaveformURL string `json:"waveform_url" db:"waveform_url"`

	// Propriétaire et permissions
	UploaderID   int64    `json:"uploader_id" db:"uploader_id"`
	IsPublic     bool     `json:"is_public" db:"is_public"`
	RequiredRole UserRole `json:"required_role" db:"required_role"`

	// Statistiques
	PlayCount     int64 `json:"play_count" db:"play_count"`
	LikeCount     int64 `json:"like_count" db:"like_count"`
	DownloadCount int64 `json:"download_count" db:"download_count"`

	// Métadonnées
	Tags      []string   `json:"tags" db:"tags"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Relations
	Uploader *User `json:"uploader,omitempty"`
}

// StreamAnalytics contient les analytics détaillées d'un stream
type StreamAnalytics struct {
	StreamID int64 `json:"stream_id" db:"stream_id"`

	// Statistiques d'audience
	TotalListeners    int64     `json:"total_listeners" db:"total_listeners"`
	UniqueListeners   int64     `json:"unique_listeners" db:"unique_listeners"`
	AverageListeners  float64   `json:"average_listeners" db:"average_listeners"`
	PeakListeners     int       `json:"peak_listeners" db:"peak_listeners"`
	PeakListenersTime time.Time `json:"peak_listeners_time" db:"peak_listeners_time"`

	// Géographie
	CountryStats map[string]int64 `json:"country_stats" db:"country_stats"`
	CityStats    map[string]int64 `json:"city_stats" db:"city_stats"`

	// Engagement
	ChatMessages     int64   `json:"chat_messages" db:"chat_messages"`
	Reactions        int64   `json:"reactions" db:"reactions"`
	Shares           int64   `json:"shares" db:"shares"`
	AverageWatchTime float64 `json:"average_watch_time" db:"average_watch_time"`

	// Qualité technique
	BufferingEvents int64   `json:"buffering_events" db:"buffering_events"`
	ErrorCount      int64   `json:"error_count" db:"error_count"`
	AverageLatency  float64 `json:"average_latency" db:"average_latency"`

	// Revenue (si applicable)
	Revenue   float64 `json:"revenue" db:"revenue"`
	Donations float64 `json:"donations" db:"donations"`

	// Timestamps
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// StreamListener représente un auditeur connecté à un stream
type StreamListener struct {
	ID       int64 `json:"id" db:"id"`
	StreamID int64 `json:"stream_id" db:"stream_id"`
	UserID   int64 `json:"user_id" db:"user_id"`

	// Information de connexion
	IPAddress string `json:"ip_address" db:"ip_address"`
	UserAgent string `json:"user_agent" db:"user_agent"`
	Country   string `json:"country" db:"country"`
	City      string `json:"city" db:"city"`

	// Qualité et performance
	Quality      StreamQuality `json:"quality" db:"quality"`
	Latency      int64         `json:"latency" db:"latency"` // en ms
	BufferHealth float64       `json:"buffer_health" db:"buffer_health"`

	// Timing
	JoinedAt time.Time  `json:"joined_at" db:"joined_at"`
	LeftAt   *time.Time `json:"left_at,omitempty" db:"left_at"`
	Duration int64      `json:"duration" db:"duration"` // temps d'écoute en secondes

	// Relations
	User   *User   `json:"user,omitempty"`
	Stream *Stream `json:"stream,omitempty"`
}

// StreamPlaylist représente une playlist pour un stream
type StreamPlaylist struct {
	ID          int64  `json:"id" db:"id"`
	UUID        string `json:"uuid" db:"uuid"`
	StreamID    int64  `json:"stream_id" db:"stream_id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`

	// Configuration
	IsActive    bool `json:"is_active" db:"is_active"`
	Shuffle     bool `json:"shuffle" db:"shuffle"`
	Repeat      bool `json:"repeat" db:"repeat"`
	AutoAdvance bool `json:"auto_advance" db:"auto_advance"`

	// Lecture en cours
	CurrentTrackIndex int   `json:"current_track_index" db:"current_track_index"`
	CurrentPosition   int64 `json:"current_position" db:"current_position"` // position en secondes

	// Métadonnées
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Relations
	Stream *Stream  `json:"stream,omitempty"`
	Tracks []*Track `json:"tracks,omitempty"`
}

// NewStream crée un nouveau stream
func NewStream(title, description string, streamerID int64) (*Stream, error) {
	stream := &Stream{
		Title:           title,
		Description:     description,
		StreamerID:      streamerID,
		RequiredRole:    RoleUser,
		IsPrivate:       false,
		Status:          StreamStatusScheduled,
		Quality:         Quality128kbps,
		Bitrate:         128000,
		SampleRate:      44100,
		MaxListeners:    1000,
		EnableChat:      true,
		EnableRecording: true,
		AutoArchive:     true,
		MaxDuration:     4 * 3600, // 4 heures
		Category:        "music",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := stream.Validate(); err != nil {
		return nil, err
	}

	return stream, nil
}

// Validate vérifie que le stream est valide
func (s *Stream) Validate() error {
	if s.Title == "" {
		return errors.New("stream title is required")
	}

	if len(s.Title) < 3 {
		return errors.New("stream title must be at least 3 characters long")
	}

	if len(s.Title) > 200 {
		return errors.New("stream title must be less than 200 characters long")
	}

	if s.StreamerID == 0 {
		return errors.New("streamer ID is required")
	}

	if err := s.ValidateStatus(); err != nil {
		return err
	}

	if err := s.ValidateQuality(); err != nil {
		return err
	}

	return nil
}

// ValidateStatus vérifie que le statut est valide
func (s *Stream) ValidateStatus() error {
	switch s.Status {
	case StreamStatusScheduled, StreamStatusLive, StreamStatusPaused, StreamStatusEnded, StreamStatusRecording, StreamStatusArchived:
		return nil
	default:
		return errors.New("invalid stream status")
	}
}

// ValidateQuality vérifie que la qualité est valide
func (s *Stream) ValidateQuality() error {
	switch s.Quality {
	case Quality64kbps, Quality128kbps, Quality256kbps, Quality320kbps, QualityLossless:
		return nil
	default:
		return errors.New("invalid stream quality")
	}
}

// CanUserListen vérifie si un utilisateur peut écouter le stream
func (s *Stream) CanUserListen(user *User) bool {
	// Stream privé avec mot de passe ou permissions
	if s.IsPrivate && !user.HasAnyRole(s.RequiredRole, RoleAdmin, RoleSuperAdmin) {
		return false
	}

	// Vérifier la capacité maximale
	if s.CurrentListeners >= s.MaxListeners && !user.IsPremium() {
		return false
	}

	return true
}

// Start démarre le stream
func (s *Stream) Start() error {
	if s.Status != StreamStatusScheduled {
		return errors.New("stream must be scheduled to start")
	}

	now := time.Now()
	s.Status = StreamStatusLive
	s.StartedAt = &now
	s.UpdatedAt = now

	return nil
}

// Pause met en pause le stream
func (s *Stream) Pause() error {
	if s.Status != StreamStatusLive {
		return errors.New("only live streams can be paused")
	}

	s.Status = StreamStatusPaused
	s.UpdatedAt = time.Now()

	return nil
}

// Resume reprend le stream en pause
func (s *Stream) Resume() error {
	if s.Status != StreamStatusPaused {
		return errors.New("only paused streams can be resumed")
	}

	s.Status = StreamStatusLive
	s.UpdatedAt = time.Now()

	return nil
}

// End termine le stream
func (s *Stream) End() error {
	if s.Status != StreamStatusLive && s.Status != StreamStatusPaused {
		return errors.New("only live or paused streams can be ended")
	}

	now := time.Now()
	s.Status = StreamStatusEnded
	s.EndedAt = &now
	s.UpdatedAt = now

	// Calculer la durée totale
	if s.StartedAt != nil {
		s.Duration = int64(now.Sub(*s.StartedAt).Seconds())
	}

	return nil
}

// AddListener ajoute un auditeur au stream
func (s *Stream) AddListener() {
	s.CurrentListeners++
	if s.CurrentListeners > s.MaxListeners {
		s.MaxListeners = s.CurrentListeners
	}
	s.TotalViews++
	s.UpdatedAt = time.Now()
}

// RemoveListener retire un auditeur du stream
func (s *Stream) RemoveListener() {
	if s.CurrentListeners > 0 {
		s.CurrentListeners--
	}
	s.UpdatedAt = time.Now()
}

// IsLive vérifie si le stream est en direct
func (s *Stream) IsLive() bool {
	return s.Status == StreamStatusLive
}

// CanBeControlledBy vérifie si un utilisateur peut contrôler le stream
func (s *Stream) CanBeControlledBy(userID int64, userRole UserRole) bool {
	// Le streamer peut toujours contrôler son stream
	if s.StreamerID == userID {
		return true
	}

	// Les admins peuvent contrôler tous les streams
	if userRole == RoleAdmin || userRole == RoleSuperAdmin {
		return true
	}

	return false
}

// NewTrack crée une nouvelle piste
func NewTrack(title, artist string, uploaderID int64) (*Track, error) {
	track := &Track{
		Title:        title,
		Artist:       artist,
		UploaderID:   uploaderID,
		IsPublic:     true,
		RequiredRole: RoleUser,
		Format:       "mp3",
		Bitrate:      320000,
		SampleRate:   44100,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := track.Validate(); err != nil {
		return nil, err
	}

	return track, nil
}

// Validate vérifie que la piste est valide
func (t *Track) Validate() error {
	if t.Title == "" {
		return errors.New("track title is required")
	}

	if len(t.Title) > 200 {
		return errors.New("track title must be less than 200 characters long")
	}

	if t.Artist == "" {
		return errors.New("track artist is required")
	}

	if len(t.Artist) > 100 {
		return errors.New("track artist must be less than 100 characters long")
	}

	if t.UploaderID == 0 {
		return errors.New("uploader ID is required")
	}

	return nil
}

// CanBePlayedBy vérifie si une piste peut être jouée par un utilisateur
func (t *Track) CanBePlayedBy(user *User) bool {
	if !t.IsPublic && !user.HasAnyRole(t.RequiredRole, RoleAdmin, RoleSuperAdmin) {
		return false
	}

	return true
}

// IncrementPlayCount incrémente le compteur de lectures
func (t *Track) IncrementPlayCount() {
	t.PlayCount++
	t.UpdatedAt = time.Now()
}

// GetDurationFormatted retourne la durée formatée
func (t *Track) GetDurationFormatted() string {
	minutes := t.Duration / 60
	seconds := t.Duration % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
