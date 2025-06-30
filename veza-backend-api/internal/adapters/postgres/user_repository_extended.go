package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/core/domain/entities"
	"github.com/okinrev/veza-web-app/internal/core/domain/repositories"
)

// ============================================================================
// PRÉFÉRENCES UTILISATEUR
// ============================================================================

// UpdatePreferences met à jour les préférences d'un utilisateur
func (r *userRepositoryComplete) UpdatePreferences(ctx context.Context, userID int64, preferences *repositories.UserPreferences) error {
	// Sérialiser les préférences en JSON
	notifJSON, err := json.Marshal(preferences.NotificationSettings)
	if err != nil {
		return fmt.Errorf("sérialisation notifications: %w", err)
	}

	privacyJSON, err := json.Marshal(preferences.PrivacySettings)
	if err != nil {
		return fmt.Errorf("sérialisation privacy: %w", err)
	}

	audioJSON, err := json.Marshal(preferences.AudioSettings)
	if err != nil {
		return fmt.Errorf("sérialisation audio: %w", err)
	}

	chatJSON, err := json.Marshal(preferences.ChatSettings)
	if err != nil {
		return fmt.Errorf("sérialisation chat: %w", err)
	}

	query := `
		INSERT INTO user_preferences (
			user_id, theme, language, timezone, 
			notification_settings, privacy_settings, 
			audio_settings, chat_settings, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id) DO UPDATE SET
			theme = EXCLUDED.theme,
			language = EXCLUDED.language,
			timezone = EXCLUDED.timezone,
			notification_settings = EXCLUDED.notification_settings,
			privacy_settings = EXCLUDED.privacy_settings,
			audio_settings = EXCLUDED.audio_settings,
			chat_settings = EXCLUDED.chat_settings,
			updated_at = EXCLUDED.updated_at
	`

	_, err = r.db.ExecContext(ctx, query,
		userID, preferences.Theme, preferences.Language, preferences.Timezone,
		notifJSON, privacyJSON, audioJSON, chatJSON, time.Now(),
	)

	if err != nil {
		return fmt.Errorf("mise à jour préférences: %w", err)
	}

	// Invalider le cache
	r.cache.Delete(ctx, fmt.Sprintf("user:preferences:%d", userID))
	return nil
}

// GetPreferences récupère les préférences d'un utilisateur
func (r *userRepositoryComplete) GetPreferences(ctx context.Context, userID int64) (*repositories.UserPreferences, error) {
	// Tentative cache
	cacheKey := fmt.Sprintf("user:preferences:%d", userID)
	var preferences repositories.UserPreferences
	if err := r.cache.Get(ctx, cacheKey, &preferences); err == nil {
		return &preferences, nil
	}

	query := `
		SELECT user_id, theme, language, timezone,
		       notification_settings, privacy_settings,
		       audio_settings, chat_settings, updated_at
		FROM user_preferences
		WHERE user_id = $1
	`

	var notifJSON, privacyJSON, audioJSON, chatJSON []byte
	pref := &repositories.UserPreferences{}

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&pref.UserID, &pref.Theme, &pref.Language, &pref.Timezone,
		&notifJSON, &privacyJSON, &audioJSON, &chatJSON, &pref.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Créer des préférences par défaut
			return r.createDefaultPreferences(ctx, userID)
		}
		return nil, fmt.Errorf("récupération préférences: %w", err)
	}

	// Désérialiser les JSON
	if err := json.Unmarshal(notifJSON, &pref.NotificationSettings); err != nil {
		return nil, fmt.Errorf("désérialisation notifications: %w", err)
	}
	if err := json.Unmarshal(privacyJSON, &pref.PrivacySettings); err != nil {
		return nil, fmt.Errorf("désérialisation privacy: %w", err)
	}
	if err := json.Unmarshal(audioJSON, &pref.AudioSettings); err != nil {
		return nil, fmt.Errorf("désérialisation audio: %w", err)
	}
	if err := json.Unmarshal(chatJSON, &pref.ChatSettings); err != nil {
		return nil, fmt.Errorf("désérialisation chat: %w", err)
	}

	// Cache du résultat
	r.cache.Set(ctx, cacheKey, pref, 30*time.Minute)
	return pref, nil
}

// createDefaultPreferences crée des préférences par défaut
func (r *userRepositoryComplete) createDefaultPreferences(ctx context.Context, userID int64) (*repositories.UserPreferences, error) {
	preferences := &repositories.UserPreferences{
		UserID:   userID,
		Theme:    "light",
		Language: "en",
		Timezone: "UTC",
		NotificationSettings: repositories.NotificationSettings{
			EmailNotifications:   true,
			PushNotifications:    true,
			SoundNotifications:   true,
			DesktopNotifications: true,
			NewMessages:          true,
			Mentions:             true,
			StreamStarted:        true,
			StreamInvites:        true,
			RoomInvites:          true,
			FriendRequests:       true,
			SystemAnnouncements:  true,
			QuietHoursEnabled:    false,
			QuietHoursStart:      "22:00",
			QuietHoursEnd:        "08:00",
		},
		PrivacySettings: repositories.PrivacySettings{
			ProfileVisibility:   "public",
			OnlineStatusVisible: true,
			LastSeenVisible:     true,
			AllowDirectMessages: "everyone",
			AllowRoomInvites:    "everyone",
			AllowStreamInvites:  "everyone",
			ShowInSearchResults: true,
			ShowEmail:           false,
			ShowRealName:        false,
		},
		AudioSettings: repositories.AudioSettings{
			DefaultQuality:     entities.StreamQualityHigh,
			AutoAdjustQuality:  true,
			Volume:             0.8,
			Muted:              false,
			EnableEqualizer:    false,
			EqualizerPreset:    "flat",
			EnableSpatialAudio: false,
			BufferSize:         5,
		},
		ChatSettings: repositories.ChatSettings{
			ShowTimestamps:       true,
			Show24HourTime:       false,
			EnableEmojis:         true,
			EnableAnimatedEmojis: true,
			FontSize:             14,
			MessageGrouping:      true,
			CompactMode:          false,
			FilterWords:          []string{},
			AllowMentions:        true,
			AllowDirectMentions:  true,
		},
		UpdatedAt: time.Now(),
	}

	// Sauvegarder
	if err := r.UpdatePreferences(ctx, userID, preferences); err != nil {
		return nil, err
	}

	return preferences, nil
}

// ============================================================================
// GESTION DES SESSIONS
// ============================================================================

// CreateSession crée une nouvelle session utilisateur
func (r *userRepositoryComplete) CreateSession(ctx context.Context, session *repositories.UserSession) error {
	query := `
		INSERT INTO user_sessions (
			user_id, session_token, refresh_token, device_info,
			ip_address, user_agent, location, is_active,
			last_activity, expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		session.UserID, session.SessionToken, session.RefreshToken,
		session.DeviceInfo, session.IPAddress, session.UserAgent,
		session.Location, session.IsActive, session.LastActivity,
		session.ExpiresAt, session.CreatedAt,
	).Scan(&session.ID)

	if err != nil {
		return fmt.Errorf("création session: %w", err)
	}

	return nil
}

// GetSession récupère une session par token
func (r *userRepositoryComplete) GetSession(ctx context.Context, sessionToken string) (*repositories.UserSession, error) {
	query := `
		SELECT s.id, s.user_id, s.session_token, s.refresh_token,
		       s.device_info, s.ip_address, s.user_agent, s.location,
		       s.is_active, s.last_activity, s.expires_at, s.created_at,
		       u.username, u.email, u.role, u.status
		FROM user_sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.session_token = $1 AND s.is_active = true AND s.expires_at > NOW()
	`

	session := &repositories.UserSession{}
	user := &entities.User{}

	err := r.db.QueryRowContext(ctx, query, sessionToken).Scan(
		&session.ID, &session.UserID, &session.SessionToken, &session.RefreshToken,
		&session.DeviceInfo, &session.IPAddress, &session.UserAgent, &session.Location,
		&session.IsActive, &session.LastActivity, &session.ExpiresAt, &session.CreatedAt,
		&user.Username, &user.Email, &user.Role, &user.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("récupération session: %w", err)
	}

	user.ID = session.UserID
	session.User = user

	return session, nil
}

// UpdateSession met à jour l'activité d'une session
func (r *userRepositoryComplete) UpdateSession(ctx context.Context, sessionToken string, lastActivity time.Time) error {
	query := `
		UPDATE user_sessions SET 
			last_activity = $2
		WHERE session_token = $1 AND is_active = true
	`

	result, err := r.db.ExecContext(ctx, query, sessionToken, lastActivity)
	if err != nil {
		return fmt.Errorf("mise à jour session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour session: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session non trouvée")
	}

	return nil
}

// InvalidateSession invalide une session
func (r *userRepositoryComplete) InvalidateSession(ctx context.Context, sessionToken string) error {
	query := `
		UPDATE user_sessions SET 
			is_active = false
		WHERE session_token = $1
	`

	result, err := r.db.ExecContext(ctx, query, sessionToken)
	if err != nil {
		return fmt.Errorf("invalidation session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification invalidation session: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session non trouvée")
	}

	return nil
}

// InvalidateAllUserSessions invalide toutes les sessions d'un utilisateur
func (r *userRepositoryComplete) InvalidateAllUserSessions(ctx context.Context, userID int64) error {
	query := `
		UPDATE user_sessions SET 
			is_active = false
		WHERE user_id = $1 AND is_active = true
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("invalidation sessions utilisateur: %w", err)
	}

	return nil
}

// GetUserSessions récupère toutes les sessions d'un utilisateur
func (r *userRepositoryComplete) GetUserSessions(ctx context.Context, userID int64) ([]*repositories.UserSession, error) {
	query := `
		SELECT id, user_id, session_token, device_info, ip_address,
		       user_agent, location, is_active, last_activity, 
		       expires_at, created_at
		FROM user_sessions
		WHERE user_id = $1
		ORDER BY last_activity DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("récupération sessions utilisateur: %w", err)
	}
	defer rows.Close()

	var sessions []*repositories.UserSession
	for rows.Next() {
		session := &repositories.UserSession{}
		err := rows.Scan(
			&session.ID, &session.UserID, &session.SessionToken,
			&session.DeviceInfo, &session.IPAddress, &session.UserAgent,
			&session.Location, &session.IsActive, &session.LastActivity,
			&session.ExpiresAt, &session.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// ============================================================================
// AUDIT ET LOGS
// ============================================================================

// CreateAuditLog crée un log d'audit
func (r *userRepositoryComplete) CreateAuditLog(ctx context.Context, log *repositories.UserAuditLog) error {
	query := `
		INSERT INTO user_audit_logs (
			user_id, action, resource, resource_id, details,
			ip_address, user_agent, success, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		log.UserID, log.Action, log.Resource, log.ResourceID,
		log.Details, log.IPAddress, log.UserAgent, log.Success,
		log.ErrorMessage, log.CreatedAt,
	).Scan(&log.ID)

	if err != nil {
		r.logger.Error("Erreur création audit log", zap.Error(err))
		// Ne pas retourner d'erreur pour ne pas bloquer l'opération principale
		return nil
	}

	return nil
}

// GetUserAuditLogs récupère les logs d'audit d'un utilisateur
func (r *userRepositoryComplete) GetUserAuditLogs(ctx context.Context, userID int64, limit, offset int) ([]*repositories.UserAuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, user_id, action, resource, resource_id, details,
		       ip_address, user_agent, success, error_message, created_at
		FROM user_audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("récupération audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*repositories.UserAuditLog
	for rows.Next() {
		log := &repositories.UserAuditLog{}
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Action, &log.Resource,
			&log.ResourceID, &log.Details, &log.IPAddress,
			&log.UserAgent, &log.Success, &log.ErrorMessage,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// ============================================================================
// RELATIONS ET CONTACTS
// ============================================================================

// AddContact ajoute un contact
func (r *userRepositoryComplete) AddContact(ctx context.Context, userID, contactID int64) error {
	query := `
		INSERT INTO user_contacts (user_id, contact_id, status, created_at, updated_at)
		VALUES ($1, $2, 'pending', $3, $3)
		ON CONFLICT (user_id, contact_id) DO NOTHING
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, contactID, now)
	if err != nil {
		return fmt.Errorf("ajout contact: %w", err)
	}

	return nil
}

// RemoveContact supprime un contact
func (r *userRepositoryComplete) RemoveContact(ctx context.Context, userID, contactID int64) error {
	query := `DELETE FROM user_contacts WHERE user_id = $1 AND contact_id = $2`

	result, err := r.db.ExecContext(ctx, query, userID, contactID)
	if err != nil {
		return fmt.Errorf("suppression contact: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification suppression contact: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("contact non trouvé")
	}

	return nil
}

// GetContacts récupère les contacts d'un utilisateur
func (r *userRepositoryComplete) GetContacts(ctx context.Context, userID int64) ([]*entities.User, error) {
	query := `
		SELECT u.id, u.uuid, u.username, u.email, u.first_name, u.last_name,
		       u.display_name, u.avatar, u.role, u.status, u.is_online,
		       u.last_seen, u.created_at
		FROM users u
		JOIN user_contacts c ON u.id = c.contact_id
		WHERE c.user_id = $1 AND c.status = 'accepted' AND u.deleted_at IS NULL
		ORDER BY u.display_name, u.username
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("récupération contacts: %w", err)
	}
	defer rows.Close()

	var contacts []*entities.User
	for rows.Next() {
		user := &entities.User{}
		err := rows.Scan(
			&user.ID, &user.UUID, &user.Username, &user.Email,
			&user.FirstName, &user.LastName, &user.DisplayName,
			&user.Avatar, &user.Role, &user.Status, &user.IsOnline,
			&user.LastSeen, &user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan contact: %w", err)
		}
		contacts = append(contacts, user)
	}

	return contacts, nil
}

// BlockUser bloque un utilisateur
func (r *userRepositoryComplete) BlockUser(ctx context.Context, userID, blockedUserID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("début transaction: %w", err)
	}
	defer tx.Rollback()

	// Ajouter le blocage
	query1 := `
		INSERT INTO user_blocks (user_id, blocked_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, blocked_id) DO NOTHING
	`
	_, err = tx.ExecContext(ctx, query1, userID, blockedUserID, time.Now())
	if err != nil {
		return fmt.Errorf("blocage utilisateur: %w", err)
	}

	// Supprimer le contact s'il existe
	query2 := `DELETE FROM user_contacts WHERE user_id = $1 AND contact_id = $2`
	_, err = tx.ExecContext(ctx, query2, userID, blockedUserID)
	if err != nil {
		return fmt.Errorf("suppression contact lors du blocage: %w", err)
	}

	return tx.Commit()
}

// UnblockUser débloque un utilisateur
func (r *userRepositoryComplete) UnblockUser(ctx context.Context, userID, blockedUserID int64) error {
	query := `DELETE FROM user_blocks WHERE user_id = $1 AND blocked_id = $2`

	result, err := r.db.ExecContext(ctx, query, userID, blockedUserID)
	if err != nil {
		return fmt.Errorf("déblocage utilisateur: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification déblocage: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("blocage non trouvé")
	}

	return nil
}

// GetBlockedUsers récupère les utilisateurs bloqués
func (r *userRepositoryComplete) GetBlockedUsers(ctx context.Context, userID int64) ([]*entities.User, error) {
	query := `
		SELECT u.id, u.uuid, u.username, u.display_name, u.avatar,
		       u.role, u.status, b.created_at as blocked_at
		FROM users u
		JOIN user_blocks b ON u.id = b.blocked_id
		WHERE b.user_id = $1 AND u.deleted_at IS NULL
		ORDER BY b.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("récupération utilisateurs bloqués: %w", err)
	}
	defer rows.Close()

	var blockedUsers []*entities.User
	for rows.Next() {
		user := &entities.User{}
		var blockedAt time.Time
		err := rows.Scan(
			&user.ID, &user.UUID, &user.Username, &user.DisplayName,
			&user.Avatar, &user.Role, &user.Status, &blockedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan utilisateur bloqué: %w", err)
		}
		blockedUsers = append(blockedUsers, user)
	}

	return blockedUsers, nil
}

// IsBlocked vérifie si un utilisateur est bloqué
func (r *userRepositoryComplete) IsBlocked(ctx context.Context, userID, otherUserID int64) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_blocks 
			WHERE (user_id = $1 AND blocked_id = $2) OR (user_id = $2 AND blocked_id = $1)
		)
	`

	err := r.db.QueryRowContext(ctx, query, userID, otherUserID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("vérification blocage: %w", err)
	}

	return exists, nil
}
