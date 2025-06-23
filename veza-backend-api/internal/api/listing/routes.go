package listing

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	listings := router.Group("/listings")
	
	// Routes publiques
	listings.GET("", handler.GetAllListings)
	listings.GET("/:id", handler.GetListingByID)
	
	// Routes protégées
	protected := listings.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		protected.POST("", handler.CreateListing)
		protected.DELETE("/:id", handler.DeleteListing)
	}
}
