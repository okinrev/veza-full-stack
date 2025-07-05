package services

import (
	"context"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
)

// SearchService définit les opérations de recherche globales
type SearchService interface {
	// Recherche globale
	Search(ctx context.Context, query string, filters SearchFilters) (*SearchResults, error)

	// Recherche par type
	SearchUsers(ctx context.Context, query string, limit int) ([]*entities.User, error)
	SearchRooms(ctx context.Context, query string, limit int) ([]*entities.Room, error)
	SearchTracks(ctx context.Context, query string, limit int) ([]*entities.Track, error)
	SearchListings(ctx context.Context, query string, limit int) ([]*entities.Listing, error)

	// Suggestions
	GetSuggestions(ctx context.Context, query string, limit int) ([]string, error)

	// Statistiques de recherche
	GetSearchStats(ctx context.Context) (*SearchStats, error)
}

// SearchFilters filtres pour la recherche globale
type SearchFilters struct {
	Types      []string    `json:"types,omitempty"`      // users, rooms, tracks, listings
	Categories []string    `json:"categories,omitempty"` // catégories spécifiques
	PriceRange *PriceRange `json:"price_range,omitempty"`
	DateRange  *DateRange  `json:"date_range,omitempty"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
}

// PriceRange plage de prix pour la recherche
type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// DateRange plage de dates pour la recherche
type DateRange struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

// SearchResults résultats de recherche globale
type SearchResults struct {
	Users    []*entities.User    `json:"users"`
	Rooms    []*entities.Room    `json:"rooms"`
	Tracks   []*entities.Track   `json:"tracks"`
	Listings []*entities.Listing `json:"listings"`
	Total    int64               `json:"total"`
}

// SearchStats statistiques de recherche
type SearchStats struct {
	TotalSearches  int64            `json:"total_searches"`
	PopularQueries []string         `json:"popular_queries"`
	SearchByType   map[string]int64 `json:"search_by_type"`
	AverageResults float64          `json:"average_results"`
}
