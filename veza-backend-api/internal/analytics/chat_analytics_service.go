package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ChatAnalyticsService service d'analytics pour le chat
type ChatAnalyticsService struct {
	db     *sql.DB
	logger *zap.Logger
	cache  EngagementCache
}

// ChatMetrics métriques du chat
type ChatMetrics struct {
	TotalMessages        int64                    `json:"total_messages"`
	ActiveRooms          int64                    `json:"active_rooms"`
	ActiveUsers          int64                    `json:"active_users"`
	MessagesPerUser      float64                  `json:"avg_messages_per_user"`
	MessagesPerRoom      float64                  `json:"avg_messages_per_room"`
	MessageLength        float64                  `json:"avg_message_length"`
	ResponseTime         float64                  `json:"avg_response_time_minutes"`
	PeakConcurrentUsers  int64                    `json:"peak_concurrent_users"`
	MessageFrequency     map[string]int64         `json:"message_frequency_by_hour"`
	TopRooms             []RoomActivity           `json:"top_rooms"`
	TopUsers             []UserActivity           `json:"top_users"`
	MessageTypes         map[string]int64         `json:"message_types"`
	EmojiUsage           map[string]int64         `json:"emoji_usage"`
	ModeratedContent     int64                    `json:"moderated_content"`
	FileShares           int64                    `json:"file_shares"`
	Reactions            int64                    `json:"total_reactions"`
	ThreadActivity       int64                    `json:"thread_activity"`
	UserGrowth           []DailyGrowth            `json:"user_growth"`
	RetentionMetrics     UserRetentionMetrics     `json:"retention_metrics"`
}

// RoomActivity activité d'une room
type RoomActivity struct {
	RoomID            string  `json:"room_id"`
	RoomName          string  `json:"room_name"`
	MessageCount      int64   `json:"message_count"`
	UniqueUsers       int64   `json:"unique_users"`
	AvgSessionLength  float64 `json:"avg_session_length_minutes"`
	LastActivity      time.Time `json:"last_activity"`
	IsActive          bool    `json:"is_active"`
}

// UserActivity activité d'un utilisateur
type UserActivity struct {
	UserID           string    `json:"user_id"`
	Username         string    `json:"username"`
	MessageCount     int64     `json:"message_count"`
	RoomsVisited     int64     `json:"rooms_visited"`
	SessionLength    float64   `json:"session_length_minutes"`
	LastSeen         time.Time `json:"last_seen"`
	EngagementLevel  string    `json:"engagement_level"` // high, medium, low
}

// UserRetentionMetrics métriques de rétention
type UserRetentionMetrics struct {
	Day1Retention  float64 `json:"day1_retention"`
	Day7Retention  float64 `json:"day7_retention"`
	Day30Retention float64 `json:"day30_retention"`
	ChurnRate      float64 `json:"churn_rate"`
}

// DailyGrowth croissance quotidienne
type DailyGrowth struct {
	Date        time.Time `json:"date"`
	NewUsers    int64     `json:"new_users"`
	ActiveUsers int64     `json:"active_users"`
	Messages    int64     `json:"messages"`
}

// NewChatAnalyticsService crée un nouveau service d'analytics chat
func NewChatAnalyticsService(db *sql.DB, logger *zap.Logger, cache EngagementCache) *ChatAnalyticsService {
	return &ChatAnalyticsService{
		db:     db,
		logger: logger,
		cache:  cache,
	}
}

// GetChatMetrics retourne les métriques complètes du chat
func (s *ChatAnalyticsService) GetChatMetrics(ctx context.Context, dateRange DateRange) (*ChatMetrics, error) {
	metrics := &ChatMetrics{}

	// Messages totaux
	var err error
	metrics.TotalMessages, err = s.getTotalMessages(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Rooms actives
	metrics.ActiveRooms, err = s.getActiveRooms(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Utilisateurs actifs
	metrics.ActiveUsers, err = s.getActiveUsers(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Moyennes
	metrics.MessagesPerUser = float64(metrics.TotalMessages) / float64(max(metrics.ActiveUsers, 1))
	metrics.MessagesPerRoom = float64(metrics.TotalMessages) / float64(max(metrics.ActiveRooms, 1))

	// Longueur moyenne des messages
	metrics.MessageLength, err = s.getAverageMessageLength(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Temps de réponse moyen
	metrics.ResponseTime, err = s.getAverageResponseTime(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Top rooms et users
	metrics.TopRooms, err = s.getTopRooms(ctx, dateRange, 10)
	if err != nil {
		return nil, err
	}

	metrics.TopUsers, err = s.getTopUsers(ctx, dateRange, 10)
	if err != nil {
		return nil, err
	}

	// Fréquence des messages par heure
	metrics.MessageFrequency, err = s.getMessageFrequencyByHour(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Types de messages
	metrics.MessageTypes, err = s.getMessageTypes(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Contenu modéré
	metrics.ModeratedContent, err = s.getModeratedContent(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Partages de fichiers
	metrics.FileShares, err = s.getFileShares(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	// Réactions totales
	metrics.Reactions, err = s.getTotalReactions(ctx, dateRange)
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

// TrackMessage enregistre un message pour l'analytics
func (s *ChatAnalyticsService) TrackMessage(ctx context.Context, userID, roomID, content string, messageType string) error {
	query := `
		INSERT INTO chat_analytics (
			user_id, room_id, message_type, content_length, timestamp
		) VALUES ($1, $2, $3, $4, NOW())`

	_, err := s.db.ExecContext(ctx, query, userID, roomID, messageType, len(content))
	if err != nil {
		s.logger.Error("Failed to track chat message", zap.Error(err))
		return fmt.Errorf("failed to track chat message: %w", err)
	}

	return nil
}

// ============================================================================
// MÉTHODES PRIVÉES
// ============================================================================

func (s *ChatAnalyticsService) getTotalMessages(ctx context.Context, dateRange DateRange) (int64, error) {
	query := `SELECT COUNT(*) FROM chat_analytics WHERE timestamp >= $1 AND timestamp <= $2`
	
	var count int64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&count)
	return count, err
}

func (s *ChatAnalyticsService) getActiveRooms(ctx context.Context, dateRange DateRange) (int64, error) {
	query := `SELECT COUNT(DISTINCT room_id) FROM chat_analytics WHERE timestamp >= $1 AND timestamp <= $2`
	
	var count int64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&count)
	return count, err
}

func (s *ChatAnalyticsService) getActiveUsers(ctx context.Context, dateRange DateRange) (int64, error) {
	query := `SELECT COUNT(DISTINCT user_id) FROM chat_analytics WHERE timestamp >= $1 AND timestamp <= $2`
	
	var count int64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&count)
	return count, err
}

func (s *ChatAnalyticsService) getAverageMessageLength(ctx context.Context, dateRange DateRange) (float64, error) {
	query := `SELECT AVG(content_length) FROM chat_analytics WHERE timestamp >= $1 AND timestamp <= $2`
	
	var avg sql.NullFloat64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&avg)
	if !avg.Valid {
		return 0, err
	}
	return avg.Float64, err
}

func (s *ChatAnalyticsService) getAverageResponseTime(ctx context.Context, dateRange DateRange) (float64, error) {
	// Calcul du temps de réponse moyen entre messages dans une room
	query := `
		WITH message_gaps AS (
			SELECT 
				room_id,
				timestamp,
				LAG(timestamp) OVER (PARTITION BY room_id ORDER BY timestamp) as prev_timestamp
			FROM chat_analytics 
			WHERE timestamp >= $1 AND timestamp <= $2
		)
		SELECT AVG(EXTRACT(EPOCH FROM (timestamp - prev_timestamp)) / 60.0)
		FROM message_gaps 
		WHERE prev_timestamp IS NOT NULL 
		AND EXTRACT(EPOCH FROM (timestamp - prev_timestamp)) < 3600` // < 1 hour

	var avg sql.NullFloat64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&avg)
	if !avg.Valid {
		return 0, err
	}
	return avg.Float64, err
}

func (s *ChatAnalyticsService) getTopRooms(ctx context.Context, dateRange DateRange, limit int) ([]RoomActivity, error) {
	query := `
		SELECT 
			ca.room_id,
			COALESCE(r.name, 'Unknown') as room_name,
			COUNT(*) as message_count,
			COUNT(DISTINCT ca.user_id) as unique_users,
			MAX(ca.timestamp) as last_activity
		FROM chat_analytics ca
		LEFT JOIN rooms r ON r.id = CAST(ca.room_id AS INTEGER)
		WHERE ca.timestamp >= $1 AND ca.timestamp <= $2
		GROUP BY ca.room_id, r.name
		ORDER BY message_count DESC
		LIMIT $3`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []RoomActivity
	for rows.Next() {
		var room RoomActivity
		err := rows.Scan(
			&room.RoomID,
			&room.RoomName,
			&room.MessageCount,
			&room.UniqueUsers,
			&room.LastActivity,
		)
		if err != nil {
			continue
		}
		
		room.IsActive = time.Since(room.LastActivity) < 24*time.Hour
		rooms = append(rooms, room)
	}

	return rooms, nil
}

func (s *ChatAnalyticsService) getTopUsers(ctx context.Context, dateRange DateRange, limit int) ([]UserActivity, error) {
	query := `
		SELECT 
			ca.user_id,
			COALESCE(u.username, 'Unknown') as username,
			COUNT(*) as message_count,
			COUNT(DISTINCT ca.room_id) as rooms_visited,
			MAX(ca.timestamp) as last_seen
		FROM chat_analytics ca
		LEFT JOIN users u ON u.id = ca.user_id
		WHERE ca.timestamp >= $1 AND ca.timestamp <= $2
		GROUP BY ca.user_id, u.username
		ORDER BY message_count DESC
		LIMIT $3`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserActivity
	for rows.Next() {
		var user UserActivity
		err := rows.Scan(
			&user.UserID,
			&user.Username,
			&user.MessageCount,
			&user.RoomsVisited,
			&user.LastSeen,
		)
		if err != nil {
			continue
		}

		// Déterminer le niveau d'engagement
		if user.MessageCount >= 100 {
			user.EngagementLevel = "high"
		} else if user.MessageCount >= 20 {
			user.EngagementLevel = "medium"
		} else {
			user.EngagementLevel = "low"
		}

		users = append(users, user)
	}

	return users, nil
}

func (s *ChatAnalyticsService) getMessageFrequencyByHour(ctx context.Context, dateRange DateRange) (map[string]int64, error) {
	query := `
		SELECT 
			EXTRACT(HOUR FROM timestamp) as hour,
			COUNT(*) as count
		FROM chat_analytics 
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY EXTRACT(HOUR FROM timestamp)
		ORDER BY hour`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	frequency := make(map[string]int64)
	for rows.Next() {
		var hour int
		var count int64
		if err := rows.Scan(&hour, &count); err != nil {
			continue
		}
		frequency[fmt.Sprintf("%02d:00", hour)] = count
	}

	return frequency, nil
}

func (s *ChatAnalyticsService) getMessageTypes(ctx context.Context, dateRange DateRange) (map[string]int64, error) {
	query := `
		SELECT message_type, COUNT(*)
		FROM chat_analytics 
		WHERE timestamp >= $1 AND timestamp <= $2
		GROUP BY message_type`

	rows, err := s.db.QueryContext(ctx, query, dateRange.Start, dateRange.End)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	types := make(map[string]int64)
	for rows.Next() {
		var messageType string
		var count int64
		if err := rows.Scan(&messageType, &count); err != nil {
			continue
		}
		types[messageType] = count
	}

	return types, nil
}

func (s *ChatAnalyticsService) getModeratedContent(ctx context.Context, dateRange DateRange) (int64, error) {
	query := `
		SELECT COUNT(*) 
		FROM moderation_logs 
		WHERE created_at >= $1 AND created_at <= $2 AND action IN ('delete', 'hide', 'warn')`
	
	var count int64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&count)
	return count, err
}

func (s *ChatAnalyticsService) getFileShares(ctx context.Context, dateRange DateRange) (int64, error) {
	query := `
		SELECT COUNT(*) 
		FROM chat_analytics 
		WHERE timestamp >= $1 AND timestamp <= $2 AND message_type = 'file'`
	
	var count int64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&count)
	return count, err
}

func (s *ChatAnalyticsService) getTotalReactions(ctx context.Context, dateRange DateRange) (int64, error) {
	query := `
		SELECT COUNT(*) 
		FROM message_reactions 
		WHERE created_at >= $1 AND created_at <= $2`
	
	var count int64
	err := s.db.QueryRowContext(ctx, query, dateRange.Start, dateRange.End).Scan(&count)
	return count, err
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
