package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// PostgreSQLStorage implémentation PostgreSQL du stockage de notifications
type PostgreSQLStorage struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewPostgreSQLStorage crée un nouveau stockage PostgreSQL
func NewPostgreSQLStorage(db *sql.DB, logger *zap.Logger) *PostgreSQLStorage {
	return &PostgreSQLStorage{
		db:     db,
		logger: logger,
	}
}

// Store stocke une notification en base de données
func (s *PostgreSQLStorage) Store(ctx context.Context, notification *Notification) error {
	query := `
		INSERT INTO notifications (
			id, type, user_id, title, message, data, priority, channels,
			created_at, expires_at, source, tags, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)`

	// Sérialiser les données JSON
	dataJSON, err := json.Marshal(notification.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	channelsJSON, err := json.Marshal(notification.Channels)
	if err != nil {
		return fmt.Errorf("failed to marshal channels: %w", err)
	}

	tagsJSON, err := json.Marshal(notification.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	metadataJSON, err := json.Marshal(notification.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query,
		notification.ID,
		string(notification.Type),
		notification.UserID,
		notification.Title,
		notification.Message,
		dataJSON,
		string(notification.Priority),
		channelsJSON,
		notification.CreatedAt,
		notification.ExpiresAt,
		notification.Source,
		tagsJSON,
		metadataJSON,
	)

	if err != nil {
		s.logger.Error("Failed to store notification", zap.Error(err))
		return fmt.Errorf("failed to store notification: %w", err)
	}

	s.logger.Debug("Notification stored successfully",
		zap.String("id", notification.ID),
		zap.String("user_id", notification.UserID))

	return nil
}

// GetByUser récupère les notifications d'un utilisateur
func (s *PostgreSQLStorage) GetByUser(ctx context.Context, userID string, limit, offset int) ([]*Notification, error) {
	query := `
		SELECT 
			id, type, user_id, title, message, data, priority, channels,
			created_at, expires_at, read_at, delivered_at, source, tags, metadata
		FROM notifications 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		notification, err := s.scanNotification(rows)
		if err != nil {
			s.logger.Error("Failed to scan notification", zap.Error(err))
			continue
		}
		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return notifications, nil
}

// MarkAsRead marque une notification comme lue
func (s *PostgreSQLStorage) MarkAsRead(ctx context.Context, notificationID, userID string) error {
	query := `
		UPDATE notifications 
		SET read_at = NOW() 
		WHERE id = $1 AND user_id = $2 AND read_at IS NULL`

	result, err := s.db.ExecContext(ctx, query, notificationID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found or already read")
	}

	s.logger.Debug("Notification marked as read",
		zap.String("id", notificationID),
		zap.String("user_id", userID))

	return nil
}

// GetUnreadCount retourne le nombre de notifications non lues
func (s *PostgreSQLStorage) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM notifications 
		WHERE user_id = $1 AND read_at IS NULL AND (expires_at IS NULL OR expires_at > NOW())`

	var count int
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// DeleteExpired supprime les notifications expirées
func (s *PostgreSQLStorage) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM notifications WHERE expires_at IS NOT NULL AND expires_at < NOW()`

	result, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired notifications: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		s.logger.Info("Deleted expired notifications",
			zap.Int64("count", rowsAffected))
	}

	return nil
}

// scanNotification scanne une ligne de résultat vers une notification
func (s *PostgreSQLStorage) scanNotification(rows *sql.Rows) (*Notification, error) {
	var notification Notification
	var dataJSON, channelsJSON, tagsJSON, metadataJSON []byte
	var readAt, deliveredAt, expiresAt sql.NullTime

	err := rows.Scan(
		&notification.ID,
		&notification.Type,
		&notification.UserID,
		&notification.Title,
		&notification.Message,
		&dataJSON,
		&notification.Priority,
		&channelsJSON,
		&notification.CreatedAt,
		&expiresAt,
		&readAt,
		&deliveredAt,
		&notification.Source,
		&tagsJSON,
		&metadataJSON,
	)
	if err != nil {
		return nil, err
	}

	// Désérialiser les données JSON
	if len(dataJSON) > 0 {
		if err := json.Unmarshal(dataJSON, &notification.Data); err != nil {
			s.logger.Warn("Failed to unmarshal notification data", zap.Error(err))
		}
	}

	if len(channelsJSON) > 0 {
		if err := json.Unmarshal(channelsJSON, &notification.Channels); err != nil {
			s.logger.Warn("Failed to unmarshal notification channels", zap.Error(err))
		}
	}

	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &notification.Tags); err != nil {
			s.logger.Warn("Failed to unmarshal notification tags", zap.Error(err))
		}
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &notification.Metadata); err != nil {
			s.logger.Warn("Failed to unmarshal notification metadata", zap.Error(err))
		}
	}

	// Gérer les timestamps nullables
	if readAt.Valid {
		notification.ReadAt = &readAt.Time
	}
	if deliveredAt.Valid {
		notification.DeliveredAt = &deliveredAt.Time
	}
	if expiresAt.Valid {
		notification.ExpiresAt = &expiresAt.Time
	}

	return &notification, nil
}

// ============================================================================
// MÉTHODES ADDITIONNELLES POUR L'IN-APP NOTIFICATION CENTER
// ============================================================================

// GetNotificationsByType récupère les notifications par type
func (s *PostgreSQLStorage) GetNotificationsByType(ctx context.Context, userID string, notificationType NotificationType, limit, offset int) ([]*Notification, error) {
	query := `
		SELECT 
			id, type, user_id, title, message, data, priority, channels,
			created_at, expires_at, read_at, delivered_at, source, tags, metadata
		FROM notifications 
		WHERE user_id = $1 AND type = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := s.db.QueryContext(ctx, query, userID, string(notificationType), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications by type: %w", err)
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		notification, err := s.scanNotification(rows)
		if err != nil {
			s.logger.Error("Failed to scan notification", zap.Error(err))
			continue
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// MarkAllAsRead marque toutes les notifications d'un utilisateur comme lues
func (s *PostgreSQLStorage) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `
		UPDATE notifications 
		SET read_at = NOW() 
		WHERE user_id = $1 AND read_at IS NULL`

	result, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	s.logger.Info("All notifications marked as read",
		zap.String("user_id", userID),
		zap.Int64("count", rowsAffected))

	return nil
}

// GetNotificationStats récupère les statistiques de notifications
func (s *PostgreSQLStorage) GetNotificationStats(ctx context.Context, userID string) (*UserNotificationStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN read_at IS NULL THEN 1 END) as unread,
			COUNT(CASE WHEN read_at IS NOT NULL THEN 1 END) as read,
			COUNT(CASE WHEN priority = 'high' OR priority = 'critical' OR priority = 'emergency' THEN 1 END) as high_priority
		FROM notifications 
		WHERE user_id = $1 AND (expires_at IS NULL OR expires_at > NOW())`

	var stats UserNotificationStats
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.Total,
		&stats.Unread,
		&stats.Read,
		&stats.HighPriority,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification stats: %w", err)
	}

	// Statistiques par type
	typeQuery := `
		SELECT type, COUNT(*), COUNT(CASE WHEN read_at IS NULL THEN 1 END)
		FROM notifications 
		WHERE user_id = $1 AND (expires_at IS NULL OR expires_at > NOW())
		GROUP BY type`

	rows, err := s.db.QueryContext(ctx, typeQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get type stats: %w", err)
	}
	defer rows.Close()

	stats.ByType = make(map[NotificationType]*TypeNotificationStats)
	for rows.Next() {
		var notificationType string
		var total, unread int
		
		if err := rows.Scan(&notificationType, &total, &unread); err != nil {
			continue
		}

		stats.ByType[NotificationType(notificationType)] = &TypeNotificationStats{
			Total:  total,
			Unread: unread,
			Read:   total - unread,
		}
	}

	return &stats, nil
}

// UserNotificationStats statistiques des notifications d'un utilisateur
type UserNotificationStats struct {
	Total        int                                        `json:"total"`
	Unread       int                                        `json:"unread"`
	Read         int                                        `json:"read"`
	HighPriority int                                        `json:"high_priority"`
	ByType       map[NotificationType]*TypeNotificationStats `json:"by_type"`
}

// TypeNotificationStats statistiques par type de notification
type TypeNotificationStats struct {
	Total  int `json:"total"`
	Unread int `json:"unread"`
	Read   int `json:"read"`
}

// DeleteOldNotifications supprime les anciennes notifications
func (s *PostgreSQLStorage) DeleteOldNotifications(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	
	query := `DELETE FROM notifications WHERE created_at < $1`

	result, err := s.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to delete old notifications: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		s.logger.Info("Deleted old notifications",
			zap.Int64("count", rowsAffected),
			zap.Duration("older_than", olderThan))
	}

	return nil
}
