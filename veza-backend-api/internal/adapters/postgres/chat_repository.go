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
		room.CreatorID, room.MaxMembers, room.Password,
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
		&roomPtr.Password, &tags, &roomPtr.Settings,
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
	if err := r.cache.Set(ctx, cacheKey, roomPtr, 15*time.Minute); err != nil {
		r.logger.Error("Failed to cache room", zap.Error(err))
	}
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
		&roomPtr.Password, &tags, &roomPtr.Settings,
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
		room.MaxMembers, room.Password, pq.Array(room.Tags),
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
		// Conversion de RoomRole vers UserRole si nécessaire
		room.RequiredRole = entities.UserRole(userRole)

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

// GetRoomModerators récupère les modérateurs d'une room
func (r *chatRepository) GetRoomModerators(ctx context.Context, roomID int64) ([]*repositories.RoomMember, error) {
	query := `
		SELECT rm.id, rm.room_id, rm.user_id, rm.role, rm.joined_at, rm.last_seen,
		       u.username, u.display_name, u.avatar, u.is_online
		FROM room_members rm
		JOIN users u ON rm.user_id = u.id
		WHERE rm.room_id = $1 AND rm.status = 'active' 
		      AND (rm.role = 'moderator' OR rm.role = 'admin' OR rm.role = 'owner')
		ORDER BY rm.role DESC, rm.joined_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("récupération modérateurs room: %w", err)
	}
	defer rows.Close()

	var moderators []*repositories.RoomMember
	for rows.Next() {
		member := &repositories.RoomMember{}
		user := &entities.User{}

		err := rows.Scan(
			&member.ID, &member.RoomID, &member.UserID, &member.Role,
			&member.JoinedAt, &member.LastSeen, &user.Username,
			&user.DisplayName, &user.Avatar, &member.IsOnline,
		)
		if err != nil {
			return nil, fmt.Errorf("scan modérateur: %w", err)
		}

		user.ID = member.UserID
		member.User = user
		moderators = append(moderators, member)
	}

	return moderators, nil
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

// AcceptInvitation accepte une invitation à rejoindre une room
func (r *chatRepository) AcceptInvitation(ctx context.Context, code string, userID int64) error {
	// Récupérer l'invitation
	query := `
		SELECT id, room_id, inviter_id, uses, max_uses, expires_at, status
		FROM room_invitations
		WHERE code = $1 AND status = 'pending'
	`

	var invitation struct {
		ID        int64
		RoomID    int64
		InviterID int64
		Uses      int
		MaxUses   int
		ExpiresAt *time.Time
		Status    string
	}

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&invitation.ID, &invitation.RoomID, &invitation.InviterID,
		&invitation.Uses, &invitation.MaxUses, &invitation.ExpiresAt, &invitation.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("invitation non trouvée ou expirée")
		}
		return fmt.Errorf("récupération invitation: %w", err)
	}

	// Vérifier l'expiration
	if invitation.ExpiresAt != nil && invitation.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("invitation expirée")
	}

	// Vérifier le nombre d'utilisations
	if invitation.MaxUses > 0 && invitation.Uses >= invitation.MaxUses {
		return fmt.Errorf("invitation épuisée")
	}

	// Vérifier si l'utilisateur est déjà dans la room
	isInRoom, err := r.IsUserInRoom(ctx, invitation.RoomID, userID)
	if err != nil {
		return fmt.Errorf("vérification appartenance room: %w", err)
	}

	if isInRoom {
		return fmt.Errorf("utilisateur déjà membre de la room")
	}

	// Transaction pour accepter l'invitation et rejoindre la room
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("début transaction: %w", err)
	}
	defer tx.Rollback()

	// Mettre à jour l'invitation
	updateQuery := `
		UPDATE room_invitations SET
			uses = uses + 1, status = 'accepted', updated_at = $2
		WHERE id = $1
	`

	_, err = tx.ExecContext(ctx, updateQuery, invitation.ID, time.Now())
	if err != nil {
		return fmt.Errorf("mise à jour invitation: %w", err)
	}

	// Ajouter l'utilisateur à la room
	if err := r.JoinRoom(ctx, invitation.RoomID, userID); err != nil {
		return fmt.Errorf("ajout utilisateur à la room: %w", err)
	}

	// Valider la transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("validation transaction: %w", err)
	}

	// Invalider les caches
	r.invalidateRoomCaches(ctx, invitation.RoomID, "")

	return nil
}

// AddReaction ajoute une réaction à un message
func (r *chatRepository) AddReaction(ctx context.Context, messageID, userID int64, emoji string) error {
	query := `
		INSERT INTO message_reactions (message_id, user_id, emoji, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (message_id, user_id, emoji) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, messageID, userID, emoji, time.Now())
	if err != nil {
		return fmt.Errorf("ajout réaction: %w", err)
	}

	// Invalider les caches de réactions
	r.cache.Delete(ctx, fmt.Sprintf("message:reactions:%d", messageID))
	return nil
}

// RemoveReaction retire une réaction d'un message
func (r *chatRepository) RemoveReaction(ctx context.Context, messageID, userID int64, emoji string) error {
	query := `
		DELETE FROM message_reactions 
		WHERE message_id = $1 AND user_id = $2 AND emoji = $3
	`

	result, err := r.db.ExecContext(ctx, query, messageID, userID, emoji)
	if err != nil {
		return fmt.Errorf("suppression réaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification suppression réaction: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("réaction non trouvée")
	}

	// Invalider les caches de réactions
	r.cache.Delete(ctx, fmt.Sprintf("message:reactions:%d", messageID))
	return nil
}

// GetMessageReactions récupère les réactions d'un message
func (r *chatRepository) GetMessageReactions(ctx context.Context, messageID int64) ([]*repositories.MessageReaction, error) {
	query := `
		SELECT mr.id, mr.message_id, mr.user_id, mr.emoji, mr.created_at,
		       u.username, u.display_name, u.avatar
		FROM message_reactions mr
		JOIN users u ON mr.user_id = u.id
		WHERE mr.message_id = $1
		ORDER BY mr.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("récupération réactions: %w", err)
	}
	defer rows.Close()

	var reactions []*repositories.MessageReaction
	for rows.Next() {
		reaction := &repositories.MessageReaction{}
		user := &entities.User{}

		err := rows.Scan(
			&reaction.ID, &reaction.MessageID, &reaction.UserID, &reaction.Emoji, &reaction.CreatedAt,
			&user.Username, &user.DisplayName, &user.Avatar,
		)
		if err != nil {
			return nil, fmt.Errorf("scan réaction: %w", err)
		}

		user.ID = reaction.UserID
		reaction.User = user
		reactions = append(reactions, reaction)
	}

	return reactions, nil
}

// CreateInvitation crée une invitation à rejoindre une room
func (r *chatRepository) CreateInvitation(ctx context.Context, invitation *repositories.RoomInvitation) error {
	query := `
		INSERT INTO room_invitations (room_id, inviter_id, invitee_id, code, uses, max_uses, expires_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		invitation.RoomID, invitation.InviterID, invitation.InviteeID,
		invitation.Code, invitation.Uses, invitation.MaxUses,
		invitation.ExpiresAt, string(invitation.Status),
		invitation.CreatedAt, invitation.UpdatedAt,
	).Scan(&invitation.ID, &invitation.CreatedAt, &invitation.UpdatedAt)

	if err != nil {
		return fmt.Errorf("création invitation: %w", err)
	}

	return nil
}

// GetInvitation récupère une invitation par code
func (r *chatRepository) GetInvitation(ctx context.Context, code string) (*repositories.RoomInvitation, error) {
	query := `
		SELECT id, room_id, inviter_id, invitee_id, code, uses, max_uses, expires_at, status, created_at, updated_at
		FROM room_invitations
		WHERE code = $1
	`

	invitation := &repositories.RoomInvitation{}
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&invitation.ID, &invitation.RoomID, &invitation.InviterID, &invitation.InviteeID,
		&invitation.Code, &invitation.Uses, &invitation.MaxUses, &invitation.ExpiresAt,
		&invitation.Status, &invitation.CreatedAt, &invitation.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("récupération invitation: %w", err)
	}

	return invitation, nil
}

// DeclineInvitation refuse une invitation
func (r *chatRepository) DeclineInvitation(ctx context.Context, code string, userID int64) error {
	query := `
		UPDATE room_invitations SET
			status = 'declined', updated_at = $2
		WHERE code = $1 AND status = 'pending'
	`

	result, err := r.db.ExecContext(ctx, query, code, time.Now())
	if err != nil {
		return fmt.Errorf("refus invitation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification refus invitation: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invitation non trouvée ou déjà traitée")
	}

	return nil
}

// GetUserInvitations récupère les invitations d'un utilisateur
func (r *chatRepository) GetUserInvitations(ctx context.Context, userID int64) ([]*repositories.RoomInvitation, error) {
	query := `
		SELECT id, room_id, inviter_id, invitee_id, code, uses, max_uses, expires_at, status, created_at, updated_at
		FROM room_invitations
		WHERE invitee_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("récupération invitations utilisateur: %w", err)
	}
	defer rows.Close()

	var invitations []*repositories.RoomInvitation
	for rows.Next() {
		invitation := &repositories.RoomInvitation{}
		err := rows.Scan(
			&invitation.ID, &invitation.RoomID, &invitation.InviterID, &invitation.InviteeID,
			&invitation.Code, &invitation.Uses, &invitation.MaxUses, &invitation.ExpiresAt,
			&invitation.Status, &invitation.CreatedAt, &invitation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan invitation: %w", err)
		}
		invitations = append(invitations, invitation)
	}

	return invitations, nil
}

// GetRoomInvitations récupère les invitations d'une room
func (r *chatRepository) GetRoomInvitations(ctx context.Context, roomID int64) ([]*repositories.RoomInvitation, error) {
	query := `
		SELECT id, room_id, inviter_id, invitee_id, code, uses, max_uses, expires_at, status, created_at, updated_at
		FROM room_invitations
		WHERE room_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("récupération invitations room: %w", err)
	}
	defer rows.Close()

	var invitations []*repositories.RoomInvitation
	for rows.Next() {
		invitation := &repositories.RoomInvitation{}
		err := rows.Scan(
			&invitation.ID, &invitation.RoomID, &invitation.InviterID, &invitation.InviteeID,
			&invitation.Code, &invitation.Uses, &invitation.MaxUses, &invitation.ExpiresAt,
			&invitation.Status, &invitation.CreatedAt, &invitation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan invitation: %w", err)
		}
		invitations = append(invitations, invitation)
	}

	return invitations, nil
}

// CreateMessage crée un nouveau message
func (r *chatRepository) CreateMessage(ctx context.Context, message *entities.ChatMessage) error {
	query := `
		INSERT INTO chat_messages (
			uuid, room_id, user_id, type, content, status, is_edited, is_system,
			parent_id, thread_count, is_flagged, flag_reason, moderated_by,
			reaction_count, mentioned_users, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		message.UUID, message.RoomID, message.UserID, string(message.Type),
		message.Content, string(message.Status), message.IsEdited, message.IsSystem,
		message.ParentID, message.ThreadCount, message.IsFlagged, message.FlagReason,
		message.ModeratedBy, message.ReactionCount, pq.Array(message.MentionedUsers),
		message.CreatedAt, message.UpdatedAt,
	).Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt)

	if err != nil {
		return fmt.Errorf("création message: %w", err)
	}

	// Invalider les caches
	r.cache.Delete(ctx, fmt.Sprintf("room:messages:%d", message.RoomID))
	return nil
}

// GetMessageByID récupère un message par ID
func (r *chatRepository) GetMessageByID(ctx context.Context, messageID int64) (*entities.ChatMessage, error) {
	query := `
		SELECT id, uuid, room_id, user_id, type, content, status, is_edited, is_system,
		       parent_id, thread_count, is_flagged, flag_reason, moderated_by,
		       reaction_count, mentioned_users, created_at, updated_at, edited_at, deleted_at
		FROM chat_messages
		WHERE id = $1 AND deleted_at IS NULL
	`

	message := &entities.ChatMessage{}
	var mentionedUsers pq.Int64Array

	err := r.db.QueryRowContext(ctx, query, messageID).Scan(
		&message.ID, &message.UUID, &message.RoomID, &message.UserID,
		&message.Type, &message.Content, &message.Status, &message.IsEdited,
		&message.IsSystem, &message.ParentID, &message.ThreadCount,
		&message.IsFlagged, &message.FlagReason, &message.ModeratedBy,
		&message.ReactionCount, &mentionedUsers, &message.CreatedAt,
		&message.UpdatedAt, &message.EditedAt, &message.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("récupération message: %w", err)
	}

	message.MentionedUsers = []int64(mentionedUsers)
	return message, nil
}

// UpdateMessage met à jour un message
func (r *chatRepository) UpdateMessage(ctx context.Context, message *entities.ChatMessage) error {
	query := `
		UPDATE chat_messages SET
			content = $2, is_edited = $3, edited_at = $4, updated_at = $5
		WHERE id = $1 AND deleted_at IS NULL
	`

	message.UpdatedAt = time.Now()
	if message.IsEdited {
		message.EditedAt = &message.UpdatedAt
	}

	result, err := r.db.ExecContext(ctx, query,
		message.ID, message.Content, message.IsEdited,
		message.EditedAt, message.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("mise à jour message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour message: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message non trouvé")
	}

	return nil
}

// DeleteMessage supprime un message (soft delete)
func (r *chatRepository) DeleteMessage(ctx context.Context, messageID, userID int64) error {
	query := `
		UPDATE chat_messages SET
			deleted_at = $3, updated_at = $3
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, messageID, userID, time.Now())
	if err != nil {
		return fmt.Errorf("suppression message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification suppression message: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message non trouvé ou non autorisé")
	}

	return nil
}

// GetRoomMessages récupère les messages d'une room avec pagination
func (r *chatRepository) GetRoomMessages(ctx context.Context, roomID int64, limit, offset int) ([]*entities.ChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT id, uuid, room_id, user_id, type, content, status, is_edited, is_system,
		       parent_id, thread_count, is_flagged, flag_reason, moderated_by,
		       reaction_count, mentioned_users, created_at, updated_at, edited_at
		FROM chat_messages
		WHERE room_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("récupération messages room: %w", err)
	}
	defer rows.Close()

	var messages []*entities.ChatMessage
	for rows.Next() {
		message := &entities.ChatMessage{}
		var mentionedUsers pq.Int64Array

		err := rows.Scan(
			&message.ID, &message.UUID, &message.RoomID, &message.UserID,
			&message.Type, &message.Content, &message.Status, &message.IsEdited,
			&message.IsSystem, &message.ParentID, &message.ThreadCount,
			&message.IsFlagged, &message.FlagReason, &message.ModeratedBy,
			&message.ReactionCount, &mentionedUsers, &message.CreatedAt,
			&message.UpdatedAt, &message.EditedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		message.MentionedUsers = []int64(mentionedUsers)
		messages = append(messages, message)
	}

	return messages, nil
}

// SearchMessages recherche des messages dans une room
func (r *chatRepository) SearchMessages(ctx context.Context, roomID int64, query string, limit, offset int) ([]*entities.ChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	sqlQuery := `
		SELECT id, uuid, room_id, user_id, type, content, status, is_edited, is_system,
		       parent_id, thread_count, is_flagged, flag_reason, moderated_by,
		       reaction_count, mentioned_users, created_at, updated_at, edited_at
		FROM chat_messages
		WHERE room_id = $1 AND deleted_at IS NULL AND content ILIKE $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	searchTerm := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, sqlQuery, roomID, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("recherche messages: %w", err)
	}
	defer rows.Close()

	var messages []*entities.ChatMessage
	for rows.Next() {
		message := &entities.ChatMessage{}
		var mentionedUsers pq.Int64Array

		err := rows.Scan(
			&message.ID, &message.UUID, &message.RoomID, &message.UserID,
			&message.Type, &message.Content, &message.Status, &message.IsEdited,
			&message.IsSystem, &message.ParentID, &message.ThreadCount,
			&message.IsFlagged, &message.FlagReason, &message.ModeratedBy,
			&message.ReactionCount, &mentionedUsers, &message.CreatedAt,
			&message.UpdatedAt, &message.EditedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		message.MentionedUsers = []int64(mentionedUsers)
		messages = append(messages, message)
	}

	return messages, nil
}

// GetUserMessages récupère les messages d'un utilisateur
func (r *chatRepository) GetUserMessages(ctx context.Context, userID int64, limit, offset int) ([]*entities.ChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT id, uuid, room_id, user_id, type, content, status, is_edited, is_system,
		       parent_id, thread_count, is_flagged, flag_reason, moderated_by,
		       reaction_count, mentioned_users, created_at, updated_at, edited_at
		FROM chat_messages
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("récupération messages utilisateur: %w", err)
	}
	defer rows.Close()

	var messages []*entities.ChatMessage
	for rows.Next() {
		message := &entities.ChatMessage{}
		var mentionedUsers pq.Int64Array

		err := rows.Scan(
			&message.ID, &message.UUID, &message.RoomID, &message.UserID,
			&message.Type, &message.Content, &message.Status, &message.IsEdited,
			&message.IsSystem, &message.ParentID, &message.ThreadCount,
			&message.IsFlagged, &message.FlagReason, &message.ModeratedBy,
			&message.ReactionCount, &mentionedUsers, &message.CreatedAt,
			&message.UpdatedAt, &message.EditedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		message.MentionedUsers = []int64(mentionedUsers)
		messages = append(messages, message)
	}

	return messages, nil
}

// GetMessageReplies récupère les réponses à un message
func (r *chatRepository) GetMessageReplies(ctx context.Context, messageID int64, limit, offset int) ([]*entities.ChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT id, uuid, room_id, user_id, type, content, status, is_edited, is_system,
		       parent_id, thread_count, is_flagged, flag_reason, moderated_by,
		       reaction_count, mentioned_users, created_at, updated_at, edited_at
		FROM chat_messages
		WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, messageID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("récupération réponses: %w", err)
	}
	defer rows.Close()

	var messages []*entities.ChatMessage
	for rows.Next() {
		message := &entities.ChatMessage{}
		var mentionedUsers pq.Int64Array

		err := rows.Scan(
			&message.ID, &message.UUID, &message.RoomID, &message.UserID,
			&message.Type, &message.Content, &message.Status, &message.IsEdited,
			&message.IsSystem, &message.ParentID, &message.ThreadCount,
			&message.IsFlagged, &message.FlagReason, &message.ModeratedBy,
			&message.ReactionCount, &mentionedUsers, &message.CreatedAt,
			&message.UpdatedAt, &message.EditedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		message.MentionedUsers = []int64(mentionedUsers)
		messages = append(messages, message)
	}

	return messages, nil
}

// CreateModerationLog crée un log de modération
func (r *chatRepository) CreateModerationLog(ctx context.Context, log *repositories.ModerationLog) error {
	query := `
		INSERT INTO moderation_logs (
			room_id, moderator_id, target_user_id, message_id, action, reason, duration, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var durationSeconds *int64
	if log.Duration != nil {
		duration := int64(*log.Duration / int64(time.Second))
		durationSeconds = &duration
	}

	err := r.db.QueryRowContext(ctx, query,
		log.RoomID, log.ModeratorID, log.TargetUserID, log.MessageID,
		log.Action, log.Reason, durationSeconds, log.CreatedAt,
	).Scan(&log.ID)

	if err != nil {
		return fmt.Errorf("création log modération: %w", err)
	}

	return nil
}

// GetActiveUsers récupère les utilisateurs les plus actifs
func (r *chatRepository) GetActiveUsers(ctx context.Context, roomID int64, period time.Duration, limit int) ([]*entities.User, error) {
	query := `
		SELECT DISTINCT u.id, u.username, u.display_name, u.avatar, u.is_online,
		       COUNT(m.id) as message_count
		FROM users u
		JOIN room_members rm ON u.id = rm.user_id AND rm.room_id = $1 AND rm.status = 'active'
		LEFT JOIN chat_messages m ON u.id = m.user_id AND m.room_id = $1 
		       AND m.created_at >= NOW() - INTERVAL '1 day' * $2
		GROUP BY u.id, u.username, u.display_name, u.avatar, u.is_online
		ORDER BY message_count DESC, u.username ASC
		LIMIT $3
	`

	days := int(period.Hours() / 24)
	rows, err := r.db.QueryContext(ctx, query, roomID, days, limit)
	if err != nil {
		return nil, fmt.Errorf("récupération utilisateurs actifs: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		user := &entities.User{}
		var messageCount int64

		err := rows.Scan(
			&user.ID, &user.Username, &user.DisplayName, &user.Avatar, &user.IsOnline, &messageCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan utilisateur actif: %w", err)
		}

		users = append(users, user)
	}

	return users, nil
}

// GetChatStats récupère les statistiques générales du chat
func (r *chatRepository) GetChatStats(ctx context.Context) (*repositories.ChatStats, error) {
	query := `
		SELECT 
			COUNT(DISTINCT r.id) as total_rooms,
			COUNT(DISTINCT m.id) as total_messages,
			COUNT(DISTINCT u.id) as total_users,
			COUNT(DISTINCT CASE WHEN r.status = 'active' THEN r.id END) as active_rooms,
			COUNT(DISTINCT CASE WHEN u.is_online THEN u.id END) as online_users,
			COUNT(DISTINCT CASE WHEN m.created_at >= NOW() - INTERVAL '1 day' THEN m.id END) as today_messages,
			COUNT(DISTINCT CASE WHEN m.created_at >= NOW() - INTERVAL '1 day' THEN m.user_id END) as today_users,
			COUNT(DISTINCT CASE WHEN rm.member_count > 10 THEN r.id END) as popular_rooms,
			NOW() as updated_at
		FROM chat_rooms r
		LEFT JOIN chat_messages m ON r.id = m.room_id AND m.deleted_at IS NULL
		LEFT JOIN users u ON m.user_id = u.id
		LEFT JOIN (
			SELECT room_id, COUNT(user_id) as member_count 
			FROM room_members 
			WHERE status = 'active' 
			GROUP BY room_id
		) rm ON r.id = rm.room_id
		WHERE r.deleted_at IS NULL
	`

	stats := &repositories.ChatStats{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalRooms, &stats.TotalMessages, &stats.TotalUsers,
		&stats.ActiveRooms, &stats.OnlineUsers, &stats.TodayMessages,
		&stats.TodayUsers, &stats.PopularRooms, &stats.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("récupération statistiques chat: %w", err)
	}

	return stats, nil
}

// GetModerationLogs récupère les logs de modération
func (r *chatRepository) GetModerationLogs(ctx context.Context, roomID int64, limit, offset int) ([]*repositories.ModerationLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT id, room_id, moderator_id, target_user_id, message_id, action, reason, duration, created_at
		FROM moderation_logs
		WHERE room_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("récupération logs modération: %w", err)
	}
	defer rows.Close()

	var logs []*repositories.ModerationLog
	for rows.Next() {
		log := &repositories.ModerationLog{}
		var durationSeconds *int64

		err := rows.Scan(
			&log.ID, &log.RoomID, &log.ModeratorID, &log.TargetUserID,
			&log.MessageID, &log.Action, &log.Reason, &durationSeconds, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan log modération: %w", err)
		}

		if durationSeconds != nil {
			duration := int64(*durationSeconds)
			log.Duration = &duration
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetOnlineUsers récupère les utilisateurs en ligne dans une room
func (r *chatRepository) GetOnlineUsers(ctx context.Context, roomID int64) ([]*entities.User, error) {
	query := `
		SELECT u.id, u.username, u.display_name, u.avatar, u.is_online
		FROM users u
		JOIN room_members rm ON u.id = rm.user_id AND rm.room_id = $1 AND rm.status = 'active'
		WHERE u.is_online = true
		ORDER BY u.username ASC
	`

	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("récupération utilisateurs en ligne: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		user := &entities.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.DisplayName, &user.Avatar, &user.IsOnline,
		)
		if err != nil {
			return nil, fmt.Errorf("scan utilisateur en ligne: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// PinMessage épingle un message
func (r *chatRepository) PinMessage(ctx context.Context, messageID, userID int64) error {
	query := `
		UPDATE chat_messages 
		SET is_pinned = true, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, messageID, time.Now())
	if err != nil {
		return fmt.Errorf("épinglage message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification épinglage: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message non trouvé")
	}

	// Invalider les caches de messages
	r.invalidateMessageCaches(ctx, messageID)
	return nil
}

// UnpinMessage désépingle un message
func (r *chatRepository) UnpinMessage(ctx context.Context, messageID, userID int64) error {
	query := `
		UPDATE chat_messages 
		SET is_pinned = false, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, messageID, time.Now())
	if err != nil {
		return fmt.Errorf("désépinglage message: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification désépinglage: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("message non trouvé")
	}

	// Invalider les caches de messages
	r.invalidateMessageCaches(ctx, messageID)
	return nil
}

// GetRoomMessagesAfter récupère les messages après un timestamp
func (r *chatRepository) GetRoomMessagesAfter(ctx context.Context, roomID int64, after time.Time, limit int) ([]*entities.ChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT id, uuid, room_id, user_id, type, content, status, is_edited, is_system,
		       parent_id, thread_count, is_flagged, flag_reason, moderated_by,
		       reaction_count, mentioned_users, created_at, updated_at, edited_at
		FROM chat_messages
		WHERE room_id = $1 AND created_at > $2 AND deleted_at IS NULL
		ORDER BY created_at ASC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, roomID, after, limit)
	if err != nil {
		return nil, fmt.Errorf("récupération messages après timestamp: %w", err)
	}
	defer rows.Close()

	var messages []*entities.ChatMessage
	for rows.Next() {
		message := &entities.ChatMessage{}
		var mentionedUsers pq.Int64Array

		err := rows.Scan(
			&message.ID, &message.UUID, &message.RoomID, &message.UserID,
			&message.Type, &message.Content, &message.Status, &message.IsEdited,
			&message.IsSystem, &message.ParentID, &message.ThreadCount,
			&message.IsFlagged, &message.FlagReason, &message.ModeratedBy,
			&message.ReactionCount, &mentionedUsers, &message.CreatedAt,
			&message.UpdatedAt, &message.EditedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		message.MentionedUsers = []int64(mentionedUsers)
		messages = append(messages, message)
	}

	return messages, nil
}

// GetRoomMessagesBefore récupère les messages avant un timestamp
func (r *chatRepository) GetRoomMessagesBefore(ctx context.Context, roomID int64, before time.Time, limit int) ([]*entities.ChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT id, uuid, room_id, user_id, type, content, status, is_edited, is_system,
		       parent_id, thread_count, is_flagged, flag_reason, moderated_by,
		       reaction_count, mentioned_users, created_at, updated_at, edited_at
		FROM chat_messages
		WHERE room_id = $1 AND created_at < $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, roomID, before, limit)
	if err != nil {
		return nil, fmt.Errorf("récupération messages avant timestamp: %w", err)
	}
	defer rows.Close()

	var messages []*entities.ChatMessage
	for rows.Next() {
		message := &entities.ChatMessage{}
		var mentionedUsers pq.Int64Array

		err := rows.Scan(
			&message.ID, &message.UUID, &message.RoomID, &message.UserID,
			&message.Type, &message.Content, &message.Status, &message.IsEdited,
			&message.IsSystem, &message.ParentID, &message.ThreadCount,
			&message.IsFlagged, &message.FlagReason, &message.ModeratedBy,
			&message.ReactionCount, &mentionedUsers, &message.CreatedAt,
			&message.UpdatedAt, &message.EditedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		message.MentionedUsers = []int64(mentionedUsers)
		messages = append(messages, message)
	}

	return messages, nil
}

// GetPinnedMessages récupère les messages épinglés d'une room
func (r *chatRepository) GetPinnedMessages(ctx context.Context, roomID int64) ([]*entities.ChatMessage, error) {
	query := `
		SELECT m.id, m.uuid, m.room_id, m.user_id, m.content, m.type,
		       m.parent_id, m.is_edited, m.created_at, m.updated_at,
		       u.username, u.display_name, u.avatar
		FROM chat_messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.room_id = $1 AND m.is_flagged = false AND m.deleted_at IS NULL
		ORDER BY m.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("récupération messages épinglés: %w", err)
	}
	defer rows.Close()

	var messages []*entities.ChatMessage
	for rows.Next() {
		message := &entities.ChatMessage{}
		var parentID sql.NullInt64
		var isEdited bool

		err := rows.Scan(
			&message.ID, &message.UUID, &message.RoomID, &message.UserID,
			&message.Content, &message.Type, &parentID, &isEdited,
			&message.CreatedAt, &message.UpdatedAt,
			&message.User.Username, &message.User.DisplayName, &message.User.Avatar,
		)

		if err != nil {
			return nil, fmt.Errorf("scan messages épinglés: %w", err)
		}

		if parentID.Valid {
			message.ParentID = &parentID.Int64
		}
		message.IsEdited = isEdited

		messages = append(messages, message)
	}

	return messages, nil
}

// ReportMessage signale un message
func (r *chatRepository) ReportMessage(ctx context.Context, report *repositories.MessageReport) error {
	query := `
		INSERT INTO message_reports (
			message_id, reporter_id, reason, description, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		report.MessageID, report.ReporterID, report.Reason,
		report.Description, report.Status, report.CreatedAt,
	).Scan(&report.ID)

	if err != nil {
		return fmt.Errorf("création signalement: %w", err)
	}

	return nil
}

// GetReports récupère les signalements
func (r *chatRepository) GetReports(ctx context.Context, roomID int64, status repositories.ReportStatus, limit, offset int) ([]*repositories.MessageReport, error) {
	query := `
		SELECT r.id, r.message_id, r.reporter_id, r.reason, r.description,
		       r.status, r.moderator_id, r.created_at, r.updated_at
		FROM message_reports r
		JOIN chat_messages m ON r.message_id = m.id
		WHERE m.room_id = $1 AND r.status = $2
		ORDER BY r.created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, roomID, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("récupération signalements: %w", err)
	}
	defer rows.Close()

	var reports []*repositories.MessageReport
	for rows.Next() {
		report := &repositories.MessageReport{}
		err := rows.Scan(
			&report.ID, &report.MessageID, &report.ReporterID,
			&report.Reason, &report.Description, &report.Status,
			&report.ModeratorID, &report.CreatedAt, &report.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan signalement: %w", err)
		}
		reports = append(reports, report)
	}

	return reports, nil
}

// UpdateReportStatus met à jour le statut d'un signalement
func (r *chatRepository) UpdateReportStatus(ctx context.Context, reportID int64, status repositories.ReportStatus, moderatorID int64) error {
	query := `
		UPDATE message_reports 
		SET status = $2, moderator_id = $3, updated_at = $4
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, reportID, status, moderatorID, time.Now())
	if err != nil {
		return fmt.Errorf("mise à jour statut signalement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("signalement non trouvé")
	}

	return nil
}

// UpdateUserPresence met à jour la présence d'un utilisateur dans une room
func (r *chatRepository) UpdateUserPresence(ctx context.Context, roomID, userID int64, isTyping bool) error {
	query := `
		UPDATE room_members 
		SET last_seen = $3, is_typing = $4
		WHERE room_id = $1 AND user_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, roomID, userID, time.Now(), isTyping)
	if err != nil {
		return fmt.Errorf("mise à jour présence: %w", err)
	}

	return nil
}

// GetTypingUsers récupère les utilisateurs en train de taper
func (r *chatRepository) GetTypingUsers(ctx context.Context, roomID int64) ([]*entities.User, error) {
	query := `
		SELECT u.id, u.username, u.display_name, u.avatar
		FROM users u
		JOIN room_members rm ON u.id = rm.user_id AND rm.room_id = $1 AND rm.status = 'active'
		WHERE rm.is_typing = true AND rm.last_seen >= NOW() - INTERVAL '30 seconds'
		ORDER BY rm.last_seen DESC
	`

	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("récupération utilisateurs tapant: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		user := &entities.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.DisplayName, &user.Avatar,
		)
		if err != nil {
			return nil, fmt.Errorf("scan utilisateur tapant: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// GetRoomActivity récupère l'activité récente d'une room
func (r *chatRepository) GetRoomActivity(ctx context.Context, roomID int64, since time.Time) (*repositories.RoomActivity, error) {
	query := `
		SELECT 
			$1 as room_id,
			COUNT(m.id) as message_count,
			COUNT(DISTINCT m.user_id) as active_user_count,
			MAX(rm.member_count) as peak_online_users,
			AVG(rm.member_count) as average_online_users,
			MAX(m.created_at) as last_activity,
			$2 as period
		FROM chat_messages m
		LEFT JOIN (
			SELECT room_id, COUNT(user_id) as member_count
			FROM room_members 
			WHERE room_id = $1 AND status = 'active'
			GROUP BY room_id
		) rm ON m.room_id = rm.room_id
		WHERE m.room_id = $1 AND m.created_at >= $3
	`

	activity := &repositories.RoomActivity{}
	err := r.db.QueryRowContext(ctx, query, roomID, time.Since(since), since).Scan(
		&activity.RoomID, &activity.MessageCount, &activity.ActiveUserCount,
		&activity.PeakOnlineUsers, &activity.AverageOnlineUsers,
		&activity.LastActivity, &activity.Period,
	)

	if err != nil {
		return nil, fmt.Errorf("récupération activité room: %w", err)
	}

	return activity, nil
}

// GetUserChatStats récupère les statistiques de chat d'un utilisateur
func (r *chatRepository) GetUserChatStats(ctx context.Context, userID int64) (*repositories.UserChatStats, error) {
	query := `
		SELECT 
			$1 as user_id,
			COUNT(m.id) as total_messages,
			COUNT(DISTINCT rm.room_id) as rooms_joined,
			COUNT(DISTINCT CASE WHEN r.creator_id = $1 THEN r.id END) as rooms_created,
			COUNT(DISTINCT mr.id) as reactions_given,
			COUNT(DISTINCT mr2.id) as reactions_received,
			AVG(daily_count) as average_per_day,
			MAX(CASE WHEN daily_count = (SELECT MAX(daily_count) FROM (
				SELECT COUNT(*) as daily_count FROM chat_messages 
				WHERE user_id = $1 GROUP BY DATE(created_at)
			) t) THEN room_id END) as most_active_room,
			MAX(m.created_at) as last_message,
			NOW() as updated_at
		FROM chat_messages m
		LEFT JOIN room_members rm ON m.user_id = rm.user_id
		LEFT JOIN chat_rooms r ON rm.room_id = r.id
		LEFT JOIN message_reactions mr ON m.user_id = mr.user_id
		LEFT JOIN message_reactions mr2 ON m.id = mr2.message_id
		LEFT JOIN (
			SELECT DATE(created_at) as date, COUNT(*) as daily_count, room_id
			FROM chat_messages 
			WHERE user_id = $1 
			GROUP BY DATE(created_at), room_id
		) daily ON m.user_id = daily.date
		WHERE m.user_id = $1
	`

	stats := &repositories.UserChatStats{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.UserID, &stats.TotalMessages, &stats.RoomsJoined,
		&stats.RoomsCreated, &stats.ReactionsGiven, &stats.ReactionsReceived,
		&stats.AveragePerDay, &stats.MostActiveRoom, &stats.LastMessage, &stats.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("récupération statistiques utilisateur: %w", err)
	}

	return stats, nil
}

// GetRoomUsageStats récupère les statistiques d'utilisation d'une room
func (r *chatRepository) GetRoomUsageStats(ctx context.Context, roomID int64, period time.Duration) (*repositories.RoomUsageStats, error) {
	query := `
		SELECT 
			$1 as room_id,
			$2 as period,
			COUNT(m.id) as message_count,
			COUNT(DISTINCT m.user_id) as unique_users,
			MAX(concurrent_users) as peak_concurrent_users,
			AVG(session_time) as average_session_time,
			EXTRACT(HOUR FROM m.created_at) as most_active_hour,
			NOW() as updated_at
		FROM chat_messages m
		LEFT JOIN (
			SELECT room_id, COUNT(DISTINCT user_id) as concurrent_users
			FROM room_members 
			WHERE room_id = $1 AND status = 'active'
			GROUP BY room_id
		) cu ON m.room_id = cu.room_id
		LEFT JOIN (
			SELECT room_id, AVG(EXTRACT(EPOCH FROM (last_seen - joined_at))) as session_time
			FROM room_members 
			WHERE room_id = $1 AND status = 'active'
			GROUP BY room_id
		) st ON m.room_id = st.room_id
		WHERE m.room_id = $1 AND m.created_at >= NOW() - $2
	`

	stats := &repositories.RoomUsageStats{}
	err := r.db.QueryRowContext(ctx, query, roomID, period).Scan(
		&stats.RoomID, &stats.Period, &stats.MessageCount, &stats.UniqueUsers,
		&stats.PeakConcurrentUsers, &stats.AverageSessionTime, &stats.MostActiveHour, &stats.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("récupération statistiques utilisation: %w", err)
	}

	return stats, nil
}

// GetPopularRooms récupère les rooms les plus populaires
func (r *chatRepository) GetPopularRooms(ctx context.Context, period time.Duration, limit int) ([]*entities.ChatRoom, error) {
	query := `
		SELECT r.id, r.uuid, r.name, r.slug, r.description, r.topic,
		       r.type, r.privacy, r.status, r.creator_id, r.max_members,
		       r.password_hash, r.tags, r.settings, r.created_at, r.updated_at,
		       COUNT(rm.user_id) as member_count
		FROM chat_rooms r
		LEFT JOIN room_members rm ON r.id = rm.room_id AND rm.status = 'active'
		WHERE r.deleted_at IS NULL
		GROUP BY r.id
		ORDER BY member_count DESC, r.created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("récupération rooms populaires: %w", err)
	}
	defer rows.Close()

	var rooms []*entities.ChatRoom
	for rows.Next() {
		room := &entities.ChatRoom{}
		var tags pq.StringArray

		err := rows.Scan(
			&room.ID, &room.UUID, &room.Name, &room.Slug,
			&room.Description, &room.Topic, &room.Type, &room.Privacy,
			&room.Status, &room.CreatorID, &room.MaxMembers,
			&room.Password, &tags, &room.Settings,
			&room.CreatedAt, &room.UpdatedAt, &room.MemberCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan room populaire: %w", err)
		}

		room.Tags = []string(tags)
		rooms = append(rooms, room)
	}

	return rooms, nil
}

// invalidateMessageCaches invalide les caches liés aux messages
func (r *chatRepository) invalidateMessageCaches(ctx context.Context, messageID int64) {
	keys := []string{
		fmt.Sprintf("message:id:%d", messageID),
		fmt.Sprintf("message:reactions:%d", messageID),
	}

	for _, key := range keys {
		r.cache.Delete(ctx, key)
	}
}

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
