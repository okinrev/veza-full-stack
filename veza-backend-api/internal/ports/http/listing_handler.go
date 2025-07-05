package http

import (
	"net/http"
)

// ListingHandler définit les opérations HTTP pour les annonces
type ListingHandler interface {
	// CRUD basique
	Create(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)

	// Recherche et listing
	List(w http.ResponseWriter, r *http.Request)
	Search(w http.ResponseWriter, r *http.Request)

	// Gestion des offres
	AddOffer(w http.ResponseWriter, r *http.Request)
	GetOffers(w http.ResponseWriter, r *http.Request)
	AcceptOffer(w http.ResponseWriter, r *http.Request)

	// Statistiques
	GetListingStats(w http.ResponseWriter, r *http.Request)
}
