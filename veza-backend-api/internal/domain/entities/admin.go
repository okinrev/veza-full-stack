package entities

import (
	"time"
)

// AdminRole représente le rôle d'un administrateur
type AdminRole string

const (
	AdminRoleSuperAdmin AdminRole = "super_admin"
	AdminRoleAdmin      AdminRole = "admin"
	AdminRoleModerator  AdminRole = "moderator"
	AdminRoleSupport    AdminRole = "support"
)

// AdminStatus représente le statut d'un administrateur
type AdminStatus string

const (
	AdminStatusActive    AdminStatus = "active"
	AdminStatusInactive  AdminStatus = "inactive"
	AdminStatusSuspended AdminStatus = "suspended"
)

// Admin représente un administrateur du système
type Admin struct {
	ID          uint        `json:"id" db:"id"`
	UserID      uint        `json:"user_id" db:"user_id"`
	Role        AdminRole   `json:"role" db:"role"`
	Status      AdminStatus `json:"status" db:"status"`
	Permissions []string    `json:"permissions" db:"permissions"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`
	User        *User       `json:"user,omitempty"`
}

// AdminCreateRequest représente la requête de création d'un administrateur
type AdminCreateRequest struct {
	UserID      uint      `json:"user_id" validate:"required"`
	Role        AdminRole `json:"role" validate:"required"`
	Permissions []string  `json:"permissions,omitempty"`
}

// AdminUpdateRequest représente la requête de mise à jour d'un administrateur
type AdminUpdateRequest struct {
	Role        *AdminRole   `json:"role,omitempty"`
	Status      *AdminStatus `json:"status,omitempty"`
	Permissions []string     `json:"permissions,omitempty"`
}

// AdminFilters représente les filtres pour la recherche d'administrateurs
type AdminFilters struct {
	Role       AdminRole   `json:"role,omitempty"`
	Status     AdminStatus `json:"status,omitempty"`
	UserID     *uint       `json:"user_id,omitempty"`
	SearchTerm string      `json:"search_term,omitempty"`
}
