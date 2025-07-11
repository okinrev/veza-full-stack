package algorithmic

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes enregistre les routes pour les services algorithmiques
func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	// Groupe pour les recommandations
	recommendations := router.Group("/recommendations")
	{
		recommendations.GET("/user/:user_id", handler.GetRecommendations)
		recommendations.POST("/interaction", handler.UpdateUserInteraction)
	}

	// Groupe pour la recherche
	search := router.Group("/search")
	{
		search.GET("/similar/:track_id", handler.SearchSimilar)
		search.GET("/tags", handler.SearchByTags)
		search.POST("/build-index", handler.BuildSearchIndex)
	}

	// Groupe pour l'auto-tagging
	tagging := router.Group("/tagging")
	{
		tagging.POST("/auto-tag/:track_id", handler.AutoTagTrack)
		tagging.GET("/track/:track_id", handler.GetTrackTags)
		tagging.POST("/update-tag", handler.UpdateUserTag)
		tagging.GET("/popular", handler.GetPopularTags)
	}

	// Groupe pour le mastering
	mastering := router.Group("/mastering")
	{
		mastering.POST("/track/:track_id", handler.MasterTrack)
		mastering.GET("/profiles", handler.GetMasteringProfiles)
		mastering.GET("/result/:track_id", handler.GetMasteringResult)
	}

	// Groupe pour la séparation de stems
	stems := router.Group("/stems")
	{
		stems.POST("/separate/:track_id", handler.SeparateStems)
		stems.GET("/result/:track_id", handler.GetStemSeparationResult)
		stems.POST("/extract-center/:track_id", handler.ExtractCenterChannel)
		stems.GET("/stats/:track_id", handler.GetStemStats)
	}
}
