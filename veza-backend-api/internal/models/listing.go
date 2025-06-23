// internal/models/listing.go
package models

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// Listing represents a marketplace listing
type Listing struct {
	ID          int            `db:"id" json:"id"`
	UserID      int            `db:"user_id" json:"user_id"`
	ProductID   int            `db:"product_id" json:"product_id"` // References user_products.id
	Description string         `db:"description" json:"description"`
	State       string         `db:"state" json:"state"` // new, like_new, good, fair, poor
	Price       sql.NullInt32  `db:"price" json:"price,omitempty"`
	ExchangeFor sql.NullString `db:"exchange_for" json:"exchange_for,omitempty"`
	Images      pq.StringArray `db:"images" json:"images"`
	Status      string         `db:"status" json:"status"` // open, closed, sold
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
}

// ListingWithDetails represents a listing with user and product information
type ListingWithDetails struct {
	Listing
	Username    string         `db:"username" json:"username,omitempty"`
	UserAvatar  sql.NullString `db:"user_avatar" json:"user_avatar,omitempty"`
	ProductName string         `db:"product_name" json:"product_name,omitempty"`
	Brand       sql.NullString `db:"brand" json:"brand,omitempty"`
	Model       sql.NullString `db:"model" json:"model,omitempty"`
	OfferCount  int            `db:"offer_count" json:"offer_count,omitempty"`
}

// Offer represents an offer on a listing
type Offer struct {
	ID                int            `db:"id" json:"id"`
	ListingID         int            `db:"listing_id" json:"listing_id"`
	FromUserID        int            `db:"from_user_id" json:"from_user_id"`
	ProposedProductID int            `db:"proposed_product_id" json:"proposed_product_id"` // References user_products.id
	Message           sql.NullString `db:"message" json:"message,omitempty"`
	Status            string         `db:"status" json:"status"` // pending, accepted, rejected, withdrawn
	CounterOffer      sql.NullString `db:"counter_offer" json:"counter_offer,omitempty"`
	ExpiresAt         sql.NullTime   `db:"expires_at" json:"expires_at,omitempty"`
	ViewedAt          sql.NullTime   `db:"viewed_at" json:"viewed_at,omitempty"`
	CreatedAt         time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at" json:"updated_at"`
}

// OfferWithDetails represents an offer with user and product information
type OfferWithDetails struct {
	Offer
	FromUsername        string         `db:"from_username" json:"from_username,omitempty"`
	FromUserAvatar      sql.NullString `db:"from_user_avatar" json:"from_user_avatar,omitempty"`
	ProposedProductName string         `db:"proposed_product_name" json:"proposed_product_name,omitempty"`
	ListingTitle        string         `db:"listing_title" json:"listing_title,omitempty"`
	ListingOwnerID      int            `db:"listing_owner_id" json:"listing_owner_id,omitempty"`
}