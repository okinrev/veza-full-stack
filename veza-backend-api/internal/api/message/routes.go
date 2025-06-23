package message

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	messages := router.Group("/messages")
	
	// Routes protégées
	protected := messages.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		protected.GET("/:user_id", handler.GetDmHandler)
	}
}
