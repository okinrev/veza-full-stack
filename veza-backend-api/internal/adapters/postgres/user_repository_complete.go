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

// DeletePasswordResetToken supprime le token de réinitialisation de mot de passe
func (r *userRepositoryComplete) DeletePasswordResetToken(ctx context.Context, userID int64) error {
	query := `
		DELETE FROM password_reset_tokens 
		WHERE user_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("suppression token réinitialisation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification suppression token: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("token non trouvé")
	}

	return nil
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

// CreateUser alias pour Create
func (r *userRepositoryComplete) CreateUser(ctx context.Context, user *entities.User) error {
	return r.Create(ctx, user)
}

// GetUserByID alias pour GetByID
func (r *userRepositoryComplete) GetUserByID(ctx context.Context, id int64) (*entities.User, error) {
	return r.GetByID(ctx, id)
}

// GetUserByEmail alias pour GetByEmail
func (r *userRepositoryComplete) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	return r.GetByEmail(ctx, email)
}

// SetTwoFactorSecret définit le secret 2FA d'un utilisateur
func (r *userRepositoryComplete) SetTwoFactorSecret(ctx context.Context, userID int64, secret string) error {
	query := `
		UPDATE users 
		SET two_factor_secret = $2, updated_at = $3
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, secret, time.Now())
	if err != nil {
		return fmt.Errorf("mise à jour secret 2FA: %w", err)
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

// GetTwoFactorSecret récupère le secret 2FA d'un utilisateur
func (r *userRepositoryComplete) GetTwoFactorSecret(ctx context.Context, userID int64) (string, error) {
	query := `
		SELECT two_factor_secret 
		FROM users 
		WHERE id = $1
	`

	var secret string
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&secret)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("utilisateur non trouvé")
		}
		return "", fmt.Errorf("récupération secret 2FA: %w", err)
	}

	return secret, nil
}

// SetRecoveryCodes définit les codes de récupération 2FA
func (r *userRepositoryComplete) SetRecoveryCodes(ctx context.Context, userID int64, codes []string) error {
	query := `
		UPDATE users 
		SET two_factor_recovery_codes = $2, updated_at = $3
		WHERE id = $1
	`

	codesStr := fmt.Sprintf("%v", codes)
	result, err := r.db.ExecContext(ctx, query, userID, codesStr, time.Now())
	if err != nil {
		return fmt.Errorf("mise à jour codes de récupération: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification mise à jour codes: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	r.invalidateUserCache(ctx, userID)
	return nil
}

// CreateSession crée une nouvelle session utilisateur
func (r *userRepositoryComplete) CreateSession(ctx context.Context, session *repositories.UserSession) error {
	query := `
		INSERT INTO user_sessions (
			user_id, session_token, refresh_token, device_info, ip_address, 
			user_agent, location, is_active, last_activity, expires_at, created_at
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
		SELECT id, user_id, session_token, refresh_token, device_info, ip_address,
		       user_agent, location, is_active, last_activity, expires_at, created_at
		FROM user_sessions 
		WHERE session_token = $1 AND is_active = true
	`

	session := &repositories.UserSession{}
	err := r.db.QueryRowContext(ctx, query, sessionToken).Scan(
		&session.ID, &session.UserID, &session.SessionToken, &session.RefreshToken,
		&session.DeviceInfo, &session.IPAddress, &session.UserAgent,
		&session.Location, &session.IsActive, &session.LastActivity,
		&session.ExpiresAt, &session.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("récupération session: %w", err)
	}

	return session, nil
}

// UpdateSession met à jour une session
func (r *userRepositoryComplete) UpdateSession(ctx context.Context, sessionToken string, lastActivity time.Time) error {
	query := `
		UPDATE user_sessions 
		SET last_activity = $2, updated_at = $3
		WHERE session_token = $1
	`

	result, err := r.db.ExecContext(ctx, query, sessionToken, lastActivity, time.Now())
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
		UPDATE user_sessions 
		SET is_active = false, updated_at = $2
		WHERE session_token = $1
	`

	result, err := r.db.ExecContext(ctx, query, sessionToken, time.Now())
	if err != nil {
		return fmt.Errorf("invalidation session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification invalidation: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("session non trouvée")
	}

	return nil
}

// InvalidateAllUserSessions invalide toutes les sessions d'un utilisateur
func (r *userRepositoryComplete) InvalidateAllUserSessions(ctx context.Context, userID int64) error {
	query := `
		UPDATE user_sessions 
		SET is_active = false, updated_at = $2
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("invalidation toutes sessions: %w", err)
	}

	return nil
}

// GetUserSessions récupère toutes les sessions d'un utilisateur
func (r *userRepositoryComplete) GetUserSessions(ctx context.Context, userID int64) ([]*repositories.UserSession, error) {
	query := `
		SELECT id, user_id, session_token, refresh_token, device_info, ip_address,
		       user_agent, location, is_active, last_activity, expires_at, created_at
		FROM user_sessions 
		WHERE user_id = $1
		ORDER BY created_at DESC
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
			&session.ID, &session.UserID, &session.SessionToken, &session.RefreshToken,
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

// RevokeRefreshToken révoque un token de rafraîchissement
func (r *userRepositoryComplete) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	query := `
		UPDATE user_sessions 
		SET is_active = false, updated_at = $2
		WHERE refresh_token = $1
	`

	result, err := r.db.ExecContext(ctx, query, refreshToken, time.Now())
	if err != nil {
		return fmt.Errorf("révocation token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification révocation: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("token non trouvé")
	}

	return nil
}

// RevokeAllUserTokens révoque tous les tokens d'un utilisateur
func (r *userRepositoryComplete) RevokeAllUserTokens(ctx context.Context, userID int64) error {
	query := `
		UPDATE user_sessions 
		SET is_active = false, updated_at = $2
		WHERE user_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("révocation tous tokens: %w", err)
	}

	return nil
}

// GetUserPermissions récupère les permissions d'un utilisateur
func (r *userRepositoryComplete) GetUserPermissions(ctx context.Context, userID int64) ([]repositories.UserPermission, error) {
	query := `
		SELECT id, user_id, resource, action, granted_at, granted_by, expires_at
		FROM user_permissions 
		WHERE user_id = $1 AND (expires_at IS NULL OR expires_at > NOW())
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("récupération permissions: %w", err)
	}
	defer rows.Close()

	var permissions []repositories.UserPermission
	for rows.Next() {
		permission := repositories.UserPermission{}
		err := rows.Scan(
			&permission.ID, &permission.UserID, &permission.Resource,
			&permission.Action, &permission.GrantedAt, &permission.GrantedBy,
			&permission.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// GrantUserPermission accorde une permission à un utilisateur
func (r *userRepositoryComplete) GrantUserPermission(ctx context.Context, userID int64, permission repositories.UserPermission) error {
	query := `
		INSERT INTO user_permissions (
			user_id, resource, action, granted_at, granted_by, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		userID, permission.Resource, permission.Action,
		permission.GrantedAt, permission.GrantedBy, permission.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("octroi permission: %w", err)
	}

	return nil
}

// RevokeUserPermission révoque une permission d'un utilisateur
func (r *userRepositoryComplete) RevokeUserPermission(ctx context.Context, userID int64, permission repositories.UserPermission) error {
	query := `
		DELETE FROM user_permissions 
		WHERE user_id = $1 AND resource = $2 AND action = $3
	`

	result, err := r.db.ExecContext(ctx, query, userID, permission.Resource, permission.Action)
	if err != nil {
		return fmt.Errorf("révocation permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification révocation: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission non trouvée")
	}

	return nil
}

// IsRoomMember vérifie si un utilisateur est membre d'une room
func (r *userRepositoryComplete) IsRoomMember(ctx context.Context, roomID, userID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM room_members 
			WHERE room_id = $1 AND user_id = $2 AND status = 'active'
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, roomID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("vérification membre room: %w", err)
	}

	return exists, nil
}

// UpdatePreferences met à jour les préférences d'un utilisateur
func (r *userRepositoryComplete) UpdatePreferences(ctx context.Context, userID int64, preferences *repositories.UserPreferences) error {
	query := `
		INSERT INTO user_preferences (
			user_id, theme, language, timezone, notification_settings, 
			privacy_settings, audio_settings, chat_settings, updated_at
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

	notificationSettings := fmt.Sprintf("%v", preferences.NotificationSettings)
	privacySettings := fmt.Sprintf("%v", preferences.PrivacySettings)
	audioSettings := fmt.Sprintf("%v", preferences.AudioSettings)
	chatSettings := fmt.Sprintf("%v", preferences.ChatSettings)

	_, err := r.db.ExecContext(ctx, query,
		userID, preferences.Theme, preferences.Language, preferences.Timezone,
		notificationSettings, privacySettings, audioSettings, chatSettings,
		preferences.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("mise à jour préférences: %w", err)
	}

	return nil
}

// GetPreferences récupère les préférences d'un utilisateur
func (r *userRepositoryComplete) GetPreferences(ctx context.Context, userID int64) (*repositories.UserPreferences, error) {
	query := `
		SELECT user_id, theme, language, timezone, notification_settings,
		       privacy_settings, audio_settings, chat_settings, updated_at
		FROM user_preferences 
		WHERE user_id = $1
	`

	preferences := &repositories.UserPreferences{}
	var notificationSettings, privacySettings, audioSettings, chatSettings string

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&preferences.UserID, &preferences.Theme, &preferences.Language,
		&preferences.Timezone, &notificationSettings, &privacySettings,
		&audioSettings, &chatSettings, &preferences.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return &repositories.UserPreferences{
				UserID:    userID,
				Theme:     "light",
				Language:  "en",
				Timezone:  "UTC",
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("récupération préférences: %w", err)
	}

	return preferences, nil
}

// CreateAuditLog crée un log d'audit
func (r *userRepositoryComplete) CreateAuditLog(ctx context.Context, log *repositories.UserAuditLog) error {
	query := `
		INSERT INTO user_audit_logs (
			user_id, action, resource, resource_id, details, ip_address, 
			user_agent, success, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		log.UserID, log.Action, log.Resource, log.ResourceID,
		log.Details, log.IPAddress, log.UserAgent, log.Success,
		log.ErrorMessage, log.CreatedAt,
	).Scan(&log.ID)

	if err != nil {
		return fmt.Errorf("création log audit: %w", err)
	}

	return nil
}

// GetUserAuditLogs récupère les logs d'audit d'un utilisateur
func (r *userRepositoryComplete) GetUserAuditLogs(ctx context.Context, userID int64, limit, offset int) ([]*repositories.UserAuditLog, error) {
	query := `
		SELECT id, user_id, action, resource, resource_id, details, ip_address,
		       user_agent, success, error_message, created_at
		FROM user_audit_logs 
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("récupération logs audit: %w", err)
	}
	defer rows.Close()

	var logs []*repositories.UserAuditLog
	for rows.Next() {
		log := &repositories.UserAuditLog{}
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Action, &log.Resource, &log.ResourceID,
			&log.Details, &log.IPAddress, &log.UserAgent, &log.Success,
			&log.ErrorMessage, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan log audit: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// AddContact ajoute un contact
func (r *userRepositoryComplete) AddContact(ctx context.Context, userID, contactID int64) error {
	query := `
		INSERT INTO user_contacts (user_id, contact_id, status, created_at)
		VALUES ($1, $2, 'pending', $3)
		ON CONFLICT (user_id, contact_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, userID, contactID, time.Now())
	if err != nil {
		return fmt.Errorf("ajout contact: %w", err)
	}

	return nil
}

// RemoveContact retire un contact
func (r *userRepositoryComplete) RemoveContact(ctx context.Context, userID, contactID int64) error {
	query := `
		DELETE FROM user_contacts 
		WHERE user_id = $1 AND contact_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, userID, contactID)
	if err != nil {
		return fmt.Errorf("suppression contact: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification suppression: %w", err)
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
		       u.display_name, u.avatar, u.bio, u.role, u.status, u.is_online,
		       u.last_seen, u.created_at, u.updated_at
		FROM users u
		JOIN user_contacts uc ON u.id = uc.contact_id
		WHERE uc.user_id = $1 AND uc.status = 'accepted' AND u.deleted_at IS NULL
		ORDER BY u.display_name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("récupération contacts: %w", err)
	}
	defer rows.Close()

	var contacts []*entities.User
	for rows.Next() {
		contact := &entities.User{}
		err := rows.Scan(
			&contact.ID, &contact.UUID, &contact.Username, &contact.Email,
			&contact.FirstName, &contact.LastName, &contact.DisplayName,
			&contact.Avatar, &contact.Bio, &contact.Role, &contact.Status,
			&contact.IsOnline, &contact.LastSeen, &contact.CreatedAt, &contact.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan contact: %w", err)
		}
		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// BlockUser bloque un utilisateur
func (r *userRepositoryComplete) BlockUser(ctx context.Context, userID, blockedUserID int64) error {
	query := `
		INSERT INTO user_blocks (user_id, blocked_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, blocked_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, userID, blockedUserID, time.Now())
	if err != nil {
		return fmt.Errorf("blocage utilisateur: %w", err)
	}

	return nil
}

// UnblockUser débloque un utilisateur
func (r *userRepositoryComplete) UnblockUser(ctx context.Context, userID, blockedUserID int64) error {
	query := `
		DELETE FROM user_blocks 
		WHERE user_id = $1 AND blocked_id = $2
	`

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
		SELECT u.id, u.uuid, u.username, u.email, u.first_name, u.last_name,
		       u.display_name, u.avatar, u.bio, u.role, u.status, u.is_online,
		       u.last_seen, u.created_at, u.updated_at
		FROM users u
		JOIN user_blocks ub ON u.id = ub.blocked_id
		WHERE ub.user_id = $1 AND u.deleted_at IS NULL
		ORDER BY u.display_name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("récupération utilisateurs bloqués: %w", err)
	}
	defer rows.Close()

	var blockedUsers []*entities.User
	for rows.Next() {
		user := &entities.User{}
		err := rows.Scan(
			&user.ID, &user.UUID, &user.Username, &user.Email,
			&user.FirstName, &user.LastName, &user.DisplayName,
			&user.Avatar, &user.Bio, &user.Role, &user.Status,
			&user.IsOnline, &user.LastSeen, &user.CreatedAt, &user.UpdatedAt,
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
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_blocks 
			WHERE user_id = $1 AND blocked_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, otherUserID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("vérification blocage: %w", err)
	}

	return exists, nil
}

// ============================================================================
// MÉTHODES 2FA MANQUANTES
// ============================================================================

// EnableTwoFactor active l'authentification à deux facteurs
func (r *userRepositoryComplete) EnableTwoFactor(ctx context.Context, userID int64) error {
	query := `
		UPDATE users 
		SET two_factor_enabled = true, updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("activation 2FA: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification activation 2FA: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	r.invalidateUserCache(ctx, userID)
	return nil
}

// DisableTwoFactor désactive l'authentification à deux facteurs
func (r *userRepositoryComplete) DisableTwoFactor(ctx context.Context, userID int64) error {
	query := `
		UPDATE users 
		SET two_factor_enabled = false, two_factor_secret = '', updated_at = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, userID, time.Now())
	if err != nil {
		return fmt.Errorf("désactivation 2FA: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification désactivation 2FA: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	r.invalidateUserCache(ctx, userID)
	return nil
}

// ============================================================================
// ALIAS POUR COMPATIBILITÉ
// ============================================================================

// UpdateUser alias pour Update
func (r *userRepositoryComplete) UpdateUser(ctx context.Context, user *entities.User) error {
	return r.Update(ctx, user)
}

// ============================================================================
// TOKENS DE RÉINITIALISATION DE MOT DE PASSE
// ============================================================================

// CreatePasswordResetToken crée un token de réinitialisation de mot de passe
func (r *userRepositoryComplete) CreatePasswordResetToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO password_reset_tokens (
			user_id, token, expires_at, created_at
		) VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			token = EXCLUDED.token,
			expires_at = EXCLUDED.expires_at,
			created_at = EXCLUDED.created_at
	`

	_, err := r.db.ExecContext(ctx, query, userID, token, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("création token réinitialisation: %w", err)
	}

	return nil
}

// GetUserByPasswordResetToken récupère un utilisateur par son token de réinitialisation
func (r *userRepositoryComplete) GetUserByPasswordResetToken(ctx context.Context, token string) (*entities.User, error) {
	query := `
		SELECT u.id, u.uuid, u.username, u.email, u.password_hash, u.salt, u.first_name, u.last_name,
		       u.display_name, u.avatar, u.bio, u.role, u.status, u.email_verified, 
		       u.email_verification_token, u.two_factor_enabled, u.two_factor_secret,
		       u.is_online, u.last_seen, u.last_login_ip, u.login_attempts, u.locked_until,
		       u.timezone, u.language, u.theme, u.created_at, u.updated_at, u.deleted_at,
		       COALESCE(u.message_count, 0) as message_count,
		       COALESCE(u.stream_count, 0) as stream_count
		FROM users u
		JOIN password_reset_tokens prt ON u.id = prt.user_id
		WHERE prt.token = $1 AND prt.expires_at > NOW() AND u.deleted_at IS NULL
	`

	user := &entities.User{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
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
		return nil, fmt.Errorf("récupération utilisateur par token: %w", err)
	}

	return user, nil
}
