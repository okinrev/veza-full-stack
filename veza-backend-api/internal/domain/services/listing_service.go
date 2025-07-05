package services

import (
	"context"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
	"github.com/okinrev/veza-web-app/internal/domain/repositories"
)

// ListingService définit les opérations métier pour les annonces
type ListingService interface {
	// CRUD basique
	Create(ctx context.Context, listing *entities.Listing) error
	GetByID(ctx context.Context, id int64) (*entities.Listing, error)
	Update(ctx context.Context, listing *entities.Listing) error
	Delete(ctx context.Context, id int64) error

	// Recherche et listing
	List(ctx context.Context, filters repositories.ListingFilters) ([]*entities.Listing, error)
	Count(ctx context.Context, filters repositories.ListingFilters) (int64, error)
	Search(ctx context.Context, query string, limit int) ([]*entities.Listing, error)

	// Gestion des offres
	AddOffer(ctx context.Context, listingID, userID int64, offer *entities.Offer) error
	GetOffers(ctx context.Context, listingID int64) ([]*entities.Offer, error)
	AcceptOffer(ctx context.Context, listingID, offerID int64) error

	// Statistiques
	GetListingStats(ctx context.Context, listingID int64) (*repositories.ListingStats, error)
	GetTotalListings(ctx context.Context) (int64, error)
	GetActiveListings(ctx context.Context) (int64, error)
}
