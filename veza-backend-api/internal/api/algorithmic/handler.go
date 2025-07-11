package algorithmic

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler gère les requêtes API pour les services algorithmiques
type Handler struct {
	recommendationService *RecommendationService
	searchService         *SearchService
	taggingService        *TaggingService
	masteringService      *MasteringService
	stemService           *StemSeparationService
	logger                *zap.Logger
}

// NewHandler crée une nouvelle instance du handler
func NewHandler(
	recommendationService *RecommendationService,
	searchService *SearchService,
	taggingService *TaggingService,
	masteringService *MasteringService,
	stemService *StemSeparationService,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		recommendationService: recommendationService,
		searchService:         searchService,
		taggingService:        taggingService,
		masteringService:      masteringService,
		stemService:           stemService,
		logger:                logger,
	}
}

// ===== RECOMMENDATION ENDPOINTS =====

// GetRecommendations retourne les recommandations pour un utilisateur
func (h *Handler) GetRecommendations(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	limit := getLimit(c, 20)

	recommendations, err := h.recommendationService.GetRecommendations(userID, limit)
	if err != nil {
		h.logger.Error("Failed to get recommendations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recommendations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": recommendations,
		"user_id":         userID,
		"limit":           limit,
	})
}

// UpdateUserInteraction met à jour une interaction utilisateur
func (h *Handler) UpdateUserInteraction(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	trackID, err := getTrackID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	interactionType := c.PostForm("interaction_type")
	if interactionType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing interaction_type"})
		return
	}

	err = h.recommendationService.UpdateUserInteraction(userID, trackID, interactionType)
	if err != nil {
		h.logger.Error("Failed to update user interaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update interaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "Interaction updated successfully",
		"user_id":          userID,
		"track_id":         trackID,
		"interaction_type": interactionType,
	})
}

// ===== SEARCH ENDPOINTS =====

// SearchSimilar recherche les audios similaires
func (h *Handler) SearchSimilar(c *gin.Context) {
	trackID, err := getTrackID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	limit := getLimit(c, 10)

	results, err := h.searchService.SearchSimilar(trackID, limit)
	if err != nil {
		h.logger.Error("Failed to search similar", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search similar tracks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results":        results,
		"query_track_id": trackID,
		"limit":          limit,
	})
}

// SearchByTags recherche par tags
func (h *Handler) SearchByTags(c *gin.Context) {
	tags := c.QueryArray("tags")
	if len(tags) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No tags provided"})
		return
	}

	limit := getLimit(c, 20)

	results, err := h.searchService.SearchByTags(tags, limit)
	if err != nil {
		h.logger.Error("Failed to search by tags", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search by tags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"tags":    tags,
		"limit":   limit,
	})
}

// BuildSearchIndex construit l'index de recherche
func (h *Handler) BuildSearchIndex(c *gin.Context) {
	err := h.searchService.BuildSearchIndex()
	if err != nil {
		h.logger.Error("Failed to build search index", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build search index"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Search index built successfully",
	})
}

// ===== TAGGING ENDPOINTS =====

// AutoTagTrack tag automatiquement un track
func (h *Handler) AutoTagTrack(c *gin.Context) {
	trackID, err := getTrackID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	// Récupérer le chemin audio depuis la base de données
	audioPath, err := h.getAudioPath(trackID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Track not found or no audio path"})
		return
	}

	tags, err := h.taggingService.AutoTagTrack(trackID, audioPath)
	if err != nil {
		h.logger.Error("Failed to auto-tag track", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to auto-tag track"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"track_id": trackID,
		"tags":     tags,
	})
}

// GetTrackTags récupère les tags d'un track
func (h *Handler) GetTrackTags(c *gin.Context) {
	trackID, err := getTrackID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	tags, err := h.taggingService.GetTrackTags(trackID)
	if err != nil {
		h.logger.Error("Failed to get track tags", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get track tags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"track_id": trackID,
		"tags":     tags,
	})
}

// UpdateUserTag permet à l'utilisateur de modifier un tag
func (h *Handler) UpdateUserTag(c *gin.Context) {
	trackID, err := getTrackID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	tagName := c.PostForm("tag_name")
	if tagName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing tag_name"})
		return
	}

	action := c.PostForm("action") // "add" ou "remove"
	if action == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing action"})
		return
	}

	err = h.taggingService.UpdateUserTag(trackID, tagName, action)
	if err != nil {
		h.logger.Error("Failed to update user tag", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Tag updated successfully",
		"track_id": trackID,
		"tag_name": tagName,
		"action":   action,
	})
}

// GetPopularTags récupère les tags les plus populaires
func (h *Handler) GetPopularTags(c *gin.Context) {
	limit := getLimit(c, 50)

	tags, err := h.taggingService.GetPopularTags(limit)
	if err != nil {
		h.logger.Error("Failed to get popular tags", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get popular tags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tags":  tags,
		"limit": limit,
	})
}

// ===== MASTERING ENDPOINTS =====

// MasterTrack applique le mastering à un track
func (h *Handler) MasterTrack(c *gin.Context) {
	trackID, err := getTrackID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	profile := c.PostForm("profile")
	if profile == "" {
		profile = "streaming" // Profil par défaut
	}

	// Récupérer le chemin audio
	audioPath, err := h.getAudioPath(trackID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Track not found or no audio path"})
		return
	}

	result, err := h.masteringService.MasterTrack(trackID, audioPath, profile)
	if err != nil {
		h.logger.Error("Failed to master track", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to master track"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"track_id": trackID,
		"profile":  profile,
		"result":   result,
	})
}

// GetMasteringProfiles retourne les profils disponibles
func (h *Handler) GetMasteringProfiles(c *gin.Context) {
	profiles := h.masteringService.GetMasteringProfiles()

	c.JSON(http.StatusOK, gin.H{
		"profiles": profiles,
	})
}

// GetMasteringResult récupère le résultat du mastering
func (h *Handler) GetMasteringResult(c *gin.Context) {
	trackID, err := getTrackID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	result, err := h.masteringService.GetMasteringResult(trackID)
	if err != nil {
		h.logger.Error("Failed to get mastering result", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get mastering result"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"track_id": trackID,
		"result":   result,
	})
}

// ===== STEM SEPARATION ENDPOINTS =====

// SeparateStems sépare les stems d'un track
func (h *Handler) SeparateStems(c *gin.Context) {
	trackID, err := getTrackID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	// Récupérer le chemin audio
	audioPath, err := h.getAudioPath(trackID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Track not found or no audio path"})
		return
	}

	result, err := h.stemService.SeparateStems(trackID, audioPath)
	if err != nil {
		h.logger.Error("Failed to separate stems", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to separate stems"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"track_id": trackID,
		"result":   result,
	})
}

// GetStemSeparationResult récupère le résultat de la séparation
func (h *Handler) GetStemSeparationResult(c *gin.Context) {
	trackID, err := getTrackID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	result, err := h.stemService.GetStemSeparationResult(trackID)
	if err != nil {
		h.logger.Error("Failed to get stem separation result", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stem separation result"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"track_id": trackID,
		"result":   result,
	})
}

// ExtractCenterChannel extrait le canal central
func (h *Handler) ExtractCenterChannel(c *gin.Context) {
	trackID, err := getTrackID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	// Récupérer le chemin audio
	audioPath, err := h.getAudioPath(trackID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Track not found or no audio path"})
		return
	}

	outputPath, err := h.stemService.ExtractCenterChannel(trackID, audioPath)
	if err != nil {
		h.logger.Error("Failed to extract center channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract center channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"track_id":    trackID,
		"output_path": outputPath,
	})
}

// GetStemStats récupère les statistiques des stems
func (h *Handler) GetStemStats(c *gin.Context) {
	trackID, err := getTrackID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	stats, err := h.stemService.GetStemStats(trackID)
	if err != nil {
		h.logger.Error("Failed to get stem stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stem stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"track_id": trackID,
		"stats":    stats,
	})
}

// ===== UTILITY FUNCTIONS =====

// getUserID extrait l'ID utilisateur depuis le contexte
func getUserID(c *gin.Context) (int64, error) {
	userIDStr := c.Param("user_id")
	if userIDStr == "" {
		userIDStr = c.Query("user_id")
	}

	if userIDStr == "" {
		return 0, fmt.Errorf("user_id not provided")
	}

	return strconv.ParseInt(userIDStr, 10, 64)
}

// getTrackID extrait l'ID track depuis le contexte
func getTrackID(c *gin.Context) (int64, error) {
	trackIDStr := c.Param("track_id")
	if trackIDStr == "" {
		trackIDStr = c.PostForm("track_id")
	}

	if trackIDStr == "" {
		return 0, fmt.Errorf("track_id not provided")
	}

	return strconv.ParseInt(trackIDStr, 10, 64)
}

// getLimit extrait la limite depuis les paramètres
func getLimit(c *gin.Context, defaultLimit int) int {
	limitStr := c.Query("limit")
	if limitStr == "" {
		return defaultLimit
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return defaultLimit
	}

	return limit
}

// getAudioPath récupère le chemin audio d'un track depuis la base de données
func (h *Handler) getAudioPath(trackID int64) (string, error) {
	// Cette fonction devrait être implémentée pour récupérer le chemin audio
	// depuis la base de données. Pour l'instant, on retourne une valeur simulée.
	return fmt.Sprintf("/audio/tracks/%d.wav", trackID), nil
}
