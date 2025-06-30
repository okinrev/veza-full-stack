package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/core/domain/entities"
	"github.com/okinrev/veza-web-app/internal/core/domain/repositories"
)

// chatRepository implémentation PostgreSQL du ChatRepository
type chatRepository struct {
	db     *sql.DB
	cache  CacheService
	logger *zap.Logger
}

// NewChatRepository crée une nouvelle instance du repository chat
func NewChatRepository(db *sql.DB, cache CacheService, logger *zap.Logger) (repositories.ChatRepository, error) {
	return &chatRepository{
		db:     db,
		cache:  cache,
		logger: logger,
	}, nil
}

// ============================================================================
// GESTION DES ROOMS
// ============================================================================

// CreateRoom crée une nouvelle room de chat
func (r *chatRepository) CreateRoom(ctx context.Context, room *entities.ChatRoom) error {
	query := `
		INSERT INTO chat_rooms (
			uuid, name, slug, description, topic, type, privacy, status,
			creator_id, max_members, password_hash, tags, settings,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		room.UUID, room.Name, room.Slug, room.Description, room.Topic,
		string(room.Type), string(room.Privacy), string(room.Status),
		room.CreatorID, room.MaxMembers, room.PasswordHash,
		pq.Array(room.Tags), room.Settings, room.CreatedAt, room.UpdatedAt,
	).Scan(&room.ID, &room.CreatedAt, &room.UpdatedAt)

	if err != nil {
		r.logger.Error("Erreur création room", zap.Error(err))
		return fmt.Errorf("création room: %w", err)
	}

	// Ajouter le créateur comme admin de la room
	if err := r.JoinRoom(ctx, room.ID, room.CreatorID); err != nil {
		return fmt.Errorf("ajout créateur à la room: %w", err)
	}

	if err := r.UpdateUserRoomRole(ctx, room.ID, room.CreatorID, entities.RoomRoleAdmin); err != nil {
		return fmt.Errorf("promotion créateur admin: %w", err)
	}

	// Invalider les caches
	r.invalidateRoomCaches(ctx, room.ID, room.Slug)
	return nil
}

// GetRoomByID récupère une room par ID
func (r *chatRepository) GetRoomByID(ctx context.Context, roomID int64) (*entities.ChatRoom, error) {
	// Tentative cache
	cacheKey := fmt.Sprintf("room:id:%d", roomID)
	var room entities.ChatRoom
	if err := r.cache.Get(ctx, cacheKey, &room); err == nil {
		return &room, nil
	}

	query := `
		SELECT r.id, r.uuid, r.name, r.slug, r.description, r.topic,
		       r.type, r.privacy, r.status, r.creator_id, r.max_members,
		       r.password_hash, r.tags, r.settings, r.created_at, r.updated_at,
		       u.username as creator_username, u.display_name as creator_display_name,
		       COUNT(rm.user_id) as member_count
		FROM chat_rooms r
		JOIN users u ON r.creator_id = u.id
		LEFT JOIN room_members rm ON r.id = rm.room_id AND rm.status = 'active'
		WHERE r.id = $1 AND r.deleted_at IS NULL
		GROUP BY r.id, u.username, u.display_name
	`

	roomPtr := &entities.ChatRoom{}
	var creatorUsername, creatorDisplayName string
	var tags pq.StringArray

	err := r.db.QueryRowContext(ctx, query, roomID).Scan(
		&roomPtr.ID, &roomPtr.UUID, &roomPtr.Name, &roomPtr.Slug,
		&roomPtr.Description, &roomPtr.Topic, &roomPtr.Type, &roomPtr.Privacy,
		&roomPtr.Status, &roomPtr.CreatorID, &roomPtr.MaxMembers,
		&roomPtr.PasswordHash, &tags, &roomPtr.Settings,
		&roomPtr.CreatedAt, &roomPtr.UpdatedAt,
		&creatorUsername, &creatorDisplayName, &roomPtr.MemberCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("récupération room par ID: %w", err)
	}

	roomPtr.Tags = []string(tags)
	roomPtr.Creator = &entities.User{
		ID:          roomPtr.CreatorID,
		Username:    creatorUsername,
		DisplayName: creatorDisplayName,
	}

	// Cache du résultat
	r.cache.Set(ctx, cacheKey, roomPtr, 15*time.Minute)
	return roomPtr, nil
}

// GetRoomBySlug récupère une room par son slug
func (r *chatRepository) GetRoomBySlug(ctx context.Context, slug string) (*entities.ChatRoom, error) {
	cacheKey := fmt.Sprintf("room:slug:%s", slug)
	var room entities.ChatRoom
	if err := r.cache.Get(ctx, cacheKey, &room); err == nil {
		return &room, nil
	}

	query := `
		SELECT r.id, r.uuid, r.name, r.slug, r.description, r.topic,
		       r.type, r.privacy, r.status, r.creator_id, r.max_members,
		       r.password_hash, r.tags, r.settings, r.created_at, r.updated_at,
		       u.username as creator_username, u.display_name as creator_display_name,
		       COUNT(rm.user_id) as member_count
		FROM chat_rooms r
		JOIN users u ON r.creator_id = u.id
		LEFT JOIN room_members rm ON r.id = rm.room_id AND rm.status = 'active'
		WHERE r.slug = $1 AND r.deleted_at IS NULL
		GROUP BY r.id, u.username, u.display_name
	`

	roomPtr := &entities.ChatRoom{}
	var creatorUsername, creatorDisplayName string
	var tags pq.StringArray

	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&roomPtr.ID, &roomPtr.UUID, &roomPtr.Name, &roomPtr.Slug,
		&roomPtr.Description, &roomPtr.Topic, &roomPtr.Type, &roomPtr.Privacy,
		&roomPtr.Status, &roomPtr.CreatorID, &roomPtr.MaxMembers,
		&roomPtr.PasswordHash, &tags, &roomPtr.Settings,
		&roomPtr.CreatedAt, &roomPtr.UpdatedAt,
		&creatorUsername, &creatorDisplayName, &roomPtr.MemberCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("récupération room par slug: %w", err)
	}

	roomPtr.Tags = []string(tags)
	roomPtr.Creator = &entities.User{
		ID:          roomPtr.CreatorID,
		Username:    creatorUsername,
		DisplayName: creatorDisplayName,
	}

	r.cache.Set(ctx, cacheKey, roomPtr, 15*time.Minute)
	return roomPtr, nil
}

// UpdateRoom met à jour une room
func (r *chatRepository) UpdateRoom(ctx context.Context, room *entities.ChatRoom) error {
	query := `
		UPDATE chat_rooms SET
			name = $2, slug = $3, description = $4, topic = $5,
			type = $6, privacy = $7, status = $8, max_members = $9,
			password_hash = $10, tags = $11, settings = $12, updated_at = $13
		WHERE id = $1 AND deleted_at IS NULL
	`

	room.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		room.ID, room.Name, room.Slug, room.Description, room.Topic,
		string(room.Type), string(room.Privacy), string(room.Status),
		room.MaxMembers, room.PasswordHash, pq.Array(room.Tags),
		room.Settings, room.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("mise à jour room: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour room: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("room non trouvée")
	}

	r.invalidateRoomCaches(ctx, room.ID, room.Slug)
	return nil
}

// DeleteRoom supprime une room (soft delete)
func (r *chatRepository) DeleteRoom(ctx context.Context, roomID int64) error {
	room, err := r.GetRoomByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return fmt.Errorf("room non trouvée")
	}

	now := time.Now()
	query := `
		UPDATE chat_rooms SET 
			deleted_at = $2, updated_at = $2, status = 'deleted'
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, roomID, now)
	if err != nil {
		return fmt.Errorf("suppression room: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification suppression room: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("room non trouvée ou déjà supprimée")
	}

	r.invalidateRoomCaches(ctx, roomID, room.Slug)
	return nil
}

// ListRooms liste les rooms avec pagination et filtres
func (r *chatRepository) ListRooms(ctx context.Context, filters repositories.RoomFilters) ([]*entities.ChatRoom, error) {
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	// Construction de la requête dynamique
	where := []string{"r.deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if filters.Type != "" {
		where = append(where, fmt.Sprintf("r.type = $%d", argIndex))
		args = append(args, string(filters.Type))
		argIndex++
	}

	if filters.Privacy != "" {
		where = append(where, fmt.Sprintf("r.privacy = $%d", argIndex))
		args = append(args, string(filters.Privacy))
		argIndex++
	}

	if filters.Status != "" {
		where = append(where, fmt.Sprintf("r.status = $%d", argIndex))
		args = append(args, string(filters.Status))
		argIndex++
	}

	if filters.CreatorID > 0 {
		where = append(where, fmt.Sprintf("r.creator_id = $%d", argIndex))
		args = append(args, filters.CreatorID)
		argIndex++
	}

	if len(filters.Tags) > 0 {
		where = append(where, fmt.Sprintf("r.tags && $%d", argIndex))
		args = append(args, pq.Array(filters.Tags))
		argIndex++
	}

	if filters.MinMembers > 0 {
		where = append(where, fmt.Sprintf("member_count >= $%d", argIndex))
		args = append(args, filters.MinMembers)
		argIndex++
	}

	if filters.MaxMembers > 0 {
		where = append(where, fmt.Sprintf("member_count <= $%d", argIndex))
		args = append(args, filters.MaxMembers)
		argIndex++
	}

	// Tri
	orderBy := "r.created_at DESC"
	if filters.SortBy != "" {
		switch filters.SortBy {
		case "created_at":
			orderBy = "r.created_at"
		case "member_count":
			orderBy = "member_count"
		case "activity":
			orderBy = "r.updated_at"
		case "name":
			orderBy = "r.name"
		}

		if filters.SortOrder == "asc" {
			orderBy += " ASC"
		} else {
			orderBy += " DESC"
		}
	}

	query := fmt.Sprintf(`
		SELECT r.id, r.uuid, r.name, r.slug, r.description, r.topic,
		       r.type, r.privacy, r.status, r.creator_id, r.max_members,
		       r.tags, r.created_at, r.updated_at,
		       u.username as creator_username, u.display_name as creator_display_name,
		       COUNT(rm.user_id) as member_count
		FROM chat_rooms r
		JOIN users u ON r.creator_id = u.id
		LEFT JOIN room_members rm ON r.id = rm.room_id AND rm.status = 'active'
		WHERE %s
		GROUP BY r.id, u.username, u.display_name
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, strings.Join(where, " AND "), orderBy, argIndex, argIndex+1)

	args = append(args, filters.Limit, filters.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("liste rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*entities.ChatRoom
	for rows.Next() {
		room := &entities.ChatRoom{}
		var creatorUsername, creatorDisplayName string
		var tags pq.StringArray

		err := rows.Scan(
			&room.ID, &room.UUID, &room.Name, &room.Slug,
			&room.Description, &room.Topic, &room.Type, &room.Privacy,
			&room.Status, &room.CreatorID, &room.MaxMembers,
			&tags, &room.CreatedAt, &room.UpdatedAt,
			&creatorUsername, &creatorDisplayName, &room.MemberCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}

		room.Tags = []string(tags)
		room.Creator = &entities.User{
			ID:          room.CreatorID,
			Username:    creatorUsername,
			DisplayName: creatorDisplayName,
		}

		rooms = append(rooms, room)
	}

	return rooms, nil
}

// SearchRooms recherche des rooms par nom/description
func (r *chatRepository) SearchRooms(ctx context.Context, query string, limit, offset int) ([]*entities.ChatRoom, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	searchQuery := `
		SELECT r.id, r.uuid, r.name, r.slug, r.description, r.topic,
		       r.type, r.privacy, r.status, r.creator_id, r.max_members,
		       r.tags, r.created_at, r.updated_at,
		       u.username as creator_username, u.display_name as creator_display_name,
		       COUNT(rm.user_id) as member_count,
		       ts_rank(to_tsvector('english', r.name || ' ' || COALESCE(r.description, '')), plainto_tsquery('english', $1)) as rank
		FROM chat_rooms r
		JOIN users u ON r.creator_id = u.id
		LEFT JOIN room_members rm ON r.id = rm.room_id AND rm.status = 'active'
		WHERE r.deleted_at IS NULL 
		  AND r.privacy = 'public'
		  AND (
		    r.name ILIKE '%' || $1 || '%' OR
		    r.description ILIKE '%' || $1 || '%' OR
		    to_tsvector('english', r.name || ' ' || COALESCE(r.description, '')) @@ plainto_tsquery('english', $1)
		  )
		GROUP BY r.id, u.username, u.display_name
		ORDER BY rank DESC, member_count DESC, r.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, searchQuery, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("recherche rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*entities.ChatRoom
	for rows.Next() {
		room := &entities.ChatRoom{}
		var creatorUsername, creatorDisplayName string
		var tags pq.StringArray
		var rank float64

		err := rows.Scan(
			&room.ID, &room.UUID, &room.Name, &room.Slug,
			&room.Description, &room.Topic, &room.Type, &room.Privacy,
			&room.Status, &room.CreatorID, &room.MaxMembers,
			&tags, &room.CreatedAt, &room.UpdatedAt,
			&creatorUsername, &creatorDisplayName, &room.MemberCount, &rank,
		)
		if err != nil {
			return nil, fmt.Errorf("scan room recherche: %w", err)
		}

		room.Tags = []string(tags)
		room.Creator = &entities.User{
			ID:          room.CreatorID,
			Username:    creatorUsername,
			DisplayName: creatorDisplayName,
		}

		rooms = append(rooms, room)
	}

	return rooms, nil
}

// GetUserRooms récupère les rooms d'un utilisateur
func (r *chatRepository) GetUserRooms(ctx context.Context, userID int64, limit, offset int) ([]*entities.ChatRoom, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT r.id, r.uuid, r.name, r.slug, r.description, r.topic,
		       r.type, r.privacy, r.status, r.creator_id, r.max_members,
		       r.tags, r.created_at, r.updated_at,
		       u.username as creator_username, u.display_name as creator_display_name,
		       COUNT(rm2.user_id) as member_count,
		       rm.role as user_role, rm.joined_at
		FROM chat_rooms r
		JOIN users u ON r.creator_id = u.id
		JOIN room_members rm ON r.id = rm.room_id AND rm.user_id = $1 AND rm.status = 'active'
		LEFT JOIN room_members rm2 ON r.id = rm2.room_id AND rm2.status = 'active'
		WHERE r.deleted_at IS NULL
		GROUP BY r.id, u.username, u.display_name, rm.role, rm.joined_at
		ORDER BY rm.joined_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("rooms utilisateur: %w", err)
	}
	defer rows.Close()

	var rooms []*entities.ChatRoom
	for rows.Next() {
		room := &entities.ChatRoom{}
		var creatorUsername, creatorDisplayName string
		var tags pq.StringArray
		var userRole entities.RoomRole
		var joinedAt time.Time

		err := rows.Scan(
			&room.ID, &room.UUID, &room.Name, &room.Slug,
			&room.Description, &room.Topic, &room.Type, &room.Privacy,
			&room.Status, &room.CreatorID, &room.MaxMembers,
			&tags, &room.CreatedAt, &room.UpdatedAt,
			&creatorUsername, &creatorDisplayName, &room.MemberCount,
			&userRole, &joinedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan room utilisateur: %w", err)
		}

		room.Tags = []string(tags)
		room.Creator = &entities.User{
			ID:          room.CreatorID,
			Username:    creatorUsername,
			DisplayName: creatorDisplayName,
		}
		room.UserRole = &userRole

		rooms = append(rooms, room)
	}

	return rooms, nil
}

// GetPublicRooms récupère les rooms publiques
func (r *chatRepository) GetPublicRooms(ctx context.Context, limit, offset int) ([]*entities.ChatRoom, error) {
	filters := repositories.RoomFilters{
		Privacy:   entities.RoomPrivacyPublic,
		Status:    entities.RoomStatusActive,
		Limit:     limit,
		Offset:    offset,
		SortBy:    "member_count",
		SortOrder: "desc",
	}

	return r.ListRooms(ctx, filters)
}

// GetRoomStats récupère les statistiques d'une room
func (r *chatRepository) GetRoomStats(ctx context.Context, roomID int64) (*repositories.RoomStats, error) {
	query := `
		SELECT 
			r.id as room_id,
			COUNT(DISTINCT rm.user_id) as member_count,
			COUNT(DISTINCT m.id) as message_count,
			COUNT(DISTINCT CASE WHEN u.is_online THEN rm.user_id END) as online_count,
			COUNT(DISTINCT CASE WHEN m.created_at >= NOW() - INTERVAL '1 day' THEN m.id END) as today_messages,
			COUNT(DISTINCT CASE WHEN m.created_at >= NOW() - INTERVAL '7 days' THEN m.id END) as week_messages,
			COUNT(DISTINCT CASE WHEN m.created_at >= NOW() - INTERVAL '30 days' THEN m.id END) as month_messages,
			MAX(m.created_at) as last_message,
			r.created_at,
			r.updated_at
		FROM chat_rooms r
		LEFT JOIN room_members rm ON r.id = rm.room_id AND rm.status = 'active'
		LEFT JOIN users u ON rm.user_id = u.id
		LEFT JOIN chat_messages m ON r.id = m.room_id AND m.deleted_at IS NULL
		WHERE r.id = $1 AND r.deleted_at IS NULL
		GROUP BY r.id, r.created_at, r.updated_at
	`

	stats := &repositories.RoomStats{}
	err := r.db.QueryRowContext(ctx, query, roomID).Scan(
		&stats.RoomID, &stats.MemberCount, &stats.MessageCount,
		&stats.OnlineCount, &stats.TodayMessages, &stats.WeekMessages,
		&stats.MonthMessages, &stats.LastMessage, &stats.CreatedAt,
		&stats.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("statistiques room: %w", err)
	}

	return stats, nil
}

// ============================================================================
// GESTION DES MEMBRES
// ============================================================================

// JoinRoom ajoute un utilisateur à une room
func (r *chatRepository) JoinRoom(ctx context.Context, roomID, userID int64) error {
	// Vérifier si l'utilisateur n'est pas déjà membre
	if exists, err := r.IsUserInRoom(ctx, roomID, userID); err != nil {
		return err
	} else if exists {
		return fmt.Errorf("utilisateur déjà membre de la room")
	}

	// Vérifier si l'utilisateur n'est pas banni
	if banned, err := r.IsUserBanned(ctx, roomID, userID); err != nil {
		return err
	} else if banned {
		return fmt.Errorf("utilisateur banni de cette room")
	}

	query := `
		INSERT INTO room_members (room_id, user_id, role, status, joined_at)
		VALUES ($1, $2, 'member', 'active', $3)
	`

	_, err := r.db.ExecContext(ctx, query, roomID, userID, time.Now())
	if err != nil {
		return fmt.Errorf("ajout membre room: %w", err)
	}

	// Invalider les caches
	r.cache.Delete(ctx, fmt.Sprintf("room:members:%d", roomID))
	return nil
}

// LeaveRoom retire un utilisateur d'une room
func (r *chatRepository) LeaveRoom(ctx context.Context, roomID, userID int64) error {
	query := `
		UPDATE room_members SET 
			status = 'left', left_at = $3
		WHERE room_id = $1 AND user_id = $2 AND status = 'active'
	`

	result, err := r.db.ExecContext(ctx, query, roomID, userID, time.Now())
	if err != nil {
		return fmt.Errorf("quitter room: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification quitter room: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non membre de la room")
	}

	r.cache.Delete(ctx, fmt.Sprintf("room:members:%d", roomID))
	return nil
}

// KickUser expulse un utilisateur d'une room
func (r *chatRepository) KickUser(ctx context.Context, roomID, userID, moderatorID int64, reason string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("début transaction: %w", err)
	}
	defer tx.Rollback()

	// Expulser l'utilisateur
	query1 := `
		UPDATE room_members SET 
			status = 'kicked', left_at = $4
		WHERE room_id = $1 AND user_id = $2 AND status = 'active'
	`
	_, err = tx.ExecContext(ctx, query1, roomID, userID, moderatorID, time.Now())
	if err != nil {
		return fmt.Errorf("expulsion utilisateur: %w", err)
	}

	// Log de modération
	log := &repositories.ModerationLog{
		RoomID:       roomID,
		ModeratorID:  moderatorID,
		TargetUserID: &userID,
		Action:       "kick",
		Reason:       reason,
		CreatedAt:    time.Now(),
	}

	query2 := `
		INSERT INTO moderation_logs (room_id, moderator_id, target_user_id, action, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = tx.ExecContext(ctx, query2, log.RoomID, log.ModeratorID, log.TargetUserID, log.Action, log.Reason, log.CreatedAt)
	if err != nil {
		return fmt.Errorf("log modération: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	r.cache.Delete(ctx, fmt.Sprintf("room:members:%d", roomID))
	return nil
}

// BanUser bannit un utilisateur d'une room
func (r *chatRepository) BanUser(ctx context.Context, roomID, userID, moderatorID int64, reason string, duration *time.Duration) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("début transaction: %w", err)
	}
	defer tx.Rollback()

	// Retirer l'utilisateur de la room
	query1 := `
		UPDATE room_members SET 
			status = 'banned', left_at = $4
		WHERE room_id = $1 AND user_id = $2 AND status = 'active'
	`
	_, err = tx.ExecContext(ctx, query1, roomID, userID, moderatorID, time.Now())
	if err != nil {
		return fmt.Errorf("bannissement utilisateur: %w", err)
	}

	// Créer l'entrée de ban
	var expiresAt *time.Time
	if duration != nil {
		expiry := time.Now().Add(*duration)
		expiresAt = &expiry
	}

	query2 := `
		INSERT INTO room_bans (room_id, user_id, moderator_id, reason, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (room_id, user_id) DO UPDATE SET
			moderator_id = EXCLUDED.moderator_id,
			reason = EXCLUDED.reason,
			expires_at = EXCLUDED.expires_at,
			created_at = EXCLUDED.created_at
	`
	_, err = tx.ExecContext(ctx, query2, roomID, userID, moderatorID, reason, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("création ban: %w", err)
	}

	// Log de modération
	var durationSeconds *int64
	if duration != nil {
		seconds := int64(duration.Seconds())
		durationSeconds = &seconds
	}

	query3 := `
		INSERT INTO moderation_logs (room_id, moderator_id, target_user_id, action, reason, duration, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = tx.ExecContext(ctx, query3, roomID, moderatorID, userID, "ban", reason, durationSeconds, time.Now())
	if err != nil {
		return fmt.Errorf("log modération ban: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction ban: %w", err)
	}

	r.cache.Delete(ctx, fmt.Sprintf("room:members:%d", roomID))
	r.cache.Delete(ctx, fmt.Sprintf("user:banned:%d:%d", roomID, userID))
	return nil
}

// UnbanUser débannit un utilisateur
func (r *chatRepository) UnbanUser(ctx context.Context, roomID, userID, moderatorID int64) error {
	query := `DELETE FROM room_bans WHERE room_id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, roomID, userID)
	if err != nil {
		return fmt.Errorf("débannissement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification débannissement: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("ban non trouvé")
	}

	// Log de modération
	query2 := `
		INSERT INTO moderation_logs (room_id, moderator_id, target_user_id, action, reason, created_at)
		VALUES ($1, $2, $3, 'unban', 'Débannissement', $4)
	`
	_, err = r.db.ExecContext(ctx, query2, roomID, moderatorID, userID, time.Now())
	if err != nil {
		r.logger.Error("Erreur log débannissement", zap.Error(err))
	}

	r.cache.Delete(ctx, fmt.Sprintf("user:banned:%d:%d", roomID, userID))
	return nil
}

// GetRoomMembers récupère les membres d'une room
func (r *chatRepository) GetRoomMembers(ctx context.Context, roomID int64, limit, offset int) ([]*repositories.RoomMember, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT rm.id, rm.room_id, rm.user_id, rm.role, rm.joined_at, rm.last_seen,
		       u.username, u.display_name, u.avatar, u.is_online
		FROM room_members rm
		JOIN users u ON rm.user_id = u.id
		WHERE rm.room_id = $1 AND rm.status = 'active'
		ORDER BY rm.joined_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("membres room: %w", err)
	}
	defer rows.Close()

	var members []*repositories.RoomMember
	for rows.Next() {
		member := &repositories.RoomMember{}
		user := &entities.User{}

		err := rows.Scan(
			&member.ID, &member.RoomID, &member.UserID, &member.Role,
			&member.JoinedAt, &member.LastSeen, &user.Username,
			&user.DisplayName, &user.Avatar, &member.IsOnline,
		)
		if err != nil {
			return nil, fmt.Errorf("scan membre: %w", err)
		}

		user.ID = member.UserID
		member.User = user
		members = append(members, member)
	}

	return members, nil
}

// IsUserInRoom vérifie si un utilisateur est dans une room
func (r *chatRepository) IsUserInRoom(ctx context.Context, roomID, userID int64) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM room_members 
			WHERE room_id = $1 AND user_id = $2 AND status = 'active'
		)
	`

	err := r.db.QueryRowContext(ctx, query, roomID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("vérification membre room: %w", err)
	}

	return exists, nil
}

// IsUserBanned vérifie si un utilisateur est banni d'une room
func (r *chatRepository) IsUserBanned(ctx context.Context, roomID, userID int64) (bool, error) {
	cacheKey := fmt.Sprintf("user:banned:%d:%d", roomID, userID)
	var banned bool
	if err := r.cache.Get(ctx, cacheKey, &banned); err == nil {
		return banned, nil
	}

	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1 FROM room_bans 
			WHERE room_id = $1 AND user_id = $2 
			  AND (expires_at IS NULL OR expires_at > NOW())
		)
	`

	err := r.db.QueryRowContext(ctx, query, roomID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("vérification ban: %w", err)
	}

	// Cache du résultat
	r.cache.Set(ctx, cacheKey, exists, 5*time.Minute)
	return exists, nil
}

// GetUserRoomRole récupère le rôle d'un utilisateur dans une room
func (r *chatRepository) GetUserRoomRole(ctx context.Context, roomID, userID int64) (entities.RoomRole, error) {
	var role entities.RoomRole
	query := `
		SELECT role FROM room_members 
		WHERE room_id = $1 AND user_id = $2 AND status = 'active'
	`

	err := r.db.QueryRowContext(ctx, query, roomID, userID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("utilisateur non membre de la room")
		}
		return "", fmt.Errorf("récupération rôle: %w", err)
	}

	return role, nil
}

// UpdateUserRoomRole met à jour le rôle d'un utilisateur dans une room
func (r *chatRepository) UpdateUserRoomRole(ctx context.Context, roomID, userID int64, role entities.RoomRole) error {
	query := `
		UPDATE room_members SET role = $3
		WHERE room_id = $1 AND user_id = $2 AND status = 'active'
	`

	result, err := r.db.ExecContext(ctx, query, roomID, userID, string(role))
	if err != nil {
		return fmt.Errorf("mise à jour rôle: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour rôle: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("membre non trouvé")
	}

	return nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// invalidateRoomCaches invalide les caches d'une room
func (r *chatRepository) invalidateRoomCaches(ctx context.Context, roomID int64, slug string) {
	keys := []string{
		fmt.Sprintf("room:id:%d", roomID),
		fmt.Sprintf("room:slug:%s", slug),
		fmt.Sprintf("room:members:%d", roomID),
		fmt.Sprintf("room:stats:%d", roomID),
	}

	for _, key := range keys {
		r.cache.Delete(ctx, key)
	}
}
