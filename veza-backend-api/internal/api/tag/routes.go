package tag

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	tags := router.Group("/tags")
	{
		tags.GET("", handler.GetAllTags)
		tags.GET("/search", handler.SearchTags)
	}
}
