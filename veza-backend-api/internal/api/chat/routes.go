package chat

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func RegisterRoutes(r *gin.Engine, h *Handler, jwtSecret string) {
	// Routes de chat avec le bon préfixe /api/v1/chat
	chatGroup := r.Group("/api/v1/chat")
	{
		// Routes pour les salons
		chatGroup.GET("/rooms", middleware.JWTAuthMiddleware(jwtSecret), h.GetPublicRoomsHandler)
		chatGroup.POST("/rooms", middleware.JWTAuthMiddleware(jwtSecret), h.CreateRoomHandler)
		chatGroup.GET("/rooms/:room/messages", middleware.JWTAuthMiddleware(jwtSecret), h.GetRoomMessagesHandler)
		chatGroup.POST("/rooms/:room/messages", middleware.JWTAuthMiddleware(jwtSecret), h.SendRoomMessageHandler)

		// Routes pour les conversations directes
		chatGroup.GET("/conversations", middleware.JWTAuthMiddleware(jwtSecret), h.GetConversationsHandler)
		chatGroup.GET("/dm/:user_id", middleware.JWTAuthMiddleware(jwtSecret), h.GetDmHandler)
		chatGroup.POST("/dm/:user_id", middleware.JWTAuthMiddleware(jwtSecret), h.SendDMHandler)

		// Route pour les messages non lus
		chatGroup.GET("/unread", middleware.JWTAuthMiddleware(jwtSecret), h.GetUnreadMessagesHandler)
	}
}

// SetupRoutes configure les routes de chat pour un RouterGroup (compatibilité)
func SetupRoutes(router *gin.RouterGroup, h *Handler, jwtSecret string) {
	chatGroup := router.Group("/chat")
	{
		// Routes pour les salons
		chatGroup.GET("/rooms", middleware.JWTAuthMiddleware(jwtSecret), h.GetPublicRoomsHandler)
		chatGroup.POST("/rooms", middleware.JWTAuthMiddleware(jwtSecret), h.CreateRoomHandler)
		chatGroup.GET("/rooms/:room/messages", middleware.JWTAuthMiddleware(jwtSecret), h.GetRoomMessagesHandler)
		chatGroup.POST("/rooms/:room/messages", middleware.JWTAuthMiddleware(jwtSecret), h.SendRoomMessageHandler)

		// Routes pour les conversations directes
		chatGroup.GET("/conversations", middleware.JWTAuthMiddleware(jwtSecret), h.GetConversationsHandler)
		chatGroup.GET("/dm/:user_id", middleware.JWTAuthMiddleware(jwtSecret), h.GetDmHandler)
		chatGroup.POST("/dm/:user_id", middleware.JWTAuthMiddleware(jwtSecret), h.SendDMHandler)

		// Route pour les messages non lus
		chatGroup.GET("/unread", middleware.JWTAuthMiddleware(jwtSecret), h.GetUnreadMessagesHandler)
	}
}
