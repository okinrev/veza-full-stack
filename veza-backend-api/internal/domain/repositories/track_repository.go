package repositories

import (
	"context"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
)

// TrackRepository définit les opérations de persistance pour les pistes audio
type TrackRepository interface {
	// CRUD basique
	Create(ctx context.Context, track *entities.Track) error
	GetByID(ctx context.Context, id int64) (*entities.Track, error)
	Update(ctx context.Context, track *entities.Track) error
	Delete(ctx context.Context, id int64) error

	// Recherche et listing
	List(ctx context.Context, filters TrackFilters) ([]*entities.Track, error)
	Count(ctx context.Context, filters TrackFilters) (int64, error)
	Search(ctx context.Context, query string, limit int) ([]*entities.Track, error)

	// Gestion des playlists
	AddToPlaylist(ctx context.Context, trackID, playlistID int64) error
	RemoveFromPlaylist(ctx context.Context, trackID, playlistID int64) error
	GetPlaylistTracks(ctx context.Context, playlistID int64) ([]*entities.Track, error)

	// Statistiques
	GetTrackStats(ctx context.Context, trackID int64) (*TrackStats, error)
	GetTotalTracks(ctx context.Context) (int64, error)
	GetTracksToday(ctx context.Context) (int64, error)
}

// TrackFilters filtres pour la recherche de pistes
type TrackFilters struct {
	ArtistID      *int64                `json:"artist_id,omitempty"`
	AlbumID       *int64                `json:"album_id,omitempty"`
	Genre         *string               `json:"genre,omitempty"`
	Status        *entities.TrackStatus `json:"status,omitempty"`
	CreatedAfter  *int64                `json:"created_after,omitempty"`
	CreatedBefore *int64                `json:"created_before,omitempty"`
	Search        string                `json:"search,omitempty"`

	// Pagination
	Limit  int `json:"limit"`
	Offset int `json:"offset"`

	// Tri
	OrderBy string `json:"order_by"` // id, title, artist, created_at, duration
	Order   string `json:"order"`    // asc, desc
}

// TrackStats statistiques d'une piste
type TrackStats struct {
	TrackID       int64 `json:"track_id"`
	PlayCount     int64 `json:"play_count"`
	LikeCount     int64 `json:"like_count"`
	ShareCount    int64 `json:"share_count"`
	DownloadCount int64 `json:"download_count"`
}
