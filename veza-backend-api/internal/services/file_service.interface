// internal/services/file_service.go

package services

import (
	"github.com/gorilla/mux"
	"veza-web-app/internal/handlers"
	"veza-web-app/internal/api/middleware"
)

func RegisterFileRoutes(r *mux.Router) {
	// Sous-routes pour les fichiers d’un produit
	productFiles := r.PathPrefix("/products/{id}/files").Subrouter()
	productFiles.Use(middleware.JWTAuthMiddleware)

	productFiles.HandleFunc("", handlers.UploadFileHandler).Methods("POST")
	productFiles.HandleFunc("", handlers.ListProductFiles).Methods("GET")

	// Sous-routes pour accéder directement à un fichier
	files := r.PathPrefix("/files").Subrouter()
	files.Use(middleware.JWTAuthMiddleware)

	files.HandleFunc("/{id:[0-9]+}", handlers.DownloadFileHandler).Methods("GET")
}
