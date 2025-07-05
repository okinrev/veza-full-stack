package http

import (
	"net/http"
)

// UserHandler définit les opérations HTTP pour les utilisateurs
type UserHandler interface {
	// CRUD basique
	Create(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	GetByUsername(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)

	// Recherche et listing
	List(w http.ResponseWriter, r *http.Request)
	Search(w http.ResponseWriter, r *http.Request)

	// Profil utilisateur
	GetProfile(w http.ResponseWriter, r *http.Request)
	UpdateProfile(w http.ResponseWriter, r *http.Request)
	GetUserStats(w http.ResponseWriter, r *http.Request)

	// Gestion des sessions
	GetSessions(w http.ResponseWriter, r *http.Request)
	RevokeSession(w http.ResponseWriter, r *http.Request)
	RevokeAllSessions(w http.ResponseWriter, r *http.Request)
}
