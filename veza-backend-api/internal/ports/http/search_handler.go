package http

import (
	"net/http"
)

// SearchHandler définit les opérations HTTP pour la recherche
type SearchHandler interface {
	// Recherche globale
	Search(w http.ResponseWriter, r *http.Request)

	// Recherche par type
	SearchUsers(w http.ResponseWriter, r *http.Request)
	SearchRooms(w http.ResponseWriter, r *http.Request)
	SearchTracks(w http.ResponseWriter, r *http.Request)
	SearchListings(w http.ResponseWriter, r *http.Request)

	// Suggestions
	GetSuggestions(w http.ResponseWriter, r *http.Request)

	// Statistiques de recherche
	GetSearchStats(w http.ResponseWriter, r *http.Request)
}
