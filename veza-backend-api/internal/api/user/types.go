package user

import (
	"database/sql"
	"time"
)

// User represents a user with password (for auth)
type User struct {
	ID          int            `db:"id" json:"id"`
	Username    string         `db:"username" json:"username"`
	Email       string         `db:"email" json:"email"`
	Password    string         `db:"password_hash" json:"-"` // Never serialize password
	FirstName   sql.NullString `db:"first_name" json:"first_name,omitempty"`
	LastName    sql.NullString `db:"last_name" json:"last_name,omitempty"`
	Bio         sql.NullString `db:"bio" json:"bio,omitempty"`
	Avatar      sql.NullString `db:"avatar" json:"avatar,omitempty"`
	Role        string         `db:"role" json:"role"`
	IsActive    bool           `db:"is_active" json:"is_active"`
	IsVerified  bool           `db:"is_verified" json:"is_verified"`
	LastLoginAt sql.NullTime   `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
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

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Role      string `json:"role,omitempty"`
}

// UpdateUserRequest represents a request to update user data
type UpdateUserRequest struct {
	Username   *string `json:"username,omitempty"`
	Email      *string `json:"email,omitempty"`
	FirstName  *string `json:"first_name,omitempty"`
	LastName   *string `json:"last_name,omitempty"`
	Bio        *string `json:"bio,omitempty"`
	Avatar     *string `json:"avatar,omitempty"`
	IsActive   *bool   `json:"is_active,omitempty"`
	IsVerified *bool   `json:"is_verified,omitempty"`
	Role       *string `json:"role,omitempty"`
}

