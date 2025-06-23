// internal/services/track_service.go
package services

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/models"
	"github.com/okinrev/veza-web-app/internal/utils"
)

type TrackService interface {
	CreateTrack(req CreateTrackRequest) (*models.Track, error)
	GetTrack(trackID, userID int) (*models.Track, error)
	UpdateTrack(trackID, userID int, req UpdateTrackRequest) (*models.Track, error)
	DeleteTrack(trackID, userID int) error
	ListTracks(page, limit int, showPrivate bool, userID int) ([]models.Track, int, error)
	SearchTracks(query string, tags []string, userID int, limit int) ([]models.Track, error)
	GetUserTracks(userID, page, limit int) ([]models.Track, int, error)
	ValidateAudioFile(filename string, size int64) error
	GenerateStreamURL(filename string, userID int) (string, error)
	GetTrackStats(trackID int) (*TrackStats, error)
}

type trackService struct {
	db        *database.DB
	jwtSecret string
}

func NewTrackService(db *database.DB, jwtSecret string) TrackService {
	return &trackService{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// Request/Response types
type CreateTrackRequest struct {
	Title           string   `json:"title" validate:"required"`
	Artist          string   `json:"artist" validate:"required"`
	Filename        string   `json:"filename" validate:"required"`
	DurationSeconds *int     `json:"duration_seconds"`
	Tags            []string `json:"tags"`
	IsPublic        bool     `json:"is_public"`
	UploaderID      int      `json:"uploader_id" validate:"required"`
}

type UpdateTrackRequest struct {
	Title    *string   `json:"title,omitempty"`
	Artist   *string   `json:"artist,omitempty"`
	Tags     *[]string `json:"tags,omitempty"`
	IsPublic *bool     `json:"is_public,omitempty"`
}

type TrackStats struct {
	TrackID     int    `json:"track_id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Duration    int    `json:"duration_seconds"`
	PlayCount   int    `json:"play_count"`
	FileSize    int64  `json:"file_size"`
	Format      string `json:"format"`
	CreatedAt   string `json:"created_at"`
}

const (
	MaxAudioSize     = 100 << 20 // 100MB
	MaxAudioDuration = 600       // 10 minutes in seconds
)

// CreateTrack creates a new track record
func (s *trackService) CreateTrack(req CreateTrackRequest) (*models.Track, error) {
	// Validate audio file
	if err := s.ValidateAudioFile(req.Filename, 0); err != nil {
		return nil, fmt.Errorf("invalid audio file: %w", err)
	}

	// Insert track into database
	var track models.Track
	err := s.db.QueryRow(`
		INSERT INTO tracks (title, artist, filename, duration_seconds, tags, is_public, uploader_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, title, artist, filename, duration_seconds, tags, is_public, uploader_id, created_at, updated_at
	`, req.Title, req.Artist, req.Filename, req.DurationSeconds, pq.Array(req.Tags), req.IsPublic, req.UploaderID).Scan(
		&track.ID, &track.Title, &track.Artist, &track.Filename,
		&track.DurationSeconds, pq.Array(&track.Tags), &track.IsPublic,
		&track.UploaderID, &track.CreatedAt, &track.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create track: %w", err)
	}

	return &track, nil
}

// GetTrack retrieves a track by ID with permission checking
func (s *trackService) GetTrack(trackID, userID int) (*models.Track, error) {
	var track models.Track
	err := s.db.QueryRow(`
		SELECT t.id, t.title, t.artist, t.filename, t.duration_seconds, t.tags, 
		       t.is_public, t.uploader_id, t.created_at, t.updated_at
		FROM tracks t
		WHERE t.id = $1 AND (t.is_public = true OR t.uploader_id = $2)
	`, trackID, userID).Scan(
		&track.ID, &track.Title, &track.Artist, &track.Filename,
		&track.DurationSeconds, pq.Array(&track.Tags), &track.IsPublic,
		&track.UploaderID, &track.CreatedAt, &track.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("track not found: %w", err)
	}

	return &track, nil
}

// UpdateTrack updates a track's metadata
func (s *trackService) UpdateTrack(trackID, userID int, req UpdateTrackRequest) (*models.Track, error) {
	// Verify ownership
	var ownerID int
	err := s.db.QueryRow("SELECT uploader_id FROM tracks WHERE id = $1", trackID).Scan(&ownerID)
	if err != nil {
		return nil, fmt.Errorf("track not found")
	}

	if ownerID != userID {
		return nil, fmt.Errorf("not authorized to update this track")
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argCount := 1

	if req.Title != nil {
		setParts = append(setParts, "title = $"+strconv.Itoa(argCount))
		args = append(args, strings.TrimSpace(*req.Title))
		argCount++
	}
	if req.Artist != nil {
		setParts = append(setParts, "artist = $"+strconv.Itoa(argCount))
		args = append(args, strings.TrimSpace(*req.Artist))
		argCount++
	}
	if req.Tags != nil {
		setParts = append(setParts, "tags = $"+strconv.Itoa(argCount))
		args = append(args, pq.Array(*req.Tags))
		argCount++
	}
	if req.IsPublic != nil {
		setParts = append(setParts, "is_public = $"+strconv.Itoa(argCount))
		args = append(args, *req.IsPublic)
		argCount++
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Add updated_at and track_id
	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, trackID)

	query := "UPDATE tracks SET " + strings.Join(setParts, ", ") + " WHERE id = $" + strconv.Itoa(argCount)

	_, err = s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update track: %w", err)
	}

	// Return updated track
	return s.GetTrack(trackID, userID)
}

// DeleteTrack deletes a track and its file
func (s *trackService) DeleteTrack(trackID, userID int) error {
	// Get track details for ownership verification
	var ownerID int
	var filename string
	err := s.db.QueryRow("SELECT uploader_id, filename FROM tracks WHERE id = $1", trackID).Scan(&ownerID, &filename)
	if err != nil {
		return fmt.Errorf("track not found")
	}

	if ownerID != userID {
		return fmt.Errorf("not authorized to delete this track")
	}

	// Delete from database
	_, err = s.db.Exec("DELETE FROM tracks WHERE id = $1", trackID)
	if err != nil {
		return fmt.Errorf("failed to delete track from database: %w", err)
	}

	// Note: File deletion should be handled by the handler/controller layer
	// to avoid tight coupling with filesystem operations

	return nil
}

// ListTracks returns a list of tracks with pagination
func (s *trackService) ListTracks(page, limit int, showPrivate bool, userID int) ([]models.Track, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query based on permissions
	baseQuery := `
		SELECT t.id, t.title, t.artist, t.filename, t.duration_seconds, t.tags, 
		       t.is_public, t.uploader_id, t.created_at, t.updated_at
		FROM tracks t
	`
	countQuery := `SELECT COUNT(*) FROM tracks t`

	whereClause := ""
	args := []interface{}{}

	// Apply visibility filters
	if showPrivate && userID > 0 {
		// Only show user's own tracks if requesting private
		whereClause = " WHERE t.uploader_id = $1"
		args = append(args, userID)
	} else {
		// Only public tracks
		whereClause = " WHERE t.is_public = true"
	}

	// Get total count
	var total int
	err := s.db.QueryRow(countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tracks: %w", err)
	}

	// Get tracks
	orderClause := " ORDER BY t.created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + " OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.db.Query(baseQuery+whereClause+orderClause, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve tracks: %w", err)
	}
	defer rows.Close()

	var tracks []models.Track
	for rows.Next() {
		var track models.Track
		err := rows.Scan(
			&track.ID, &track.Title, &track.Artist, &track.Filename,
			&track.DurationSeconds, pq.Array(&track.Tags), &track.IsPublic,
			&track.UploaderID, &track.CreatedAt, &track.UpdatedAt,
		)
		if err != nil {
			continue
		}
		tracks = append(tracks, track)
	}

	return tracks, total, nil
}

// SearchTracks searches for tracks with filters
func (s *trackService) SearchTracks(query string, tags []string, userID int, limit int) ([]models.Track, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	baseQuery := `
		SELECT t.id, t.title, t.artist, t.filename, t.duration_seconds, t.tags, 
		       t.is_public, t.uploader_id, t.created_at, t.updated_at
		FROM tracks t
		WHERE t.is_public = true
	`

	conditions := []string{}
	args := []interface{}{}
	argCount := 1

	if query != "" {
		conditions = append(conditions, "(LOWER(t.title) LIKE LOWER($"+strconv.Itoa(argCount)+") OR LOWER(t.artist) LIKE LOWER($"+strconv.Itoa(argCount)+"))")
		args = append(args, "%"+query+"%")
		argCount++
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			conditions = append(conditions, "$"+strconv.Itoa(argCount)+" = ANY(t.tags)")
			args = append(args, tag)
			argCount++
		}
	}

	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY t.created_at DESC LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search tracks: %w", err)
	}
	defer rows.Close()

	var tracks []models.Track
	for rows.Next() {
		var track models.Track
		err := rows.Scan(
			&track.ID, &track.Title, &track.Artist, &track.Filename,
			&track.DurationSeconds, pq.Array(&track.Tags), &track.IsPublic,
			&track.UploaderID, &track.CreatedAt, &track.UpdatedAt,
		)
		if err != nil {
			continue
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// GetUserTracks returns tracks uploaded by a specific user
func (s *trackService) GetUserTracks(userID, page, limit int) ([]models.Track, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get total count
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM tracks WHERE uploader_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user tracks: %w", err)
	}

	// Get tracks
	rows, err := s.db.Query(`
		SELECT id, title, artist, filename, duration_seconds, tags, is_public, uploader_id, created_at, updated_at
		FROM tracks 
		WHERE uploader_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)

	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve user tracks: %w", err)
	}
	defer rows.Close()

	var tracks []models.Track
	for rows.Next() {
		var track models.Track
		err := rows.Scan(
			&track.ID, &track.Title, &track.Artist, &track.Filename,
			&track.DurationSeconds, pq.Array(&track.Tags), &track.IsPublic,
			&track.UploaderID, &track.CreatedAt, &track.UpdatedAt,
		)
		if err != nil {
			continue
		}
		tracks = append(tracks, track)
	}

	return tracks, total, nil
}

// ValidateAudioFile validates audio file format and constraints
func (s *trackService) ValidateAudioFile(filename string, size int64) error {
	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := []string{".mp3", ".wav", ".flac", ".ogg", ".m4a", ".aac"}
	validExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			validExt = true
			break
		}
	}

	if !validExt {
		return fmt.Errorf("unsupported audio format: %s", ext)
	}

	// Validate file size if provided
	if size > 0 && size > MaxAudioSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", MaxAudioSize)
	}

	return nil
}

// GenerateStreamURL creates a signed URL for audio streaming
func (s *trackService) GenerateStreamURL(filename string, userID int) (string, error) {
	// Verify track exists and user has access
	var trackExists, isPublic bool
	var uploaderID int
	err := s.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM tracks WHERE filename = $1), is_public, uploader_id
		FROM tracks WHERE filename = $1
	`, filename).Scan(&trackExists, &isPublic, &uploaderID)

	if !trackExists {
		return "", fmt.Errorf("track not found")
	}

	// Check access permissions
	if !isPublic && uploaderID != userID {
		return "", fmt.Errorf("access denied to private track")
	}

	// Generate signed URL
	signedURL, err := utils.GenerateSignedURL(filename, userID, s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return signedURL, nil
}

// GetTrackStats returns statistics for a track
func (s *trackService) GetTrackStats(trackID int) (*TrackStats, error) {
	var stats TrackStats
	err := s.db.QueryRow(`
		SELECT id, title, artist, COALESCE(duration_seconds, 0), created_at
		FROM tracks WHERE id = $1 AND is_public = true
	`, trackID).Scan(&stats.TrackID, &stats.Title, &stats.Artist, &stats.Duration, &stats.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("track not found: %w", err)
	}

	// Note: PlayCount would require a separate plays tracking table
	stats.PlayCount = 0

	return &stats, nil
}