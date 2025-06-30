package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/core/domain/entities"
	"github.com/okinrev/veza-web-app/internal/core/domain/repositories"
)

// userRepositoryComplete implémentation PostgreSQL complète du UserRepository
type userRepositoryComplete struct {
	db     *sql.DB
	cache  CacheService
	logger *zap.Logger
}

// CacheService interface pour le cache Redis
type CacheService interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
}

// NewUserRepositoryComplete crée une nouvelle instance complète du repository utilisateur
func NewUserRepositoryComplete(db *sql.DB, cache CacheService, logger *zap.Logger) (repositories.UserRepository, error) {
	return &userRepositoryComplete{
		db:     db,
		cache:  cache,
		logger: logger,
	}, nil
}

// ============================================================================
// CRUD DE BASE
// ============================================================================

// Create crée un nouvel utilisateur
func (r *userRepositoryComplete) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (
			uuid, username, email, password_hash, salt, first_name, last_name, 
			display_name, avatar, bio, role, status, email_verified, 
			email_verification_token, two_factor_enabled, two_factor_secret,
			is_online, last_seen, timezone, language, theme, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23
		) RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		user.UUID, user.Username, user.Email, user.PasswordHash, user.Salt,
		user.FirstName, user.LastName, user.DisplayName, user.Avatar, user.Bio,
		string(user.Role), string(user.Status), user.EmailVerified,
		user.EmailVerificationToken, user.TwoFactorEnabled, user.TwoFactorSecret,
		user.IsOnline, user.LastSeen, user.Timezone, user.Language, user.Theme,
		user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		r.logger.Error("Erreur création utilisateur", zap.Error(err))
		return fmt.Errorf("création utilisateur: %w", err)
	}

	// Invalider le cache
	r.invalidateUserCaches(ctx, user.ID, user.Username, user.Email)
	return nil
}

// GetByID récupère un utilisateur par son ID
func (r *userRepositoryComplete) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	// Tentative cache
	cacheKey := fmt.Sprintf("user:id:%d", id)
	var user entities.User
	if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	query := `
		SELECT id, uuid, username, email, password_hash, salt, first_name, last_name,
		       display_name, avatar, bio, role, status, email_verified, 
		       email_verification_token, two_factor_enabled, two_factor_secret,
		       is_online, last_seen, last_login_ip, login_attempts, locked_until,
		       timezone, language, theme, created_at, updated_at, deleted_at,
		       COALESCE(message_count, 0) as message_count,
		       COALESCE(stream_count, 0) as stream_count
		FROM users 
		LEFT JOIN user_stats USING(id)
		WHERE id = $1 AND deleted_at IS NULL
	`

	userPtr := &entities.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&userPtr.ID, &userPtr.UUID, &userPtr.Username, &userPtr.Email,
		&userPtr.PasswordHash, &userPtr.Salt, &userPtr.FirstName, &userPtr.LastName,
		&userPtr.DisplayName, &userPtr.Avatar, &userPtr.Bio, &userPtr.Role, &userPtr.Status,
		&userPtr.EmailVerified, &userPtr.EmailVerificationToken,
		&userPtr.TwoFactorEnabled, &userPtr.TwoFactorSecret,
		&userPtr.IsOnline, &userPtr.LastSeen, &userPtr.LastLoginIP,
		&userPtr.LoginAttempts, &userPtr.LockedUntil, &userPtr.Timezone,
		&userPtr.Language, &userPtr.Theme, &userPtr.CreatedAt, &userPtr.UpdatedAt,
		&userPtr.DeletedAt, &userPtr.MessageCount, &userPtr.StreamCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Erreur récupération utilisateur par ID", zap.Error(err), zap.Int64("user_id", id))
		return nil, fmt.Errorf("récupération utilisateur: %w", err)
	}

	// Cache du résultat
	r.cache.Set(ctx, cacheKey, userPtr, 15*time.Minute)
	return userPtr, nil
}

// GetByUUID récupère un utilisateur par son UUID
func (r *userRepositoryComplete) GetByUUID(ctx context.Context, uuid string) (*entities.User, error) {
	query := `
		SELECT id, uuid, username, email, password_hash, salt, first_name, last_name,
		       display_name, avatar, bio, role, status, email_verified, 
		       email_verification_token, two_factor_enabled, two_factor_secret,
		       is_online, last_seen, last_login_ip, login_attempts, locked_until,
		       timezone, language, theme, created_at, updated_at, deleted_at,
		       COALESCE(message_count, 0) as message_count,
		       COALESCE(stream_count, 0) as stream_count
		FROM users 
		LEFT JOIN user_stats USING(id)
		WHERE uuid = $1 AND deleted_at IS NULL
	`

	user := &entities.User{}
	err := r.db.QueryRowContext(ctx, query, uuid).Scan(
		&user.ID, &user.UUID, &user.Username, &user.Email,
		&user.PasswordHash, &user.Salt, &user.FirstName, &user.LastName,
		&user.DisplayName, &user.Avatar, &user.Bio, &user.Role, &user.Status,
		&user.EmailVerified, &user.EmailVerificationToken,
		&user.TwoFactorEnabled, &user.TwoFactorSecret,
		&user.IsOnline, &user.LastSeen, &user.LastLoginIP,
		&user.LoginAttempts, &user.LockedUntil, &user.Timezone,
		&user.Language, &user.Theme, &user.CreatedAt, &user.UpdatedAt,
		&user.DeletedAt, &user.MessageCount, &user.StreamCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("récupération utilisateur par UUID: %w", err)
	}

	return user, nil
}

// GetByEmail récupère un utilisateur par son email
func (r *userRepositoryComplete) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	cacheKey := fmt.Sprintf("user:email:%s", email)
	var user entities.User
	if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	query := `
		SELECT id, uuid, username, email, password_hash, salt, first_name, last_name,
		       display_name, avatar, bio, role, status, email_verified, 
		       email_verification_token, two_factor_enabled, two_factor_secret,
		       is_online, last_seen, last_login_ip, login_attempts, locked_until,
		       timezone, language, theme, created_at, updated_at, deleted_at,
		       COALESCE(message_count, 0) as message_count,
		       COALESCE(stream_count, 0) as stream_count
		FROM users 
		LEFT JOIN user_stats USING(id)
		WHERE email = $1 AND deleted_at IS NULL
	`

	userPtr := &entities.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&userPtr.ID, &userPtr.UUID, &userPtr.Username, &userPtr.Email,
		&userPtr.PasswordHash, &userPtr.Salt, &userPtr.FirstName, &userPtr.LastName,
		&userPtr.DisplayName, &userPtr.Avatar, &userPtr.Bio, &userPtr.Role, &userPtr.Status,
		&userPtr.EmailVerified, &userPtr.EmailVerificationToken,
		&userPtr.TwoFactorEnabled, &userPtr.TwoFactorSecret,
		&userPtr.IsOnline, &userPtr.LastSeen, &userPtr.LastLoginIP,
		&userPtr.LoginAttempts, &userPtr.LockedUntil, &userPtr.Timezone,
		&userPtr.Language, &userPtr.Theme, &userPtr.CreatedAt, &userPtr.UpdatedAt,
		&userPtr.DeletedAt, &userPtr.MessageCount, &userPtr.StreamCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("récupération utilisateur par email: %w", err)
	}

	r.cache.Set(ctx, cacheKey, userPtr, 15*time.Minute)
	return userPtr, nil
}

// GetByUsername récupère un utilisateur par son nom d'utilisateur
func (r *userRepositoryComplete) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	cacheKey := fmt.Sprintf("user:username:%s", username)
	var user entities.User
	if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
		return &user, nil
	}

	query := `
		SELECT id, uuid, username, email, password_hash, salt, first_name, last_name,
		       display_name, avatar, bio, role, status, email_verified, 
		       email_verification_token, two_factor_enabled, two_factor_secret,
		       is_online, last_seen, last_login_ip, login_attempts, locked_until,
		       timezone, language, theme, created_at, updated_at, deleted_at,
		       COALESCE(message_count, 0) as message_count,
		       COALESCE(stream_count, 0) as stream_count
		FROM users 
		LEFT JOIN user_stats USING(id)
		WHERE username = $1 AND deleted_at IS NULL
	`

	userPtr := &entities.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&userPtr.ID, &userPtr.UUID, &userPtr.Username, &userPtr.Email,
		&userPtr.PasswordHash, &userPtr.Salt, &userPtr.FirstName, &userPtr.LastName,
		&userPtr.DisplayName, &userPtr.Avatar, &userPtr.Bio, &userPtr.Role, &userPtr.Status,
		&userPtr.EmailVerified, &userPtr.EmailVerificationToken,
		&userPtr.TwoFactorEnabled, &userPtr.TwoFactorSecret,
		&userPtr.IsOnline, &userPtr.LastSeen, &userPtr.LastLoginIP,
		&userPtr.LoginAttempts, &userPtr.LockedUntil, &userPtr.Timezone,
		&userPtr.Language, &userPtr.Theme, &userPtr.CreatedAt, &userPtr.UpdatedAt,
		&userPtr.DeletedAt, &userPtr.MessageCount, &userPtr.StreamCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("récupération utilisateur par username: %w", err)
	}

	r.cache.Set(ctx, cacheKey, userPtr, 15*time.Minute)
	return userPtr, nil
}

// Update met à jour un utilisateur
func (r *userRepositoryComplete) Update(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE users SET
			username = $2, email = $3, first_name = $4, last_name = $5,
			display_name = $6, avatar = $7, bio = $8, role = $9, status = $10,
			email_verified = $11, two_factor_enabled = $12, is_online = $13,
			last_seen = $14, timezone = $15, language = $16, theme = $17,
			updated_at = $18
		WHERE id = $1 AND deleted_at IS NULL
	`

	user.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.FirstName, user.LastName,
		user.DisplayName, user.Avatar, user.Bio, string(user.Role), string(user.Status),
		user.EmailVerified, user.TwoFactorEnabled, user.IsOnline,
		user.LastSeen, user.Timezone, user.Language, user.Theme,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("mise à jour utilisateur: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	// Invalider le cache
	r.invalidateUserCaches(ctx, user.ID, user.Username, user.Email)
	return nil
}

// Delete supprime définitivement un utilisateur
func (r *userRepositoryComplete) Delete(ctx context.Context, id int64) error {
	// Récupérer les infos pour invalider le cache
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("utilisateur non trouvé")
	}

	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("suppression utilisateur: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification suppression: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	// Invalider le cache
	r.invalidateUserCaches(ctx, user.ID, user.Username, user.Email)
	return nil
}

// SoftDelete effectue une suppression logique
func (r *userRepositoryComplete) SoftDelete(ctx context.Context, id int64) error {
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("utilisateur non trouvé")
	}

	now := time.Now()
	query := `
		UPDATE users SET 
			deleted_at = $2, updated_at = $2, status = 'deleted'
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id, now)
	if err != nil {
		return fmt.Errorf("suppression logique utilisateur: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification suppression logique: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé ou déjà supprimé")
	}

	// Invalider le cache
	r.invalidateUserCaches(ctx, user.ID, user.Username, user.Email)
	return nil
}

// ============================================================================
// RECHERCHE ET LISTING
// ============================================================================

// List récupère une liste d'utilisateurs avec pagination
func (r *userRepositoryComplete) List(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, uuid, username, email, first_name, last_name,
		       display_name, avatar, bio, role, status, email_verified,
		       is_online, last_seen, timezone, language, theme,
		       created_at, updated_at
		FROM users 
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("liste utilisateurs: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		user := &entities.User{}
		err := rows.Scan(
			&user.ID, &user.UUID, &user.Username, &user.Email,
			&user.FirstName, &user.LastName, &user.DisplayName,
			&user.Avatar, &user.Bio, &user.Role, &user.Status,
			&user.EmailVerified, &user.IsOnline, &user.LastSeen,
			&user.Timezone, &user.Language, &user.Theme,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan utilisateur: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// Search recherche des utilisateurs par query
func (r *userRepositoryComplete) Search(ctx context.Context, query string, limit, offset int) ([]*entities.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	searchQuery := `
		SELECT id, uuid, username, email, first_name, last_name,
		       display_name, avatar, bio, role, status, email_verified,
		       is_online, last_seen, created_at, updated_at
		FROM users 
		WHERE deleted_at IS NULL 
		  AND (
		    username ILIKE '%' || $1 || '%' OR
		    email ILIKE '%' || $1 || '%' OR
		    first_name ILIKE '%' || $1 || '%' OR
		    last_name ILIKE '%' || $1 || '%' OR
		    display_name ILIKE '%' || $1 || '%'
		  )
		ORDER BY 
		  CASE 
		    WHEN username ILIKE $1 || '%' THEN 1
		    WHEN display_name ILIKE $1 || '%' THEN 2
		    ELSE 3
		  END,
		  username
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, searchQuery, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("recherche utilisateurs: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		user := &entities.User{}
		err := rows.Scan(
			&user.ID, &user.UUID, &user.Username, &user.Email,
			&user.FirstName, &user.LastName, &user.DisplayName,
			&user.Avatar, &user.Bio, &user.Role, &user.Status,
			&user.EmailVerified, &user.IsOnline, &user.LastSeen,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan utilisateur recherche: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// GetByRole récupère les utilisateurs par rôle
func (r *userRepositoryComplete) GetByRole(ctx context.Context, role entities.UserRole, limit, offset int) ([]*entities.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT id, uuid, username, email, first_name, last_name,
		       display_name, avatar, role, status, email_verified,
		       is_online, last_seen, created_at, updated_at
		FROM users 
		WHERE role = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, string(role), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("utilisateurs par rôle: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		user := &entities.User{}
		err := rows.Scan(
			&user.ID, &user.UUID, &user.Username, &user.Email,
			&user.FirstName, &user.LastName, &user.DisplayName,
			&user.Avatar, &user.Role, &user.Status, &user.EmailVerified,
			&user.IsOnline, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan utilisateur par rôle: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// GetByStatus récupère les utilisateurs par statut
func (r *userRepositoryComplete) GetByStatus(ctx context.Context, status entities.UserStatus, limit, offset int) ([]*entities.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT id, uuid, username, email, first_name, last_name,
		       display_name, avatar, role, status, email_verified,
		       is_online, last_seen, created_at, updated_at
		FROM users 
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, string(status), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("utilisateurs par statut: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		user := &entities.User{}
		err := rows.Scan(
			&user.ID, &user.UUID, &user.Username, &user.Email,
			&user.FirstName, &user.LastName, &user.DisplayName,
			&user.Avatar, &user.Role, &user.Status, &user.EmailVerified,
			&user.IsOnline, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan utilisateur par statut: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// ============================================================================
// AUTHENTIFICATION ET SÉCURITÉ
// ============================================================================

// UpdatePassword met à jour le mot de passe hashé
func (r *userRepositoryComplete) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	query := `
		UPDATE users SET 
			password_hash = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, userID, passwordHash, time.Now())
	if err != nil {
		return fmt.Errorf("mise à jour mot de passe: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour mot de passe: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	// Invalider le cache
	r.invalidateUserCache(ctx, userID)
	return nil
}

// UpdateLoginAttempts met à jour les tentatives de connexion
func (r *userRepositoryComplete) UpdateLoginAttempts(ctx context.Context, userID int64, attempts int, lockedUntil *time.Time) error {
	query := `
		UPDATE users SET 
			login_attempts = $2, locked_until = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, userID, attempts, lockedUntil, time.Now())
	if err != nil {
		return fmt.Errorf("mise à jour tentatives connexion: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour tentatives: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	r.invalidateUserCache(ctx, userID)
	return nil
}

// ResetLoginAttempts remet à zéro les tentatives de connexion
func (r *userRepositoryComplete) ResetLoginAttempts(ctx context.Context, userID int64) error {
	query := `
		UPDATE users SET 
			login_attempts = 0, locked_until = NULL, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("reset tentatives connexion: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification reset tentatives: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	r.invalidateUserCache(ctx, userID)
	return nil
}

// UpdateLastLogin met à jour la dernière connexion
func (r *userRepositoryComplete) UpdateLastLogin(ctx context.Context, userID int64, ipAddress string) error {
	query := `
		UPDATE users SET 
			last_login_ip = $2, last_seen = $3, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, userID, ipAddress, now)
	if err != nil {
		return fmt.Errorf("mise à jour dernière connexion: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour dernière connexion: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	r.invalidateUserCache(ctx, userID)
	return nil
}

// ============================================================================
// VÉRIFICATION D'EMAIL ET 2FA
// ============================================================================

// UpdateEmailVerification met à jour la vérification d'email
func (r *userRepositoryComplete) UpdateEmailVerification(ctx context.Context, userID int64, verified bool, token string) error {
	query := `
		UPDATE users SET 
			email_verified = $2, email_verification_token = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, userID, verified, token, time.Now())
	if err != nil {
		return fmt.Errorf("mise à jour vérification email: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour email: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	r.invalidateUserCache(ctx, userID)
	return nil
}

// UpdateTwoFactor met à jour les paramètres 2FA
func (r *userRepositoryComplete) UpdateTwoFactor(ctx context.Context, userID int64, enabled bool, secret string) error {
	query := `
		UPDATE users SET 
			two_factor_enabled = $2, two_factor_secret = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, userID, enabled, secret, time.Now())
	if err != nil {
		return fmt.Errorf("mise à jour 2FA: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour 2FA: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	r.invalidateUserCache(ctx, userID)
	return nil
}

// ============================================================================
// GESTION DES RÔLES ET STATUTS
// ============================================================================

// UpdateRole met à jour le rôle d'un utilisateur
func (r *userRepositoryComplete) UpdateRole(ctx context.Context, userID int64, role entities.UserRole) error {
	query := `
		UPDATE users SET 
			role = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, userID, string(role), time.Now())
	if err != nil {
		return fmt.Errorf("mise à jour rôle: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour rôle: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	r.invalidateUserCache(ctx, userID)
	return nil
}

// UpdateStatus met à jour le statut d'un utilisateur
func (r *userRepositoryComplete) UpdateStatus(ctx context.Context, userID int64, status entities.UserStatus) error {
	query := `
		UPDATE users SET 
			status = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, userID, string(status), time.Now())
	if err != nil {
		return fmt.Errorf("mise à jour statut: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour statut: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	r.invalidateUserCache(ctx, userID)
	return nil
}

// ============================================================================
// PRÉSENCE ET ACTIVITÉ
// ============================================================================

// UpdateOnlineStatus met à jour le statut en ligne
func (r *userRepositoryComplete) UpdateOnlineStatus(ctx context.Context, userID int64, isOnline bool) error {
	query := `
		UPDATE users SET 
			is_online = $2, last_seen = $3, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, userID, isOnline, now)
	if err != nil {
		return fmt.Errorf("mise à jour statut en ligne: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour statut en ligne: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	r.invalidateUserCache(ctx, userID)
	return nil
}

// GetOnlineUsers récupère les utilisateurs en ligne
func (r *userRepositoryComplete) GetOnlineUsers(ctx context.Context, limit int) ([]*entities.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT id, uuid, username, display_name, avatar, role, status,
		       is_online, last_seen, created_at
		FROM users 
		WHERE is_online = true AND deleted_at IS NULL
		ORDER BY last_seen DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("utilisateurs en ligne: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		user := &entities.User{}
		err := rows.Scan(
			&user.ID, &user.UUID, &user.Username, &user.DisplayName,
			&user.Avatar, &user.Role, &user.Status, &user.IsOnline,
			&user.LastSeen, &user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan utilisateur en ligne: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdateLastSeen met à jour la dernière activité
func (r *userRepositoryComplete) UpdateLastSeen(ctx context.Context, userID int64, lastSeen time.Time) error {
	query := `
		UPDATE users SET 
			last_seen = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, userID, lastSeen)
	if err != nil {
		return fmt.Errorf("mise à jour dernière activité: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour dernière activité: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	// Mise à jour cache sans invalidation complète
	return nil
}

// ============================================================================
// STATISTIQUES ET ANALYTICS
// ============================================================================

// GetUserStats récupère les statistiques d'un utilisateur
func (r *userRepositoryComplete) GetUserStats(ctx context.Context, userID int64) (*repositories.UserStats, error) {
	query := `
		SELECT 
			u.id as user_id,
			COALESCE(us.message_count, 0) as message_count,
			COALESCE(us.stream_count, 0) as stream_count,
			COALESCE(us.total_stream_time, 0) as total_stream_time,
			COALESCE(us.total_listen_time, 0) as total_listen_time,
			COALESCE(us.rooms_joined, 0) as rooms_joined,
			COALESCE(us.friends_count, 0) as friends_count,
			COALESCE(us.followers_count, 0) as followers_count,
			COALESCE(us.following_count, 0) as following_count,
			COALESCE(us.likes_received, 0) as likes_received,
			COALESCE(us.likes_given, 0) as likes_given,
			EXTRACT(EPOCH FROM (NOW() - u.created_at))/86400 as account_age,
			u.last_seen as last_active,
			COALESCE(us.updated_at, u.updated_at) as updated_at
		FROM users u
		LEFT JOIN user_stats us ON u.id = us.user_id
		WHERE u.id = $1 AND u.deleted_at IS NULL
	`

	stats := &repositories.UserStats{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.UserID, &stats.MessageCount, &stats.StreamCount,
		&stats.TotalStreamTime, &stats.TotalListenTime, &stats.RoomsJoined,
		&stats.FriendsCount, &stats.FollowersCount, &stats.FollowingCount,
		&stats.LikesReceived, &stats.LikesGiven, &stats.AccountAge,
		&stats.LastActive, &stats.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("récupération statistiques utilisateur: %w", err)
	}

	return stats, nil
}

// GetUserCount retourne le nombre total d'utilisateurs
func (r *userRepositoryComplete) GetUserCount(ctx context.Context) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("comptage utilisateurs: %w", err)
	}

	return count, nil
}

// GetUserCountByRole retourne le nombre d'utilisateurs par rôle
func (r *userRepositoryComplete) GetUserCountByRole(ctx context.Context, role entities.UserRole) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users WHERE role = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, query, string(role)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("comptage utilisateurs par rôle: %w", err)
	}

	return count, nil
}

// GetUserCountByStatus retourne le nombre d'utilisateurs par statut
func (r *userRepositoryComplete) GetUserCountByStatus(ctx context.Context, status entities.UserStatus) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM users WHERE status = $1 AND deleted_at IS NULL`

	err := r.db.QueryRowContext(ctx, query, string(status)).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("comptage utilisateurs par statut: %w", err)
	}

	return count, nil
}

// GetActiveUsersCount retourne le nombre d'utilisateurs actifs depuis une date
func (r *userRepositoryComplete) GetActiveUsersCount(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	query := `
		SELECT COUNT(*) 
		FROM users 
		WHERE last_seen >= $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("comptage utilisateurs actifs: %w", err)
	}

	return count, nil
}

// ============================================================================
// VALIDATION ET VÉRIFICATION D'UNICITÉ
// ============================================================================

// EmailExists vérifie si un email existe déjà
func (r *userRepositoryComplete) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`

	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("vérification existence email: %w", err)
	}

	return exists, nil
}

// UsernameExists vérifie si un nom d'utilisateur existe déjà
func (r *userRepositoryComplete) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND deleted_at IS NULL)`

	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("vérification existence username: %w", err)
	}

	return exists, nil
}

// EmailExistsExcludingUser vérifie si un email existe (excluant un utilisateur)
func (r *userRepositoryComplete) EmailExistsExcludingUser(ctx context.Context, email string, userID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2 AND deleted_at IS NULL)`

	err := r.db.QueryRowContext(ctx, query, email, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("vérification existence email excluant utilisateur: %w", err)
	}

	return exists, nil
}

// UsernameExistsExcludingUser vérifie si un username existe (excluant un utilisateur)
func (r *userRepositoryComplete) UsernameExistsExcludingUser(ctx context.Context, username string, userID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND id != $2 AND deleted_at IS NULL)`

	err := r.db.QueryRowContext(ctx, query, username, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("vérification existence username excluant utilisateur: %w", err)
	}

	return exists, nil
}

// ============================================================================
// HELPER FUNCTIONS (Cache Invalidation)
// ============================================================================

// invalidateUserCaches invalide tous les caches d'un utilisateur
func (r *userRepositoryComplete) invalidateUserCaches(ctx context.Context, userID int64, username, email string) {
	keys := []string{
		fmt.Sprintf("user:id:%d", userID),
		fmt.Sprintf("user:username:%s", username),
		fmt.Sprintf("user:email:%s", email),
	}

	for _, key := range keys {
		r.cache.Delete(ctx, key)
	}
}

// invalidateUserCache invalide le cache par ID
func (r *userRepositoryComplete) invalidateUserCache(ctx context.Context, userID int64) {
	r.cache.Delete(ctx, fmt.Sprintf("user:id:%d", userID))
}
