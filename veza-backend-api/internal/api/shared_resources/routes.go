package shared_resources

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	resources := router.Group("/shared-resources")
	{
		// Routes publiques
		resources.GET("", handler.ListSharedResources)
		resources.GET("/search", handler.SearchSharedResources)
		resources.GET("/:filename", handler.ServeSharedFile)

		// Routes protégées
		protected := resources.Group("")
		protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
		{
			protected.POST("", handler.UploadSharedResource)
			protected.PUT("/:id", handler.UpdateSharedResource)
			protected.DELETE("/:id", handler.DeleteSharedResource)
		}
	}
}
