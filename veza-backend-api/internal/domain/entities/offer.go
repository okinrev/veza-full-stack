package entities

import (
	"time"
)

// OfferStatus représente le statut d'une offre
type OfferStatus string

const (
	OfferStatusPending   OfferStatus = "pending"
	OfferStatusAccepted  OfferStatus = "accepted"
	OfferStatusRejected  OfferStatus = "rejected"
	OfferStatusCancelled OfferStatus = "cancelled"
	OfferStatusExpired   OfferStatus = "expired"
)

// Offer représente une offre sur une annonce
type Offer struct {
	ID        uint        `json:"id" db:"id"`
	ListingID uint        `json:"listing_id" db:"listing_id"`
	UserID    uint        `json:"user_id" db:"user_id"`
	Amount    float64     `json:"amount" db:"amount"`
	Currency  string      `json:"currency" db:"currency"`
	Message   string      `json:"message" db:"message"`
	Status    OfferStatus `json:"status" db:"status"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time  `json:"deleted_at,omitempty" db:"deleted_at"`
	Listing   *Listing    `json:"listing,omitempty"`
	User      *User       `json:"user,omitempty"`
}

// OfferCreateRequest représente la requête de création d'une offre
type OfferCreateRequest struct {
	ListingID uint    `json:"listing_id" validate:"required"`
	Amount    float64 `json:"amount" validate:"required,min=0"`
	Currency  string  `json:"currency" validate:"required,len=3"`
	Message   string  `json:"message" validate:"omitempty,max=1000"`
}

// OfferUpdateRequest représente la requête de mise à jour d'une offre
type OfferUpdateRequest struct {
	Amount   *float64     `json:"amount,omitempty" validate:"omitempty,min=0"`
	Currency *string      `json:"currency,omitempty" validate:"omitempty,len=3"`
	Message  *string      `json:"message,omitempty" validate:"omitempty,max=1000"`
	Status   *OfferStatus `json:"status,omitempty"`
}
