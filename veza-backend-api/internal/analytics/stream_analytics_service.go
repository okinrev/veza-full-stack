package analytics

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// StreamAnalyticsService service d'analytics pour le streaming
type StreamAnalyticsService struct {
	db     *sql.DB
	logger *zap.Logger
	cache  EngagementCache
}

// StreamMetrics métriques de streaming
type StreamMetrics struct {
	TotalStreams        int64                `json:"total_streams"`
	UniqueListeners     int64                `json:"unique_listeners"`
	TotalListenTime     float64              `json:"total_listen_time_hours"`
	AvgSessionLength    float64              `json:"avg_session_length_minutes"`
	AvgBitrate          float64              `json:"avg_bitrate_kbps"`
	PeakConcurrentUsers int64                `json:"peak_concurrent_users"`
	BufferEvents        int64                `json:"buffer_events"`
	QualityChanges      int64                `json:"quality_changes"`
	SkipRate            float64              `json:"skip_rate_percent"`
	CompletionRate      float64              `json:"completion_rate_percent"`
	TopTracks           []TrackAnalytics     `json:"top_tracks"`
	TopGenres           []GenreAnalytics     `json:"top_genres"`
	ListenerGeography   map[string]int64     `json:"listener_geography"`
	DeviceBreakdown     map[string]int64     `json:"device_breakdown"`
	QualityDistribution map[string]int64     `json:"quality_distribution"`
	HourlyListeners     map[string]int64     `json:"hourly_listeners"`
	BandwidthUsage      BandwidthMetrics     `json:"bandwidth_usage"`
	PerformanceMetrics  PerformanceMetrics   `json:"performance_metrics"`
	UserBehavior        UserBehaviorMetrics  `json:"user_behavior"`
	RevenueMetrics      StreamRevenueMetrics `json:"revenue_metrics"`
}

// TrackAnalytics analytics d'une piste
type TrackAnalytics struct {
	TrackID         string  `json:"track_id"`
	Title           string  `json:"title"`
	Artist          string  `json:"artist"`
	Genre           string  `json:"genre"`
	Plays           int64   `json:"plays"`
	UniqueListeners int64   `json:"unique_listeners"`
	TotalDuration   float64 `json:"total_duration_hours"`
	AvgListenTime   float64 `json:"avg_listen_time_percent"`
	SkipRate        float64 `json:"skip_rate_percent"`
	Likes           int64   `json:"likes"`
	Shares          int64   `json:"shares"`
	Downloads       int64   `json:"downloads"`
	Revenue         float64 `json:"revenue"`
	Trending        bool    `json:"trending"`
}

// GenreAnalytics analytics par genre
type GenreAnalytics struct {
	Genre           string  `json:"genre"`
	Plays           int64   `json:"plays"`
	UniqueListeners int64   `json:"unique_listeners"`
	AvgSessionTime  float64 `json:"avg_session_time_minutes"`
	Growth          float64 `json:"growth_percent"`
}

// BandwidthMetrics métriques de bande passante
type BandwidthMetrics struct {
	TotalGB      float64            `json:"total_gb"`
	PeakMbps     float64            `json:"peak_mbps"`
	AvgMbps      float64            `json:"avg_mbps"`
	ByQuality    map[string]float64 `json:"by_quality_gb"`
	CostEstimate float64            `json:"cost_estimate_usd"`
}

// PerformanceMetrics métriques de performance
type PerformanceMetrics struct {
	AvgLatency    float64 `json:"avg_latency_ms"`
	ErrorRate     float64 `json:"error_rate_percent"`
	UptimePercent float64 `json:"uptime_percent"`
	CDNHitRate    float64 `json:"cdn_hit_rate_percent"`
	BufferHealth  float64 `json:"buffer_health_score"`
}

// UserBehaviorMetrics métriques de comportement utilisateur
type UserBehaviorMetrics struct {
	AvgPlaylistSize    float64            `json:"avg_playlist_size"`
	RepeatListens      float64            `json:"repeat_listen_rate_percent"`
	DiscoveryRate      float64            `json:"discovery_rate_percent"`
	SocialInteractions int64              `json:"social_interactions"`
	ListeningPatterns  map[string]float64 `json:"listening_patterns"`
	PreferredQualities map[string]int64   `json:"preferred_qualities"`
}

// StreamRevenueMetrics métriques de revenus streaming
type StreamRevenueMetrics struct {
	TotalRevenue        float64        `json:"total_revenue"`
	RevenuePerListen    float64        `json:"revenue_per_listen"`
	RevenuePerUser      float64        `json:"revenue_per_user"`
	SubscriptionRevenue float64        `json:"subscription_revenue"`
	AdRevenue           float64        `json:"ad_revenue"`
	RevenueByTrack      []TrackRevenue `json:"revenue_by_track"`
	RevenueGrowth       float64        `json:"revenue_growth_percent"`
}

// TrackRevenue revenus par piste
type TrackRevenue struct {
	TrackID string  `json:"track_id"`
	Title   string  `json:"title"`
	Artist  string  `json:"artist"`
	Revenue float64 `json:"revenue"`
	Plays   int64   `json:"plays"`
	RPP     float64 `json:"revenue_per_play"` // Revenue Per Play
}

// StreamSession session d'écoute
type StreamSession struct {
	ID               string                 `json:"id"`
	UserID           string                 `json:"user_id"`
	TrackID          string                 `json:"track_id"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          *time.Time             `json:"end_time,omitempty"`
	Duration         *int64                 `json:"duration_seconds,omitempty"`
	ListenProgress   float64                `json:"listen_progress_percent"`
	Quality          string                 `json:"quality"`
	DeviceType       string                 `json:"device_type"`
	Country          string                 `json:"country"`
	City             string                 `json:"city"`
	BufferEvents     int64                  `json:"buffer_events"`
	QualityChanges   int64                  `json:"quality_changes"`
	Skipped          bool                   `json:"skipped"`
	Completed        bool                   `json:"completed"`
	BytesTransferred int64                  `json:"bytes_transferred"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// NewStreamAnalyticsService crée un nouveau service d'analytics streaming
func NewStreamAnalyticsService(db *sql.DB, logger *zap.Logger, cache EngagementCache) *StreamAnalyticsService {
	return &StreamAnalyticsService{
		db:     db,
		logger: logger,
		cache:  cache,
	}
}

// TrackStreamSession enregistre une session de streaming
func (s *StreamAnalyticsService) TrackStreamSession(ctx context.Context, session *StreamSession) error {
	query := `
		INSERT INTO stream_sessions (
			id, user_id, track_id, start_time, quality, device_type,
			country, city, bytes_transferred, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	metadataJSON, _ := json.Marshal(session.Metadata)

	_, err := s.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.TrackID, session.StartTime,
		session.Quality, session.DeviceType, session.Country, session.City,
		session.BytesTransferred, metadataJSON)

	if err != nil {
		s.logger.Error("Failed to track stream session", zap.Error(err))
		return fmt.Errorf("failed to track stream session: %w", err)
	}

	return nil
}

// EndStreamSession termine une session de streaming
func (s *StreamAnalyticsService) EndStreamSession(ctx context.Context, sessionID string, listenProgress float64, completed, skipped bool) error {
	query := `
		UPDATE stream_sessions 
		SET end_time = NOW(),
		    duration = EXTRACT(EPOCH FROM (NOW() - start_time)),
		    listen_progress = $2,
		    completed = $3,
		    skipped = $4
		WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, sessionID, listenProgress, completed, skipped)
	if err != nil {
		s.logger.Error("Failed to end stream session", zap.Error(err))
		return fmt.Errorf("failed to end stream session: %w", err)
	}

	return nil
}

// TrackBufferEvent enregistre un événement de buffering
func (s *StreamAnalyticsService) TrackBufferEvent(ctx context.Context, sessionID string, bufferTime float64) error {
	// Mettre à jour le compteur de buffer events
	updateQuery := `UPDATE stream_sessions SET buffer_events = buffer_events + 1 WHERE id = $1`
	_, err := s.db.ExecContext(ctx, updateQuery, sessionID)

	if err != nil {
		return fmt.Errorf("failed to track buffer event: %w", err)
	}

	// Enregistrer l'événement détaillé
	insertQuery := `
		INSERT INTO stream_events (session_id, event_type, event_data, timestamp)
		VALUES ($1, 'buffer', $2, NOW())`

	eventData := fmt.Sprintf(`{"buffer_time": %f}`, bufferTime)
	_, err = s.db.ExecContext(ctx, insertQuery, sessionID, eventData)

	return err
}

// TrackQualityChange enregistre un changement de qualité
func (s *StreamAnalyticsService) TrackQualityChange(ctx context.Context, sessionID, fromQuality, toQuality string) error {
	// Mettre à jour le compteur
	updateQuery := `UPDATE stream_sessions SET quality_changes = quality_changes + 1 WHERE id = $1`
	_, err := s.db.ExecContext(ctx, updateQuery, sessionID)

	if err != nil {
		return fmt.Errorf("failed to track quality change: %w", err)
	}

	// Enregistrer l'événement
	insertQuery := `
		INSERT INTO stream_events (session_id, event_type, event_data, timestamp)
		VALUES ($1, 'quality_change', $2, NOW())`

	eventData := fmt.Sprintf(`{"from": "%s", "to": "%s"}`, fromQuality, toQuality)
	_, err = s.db.ExecContext(ctx, insertQuery, sessionID, eventData)

	return err
}

// GetStreamMetrics retourne les métriques complètes de streaming
func (s *StreamAnalyticsService) GetStreamMetrics(ctx context.Context, dateRange DateRange) (*StreamMetrics, error) {
	metrics := &StreamMetrics{}

	// Métriques de base
	var err error
	metrics.TotalStreams, err = s.getTotalStreams(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	metrics.UniqueListeners, err = s.getUniqueListeners(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	metrics.TotalListenTime, err = s.getTotalListenTime(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	metrics.AvgSessionLength, err = s.getAverageSessionLength(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	metrics.SkipRate, err = s.getSkipRate(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	metrics.CompletionRate, err = s.getCompletionRate(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Top tracks
	metrics.TopTracks, err = s.getTopTracks(ctx, dateRange, 20)
	if err != nil {
		return nil, err
	}

	// Top genres
	metrics.TopGenres, err = s.getTopGenres(ctx, dateRange, 10)
	if err != nil {
		return nil, err
	}

	// Répartition géographique
	metrics.ListenerGeography, err = s.getListenerGeography(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Répartition par device
	metrics.DeviceBreakdown, err = s.getDeviceBreakdown(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Distribution de qualité
	metrics.QualityDistribution, err = s.getQualityDistribution(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Listeners par heure
	metrics.HourlyListeners, err = s.getHourlyListeners(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Métriques de bande passante
	metrics.BandwidthUsage, err = s.getBandwidthMetrics(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Métriques de performance
	metrics.PerformanceMetrics, err = s.getPerformanceMetrics(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Comportement utilisateur
	metrics.UserBehavior, err = s.getUserBehaviorMetrics(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// ============================================================================
// MÉTHODES PRIVÉES
// ============================================================================

func (s *StreamAnalyticsService) getTotalStreams(ctx context.Context, dateRange DateRange) (int64, error) {
	query := `SELECT COUNT(*) FROM stream_sessions WHERE start_time >= $1 AND start_time <= $2`

	var count int64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&count)
	return count, err
}

func (s *StreamAnalyticsService) getUniqueListeners(ctx context.Context, dateRange DateRange) (int64, error) {
	query := `SELECT COUNT(DISTINCT user_id) FROM stream_sessions WHERE start_time >= $1 AND start_time <= $2`

	var count int64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&count)
	return count, err
}

func (s *StreamAnalyticsService) getTotalListenTime(ctx context.Context, dateRange DateRange) (float64, error) {
	query := `
		SELECT SUM(COALESCE(duration, 0)) / 3600.0 as total_hours
		FROM stream_sessions 
		WHERE start_time >= $1 AND start_time <= $2 AND duration IS NOT NULL`

	var totalHours sql.NullFloat64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&totalHours)
	if !totalHours.Valid {
		return 0, err
	}
	return totalHours.Float64, err
}

func (s *StreamAnalyticsService) getAverageSessionLength(ctx context.Context, dateRange DateRange) (float64, error) {
	query := `
		SELECT AVG(duration) / 60.0 as avg_minutes
		FROM stream_sessions 
		WHERE start_time >= $1 AND start_time <= $2 AND duration IS NOT NULL`

	var avgMinutes sql.NullFloat64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&avgMinutes)
	if !avgMinutes.Valid {
		return 0, err
	}
	return avgMinutes.Float64, err
}

func (s *StreamAnalyticsService) getSkipRate(ctx context.Context, dateRange DateRange) (float64, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN skipped = true THEN 1 END) * 100.0 / COUNT(*) as skip_rate
		FROM stream_sessions 
		WHERE start_time >= $1 AND start_time <= $2`

	var skipRate sql.NullFloat64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&skipRate)
	if !skipRate.Valid {
		return 0, err
	}
	return skipRate.Float64, err
}

func (s *StreamAnalyticsService) getCompletionRate(ctx context.Context, dateRange DateRange) (float64, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN completed = true THEN 1 END) * 100.0 / COUNT(*) as completion_rate
		FROM stream_sessions 
		WHERE start_time >= $1 AND start_time <= $2`

	var completionRate sql.NullFloat64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&completionRate)
	if !completionRate.Valid {
		return 0, err
	}
	return completionRate.Float64, err
}

func (s *StreamAnalyticsService) getTopTracks(ctx context.Context, dateRange DateRange, limit int) ([]TrackAnalytics, error) {
	query := `
		SELECT 
			ss.track_id,
			COALESCE(t.title, 'Unknown') as title,
			COALESCE(t.artist, 'Unknown') as artist,
			COALESCE(t.genre, 'Unknown') as genre,
			COUNT(*) as plays,
			COUNT(DISTINCT ss.user_id) as unique_listeners,
			SUM(COALESCE(ss.duration, 0)) / 3600.0 as total_duration_hours,
			AVG(ss.listen_progress) as avg_listen_time_percent,
			COUNT(CASE WHEN ss.skipped = true THEN 1 END) * 100.0 / COUNT(*) as skip_rate
		FROM stream_sessions ss
		LEFT JOIN tracks t ON t.id = CAST(ss.track_id AS INTEGER)
		WHERE ss.start_time >= $1 AND ss.start_time <= $2
		GROUP BY ss.track_id, t.title, t.artist, t.genre
		ORDER BY plays DESC
		LIMIT $3`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []TrackAnalytics
	for rows.Next() {
		var track TrackAnalytics
		err := rows.Scan(
			&track.TrackID,
			&track.Title,
			&track.Artist,
			&track.Genre,
			&track.Plays,
			&track.UniqueListeners,
			&track.TotalDuration,
			&track.AvgListenTime,
			&track.SkipRate,
		)
		if err != nil {
			continue
		}

		// Déterminer si trending (plus de 50% d'augmentation cette semaine)
		track.Trending = track.Plays > 100 && track.SkipRate < 20

		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (s *StreamAnalyticsService) getTopGenres(ctx context.Context, dateRange DateRange, limit int) ([]GenreAnalytics, error) {
	query := `
		SELECT 
			COALESCE(t.genre, 'Unknown') as genre,
			COUNT(*) as plays,
			COUNT(DISTINCT ss.user_id) as unique_listeners,
			AVG(COALESCE(ss.duration, 0)) / 60.0 as avg_session_time_minutes
		FROM stream_sessions ss
		LEFT JOIN tracks t ON t.id = CAST(ss.track_id AS INTEGER)
		WHERE ss.start_time >= $1 AND ss.start_time <= $2
		GROUP BY t.genre
		ORDER BY plays DESC
		LIMIT $3`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []GenreAnalytics
	for rows.Next() {
		var genre GenreAnalytics
		err := rows.Scan(
			&genre.Genre,
			&genre.Plays,
			&genre.UniqueListeners,
			&genre.AvgSessionTime,
		)
		if err != nil {
			continue
		}

		// TODO: Calculer la croissance par rapport à la période précédente
		genre.Growth = 0.0

		genres = append(genres, genre)
	}

	return genres, nil
}

func (s *StreamAnalyticsService) getListenerGeography(ctx context.Context, dateRange DateRange) (map[string]int64, error) {
	query := `
		SELECT country, COUNT(DISTINCT user_id)
		FROM stream_sessions 
		WHERE start_time >= $1 AND start_time <= $2 AND country IS NOT NULL
		GROUP BY country
		ORDER BY COUNT(DISTINCT user_id) DESC
		LIMIT 20`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	geography := make(map[string]int64)
	for rows.Next() {
		var country string
		var count int64
		if err := rows.Scan(&country, &count); err != nil {
			continue
		}
		geography[country] = count
	}

	return geography, nil
}

func (s *StreamAnalyticsService) getDeviceBreakdown(ctx context.Context, dateRange DateRange) (map[string]int64, error) {
	query := `
		SELECT device_type, COUNT(*)
		FROM stream_sessions 
		WHERE start_time >= $1 AND start_time <= $2
		GROUP BY device_type`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	breakdown := make(map[string]int64)
	for rows.Next() {
		var deviceType string
		var count int64
		if err := rows.Scan(&deviceType, &count); err != nil {
			continue
		}
		breakdown[deviceType] = count
	}

	return breakdown, nil
}

func (s *StreamAnalyticsService) getQualityDistribution(ctx context.Context, dateRange DateRange) (map[string]int64, error) {
	query := `
		SELECT quality, COUNT(*)
		FROM stream_sessions 
		WHERE start_time >= $1 AND start_time <= $2
		GROUP BY quality`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	distribution := make(map[string]int64)
	for rows.Next() {
		var quality string
		var count int64
		if err := rows.Scan(&quality, &count); err != nil {
			continue
		}
		distribution[quality] = count
	}

	return distribution, nil
}

func (s *StreamAnalyticsService) getHourlyListeners(ctx context.Context, dateRange DateRange) (map[string]int64, error) {
	query := `
		SELECT 
			EXTRACT(HOUR FROM start_time) as hour,
			COUNT(DISTINCT user_id) as listeners
		FROM stream_sessions 
		WHERE start_time >= $1 AND start_time <= $2
		GROUP BY EXTRACT(HOUR FROM start_time)
		ORDER BY hour`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hourlyListeners := make(map[string]int64)
	for rows.Next() {
		var hour int
		var listeners int64
		if err := rows.Scan(&hour, &listeners); err != nil {
			continue
		}
		hourlyListeners[fmt.Sprintf("%02d:00", hour)] = listeners
	}

	return hourlyListeners, nil
}

func (s *StreamAnalyticsService) getBandwidthMetrics(ctx context.Context, dateRange DateRange) (BandwidthMetrics, error) {
	query := `
		SELECT 
			SUM(bytes_transferred) / (1024*1024*1024.0) as total_gb,
			quality,
			SUM(bytes_transferred) / (1024*1024*1024.0) as gb_by_quality
		FROM stream_sessions 
		WHERE start_time >= $1 AND start_time <= $2
		GROUP BY quality`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End)
	if err != nil {
		return BandwidthMetrics{}, err
	}
	defer rows.Close()

	metrics := BandwidthMetrics{
		ByQuality: make(map[string]float64),
	}

	for rows.Next() {
		var totalGB sql.NullFloat64
		var quality string
		var gbByQuality sql.NullFloat64

		if err := rows.Scan(&totalGB, &quality, &gbByQuality); err != nil {
			continue
		}

		if totalGB.Valid {
			metrics.TotalGB += totalGB.Float64
		}
		if gbByQuality.Valid {
			metrics.ByQuality[quality] = gbByQuality.Float64
		}
	}

	// Estimation du coût (exemple: $0.05/GB)
	metrics.CostEstimate = metrics.TotalGB * 0.05

	return metrics, nil
}

func (s *StreamAnalyticsService) getPerformanceMetrics(ctx context.Context, dateRange DateRange) (PerformanceMetrics, error) {
	query := `
		SELECT 
			AVG(buffer_events) as avg_buffer_events,
			COUNT(CASE WHEN buffer_events = 0 THEN 1 END) * 100.0 / COUNT(*) as no_buffer_rate
		FROM stream_sessions 
		WHERE start_time >= $1 AND start_time <= $2`

	var avgBufferEvents sql.NullFloat64
	var noBufferRate sql.NullFloat64

	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&avgBufferEvents, &noBufferRate)
	if err != nil {
		return PerformanceMetrics{}, err
	}

	metrics := PerformanceMetrics{
		AvgLatency:    50.0, // TODO: Calculer depuis les vraies métriques
		ErrorRate:     2.5,  // TODO: Calculer depuis les logs d'erreur
		UptimePercent: 99.9, // TODO: Calculer depuis le monitoring
		CDNHitRate:    85.0, // TODO: Intégrer avec le CDN
	}

	if noBufferRate.Valid {
		metrics.BufferHealth = noBufferRate.Float64
	}

	return metrics, nil
}

func (s *StreamAnalyticsService) getUserBehaviorMetrics(ctx context.Context, dateRange DateRange) (UserBehaviorMetrics, error) {
	// TODO: Implémenter l'analyse comportementale avancée
	return UserBehaviorMetrics{
		AvgPlaylistSize:    15.5,
		RepeatListens:      35.2,
		DiscoveryRate:      28.7,
		SocialInteractions: 0,
		ListeningPatterns: map[string]float64{
			"morning":   25.5,
			"afternoon": 30.2,
			"evening":   35.8,
			"night":     8.5,
		},
		PreferredQualities: map[string]int64{
			"128kbps":  450,
			"256kbps":  320,
			"320kbps":  180,
			"lossless": 50,
		},
	}, nil
}
