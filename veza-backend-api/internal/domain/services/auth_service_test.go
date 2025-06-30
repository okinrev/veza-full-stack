package services

import (
	"context"
	"testing"
	"time"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
)

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

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	user.ID = int64(len(m.users) + 1)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.Email] = user
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

func TestNewAuthService(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewAuthService(mockRepo)

	if service == nil {
		t.Error("NewAuthService returned nil")
	}
}

func TestAuthService_Register(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewAuthService(mockRepo)

	tests := []struct {
		name     string
		req      RegisterRequest
		wantErr  bool
		errorMsg string
	}{
		{
			name: "Valid registration",
			req: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "Password123!",
			},
			wantErr: false,
		},
		{
			name: "Invalid email",
			req: RegisterRequest{
				Username: "testuser",
				Email:    "invalid-email",
				Password: "Password123!",
			},
			wantErr:  true,
			errorMsg: "email invalide",
		},
		{
			name: "Weak password",
			req: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "123",
			},
			wantErr:  true,
			errorMsg: "mot de passe invalide",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.Register(context.Background(), tt.req)

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

			if resp == nil {
				t.Error("Response is nil")
				return
			}

			if resp.User == nil {
				t.Error("User in response is nil")
				return
			}

			if resp.User.Username != tt.req.Username {
				t.Errorf("Username = %v, want %v", resp.User.Username, tt.req.Username)
			}

			if resp.User.Email != tt.req.Email {
				t.Errorf("Email = %v, want %v", resp.User.Email, tt.req.Email)
			}

			if resp.AccessToken == "" {
				t.Error("AccessToken is empty")
			}

			if resp.RefreshToken == "" {
				t.Error("RefreshToken is empty")
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	mockRepo := NewMockUserRepository()
	service := NewAuthService(mockRepo)

	// Créer un utilisateur de test
	user, _ := entities.NewUser("testuser", "test@example.com", "Password123!")
	mockRepo.Create(context.Background(), user)

	tests := []struct {
		name     string
		req      LoginRequest
		wantErr  bool
		errorMsg string
	}{
		{
			name: "Valid login",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "Password123!",
			},
			wantErr: false,
		},
		{
			name: "Invalid email",
			req: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "Password123!",
			},
			wantErr:  true,
			errorMsg: "utilisateur non trouvé",
		},
		{
			name: "Invalid password",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "WrongPassword!",
			},
			wantErr:  true,
			errorMsg: "mot de passe incorrect",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.Login(context.Background(), tt.req)

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

			if resp == nil {
				t.Error("Response is nil")
				return
			}

			if resp.User.Email != tt.req.Email {
				t.Errorf("Email = %v, want %v", resp.User.Email, tt.req.Email)
			}

			if resp.AccessToken == "" {
				t.Error("AccessToken is empty")
			}

			if resp.RefreshToken == "" {
				t.Error("RefreshToken is empty")
			}
		})
	}
}
