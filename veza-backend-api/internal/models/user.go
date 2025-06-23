// internal/models/user.go
package models

import (
	"database/sql"
	"fmt"
	"time"
)

// User represents a user in the system
type User struct {
	ID           int            `db:"id" json:"id"`
	Username     string         `db:"username" json:"username"`
	Email        string         `db:"email" json:"email"`
	PasswordHash string         `db:"password_hash" json:"-"` // Never serialize password
	FirstName    sql.NullString `db:"first_name" json:"first_name,omitempty"`
	LastName     sql.NullString `db:"last_name" json:"last_name,omitempty"`
	Bio          sql.NullString `db:"bio" json:"bio,omitempty"`
	Avatar       sql.NullString `db:"avatar" json:"avatar,omitempty"`
	Role         string         `db:"role" json:"role"` // user, admin, super_admin
	IsActive     bool           `db:"is_active" json:"is_active"`
	IsVerified   bool           `db:"is_verified" json:"is_verified"`
	LastLoginAt  sql.NullTime   `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
}

// UserResponse represents user data without sensitive information
type UserResponse struct {
	ID          int            `json:"id"`
	Username    string         `json:"username"`
	Email       string         `json:"email"`
	FirstName   sql.NullString `json:"first_name,omitempty"`
	LastName    sql.NullString `json:"last_name,omitempty"`
	Bio         sql.NullString `json:"bio,omitempty"`
	Avatar      sql.NullString `json:"avatar,omitempty"`
	Role        string         `json:"role"`
	IsActive    bool           `json:"is_active"`
	IsVerified  bool           `json:"is_verified"`
	LastLoginAt sql.NullTime   `json:"last_login_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ToResponse converts User to UserResponse (removing sensitive data)
func (u *User) ToResponse() *UserResponse {
	fmt.Printf("🔄 Conversion de l'utilisateur en réponse: %+v\n", u)
	response := &UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Bio:         u.Bio,
		Avatar:      u.Avatar,
		Role:        u.Role,
		IsActive:    u.IsActive,
		IsVerified:  u.IsVerified,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
	fmt.Printf("✅ Réponse utilisateur générée: %+v\n", response)
	return response
}

// RefreshToken represents a JWT refresh token
type RefreshToken struct {
	ID        int       `db:"id" json:"id"`
	UserID    int       `db:"user_id" json:"user_id"`
	Token     string    `db:"token" json:"token"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
