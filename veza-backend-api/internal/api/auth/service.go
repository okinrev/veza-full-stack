package auth

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/models"
	"github.com/okinrev/veza-web-app/internal/utils"
)

type Service struct {
	db        *database.DB
	jwtSecret string
}

func NewService(db *database.DB, jwtSecret string) *Service {
	return &Service{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         interface{} `json:"user"`
	ExpiresIn    int64       `json:"expires_in"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (s *Service) Register(req RegisterRequest) (*models.User, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)

	// Check email exists
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("email already exists")
	}

	// Check username exists
	err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", req.Username).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
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
		INSERT INTO users (username, email, password_hash, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, 'user', true, NOW(), NOW()) 
		RETURNING id, username, email, password_hash, role, is_active, created_at, updated_at
	`, req.Username, req.Email, hashedPassword).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	fmt.Println("HASH créé :", user.PasswordHash)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func (s *Service) Login(req LoginRequest) (*LoginResponse, error) {
	fmt.Println("📥 Entrée dans service.Login avec email:", req.Email)
	fmt.Printf("🔑 JWT Secret: %s\n", s.jwtSecret)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// Vérifier si l'utilisateur existe
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
	if err != nil {
		fmt.Printf("❌ Erreur lors de la vérification de l'existence de l'utilisateur: %v\n", err)
		return nil, fmt.Errorf("database error")
	}
	fmt.Printf("🔍 Nombre d'utilisateurs trouvés avec cet email: %d\n", count)

	var user models.User
	err = s.db.QueryRow(`
		SELECT id, username, email, password_hash, role, created_at, updated_at 
		FROM users WHERE email = $1 AND is_active = true
	`, req.Email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		fmt.Printf("❌ Erreur lors de la recherche de l'utilisateur: %v\n", err)
		return nil, fmt.Errorf("invalid email or password")
	}

	fmt.Printf("👤 Utilisateur trouvé dans la base de données:\n")
	fmt.Printf("  ID: %d\n", user.ID)
	fmt.Printf("  Username: %s\n", user.Username)
	fmt.Printf("  Email: %s\n", user.Email)
	fmt.Printf("  Role: %s\n", user.Role)
	fmt.Printf("  Password Hash: %s\n", user.PasswordHash)
	fmt.Printf("  Created At: %v\n", user.CreatedAt)
	fmt.Printf("  Updated At: %v\n", user.UpdatedAt)

	checkErr := utils.CheckPasswordHash(req.Password, user.PasswordHash)

	fmt.Println("Password:", req.Password)
	fmt.Println("Hash from DB:", user.PasswordHash)
	fmt.Printf("🔒 Résultat du check bcrypt: %v\n", checkErr)
	fmt.Println("🔒 Comparing password: ", req.Password, " <=> ", user.PasswordHash)

	if checkErr != nil {
		fmt.Printf("❌ Erreur lors de la vérification du mot de passe: %v\n", checkErr)
		return nil, fmt.Errorf("invalid email or password")
	}

	fmt.Printf("🧬 Claims: ID=%d, username=%s, role=%s\n", user.ID, user.Username, user.Role)

	accessToken, refreshToken, expiresIn, err := utils.GenerateTokenPair(
		user.ID, user.Username, user.Role, s.jwtSecret,
	)
	if err != nil {
		fmt.Printf("❌ Erreur lors de la génération des tokens: %v\n", err)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Vérifier si un refresh token existe déjà
	var existingToken string
	err = s.db.QueryRow("SELECT token FROM refresh_tokens WHERE user_id = $1", user.ID).Scan(&existingToken)
	if err != nil && err != sql.ErrNoRows {
		fmt.Printf("❌ Erreur lors de la vérification du refresh token existant: %v\n", err)
		return nil, fmt.Errorf("database error")
	}
	fmt.Printf("🔑 Refresh token existant: %s\n", existingToken)

	_, err = s.db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, NOW() + INTERVAL '7 days', NOW())
		ON CONFLICT (user_id) DO UPDATE SET 
			token = EXCLUDED.token, 
			expires_at = EXCLUDED.expires_at,
			created_at = EXCLUDED.created_at
	`, user.ID, refreshToken)

	if err != nil {
		fmt.Printf("❌ Erreur lors du stockage du refresh token: %v\n", err)
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	response := &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
		ExpiresIn:    expiresIn,
	}

	fmt.Printf("✅ Login réussi pour l'utilisateur %s (ID: %d)\n", user.Username, user.ID)
	fmt.Printf("📦 Réponse du service: %+v\n", response)

	return response, nil
}

func (s *Service) RefreshToken(refreshToken string) (*TokenResponse, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT u.id, u.username, u.email, u.role, u.created_at, u.updated_at
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		WHERE rt.token = $1 AND rt.expires_at > NOW() AND u.is_active = true
	`, refreshToken).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	accessToken, expiresIn, err := utils.GenerateAccessToken(
		user.ID, user.Username, user.Role, s.jwtSecret,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	}, nil
}

func (s *Service) Logout(refreshToken string) error {
	_, err := s.db.Exec("DELETE FROM refresh_tokens WHERE token = $1", refreshToken)
	return err
}

func (s *Service) GetMe(userID int) (*models.UserResponse, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, username, email, first_name, last_name, bio, avatar, 
		       role, is_active, is_verified, last_login_at, created_at, updated_at
		FROM users WHERE id = $1 AND is_active = true
	`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName,
		&user.LastName, &user.Bio, &user.Avatar, &user.Role,
		&user.IsActive, &user.IsVerified, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user.ToResponse(), nil
}
