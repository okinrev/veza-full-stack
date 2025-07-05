package entities

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"testing"
	"time"
)

// generateSecureTestPassword génère un mot de passe sécurisé pour les tests
func generateSecureTestPassword() string {
	// Générer 16 bytes aléatoires
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	// Convertir en hex et ajouter des caractères spéciaux pour respecter les règles
	return "Test" + hex.EncodeToString(bytes)[:8] + "!"
}

// generateSecureTestEmail génère un email de test sécurisé
func generateSecureTestEmail() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return "test-" + hex.EncodeToString(bytes)[:8] + "@example.com"
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

func generateRandomEmail() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return "test-" + hex.EncodeToString(bytes) + "@example.com"
}

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid user",
			username: "testuser",
			email:    generateSecureTestEmail(),
			password: generateSecureTestPassword(),
			wantErr:  false,
		},
		{
			name:     "Invalid username too short",
			username: "ab",
			email:    generateSecureTestEmail(),
			password: generateSecureTestPassword(),
			wantErr:  true,
		},
		{
			name:     "Invalid email format",
			username: "testuser",
			email:    "invalid-email",
			password: generateSecureTestPassword(),
			wantErr:  true,
		},
		{
			name:     "Invalid password too short",
			username: "testuser",
			email:    generateSecureTestEmail(),
			password: "123",
			wantErr:  true,
		},
		{
			name:     "Invalid password no special chars",
			username: "testuser",
			email:    generateSecureTestEmail(),
			password: "TestPassword123",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.username, tt.email, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if user == nil {
					t.Error("NewUser() returned nil user")
					return
				}

				if user.Username != tt.username {
					t.Errorf("NewUser() username = %v, want %v", user.Username, tt.username)
				}

				if user.Email != tt.email {
					t.Errorf("NewUser() email = %v, want %v", user.Email, tt.email)
				}

				if user.Role != RoleUser {
					t.Errorf("NewUser() role = %v, want %v", user.Role, RoleUser)
				}

				if user.Status != StatusActive {
					t.Errorf("NewUser() status = %v, want %v", user.Status, StatusActive)
				}

				if !user.IsActive {
					t.Error("NewUser() user should be active")
				}
			}
		})
	}
}

func TestUser_ValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"Valid username", "testuser", false},
		{"Valid with numbers", "test123", false},
		{"Valid with underscore", "test_user", false},
		{"Valid with dash", "test-user", false},
		{"Empty username", "", true},
		{"Too short", "ab", true},
		{"Too long", "this_username_is_way_too_long_for_our_system", true},
		{"Invalid chars", "test@user", true},
		{"Invalid chars space", "test user", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Username: tt.username}
			err := user.ValidateUsername()

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_ValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"Valid email", "test@example.com", false},
		{"Valid email with subdomain", "test@mail.example.com", false},
		{"Valid email with numbers", "test123@example.com", false},
		{"Valid email with dots", "test.user@example.com", false},
		{"Empty email", "", true},
		{"Invalid format no @", "testexample.com", true},
		{"Invalid format no domain", "test@", true},
		{"Invalid format no TLD", "test@example", true},
		{"Invalid chars", "test@exam ple.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Email: tt.email}
			err := user.ValidateEmail()

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_ValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid password", generateSecureTestPassword(), false},
		{"Valid complex password", "MyStr0ng!Pass", false},
		{"Empty password", "", true},
		{"Too short", "Pass1!", true},
		{"No uppercase", "testpassword123!", true},
		{"No lowercase", "TESTPASSWORD123!", true},
		{"No number", "TestPassword!", true},
		{"No special char", "TestPassword123", true},
		{"Too long", "ThisPasswordIsWayTooLongForOurSystemAndShouldBeRejectedByTheValidationLogicBecauseItExceedsTheMaximumAllowedLength!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Password: tt.password}
			err := user.ValidatePassword()

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_HasRole(t *testing.T) {
	tests := []struct {
		name     string
		userRole UserRole
		testRole UserRole
		want     bool
	}{
		{"User has user role", RoleUser, RoleUser, true},
		{"Premium has user role", RolePremium, RoleUser, true},
		{"Premium has premium role", RolePremium, RolePremium, true},
		{"Admin has all roles", RoleAdmin, RoleUser, true},
		{"Admin has premium role", RoleAdmin, RolePremium, true},
		{"Admin has moderator role", RoleAdmin, RoleModerator, true},
		{"Admin has admin role", RoleAdmin, RoleAdmin, true},
		{"User doesn't have premium role", RoleUser, RolePremium, false},
		{"User doesn't have admin role", RoleUser, RoleAdmin, false},
		{"Premium doesn't have admin role", RolePremium, RoleAdmin, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.userRole}
			got := user.HasRole(tt.testRole)

			if got != tt.want {
				t.Errorf("HasRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_IsAllowedToLogin(t *testing.T) {
	tests := []struct {
		name     string
		isActive bool
		status   UserStatus
		want     bool
	}{
		{"Active user with active status", true, StatusActive, true},
		{"Active user with inactive status", true, StatusInactive, true},
		{"Inactive user", false, StatusActive, false},
		{"Suspended user", true, StatusSuspended, false},
		{"Banned user", true, StatusBanned, false},
		{"Inactive and banned", false, StatusBanned, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				IsActive: tt.isActive,
				Status:   tt.status,
			}
			got := user.IsAllowedToLogin()

			if got != tt.want {
				t.Errorf("IsAllowedToLogin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_UpdateLastLogin(t *testing.T) {
	user := &User{
		LastLoginAt: nil,
		UpdatedAt:   time.Now().Add(-time.Hour),
	}

	oldUpdatedAt := user.UpdatedAt

	user.UpdateLastLogin()

	if user.LastLoginAt == nil {
		t.Error("UpdateLastLogin() did not set LastLoginAt")
	}

	if user.UpdatedAt.Before(oldUpdatedAt) || user.UpdatedAt.Equal(oldUpdatedAt) {
		t.Error("UpdateLastLogin() did not update UpdatedAt")
	}

	// Vérifier que LastLoginAt est récent (dans les 5 secondes)
	if time.Since(*user.LastLoginAt) > 5*time.Second {
		t.Error("UpdateLastLogin() LastLoginAt is not recent")
	}
}

func TestUser_ToPublic(t *testing.T) {
	user := &User{
		ID:         123,
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "secret",
		FirstName:  "Test",
		LastName:   "User",
		Bio:        "Test bio",
		Avatar:     "avatar.jpg",
		Role:       RolePremium,
		IsVerified: true,
		CreatedAt:  time.Now(),
	}

	public := user.ToPublic()

	if public.ID != user.ID {
		t.Errorf("ToPublic() ID = %v, want %v", public.ID, user.ID)
	}

	if public.Username != user.Username {
		t.Errorf("ToPublic() Username = %v, want %v", public.Username, user.Username)
	}

	// Vérifier que le mot de passe n'est pas exposé
	// (pas de champ Password dans UserPublic)

	if public.Role != user.Role {
		t.Errorf("ToPublic() Role = %v, want %v", public.Role, user.Role)
	}

	if public.IsVerified != user.IsVerified {
		t.Errorf("ToPublic() IsVerified = %v, want %v", public.IsVerified, user.IsVerified)
	}
}

func TestUser_ToSession(t *testing.T) {
	user := &User{
		ID:        123,
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      RolePremium,
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	session := user.ToSession()

	if session.UserID != user.ID {
		t.Errorf("ToSession() UserID = %v, want %v", session.UserID, user.ID)
	}

	if session.Username != user.Username {
		t.Errorf("ToSession() Username = %v, want %v", session.Username, user.Username)
	}

	if session.Email != user.Email {
		t.Errorf("ToSession() Email = %v, want %v", session.Email, user.Email)
	}

	if session.Role != user.Role {
		t.Errorf("ToSession() Role = %v, want %v", session.Role, user.Role)
	}

	if session.IsActive != user.IsActive {
		t.Errorf("ToSession() IsActive = %v, want %v", session.IsActive, user.IsActive)
	}
}
