package chat

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func RegisterRoutes(r *gin.Engine, h *Handler, jwtSecret string) {
	chatGroup := r.Group("/chat")
	{
		// Routes pour les messages directs
		chatGroup.GET("/dm/:user_id", middleware.JWTAuthMiddleware(jwtSecret), h.GetDmHandler)

		// Routes pour les salons
		chatGroup.GET("/rooms", middleware.JWTAuthMiddleware(jwtSecret), h.GetPublicRoomsHandler)
		chatGroup.POST("/rooms", middleware.JWTAuthMiddleware(jwtSecret), h.CreateRoomHandler)
		chatGroup.GET("/rooms/:room/messages", middleware.JWTAuthMiddleware(jwtSecret), h.GetRoomMessagesHandler)
	}
}
