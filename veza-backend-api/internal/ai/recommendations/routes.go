package recommendations

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupRoutes configure les routes pour les recommandations
func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	// Groupe de routes pour les recommandations
	recommendations := router.Group("/recommendations")
	{
		// Route principale pour obtenir des recommandations
		recommendations.GET("", handler.GetRecommendations)

		// Routes pour les statistiques
		recommendations.GET("/stats", handler.GetRecommendationStats)

		// Routes pour les profils utilisateur
		users := recommendations.Group("/users")
		{
			users.GET("/:user_id/profile", handler.GetUserProfile)
			users.POST("/:user_id/activity", handler.UpdateUserActivity)
		}
	}
}

// RegisterRoutes enregistre les routes dans le router principal
func RegisterRoutes(router *gin.Engine, engine *RecommendationEngine, logger *zap.Logger, jwtSecret string) {
	handler := NewHandler(engine, logger)

	// Groupe API v1
	v1 := router.Group("/api/v1")
	SetupRoutes(v1, handler, jwtSecret)
}
