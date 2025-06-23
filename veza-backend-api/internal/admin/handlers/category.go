//file: internal/models/admin/category.go

package handlers

import "time"

// Category modèle pour les catégories
type Category struct {
	ID           int64     `json:"id" db:"id"`
	Name         string    `json:"name" db:"name" validate:"required,min=2,max=100"`
	Description  string    `json:"description" db:"description" validate:"max=500"`
	Icon         string    `json:"icon" db:"icon" validate:"max=50"`
	Color        string    `json:"color" db:"color" validate:"omitempty,hexcolor"`
	SortOrder    int       `json:"sort_order" db:"sort_order"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	
	// Champs calculés
	ProductCount int `json:"product_count,omitempty" db:"product_count"`
}

// CreateCategoryRequest structure pour la création
type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
	Icon        string `json:"icon" validate:"max=50"`
	Color       string `json:"color" validate:"omitempty,hexcolor"`
	SortOrder   int    `json:"sort_order"`
	IsActive    bool   `json:"is_active"`
}