// internal/services/user_service.go
package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/models"
	"github.com/okinrev/veza-web-app/internal/utils"
)

type UserService interface {
	GetMe(userID int) (*models.User, error)
	UpdateMe(userID int, req UpdateUserRequest) (*models.User, error)
	ChangePassword(userID int, req ChangePasswordRequest) error
	GetUsers(page, limit int, search string) ([]models.User, int, error)
	GetUsersExceptMe(userID, limit int, search string) ([]models.User, error)
	SearchUsers(query string, limit int) ([]models.User, error)
	GetUserByID(userID int) (*models.User, error)
	GetUserAvatar(userID int) (*string, error)
	UpdateUserAvatar(userID int, avatarURL string) error
	DeleteUserAvatar(userID int) error
}

type userService struct {
	db *database.DB
}

func NewUserService(db *database.DB) UserService {
	return &userService{db: db}
}

// Request types
type UpdateUserRequest struct {
	Username  *string `json:"username,omitempty"`
	Email     *string `json:"email,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Bio       *string `json:"bio,omitempty"`
	Avatar    *string `json:"avatar,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// GetMe returns the current user's profile
func (s *userService) GetMe(userID int) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, username, email, first_name, last_name, bio, avatar, role, created_at, updated_at
		FROM users WHERE id = $1 AND role != 'deleted'
	`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, 
		&user.LastName, &user.Bio, &user.Avatar, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

// UpdateMe updates the current user's profile
func (s *userService) UpdateMe(userID int, req UpdateUserRequest) (*models.User, error) {
	// Build dynamic update query
	setParts := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argCount := 1

	if req.Username != nil {
		setParts = append(setParts, "username = $"+strconv.Itoa(argCount))
		args = append(args, strings.TrimSpace(*req.Username))
		argCount++
	}
	if req.Email != nil {
		setParts = append(setParts, "email = $"+strconv.Itoa(argCount))
		args = append(args, strings.TrimSpace(strings.ToLower(*req.Email)))
		argCount++
	}
	if req.FirstName != nil {
		setParts = append(setParts, "first_name = $"+strconv.Itoa(argCount))
		args = append(args, req.FirstName)
		argCount++
	}
	if req.LastName != nil {
		setParts = append(setParts, "last_name = $"+strconv.Itoa(argCount))
		args = append(args, req.LastName)
		argCount++
	}
	if req.Bio != nil {
		setParts = append(setParts, "bio = $"+strconv.Itoa(argCount))
		args = append(args, req.Bio)
		argCount++
	}
	if req.Avatar != nil {
		setParts = append(setParts, "avatar = $"+strconv.Itoa(argCount))
		args = append(args, req.Avatar)
		argCount++
	}

	if len(setParts) == 1 { // Only updated_at
		return nil, fmt.Errorf("no fields to update")
	}

	// Add user ID as the last argument
	args = append(args, userID)

	query := "UPDATE users SET " + strings.Join(setParts, ", ") + " WHERE id = $" + strconv.Itoa(argCount)

	_, err := s.db.Exec(query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return nil, fmt.Errorf("username or email already exists")
		}
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Return updated user
	return s.GetMe(userID)
}

// ChangePassword changes the user's password
func (s *userService) ChangePassword(userID int, req ChangePasswordRequest) error {
	// Get current password hash
	var currentHash string
	err := s.db.QueryRow("SELECT password_hash FROM users WHERE id = $1 AND role != 'deleted'", userID).Scan(&currentHash)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Verify current password
	if err := utils.CheckPasswordHash(req.CurrentPassword, currentHash); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to process new password: %w", err)
	}

	// Update password
	_, err = s.db.Exec("UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2", newHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// GetUsers returns a paginated list of users
func (s *userService) GetUsers(page, limit int, search string) ([]models.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build search query
	baseQuery := `
		SELECT id, username, email, first_name, last_name, bio, avatar, role, created_at, updated_at
		FROM users
		WHERE role != 'deleted'
	`
	countQuery := "SELECT COUNT(*) FROM users WHERE role != 'deleted'"
	
	args := []interface{}{}
	argCount := 1
	
	if search != "" {
		searchClause := " AND (username ILIKE $" + strconv.Itoa(argCount) + " OR email ILIKE $" + strconv.Itoa(argCount) + 
			" OR first_name ILIKE $" + strconv.Itoa(argCount) + " OR last_name ILIKE $" + strconv.Itoa(argCount) + ")"
		baseQuery += searchClause
		countQuery += searchClause
		args = append(args, "%"+search+"%")
		argCount++
	}

	// Get total count
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users
	orderClause := " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argCount) + " OFFSET $" + strconv.Itoa(argCount+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(baseQuery+orderClause, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
			&user.Bio, &user.Avatar, &user.Role, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, total, nil
}

// GetUsersExceptMe returns all users except current user (for chat)
func (s *userService) GetUsersExceptMe(userID, limit int, search string) ([]models.User, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}

	baseQuery := `
		SELECT id, username, email, first_name, last_name, avatar, role, created_at
		FROM users 
		WHERE id != $1 AND role != 'deleted'
	`
	
	args := []interface{}{userID}
	argCount := 2

	if search != "" {
		baseQuery += " AND (username ILIKE $" + strconv.Itoa(argCount) + " OR first_name ILIKE $" + strconv.Itoa(argCount) + " OR last_name ILIKE $" + strconv.Itoa(argCount) + ")"
		args = append(args, "%"+search+"%")
		argCount++
	}

	baseQuery += " ORDER BY username ASC LIMIT $" + strconv.Itoa(argCount)
	args = append(args, limit)

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FirstName, 
			&user.LastName, &user.Avatar, &user.Role, &user.CreatedAt,
		)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// SearchUsers searches for users
func (s *userService) SearchUsers(query string, limit int) ([]models.User, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}

	rows, err := s.db.Query(`
		SELECT id, username, email, first_name, last_name, bio, avatar, role, created_at, updated_at
		FROM users 
		WHERE (username ILIKE $1 OR email ILIKE $1 OR first_name ILIKE $1 OR last_name ILIKE $1)
		  AND role != 'deleted'
		ORDER BY username ASC
		LIMIT $2
	`, "%"+query+"%", limit)

	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
			&user.Bio, &user.Avatar, &user.Role, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUserByID returns a specific user by ID
func (s *userService) GetUserByID(userID int) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, username, email, first_name, last_name, bio, avatar, role, created_at, updated_at
		FROM users WHERE id = $1 AND role != 'deleted'
	`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
		&user.Bio, &user.Avatar, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	
	return &user, nil
}

// GetUserAvatar returns avatar URL for a user
func (s *userService) GetUserAvatar(userID int) (*string, error) {
	var avatar *string
	err := s.db.QueryRow("SELECT avatar FROM users WHERE id = $1 AND role != 'deleted'", userID).Scan(&avatar)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return avatar, nil
}

// UpdateUserAvatar updates user's avatar URL
func (s *userService) UpdateUserAvatar(userID int, avatarURL string) error {
	_, err := s.db.Exec("UPDATE users SET avatar = $1, updated_at = NOW() WHERE id = $2", avatarURL, userID)
	if err != nil {
		return fmt.Errorf("failed to update avatar: %w", err)
	}
	return nil
}

// DeleteUserAvatar removes user's avatar
func (s *userService) DeleteUserAvatar(userID int) error {
	_, err := s.db.Exec("UPDATE users SET avatar = NULL, updated_at = NOW() WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to remove avatar: %w", err)
	}
	return nil
}