package notifications

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegisterRoutes enregistre toutes les routes de notifications
func RegisterRoutes(router *gin.RouterGroup, handler *Handler, logger *zap.Logger) {
	// Groupe de routes pour les notifications
	notif := router.Group("/notifications")
	
	// Routes publiques (nécessitent authentification)
	{
		// WebSocket endpoint
		notif.GET("/ws", handler.HandleWebSocket)
		
		// Endpoints de notification
		notif.POST("/send", handler.SendNotification)
		notif.GET("", handler.GetUserNotifications)
		notif.PUT("/:id/read", handler.MarkAsRead)
		notif.GET("/unread-count", handler.GetUnreadCount)
		notif.POST("/test", handler.TestNotification)
		
		// Préférences utilisateur
		notif.GET("/preferences", handler.GetUserPreferences)
		notif.PUT("/preferences", handler.UpdateUserPreferences)
		
		// Endpoints d'information
		notif.GET("/types", handler.GetNotificationTypes)
		notif.GET("/channels", handler.GetChannels)
	}
	
	// Routes d'administration (nécessitent rôle admin)
	admin := notif.Group("/admin")
	{
		admin.POST("/bulk", handler.SendBulkNotification)
		admin.POST("/broadcast", handler.BroadcastNotification)
		admin.GET("/stats", handler.GetStats)
	}
	
	logger.Info("📡 Notification routes registered successfully")
}

// RegisterWebSocketRoute enregistre uniquement la route WebSocket
func RegisterWebSocketRoute(router *gin.Engine, websocketService *WebSocketService) {
	router.GET("/ws/notifications", websocketService.HandleWebSocket)
}
