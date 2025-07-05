package http

import (
	"net/http"
)

// MessageHandler définit les opérations HTTP pour les messages
type MessageHandler interface {
	// CRUD basique
	Create(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)

	// Recherche et listing
	List(w http.ResponseWriter, r *http.Request)
	Search(w http.ResponseWriter, r *http.Request)

	// Gestion des réactions
	AddReaction(w http.ResponseWriter, r *http.Request)
	RemoveReaction(w http.ResponseWriter, r *http.Request)
	GetReactions(w http.ResponseWriter, r *http.Request)

	// Statistiques
	GetMessageStats(w http.ResponseWriter, r *http.Request)
}
