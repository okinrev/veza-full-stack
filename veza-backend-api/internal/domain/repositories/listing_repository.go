package repositories

import (
	"context"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
)

// ListingRepository définit les opérations de persistance pour les annonces
type ListingRepository interface {
	// CRUD basique
	Create(ctx context.Context, listing *entities.Listing) error
	GetByID(ctx context.Context, id int64) (*entities.Listing, error)
	Update(ctx context.Context, listing *entities.Listing) error
	Delete(ctx context.Context, id int64) error

	// Recherche et listing
	List(ctx context.Context, filters ListingFilters) ([]*entities.Listing, error)
	Count(ctx context.Context, filters ListingFilters) (int64, error)
	Search(ctx context.Context, query string, limit int) ([]*entities.Listing, error)

	// Gestion des offres
	AddOffer(ctx context.Context, listingID, userID int64, offer *entities.Offer) error
	GetOffers(ctx context.Context, listingID int64) ([]*entities.Offer, error)
	AcceptOffer(ctx context.Context, listingID, offerID int64) error

	// Statistiques
	GetListingStats(ctx context.Context, listingID int64) (*ListingStats, error)
	GetTotalListings(ctx context.Context) (int64, error)
	GetActiveListings(ctx context.Context) (int64, error)
}

// ListingFilters filtres pour la recherche d'annonces
type ListingFilters struct {
	UserID        *int64                  `json:"user_id,omitempty"`
	Category      *string                 `json:"category,omitempty"`
	Status        *entities.ListingStatus `json:"status,omitempty"`
	PriceMin      *float64                `json:"price_min,omitempty"`
	PriceMax      *float64                `json:"price_max,omitempty"`
	CreatedAfter  *int64                  `json:"created_after,omitempty"`
	CreatedBefore *int64                  `json:"created_before,omitempty"`
	Search        string                  `json:"search,omitempty"`

	// Pagination
	Limit  int `json:"limit"`
	Offset int `json:"offset"`

	// Tri
	OrderBy string `json:"order_by"` // id, title, price, created_at
	Order   string `json:"order"`    // asc, desc
}

// ListingStats statistiques d'une annonce
type ListingStats struct {
	ListingID  int64 `json:"listing_id"`
	ViewCount  int64 `json:"view_count"`
	OfferCount int64 `json:"offer_count"`
	LikeCount  int64 `json:"like_count"`
	ShareCount int64 `json:"share_count"`
}
