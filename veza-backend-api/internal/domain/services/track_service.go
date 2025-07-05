package services

import (
	"context"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
	"github.com/okinrev/veza-web-app/internal/domain/repositories"
)

// TrackService définit les opérations métier pour les pistes audio
type TrackService interface {
	// CRUD basique
	Create(ctx context.Context, track *entities.Track) error
	GetByID(ctx context.Context, id int64) (*entities.Track, error)
	Update(ctx context.Context, track *entities.Track) error
	Delete(ctx context.Context, id int64) error

	// Recherche et listing
	List(ctx context.Context, filters repositories.TrackFilters) ([]*entities.Track, error)
	Count(ctx context.Context, filters repositories.TrackFilters) (int64, error)
	Search(ctx context.Context, query string, limit int) ([]*entities.Track, error)

	// Gestion des playlists
	AddToPlaylist(ctx context.Context, trackID, playlistID int64) error
	RemoveFromPlaylist(ctx context.Context, trackID, playlistID int64) error
	GetPlaylistTracks(ctx context.Context, playlistID int64) ([]*entities.Track, error)

	// Statistiques
	GetTrackStats(ctx context.Context, trackID int64) (*repositories.TrackStats, error)
	GetTotalTracks(ctx context.Context) (int64, error)
	GetTracksToday(ctx context.Context) (int64, error)
}
