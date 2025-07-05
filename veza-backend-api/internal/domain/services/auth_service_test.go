package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"testing"
	"time"

	"github.com/okinrev/veza-web-app/internal/config"
	"github.com/okinrev/veza-web-app/internal/domain/entities"
	"github.com/okinrev/veza-web-app/internal/domain/repositories"
	"github.com/okinrev/veza-web-app/internal/infrastructure/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// generateSecureTestPassword génère un mot de passe sécurisé pour les tests
func generateSecureTestPassword() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return "Test" + base64.URLEncoding.EncodeToString(bytes)[:8] + "!"
}

// generateSecureTestEmail génère un email de test sécurisé
func generateSecureTestEmail() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return "test-" + hex.EncodeToString(bytes) + "@example.com"
}

// generateSecureTestSecret génère un secret sécurisé pour les tests
func generateSecureTestSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// MockUserRepository pour les tests
type MockUserRepository struct {
	users         map[string]*entities.User
	refreshTokens map[string]*MockRefreshToken
}

type MockRefreshToken struct {
	ID        int64
	UserID    int64
	Token     string
	ExpiresAt int64
	CreatedAt int64
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:         make(map[string]*entities.User),
		refreshTokens: make(map[string]*MockRefreshToken),
	}
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	if user, exists := m.users[email]; exists {
		return user, nil
	}
	return nil, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, userID int64) (*entities.User, error) {
	for _, user := range m.users {
		if user.ID == userID {
			return user, nil
		}
	}
	return nil, nil
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	user.ID = int64(len(m.users) + 1)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Hash le mot de passe comme le ferait un vrai repository
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}

	m.users[user.Email] = user
	return nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *entities.User) error {
	if existingUser, exists := m.users[user.Email]; exists {
		existingUser.UpdatedAt = time.Now()
		existingUser.LastLoginAt = user.LastLoginAt
		return nil
	}
	return nil
}

func (m *MockUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	for _, user := range m.users {
		if user.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	_, exists := m.users[email]
	return exists, nil
}

func (m *MockUserRepository) SaveRefreshToken(ctx context.Context, userID int64, token string, expiresAt int64) error {
	m.refreshTokens[token] = &MockRefreshToken{
		ID:        int64(len(m.refreshTokens) + 1),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().Unix(),
	}
	return nil
}

func (m *MockUserRepository) GetRefreshToken(ctx context.Context, token string) (*repositories.RefreshToken, error) {
	if tokenData, exists := m.refreshTokens[token]; exists {
		return &repositories.RefreshToken{
			ID:        tokenData.ID,
			UserID:    tokenData.UserID,
			Token:     tokenData.Token,
			ExpiresAt: tokenData.ExpiresAt,
			CreatedAt: tokenData.CreatedAt,
		}, nil
	}
	return nil, nil
}

func (m *MockUserRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	delete(m.refreshTokens, token)
	return nil
}

func (m *MockUserRepository) RevokeAllUserTokens(ctx context.Context, userID int64) error {
	for token, tokenData := range m.refreshTokens {
		if tokenData.UserID == userID {
			delete(m.refreshTokens, token)
		}
	}
	return nil
}

// Implémenter les méthodes manquantes pour l'interface UserRepository
func (m *MockUserRepository) Count(ctx context.Context, filters repositories.UserFilters) (int64, error) {
	return int64(len(m.users)), nil
}

func (m *MockUserRepository) List(ctx context.Context, filters repositories.UserFilters) ([]*entities.User, error) {
	users := make([]*entities.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *MockUserRepository) Search(ctx context.Context, query string, limit int) ([]*entities.User, error) {
	users := make([]*entities.User, 0)
	for _, user := range m.users {
		if len(users) >= limit {
			break
		}
		users = append(users, user)
	}
	return users, nil
}

func (m *MockUserRepository) Delete(ctx context.Context, userID int64) error {
	for email, user := range m.users {
		if user.ID == userID {
			delete(m.users, email)
			return nil
		}
	}
	return nil
}

func (m *MockUserRepository) GetUserStats(ctx context.Context, userID int64) (*repositories.UserStats, error) {
	return &repositories.UserStats{
		UserID:        userID,
		TotalRooms:    0,
		TotalMessages: 0,
		TotalTracks:   0,
		TotalListings: 0,
		LastActivity:  time.Now().Unix(),
	}, nil
}

func (m *MockUserRepository) GetTotalUsers(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}

func (m *MockUserRepository) GetActiveUsers(ctx context.Context) (int64, error) {
	count := int64(0)
	for _, user := range m.users {
		if user.IsActive {
			count++
		}
	}
	return count, nil
}

func (m *MockUserRepository) GetNewUsersToday(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}

// MockCacheService pour les tests
type MockCacheService struct{}

func (m *MockCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	return nil, nil
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (m *MockCacheService) ListPush(ctx context.Context, key string, values ...interface{}) error {
	return nil
}

func (m *MockCacheService) ListPop(ctx context.Context, key string) (interface{}, error) {
	return nil, nil
}

func (m *MockCacheService) ListRange(ctx context.Context, key string, start, stop int64) ([]interface{}, error) {
	return nil, nil
}

func (m *MockCacheService) ListLength(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *MockCacheService) SetAdd(ctx context.Context, key string, members ...interface{}) error {
	return nil
}

func (m *MockCacheService) SetRemove(ctx context.Context, key string, members ...interface{}) error {
	return nil
}

func (m *MockCacheService) SetMembers(ctx context.Context, key string) ([]interface{}, error) {
	return nil, nil
}

func (m *MockCacheService) SetIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return false, nil
}

func (m *MockCacheService) HashSet(ctx context.Context, key, field string, value interface{}) error {
	return nil
}

func (m *MockCacheService) HashGet(ctx context.Context, key, field string) (interface{}, error) {
	return nil, nil
}

func (m *MockCacheService) HashGetAll(ctx context.Context, key string) (map[string]interface{}, error) {
	return nil, nil
}

func (m *MockCacheService) HashDelete(ctx context.Context, key string, fields ...string) error {
	return nil
}

func (m *MockCacheService) Increment(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *MockCacheService) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return 0, nil
}

func (m *MockCacheService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

func (m *MockCacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, nil
}

func (m *MockCacheService) Pipeline() Pipeline {
	return &MockPipeline{}
}

func (m *MockCacheService) Keys(ctx context.Context, pattern string) ([]string, error) {
	return nil, nil
}

func (m *MockCacheService) DeletePattern(ctx context.Context, pattern string) error {
	return nil
}

func (m *MockCacheService) Ping(ctx context.Context) error {
	return nil
}

func (m *MockCacheService) Info(ctx context.Context) (map[string]string, error) {
	return nil, nil
}

func (m *MockCacheService) FlushDB(ctx context.Context) error {
	return nil
}

// MockPipeline pour les tests
type MockPipeline struct{}

func (m *MockPipeline) Get(key string) *PipelineResult {
	return &PipelineResult{}
}

func (m *MockPipeline) Set(key string, value interface{}, ttl time.Duration) *PipelineResult {
	return &PipelineResult{}
}

func (m *MockPipeline) Delete(key string) *PipelineResult {
	return &PipelineResult{}
}

func (m *MockPipeline) Increment(key string) *PipelineResult {
	return &PipelineResult{}
}

func (m *MockPipeline) Execute(ctx context.Context) ([]*PipelineResult, error) {
	return nil, nil
}

func TestNewAuthService(t *testing.T) {
	mockRepo := NewMockUserRepository()
	mockCache := &MockCacheService{}

	// Configuration JWT de test avec secrets sécurisés
	jwtConfig := config.JWTConfig{
		Secret:          generateSecureTestSecret(),
		ExpirationTime:  15 * time.Minute,
		RefreshTime:     7 * 24 * time.Hour,
		RefreshTTL:      7 * 24 * time.Hour,
		RefreshRotation: true,
	}

	logger, _ := zap.NewDevelopment()
	jwtService := jwt.NewJWTService(generateSecureTestSecret(), generateSecureTestSecret(), "test-issuer")

	service, err := NewAuthService(mockRepo, mockCache, jwtConfig, logger, jwtService)

	if err != nil {
		t.Errorf("NewAuthService returned error: %v", err)
	}

	if service == nil {
		t.Error("NewAuthService returned nil")
	}
}

func TestAuthService_Register(t *testing.T) {
	mockRepo := NewMockUserRepository()
	mockCache := &MockCacheService{}

	jwtConfig := config.JWTConfig{
		Secret:          generateSecureTestSecret(),
		ExpirationTime:  15 * time.Minute,
		RefreshTime:     7 * 24 * time.Hour,
		RefreshTTL:      7 * 24 * time.Hour,
		RefreshRotation: true,
	}

	logger, _ := zap.NewDevelopment()
	jwtService := jwt.NewJWTService(generateSecureTestSecret(), generateSecureTestSecret(), "test-issuer")

	service, err := NewAuthService(mockRepo, mockCache, jwtConfig, logger, jwtService)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	tests := []struct {
		name     string
		username string
		email    string
		password string
		wantErr  bool
		errorMsg string
	}{
		{
			name:     "Valid registration",
			username: "testuser",
			email:    generateSecureTestEmail(),
			password: generateSecureTestPassword(),
			wantErr:  false,
		},
		{
			name:     "Invalid email",
			username: "testuser",
			email:    "invalid-email",
			password: generateSecureTestPassword(),
			wantErr:  true,
			errorMsg: "email invalide",
		},
		{
			name:     "Weak password",
			username: "testuser",
			email:    generateSecureTestEmail(),
			password: "123",
			wantErr:  true,
			errorMsg: "mot de passe invalide",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.Register(context.Background(), tt.username, tt.email, tt.password)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Error("User is nil")
				return
			}

			if user.Username != tt.username {
				t.Errorf("Username = %v, want %v", user.Username, tt.username)
			}

			if user.Email != tt.email {
				t.Errorf("Email = %v, want %v", user.Email, tt.email)
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	mockRepo := NewMockUserRepository()
	mockCache := &MockCacheService{}

	jwtConfig := config.JWTConfig{
		Secret:          generateSecureTestSecret(),
		ExpirationTime:  15 * time.Minute,
		RefreshTime:     7 * 24 * time.Hour,
		RefreshTTL:      7 * 24 * time.Hour,
		RefreshRotation: true,
	}

	logger, _ := zap.NewDevelopment()
	jwtService := jwt.NewJWTService(generateSecureTestSecret(), generateSecureTestSecret(), "test-issuer")

	service, err := NewAuthService(mockRepo, mockCache, jwtConfig, logger, jwtService)
	if err != nil {
		t.Fatalf("Failed to create auth service: %v", err)
	}

	// Créer un utilisateur de test avec mot de passe sécurisé
	testPassword := generateSecureTestPassword()
	user, _ := entities.NewUser("testuser", generateSecureTestEmail(), testPassword)
	if err := mockRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	tests := []struct {
		name     string
		login    string
		password string
		wantErr  bool
		errorMsg string
	}{
		{
			name:     "Valid login",
			login:    user.Email,
			password: testPassword,
			wantErr:  false,
		},
		{
			name:     "Invalid email",
			login:    "nonexistent@example.com",
			password: generateSecureTestPassword(),
			wantErr:  true,
			errorMsg: "utilisateur non trouvé",
		},
		{
			name:     "Invalid password",
			login:    user.Email,
			password: "WrongPassword!",
			wantErr:  true,
			errorMsg: "mot de passe incorrect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.Login(context.Background(), tt.login, tt.password)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Error("User is nil")
				return
			}

			if user.Email != tt.login {
				t.Errorf("Email = %v, want %v", user.Email, tt.login)
			}
		})
	}
}
