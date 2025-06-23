// internal/services/auth_service.go
package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/models"
	"github.com/okinrev/veza-web-app/internal/utils"
)

type AuthService interface {
	Register(req RegisterRequest) (*models.User, error)
	Login(req LoginRequest) (*LoginResponse, error)
	RefreshToken(refreshToken string) (*TokenResponse, error)
	Logout(refreshToken string) error
	VerifyToken(tokenString string) (*TokenClaims, error)
	GenerateTokenPair(userID int, username, role string) (*TokenPair, error)
}

type authService struct {
	db        *database.DB
	jwtSecret string
}

func NewAuthService(db *database.DB, jwtSecret string) AuthService {
	return &authService{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// Request/Response types
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type TokenClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Register creates a new user account
func (s *authService) Register(req RegisterRequest) (*models.User, error) {
	// Normalize input
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)

	// Check if email already exists
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("email already exists")
	}

	// Check if username already exists
	err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", req.Username).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	var user models.User
	err = s.db.QueryRow(`
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, 'user', NOW(), NOW()) 
		RETURNING id, username, email, role, created_at, updated_at
	`, req.Username, req.Email, hashedPassword).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// Login authenticates a user and returns tokens
func (s *authService) Login(req LoginRequest) (*LoginResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// Get user from database
	var user models.User
	var passwordHash string
	err := s.db.QueryRow(`
		SELECT id, username, email, password_hash, role, created_at, updated_at 
		FROM users WHERE email = $1 AND role != 'deleted'
	`, req.Email).Scan(
		&user.ID, &user.Username, &user.Email, &passwordHash, 
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	if err := utils.CheckPasswordHash(req.Password, passwordHash); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Generate tokens
	tokenPair, err := s.GenerateTokenPair(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store refresh token
	err = s.storeRefreshToken(user.ID, tokenPair.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Update last login
	_, err = s.db.Exec("UPDATE users SET updated_at = NOW() WHERE id = $1", user.ID)
	if err != nil {
		// Log error but don't fail login
		fmt.Printf("Failed to update last login for user %d: %v\n", user.ID, err)
	}

	return &LoginResponse{
		User:         &user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// RefreshToken generates a new access token from a refresh token
func (s *authService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	// Verify refresh token and get user
	var user models.User
	err := s.db.QueryRow(`
		SELECT u.id, u.username, u.email, u.role, u.created_at, u.updated_at
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		WHERE rt.token = $1 AND rt.expires_at > NOW() AND u.role != 'deleted'
	`, refreshToken).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role, 
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	// Generate new access token
	accessToken, expiresIn, err := utils.GenerateAccessToken(user.ID, user.Username, user.Role, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	}, nil
}

// Logout invalidates the refresh token
func (s *authService) Logout(refreshToken string) error {
	_, err := s.db.Exec("DELETE FROM refresh_tokens WHERE token = $1", refreshToken)
	if err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}
	return nil
}

// VerifyToken validates and extracts claims from a JWT token
func (s *authService) VerifyToken(tokenString string) (*TokenClaims, error) {
	claims, err := utils.ValidateJWT(tokenString, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return &TokenClaims{
		UserID:   claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}

// GenerateTokenPair generates both access and refresh tokens
func (s *authService) GenerateTokenPair(userID int, username, role string) (*TokenPair, error) {
	accessToken, refreshToken, expiresIn, err := utils.GenerateTokenPair(userID, username, role, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// Helper methods
func (s *authService) storeRefreshToken(userID int, token string) error {
	_, err := s.db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, NOW() + INTERVAL '7 days', NOW())
		ON CONFLICT (user_id) DO UPDATE SET 
			token = EXCLUDED.token, 
			expires_at = EXCLUDED.expires_at,
			created_at = EXCLUDED.created_at
	`, userID, token)

	return err
}