package http

import (
	"net/http"
)

// AdminHandler définit les opérations HTTP pour l'administration
type AdminHandler interface {
	// Dashboard
	GetDashboard(w http.ResponseWriter, r *http.Request)
	GetStats(w http.ResponseWriter, r *http.Request)

	// Gestion des utilisateurs
	GetUsers(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
	BanUser(w http.ResponseWriter, r *http.Request)
	UnbanUser(w http.ResponseWriter, r *http.Request)

	// Gestion des salles
	GetRooms(w http.ResponseWriter, r *http.Request)
	GetRoom(w http.ResponseWriter, r *http.Request)
	UpdateRoom(w http.ResponseWriter, r *http.Request)
	DeleteRoom(w http.ResponseWriter, r *http.Request)

	// Gestion des messages
	GetMessages(w http.ResponseWriter, r *http.Request)
	GetMessage(w http.ResponseWriter, r *http.Request)
	DeleteMessage(w http.ResponseWriter, r *http.Request)

	// Modération
	GetModerationLogs(w http.ResponseWriter, r *http.Request)
	CreateModerationLog(w http.ResponseWriter, r *http.Request)

	// Système
	GetSystemInfo(w http.ResponseWriter, r *http.Request)
	GetLogs(w http.ResponseWriter, r *http.Request)
}
