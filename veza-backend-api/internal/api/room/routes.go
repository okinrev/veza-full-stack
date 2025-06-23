package room

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	rooms := router.Group("/rooms")
	
	// Routes publiques
	rooms.GET("", handler.GetPublicRoomsHandler)
	
	// Routes protégées
	protected := rooms.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		protected.POST("", handler.CreateRoomHandler)
		protected.GET("/:room/messages", handler.GetRoomMessagesHandler)
	}
}
