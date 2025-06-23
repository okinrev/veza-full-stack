//file: backend/routes/room.go

package services

import (
    "github.com/gorilla/mux"
    "veza-web-app/internal/handlers"
    "github.com/jmoiron/sqlx"
    "veza-web-app/internal/api/middleware"
)

func RegisterRoomRoutes(r *mux.Router, db *sqlx.DB) {
	r.Handle("/chat/rooms", middleware.JWTAuthMiddleware(handlers.GetPublicRoomsHandler(db))).Methods("GET")
	r.Handle("/chat/rooms", middleware.JWTAuthMiddleware(handlers.CreateRoomHandler(db))).Methods("POST")
	r.Handle("/chat/rooms/{room}/messages", middleware.JWTAuthMiddleware(handlers.GetRoomMessagesHandler(db))).Methods("GET")
}
