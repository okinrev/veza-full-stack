package http

import (
	"net/http"
)

// RoomHandler définit les opérations HTTP pour les salles de chat
type RoomHandler interface {
	// CRUD basique
	Create(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	GetByUUID(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)

	// Recherche et listing
	List(w http.ResponseWriter, r *http.Request)
	Search(w http.ResponseWriter, r *http.Request)

	// Gestion des membres
	AddMember(w http.ResponseWriter, r *http.Request)
	RemoveMember(w http.ResponseWriter, r *http.Request)
	GetMembers(w http.ResponseWriter, r *http.Request)
	IsMember(w http.ResponseWriter, r *http.Request)

	// Gestion des messages
	GetMessages(w http.ResponseWriter, r *http.Request)
	GetMessageCount(w http.ResponseWriter, r *http.Request)

	// Statistiques
	GetRoomStats(w http.ResponseWriter, r *http.Request)
}
