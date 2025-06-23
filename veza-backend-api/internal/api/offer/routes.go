package offer

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	offers := router.Group("/offers")
	
	// Routes protégées
	protected := offers.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		protected.POST("/listings/:id/offers", handler.CreateOffer)
		protected.POST("/:id/accept", handler.AcceptOffer)
	}
}
