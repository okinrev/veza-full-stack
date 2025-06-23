// internal/api/user/service.go
package user

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/okinrev/veza-web-app/internal/utils"
	"github.com/okinrev/veza-web-app/internal/database"
)

// Service handles user business logic
type Service struct {
	db *database.DB
}

// NewService creates a new user service
func NewService(db *database.DB) *Service {
	return &Service{
		db: db,
	}
}

// GetUsers retrieves users with pagination and optional search
func (s *Service) GetUsers(page, limit int, search string) ([]UserResponse, int, error) {
	offset := (page - 1) * limit
	
	// Build the query with optional search
	baseQuery := `
		SELECT id, email, first_name, last_name, username, avatar, bio, 
			   role, is_active, is_verified, last_login_at, created_at, updated_at
		FROM users
	`
	countQuery := "SELECT COUNT(*) FROM users"
	
	var whereClause string
	var args []interface{}
	argIndex := 1
	
	if search != "" {
		whereClause = ` WHERE (
			email ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			first_name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			last_name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR 
			username ILIKE $` + fmt.Sprintf("%d", argIndex) + `
		)`
		args = append(args, "%"+search+"%")
		argIndex++
	}
	
	// Get total count
	var total int
	err := s.db.QueryRow(countQuery + whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}
	
	// Get users
	orderClause := " ORDER BY created_at DESC"
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)
	
	query := baseQuery + whereClause + orderClause + limitClause
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()
	
	var users []UserResponse
	for rows.Next() {
		var user UserResponse
		err := rows.Scan(
			&user.ID, &user.Email, &user.FirstName, &user.LastName,
			&user.Username, &user.Avatar, &user.Bio, &user.Role,
			&user.IsActive, &user.IsVerified, &user.LastLoginAt,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	
	return users, total, nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(userID int) (*UserResponse, error) {
	query := `
		SELECT id, email, first_name, last_name, username, avatar, bio,
			   role, is_active, is_verified, last_login_at, created_at, updated_at
		FROM users 
		WHERE id = $1 AND is_active = true
	`
	
	var user UserResponse
	err := s.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName,
		&user.Username, &user.Avatar, &user.Bio, &user.Role,
		&user.IsActive, &user.IsVerified, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

// GetUserByEmail retrieves a user by email (includes password hash for auth)
func (s *Service) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, username, 
			   avatar, bio, role, is_active, is_verified, last_login_at, 
			   created_at, updated_at
		FROM users 
		WHERE email = $1
	`
	
	var user User
	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName,
		&user.LastName, &user.Username, &user.Avatar, &user.Bio,
		&user.Role, &user.IsActive, &user.IsVerified, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

// CreateUser creates a new user
func (s *Service) CreateUser(req CreateUserRequest) (*UserResponse, error) {
	// Hash the password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = "user"
	}
	
	query := `
		INSERT INTO users (email, password_hash, first_name, last_name, username, role, is_active, is_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, true, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, email, first_name, last_name, username, role, is_active, is_verified, created_at, updated_at
	`
	
	var user UserResponse
	err = s.db.QueryRow(
		query, req.Email, passwordHash, req.FirstName, req.LastName, 
		req.Username, role,
	).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName,
		&user.Username, &user.Role, &user.IsActive, &user.IsVerified,
		&user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("email already exists")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return &user, nil
}

// UpdateUser updates an existing user
func (s *Service) UpdateUser(userID int, req UpdateUserRequest) (*UserResponse, error) {
	// Build dynamic update query
	setParts := []string{"updated_at = CURRENT_TIMESTAMP"}
	args := []interface{}{}
	argIndex := 1
	
	if req.FirstName != nil {
		setParts = append(setParts, fmt.Sprintf("first_name = $%d", argIndex))
		args = append(args, req.FirstName)
		argIndex++
	}
	
	if req.LastName != nil {
		setParts = append(setParts, fmt.Sprintf("last_name = $%d", argIndex))
		args = append(args, req.LastName)
		argIndex++
	}
	
	if req.Username != nil {
		setParts = append(setParts, fmt.Sprintf("username = $%d", argIndex))
		args = append(args, req.Username)
		argIndex++
	}
	
	if req.Avatar != nil {
		setParts = append(setParts, fmt.Sprintf("avatar = $%d", argIndex))
		args = append(args, req.Avatar)
		argIndex++
	}
	
	if req.Bio != nil {
		setParts = append(setParts, fmt.Sprintf("bio = $%d", argIndex))
		args = append(args, req.Bio)
		argIndex++
	}
	
	if req.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, req.IsActive)
		argIndex++
	}
	
	if req.IsVerified != nil {
		setParts = append(setParts, fmt.Sprintf("is_verified = $%d", argIndex))
		args = append(args, req.IsVerified)
		argIndex++
	}
	
	if req.Role != nil {
		setParts = append(setParts, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, req.Role)
		argIndex++
	}
	
	// Add user ID as the last argument
	args = append(args, userID)
	
	query := fmt.Sprintf(`
		UPDATE users 
		SET %s
		WHERE id = $%d
		RETURNING id, email, first_name, last_name, username, avatar, bio,
				  role, is_active, is_verified, last_login_at, created_at, updated_at
	`, strings.Join(setParts, ", "), argIndex)
	
	var user UserResponse
	err := s.db.QueryRow(query, args...).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName,
		&user.Username, &user.Avatar, &user.Bio, &user.Role,
		&user.IsActive, &user.IsVerified, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	
	return &user, nil
}

// DeleteUser soft deletes a user (sets is_active to false)
func (s *Service) DeleteUser(userID int) error {
	query := `
		UPDATE users 
		SET is_active = false, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND is_active = true
	`
	
	result, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	return nil
}

// UpdateLastLogin updates the user's last login timestamp
func (s *Service) UpdateLastLogin(userID int) error {
	query := `
		UPDATE users 
		SET last_login_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	
	_, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	
	return nil
}

// ChangePassword updates a user's password
func (s *Service) ChangePassword(userID int, currentPassword, newPassword string) error {
	// First, get the current password hash
	var currentHash string
	err := s.db.QueryRow("SELECT password_hash FROM users WHERE id = $1", userID).Scan(&currentHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to get user password: %w", err)
	}
	
	// Verify current password
	if err := utils.CheckPasswordHash(currentPassword, currentHash); err != nil {
		return fmt.Errorf("current password is incorrect")
	}
	
	// Hash new password
	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}
	
	// Update password
	query := `
		UPDATE users 
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	
	_, err = s.db.Exec(query, newHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	
	return nil
}

// GetUserStats returns basic user statistics
func (s *Service) GetUserStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total users
	var totalUsers int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_active = true").Scan(&totalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get total users: %w", err)
	}
	stats["total_users"] = totalUsers
	
	// Verified users
	var verifiedUsers int
	err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_active = true AND is_verified = true").Scan(&verifiedUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get verified users: %w", err)
	}
	stats["verified_users"] = verifiedUsers
	
	// Active users (logged in within last 30 days)
	var activeUsers int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE is_active = true AND last_login_at > CURRENT_TIMESTAMP - INTERVAL '30 days'
	`).Scan(&activeUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	stats["active_users"] = activeUsers
	
	// New users this month
	var newUsersThisMonth int
	err = s.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE is_active = true AND created_at >= date_trunc('month', CURRENT_TIMESTAMP)
	`).Scan(&newUsersThisMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get new users this month: %w", err)
	}
	stats["new_users_this_month"] = newUsersThisMonth
	
	return stats, nil
}
