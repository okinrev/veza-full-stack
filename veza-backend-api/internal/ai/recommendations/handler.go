package recommendations

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler gère les requêtes HTTP pour les recommandations
type Handler struct {
	engine *RecommendationEngine
	logger *zap.Logger
}

// NewHandler crée un nouveau handler de recommandations
func NewHandler(engine *RecommendationEngine, logger *zap.Logger) *Handler {
	return &Handler{
		engine: engine,
		logger: logger,
	}
}

// GetRecommendations handler pour récupérer les recommandations
// @Summary Obtenir des recommandations personnalisées
// @Description Génère des recommandations de pistes, artistes et samples basées sur le profil utilisateur
// @Tags recommendations
// @Accept json
// @Produce json
// @Param user_id query int true "ID de l'utilisateur"
// @Param context query string false "Contexte (discovery, collaboration, production, learning)" default(discovery)
// @Param limit query int false "Nombre de recommandations" default(20)
// @Param freshness query number false "Fraîcheur (0-1, 0=populaire, 1=récent)" default(0.5)
// @Param filters query string false "Filtres JSON"
// @Success 200 {object} RecommendationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/recommendations [get]
func (h *Handler) GetRecommendations(c *gin.Context) {
	// Récupérer les paramètres de la requête
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// Paramètres optionnels
	context := c.DefaultQuery("context", "discovery")
	limitStr := c.DefaultQuery("limit", "20")
	freshnessStr := c.DefaultQuery("freshness", "0.5")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	freshness, err := strconv.ParseFloat(freshnessStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid freshness"})
		return
	}

	// Parser les filtres JSON
	var filters Filters
	if filtersStr := c.Query("filters"); filtersStr != "" {
		if err := json.Unmarshal([]byte(filtersStr), &filters); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filters format"})
			return
		}
	}

	// Valider les paramètres
	if err := h.validateRecommendationParams(userID, context, limit, freshness); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Construire la requête
	req := &RecommendationRequest{
		UserID:    userID,
		Context:   context,
		Limit:     limit,
		Filters:   filters,
		Freshness: freshness,
	}

	// Générer les recommandations
	recommendations, err := h.engine.GetRecommendations(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get recommendations",
			zap.Int64("user_id", userID),
			zap.String("context", context),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate recommendations"})
		return
	}

	h.logger.Info("🎯 Generated recommendations",
		zap.Int64("user_id", userID),
		zap.String("context", context),
		zap.Int("tracks_count", len(recommendations.Tracks)),
		zap.Float64("confidence", recommendations.Confidence),
	)

	c.JSON(http.StatusOK, recommendations)
}

// UpdateUserActivity handler pour mettre à jour l'activité utilisateur
// @Summary Mettre à jour l'activité utilisateur
// @Description Met à jour le profil utilisateur basé sur une nouvelle activité
// @Tags recommendations
// @Accept json
// @Produce json
// @Param user_id path int true "ID de l'utilisateur"
// @Param activity body UserActivity true "Activité utilisateur"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/recommendations/users/{user_id}/activity [post]
func (h *Handler) UpdateUserActivity(c *gin.Context) {
	// Récupérer l'ID utilisateur depuis l'URL
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// Parser l'activité depuis le body
	var activity UserActivity
	if err := c.ShouldBindJSON(&activity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid activity format"})
		return
	}

	// Valider l'activité
	if err := h.validateUserActivity(&activity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mettre à jour le profil utilisateur
	if err := h.engine.UpdateUserProfile(c.Request.Context(), userID, activity); err != nil {
		h.logger.Error("Failed to update user profile",
			zap.Int64("user_id", userID),
			zap.String("activity_type", activity.Type),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user profile"})
		return
	}

	h.logger.Info("🔄 Updated user activity",
		zap.Int64("user_id", userID),
		zap.String("activity_type", activity.Type),
		zap.Int64("track_id", activity.TrackID),
	)

	c.JSON(http.StatusOK, gin.H{"message": "activity updated successfully"})
}

// GetUserProfile handler pour récupérer le profil utilisateur
// @Summary Obtenir le profil utilisateur
// @Description Récupère le profil utilisateur pour les recommandations
// @Tags recommendations
// @Accept json
// @Produce json
// @Param user_id path int true "ID de l'utilisateur"
// @Success 200 {object} UserProfile
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/recommendations/users/{user_id}/profile [get]
func (h *Handler) GetUserProfile(c *gin.Context) {
	// Récupérer l'ID utilisateur depuis l'URL
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// Récupérer le profil utilisateur
	profile, err := h.engine.userProfileService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user profile",
			zap.Int64("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// GetRecommendationStats handler pour les statistiques de recommandations
// @Summary Obtenir les statistiques de recommandations
// @Description Récupère les statistiques d'utilisation des recommandations
// @Tags recommendations
// @Accept json
// @Produce json
// @Param user_id query int false "ID de l'utilisateur (optionnel)"
// @Param date_from query string false "Date de début (YYYY-MM-DD)"
// @Param date_to query string false "Date de fin (YYYY-MM-DD)"
// @Success 200 {object} RecommendationStats
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/recommendations/stats [get]
func (h *Handler) GetRecommendationStats(c *gin.Context) {
	// TODO: Implémenter les statistiques de recommandations
	stats := &RecommendationStats{
		TotalRecommendations: 0,
		AverageConfidence:    0.0,
		TopContexts:          []string{},
		PopularGenres:        []string{},
	}

	c.JSON(http.StatusOK, stats)
}

// validateRecommendationParams valide les paramètres de recommandation
func (h *Handler) validateRecommendationParams(userID int64, context string, limit int, freshness float64) error {
	if userID <= 0 {
		return fmt.Errorf("user_id must be positive")
	}

	validContexts := []string{"discovery", "collaboration", "production", "learning"}
	contextValid := false
	for _, validContext := range validContexts {
		if context == validContext {
			contextValid = true
			break
		}
	}
	if !contextValid {
		return fmt.Errorf("invalid context: %s", context)
	}

	if limit <= 0 || limit > 100 {
		return fmt.Errorf("limit must be between 1 and 100")
	}

	if freshness < 0 || freshness > 1 {
		return fmt.Errorf("freshness must be between 0 and 1")
	}

	return nil
}

// validateUserActivity valide l'activité utilisateur
func (h *Handler) validateUserActivity(activity *UserActivity) error {
	validTypes := []string{"listen", "like", "share", "collaborate", "purchase"}
	typeValid := false
	for _, validType := range validTypes {
		if activity.Type == validType {
			typeValid = true
			break
		}
	}
	if !typeValid {
		return fmt.Errorf("invalid activity type: %s", activity.Type)
	}

	if activity.TrackID <= 0 {
		return fmt.Errorf("track_id must be positive")
	}

	if activity.Interaction < 0 || activity.Interaction > 1 {
		return fmt.Errorf("interaction must be between 0 and 1")
	}

	if activity.Duration < 0 {
		return fmt.Errorf("duration must be positive")
	}

	return nil
}

// RecommendationStats statistiques de recommandations
type RecommendationStats struct {
	TotalRecommendations int      `json:"total_recommendations"`
	AverageConfidence    float64  `json:"average_confidence"`
	TopContexts          []string `json:"top_contexts"`
	PopularGenres        []string `json:"popular_genres"`
}

// ErrorResponse réponse d'erreur standardisée
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse réponse de succès standardisée
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
