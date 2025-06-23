package search

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	search := router.Group("/search")
	{
		search.GET("", handler.GlobalSearchHandler)
		search.GET("/advanced", handler.AdvancedSearchHandler)
		search.GET("/autocomplete", handler.AutocompleteHandler)
	}
}
