package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
	"github.com/okinrev/veza-web-app/internal/domain/repositories"
)

// userRepository implémentation PostgreSQL du UserRepository
type userRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewUserRepository crée une nouvelle instance du repository utilisateur
func NewUserRepository(db *sql.DB, logger *zap.Logger) (repositories.UserRepository, error) {
	return &userRepository{
		db:     db,
		logger: logger,
	}, nil
}

// Create crée un nouvel utilisateur
func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, first_name, last_name, bio, avatar, role, status, is_active, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	var createdAt, updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		user.Username,
		user.Email,
		user.Password,
		nullStringToPtr(user.FirstName),
		nullStringToPtr(user.LastName),
		nullStringToPtr(user.Bio),
		nullStringToPtr(user.Avatar),
		string(user.Role),
		string(user.Status),
		user.IsActive,
		user.IsVerified,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &createdAt, &updatedAt)

	if err != nil {
		r.logger.Error("Erreur création utilisateur", zap.Error(err))
		return fmt.Errorf("erreur création utilisateur: %w", err)
	}

	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	return nil
}

// GetByID récupère un utilisateur par son ID
func (r *userRepository) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, bio, avatar, 
		       role, status, is_active, is_verified, last_login_at, created_at, updated_at
		FROM users 
		WHERE id = $1
	`

	user := &entities.User{}
	var firstName, lastName, bio, avatar sql.NullString
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&firstName,
		&lastName,
		&bio,
		&avatar,
		&user.Role,
		&user.Status,
		&user.IsActive,
		&user.IsVerified,
		&lastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Erreur récupération utilisateur par ID", zap.Error(err), zap.Int64("user_id", id))
		return nil, fmt.Errorf("erreur récupération utilisateur: %w", err)
	}

	// Convertir les NullString
	user.FirstName = nullStringFromPtr(firstName)
	user.LastName = nullStringFromPtr(lastName)
	user.Bio = nullStringFromPtr(bio)
	user.Avatar = nullStringFromPtr(avatar)
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// GetByUsername récupère un utilisateur par son nom d'utilisateur
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, bio, avatar, 
		       role, status, is_active, is_verified, last_login_at, created_at, updated_at
		FROM users 
		WHERE username = $1
	`

	user := &entities.User{}
	var firstName, lastName, bio, avatar sql.NullString
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&firstName,
		&lastName,
		&bio,
		&avatar,
		&user.Role,
		&user.Status,
		&user.IsActive,
		&user.IsVerified,
		&lastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Erreur récupération utilisateur par username", zap.Error(err), zap.String("username", username))
		return nil, fmt.Errorf("erreur récupération utilisateur: %w", err)
	}

	// Convertir les NullString
	user.FirstName = nullStringFromPtr(firstName)
	user.LastName = nullStringFromPtr(lastName)
	user.Bio = nullStringFromPtr(bio)
	user.Avatar = nullStringFromPtr(avatar)
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// GetByEmail récupère un utilisateur par son email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, bio, avatar, 
		       role, status, is_active, is_verified, last_login_at, created_at, updated_at
		FROM users 
		WHERE email = $1
	`

	user := &entities.User{}
	var firstName, lastName, bio, avatar sql.NullString
	var lastLoginAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&firstName,
		&lastName,
		&bio,
		&avatar,
		&user.Role,
		&user.Status,
		&user.IsActive,
		&user.IsVerified,
		&lastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Erreur récupération utilisateur par email", zap.Error(err), zap.String("email", email))
		return nil, fmt.Errorf("erreur récupération utilisateur: %w", err)
	}

	// Convertir les NullString
	user.FirstName = nullStringFromPtr(firstName)
	user.LastName = nullStringFromPtr(lastName)
	user.Bio = nullStringFromPtr(bio)
	user.Avatar = nullStringFromPtr(avatar)
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	return user, nil
}

// Update met à jour un utilisateur
func (r *userRepository) Update(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE users 
		SET username = $2, email = $3, password_hash = $4, first_name = $5, last_name = $6, 
		    bio = $7, avatar = $8, role = $9, status = $10, is_active = $11, is_verified = $12, 
		    last_login_at = $13, updated_at = $14
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		nullStringToPtr(user.FirstName),
		nullStringToPtr(user.LastName),
		nullStringToPtr(user.Bio),
		nullStringToPtr(user.Avatar),
		string(user.Role),
		string(user.Status),
		user.IsActive,
		user.IsVerified,
		user.LastLoginAt,
		user.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Erreur mise à jour utilisateur", zap.Error(err), zap.Int64("user_id", user.ID))
		return fmt.Errorf("erreur mise à jour utilisateur: %w", err)
	}

	return nil
}

// Delete supprime un utilisateur
func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Erreur suppression utilisateur", zap.Error(err), zap.Int64("user_id", id))
		return fmt.Errorf("erreur suppression utilisateur: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur non trouvé")
	}

	return nil
}

// List liste les utilisateurs avec filtres
func (r *userRepository) List(ctx context.Context, filters repositories.UserFilters) ([]*entities.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, bio, avatar, 
		       role, status, is_active, is_verified, last_login_at, created_at, updated_at
		FROM users 
	`

	var whereConditions []string
	var args []interface{}
	argIndex := 1

	// Construire les conditions WHERE
	if filters.Role != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, string(*filters.Role))
		argIndex++
	}

	if filters.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, string(*filters.Status))
		argIndex++
	}

	if filters.IsActive != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *filters.IsActive)
		argIndex++
	}

	if filters.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(username ILIKE $%d OR email ILIKE $%d OR first_name ILIKE $%d OR last_name ILIKE $%d)", argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+filters.Search+"%")
		argIndex++
	}

	// Ajouter les conditions WHERE
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Tri
	orderBy := "created_at"
	if filters.OrderBy != "" {
		orderBy = filters.OrderBy
	}

	order := "DESC"
	if filters.Order == "asc" {
		order = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, order)

	// Pagination
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filters.Limit)
		argIndex++

		if filters.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, filters.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Erreur liste utilisateurs", zap.Error(err))
		return nil, fmt.Errorf("erreur liste utilisateurs: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		user := &entities.User{}
		var firstName, lastName, bio, avatar sql.NullString
		var lastLoginAt sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&firstName,
			&lastName,
			&bio,
			&avatar,
			&user.Role,
			&user.Status,
			&user.IsActive,
			&user.IsVerified,
			&lastLoginAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			r.logger.Error("Erreur scan utilisateur", zap.Error(err))
			return nil, fmt.Errorf("erreur scan utilisateur: %w", err)
		}

		// Convertir les NullString
		user.FirstName = nullStringFromPtr(firstName)
		user.LastName = nullStringFromPtr(lastName)
		user.Bio = nullStringFromPtr(bio)
		user.Avatar = nullStringFromPtr(avatar)
		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}

		users = append(users, user)
	}

	return users, nil
}

// Count compte les utilisateurs avec filtres
func (r *userRepository) Count(ctx context.Context, filters repositories.UserFilters) (int64, error) {
	query := "SELECT COUNT(*) FROM users"

	var whereConditions []string
	var args []interface{}
	argIndex := 1

	// Même logique que List pour les filtres
	if filters.Role != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, string(*filters.Role))
		argIndex++
	}

	if filters.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, string(*filters.Status))
		argIndex++
	}

	if filters.IsActive != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *filters.IsActive)
		argIndex++
	}

	if filters.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(username ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filters.Search+"%")
		argIndex++
	}

	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		r.logger.Error("Erreur count utilisateurs", zap.Error(err))
		return 0, fmt.Errorf("erreur count utilisateurs: %w", err)
	}

	return count, nil
}

// Search recherche des utilisateurs
func (r *userRepository) Search(ctx context.Context, query string, limit int) ([]*entities.User, error) {
	filters := repositories.UserFilters{
		Search: query,
		Limit:  limit,
	}
	return r.List(ctx, filters)
}

// ExistsByUsername vérifie si un utilisateur existe par username
func (r *userRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"

	var exists bool
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		r.logger.Error("Erreur vérification existence username", zap.Error(err))
		return false, fmt.Errorf("erreur vérification existence: %w", err)
	}

	return exists, nil
}

// ExistsByEmail vérifie si un utilisateur existe par email
func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"

	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		r.logger.Error("Erreur vérification existence email", zap.Error(err))
		return false, fmt.Errorf("erreur vérification existence: %w", err)
	}

	return exists, nil
}

// SaveRefreshToken sauvegarde un refresh token
func (r *userRepository) SaveRefreshToken(ctx context.Context, userID int64, token string, expiresAt int64) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at
	`

	_, err := r.db.ExecContext(ctx, query, userID, token, expiresAt)
	if err != nil {
		r.logger.Error("Erreur sauvegarde refresh token", zap.Error(err))
		return fmt.Errorf("erreur sauvegarde refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken récupère un refresh token
func (r *userRepository) GetRefreshToken(ctx context.Context, token string) (*repositories.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens 
		WHERE token = $1
	`

	refreshToken := &repositories.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.Token,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error("Erreur récupération refresh token", zap.Error(err))
		return nil, fmt.Errorf("erreur récupération refresh token: %w", err)
	}

	return refreshToken, nil
}

// RevokeRefreshToken révoque un refresh token
func (r *userRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	query := "DELETE FROM refresh_tokens WHERE token = $1"

	_, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		r.logger.Error("Erreur révocation refresh token", zap.Error(err))
		return fmt.Errorf("erreur révocation refresh token: %w", err)
	}

	return nil
}

// RevokeAllUserTokens révoque tous les tokens d'un utilisateur
func (r *userRepository) RevokeAllUserTokens(ctx context.Context, userID int64) error {
	query := "DELETE FROM refresh_tokens WHERE user_id = $1"

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("Erreur révocation tokens utilisateur", zap.Error(err))
		return fmt.Errorf("erreur révocation tokens utilisateur: %w", err)
	}

	return nil
}

// GetUserStats récupère les statistiques d'un utilisateur
func (r *userRepository) GetUserStats(ctx context.Context, userID int64) (*repositories.UserStats, error) {
	// Pour la Phase 1, retourner des stats de base
	// En Phase 4-5, on ajoutera les vraies stats avec rooms, messages, tracks
	return &repositories.UserStats{
		UserID:        userID,
		TotalRooms:    0,
		TotalMessages: 0,
		TotalTracks:   0,
		TotalListings: 0,
		LastActivity:  time.Now().Unix(),
	}, nil
}

// GetTotalUsers retourne le nombre total d'utilisateurs
func (r *userRepository) GetTotalUsers(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM users"

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur count total utilisateurs: %w", err)
	}

	return count, nil
}

// GetActiveUsers retourne le nombre d'utilisateurs actifs
func (r *userRepository) GetActiveUsers(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM users WHERE is_active = true AND status = 'active'"

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur count utilisateurs actifs: %w", err)
	}

	return count, nil
}

// GetNewUsersToday retourne le nombre de nouveaux utilisateurs aujourd'hui
func (r *userRepository) GetNewUsersToday(ctx context.Context) (int64, error) {
	query := "SELECT COUNT(*) FROM users WHERE DATE(created_at) = CURRENT_DATE"

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur count nouveaux utilisateurs: %w", err)
	}

	return count, nil
}

// Fonctions utilitaires pour gérer les NullString
func nullStringToPtr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullStringFromPtr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
