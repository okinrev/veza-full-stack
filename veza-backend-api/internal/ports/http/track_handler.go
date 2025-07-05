package http

import (
	"net/http"
)

// TrackHandler définit les opérations HTTP pour les pistes audio
type TrackHandler interface {
	// CRUD basique
	Create(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)

	// Recherche et listing
	List(w http.ResponseWriter, r *http.Request)
	Search(w http.ResponseWriter, r *http.Request)

	// Gestion des playlists
	AddToPlaylist(w http.ResponseWriter, r *http.Request)
	RemoveFromPlaylist(w http.ResponseWriter, r *http.Request)
	GetPlaylistTracks(w http.ResponseWriter, r *http.Request)

	// Statistiques
	GetTrackStats(w http.ResponseWriter, r *http.Request)
}
