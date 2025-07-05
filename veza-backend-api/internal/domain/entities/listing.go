package entities

import (
	"time"
)

// ListingStatus représente le statut d'une annonce
type ListingStatus string

const (
	ListingStatusActive   ListingStatus = "active"
	ListingStatusInactive ListingStatus = "inactive"
	ListingStatusSold     ListingStatus = "sold"
	ListingStatusExpired  ListingStatus = "expired"
)

// Listing représente une annonce dans le système
type Listing struct {
	ID          uint          `json:"id" db:"id"`
	UserID      uint          `json:"user_id" db:"user_id"`
	Title       string        `json:"title" db:"title"`
	Description string        `json:"description" db:"description"`
	Price       float64       `json:"price" db:"price"`
	Currency    string        `json:"currency" db:"currency"`
	Status      ListingStatus `json:"status" db:"status"`
	Category    string        `json:"category" db:"category"`
	Location    string        `json:"location" db:"location"`
	Images      []string      `json:"images" db:"images"`
	Tags        []string      `json:"tags" db:"tags"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time    `json:"deleted_at,omitempty" db:"deleted_at"`
	Offers      []Offer       `json:"offers,omitempty"`
	User        *User         `json:"user,omitempty"`
}

// ListingFilters représente les filtres pour la recherche d'annonces
type ListingFilters struct {
	Category   string        `json:"category,omitempty"`
	Location   string        `json:"location,omitempty"`
	MinPrice   *float64      `json:"min_price,omitempty"`
	MaxPrice   *float64      `json:"max_price,omitempty"`
	Status     ListingStatus `json:"status,omitempty"`
	UserID     *uint         `json:"user_id,omitempty"`
	SearchTerm string        `json:"search_term,omitempty"`
	Tags       []string      `json:"tags,omitempty"`
}

// ListingCreateRequest représente la requête de création d'une annonce
type ListingCreateRequest struct {
	Title       string   `json:"title" validate:"required,min=1,max=200"`
	Description string   `json:"description" validate:"required,min=10,max=2000"`
	Price       float64  `json:"price" validate:"required,min=0"`
	Currency    string   `json:"currency" validate:"required,len=3"`
	Category    string   `json:"category" validate:"required"`
	Location    string   `json:"location" validate:"required"`
	Images      []string `json:"images,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// ListingUpdateRequest représente la requête de mise à jour d'une annonce
type ListingUpdateRequest struct {
	Title       *string        `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description *string        `json:"description,omitempty" validate:"omitempty,min=10,max=2000"`
	Price       *float64       `json:"price,omitempty" validate:"omitempty,min=0"`
	Currency    *string        `json:"currency,omitempty" validate:"omitempty,len=3"`
	Category    *string        `json:"category,omitempty"`
	Location    *string        `json:"location,omitempty"`
	Status      *ListingStatus `json:"status,omitempty"`
	Images      []string       `json:"images,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
}
