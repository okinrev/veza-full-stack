// internal/api/track/handler.go
package track

import (
	"net/http"
	"strconv"
	"github.com/okinrev/veza-web-app/internal/utils/response"  // ADD THIS
    "github.com/okinrev/veza-web-app/internal/common"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// AddTrackWithUpload upload une nouvelle piste
func (h *Handler) AddTrackWithUpload(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	// Récupérer les données du formulaire
	title := c.PostForm("title")
	artist := c.PostForm("artist")
	tags := c.PostForm("tags")

	if title == "" {
		response.ErrorJSON(c.Writer, "Title is required", http.StatusBadRequest)
		return
	}

	// Récupérer le fichier audio
	file, fileHeader, err := c.Request.FormFile("audio")
	if err != nil {
		response.ErrorJSON(c.Writer, "Audio file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// TODO: Implémenter la logique de sauvegarde du fichier
	// Pour l'instant, réponse de placeholder
	track := map[string]interface{}{
		"id":          1,
		"title":       title,
		"artist":      artist,
		"tags":        tags,
		"filename":    fileHeader.Filename,
		"uploader_id": userID,
		"created_at":  "2025-01-01T00:00:00Z",
	}

	response.SuccessJSON(c.Writer, track, "Track uploaded successfully")
}

// ListTracks liste toutes les pistes
func (h *Handler) ListTracks(c *gin.Context) {
	// TODO: Implémenter la récupération depuis la base de données
	tracks := []map[string]interface{}{
		{
			"id":       1,
			"title":    "Sample Track",
			"artist":   "Sample Artist",
			"filename": "sample.mp3",
		},
	}

	response.SuccessJSON(c.Writer, tracks, "Tracks retrieved successfully")
}

// GetTrack récupère une piste spécifique
func (h *Handler) GetTrack(c *gin.Context) {
	idStr := c.Param("id")
	trackID, err := strconv.Atoi(idStr)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid track ID", http.StatusBadRequest)
		return
	}

	// TODO: Récupérer depuis la base de données
	track := map[string]interface{}{
		"id":       trackID,
		"title":    "Sample Track",
		"artist":   "Sample Artist",
		"filename": "sample.mp3",
	}

	response.SuccessJSON(c.Writer, track, "Track retrieved successfully")
}

// UpdateTrack met à jour une piste
func (h *Handler) UpdateTrack(c *gin.Context) {
	idStr := c.Param("id")
	trackID, err := strconv.Atoi(idStr)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid track ID", http.StatusBadRequest)
		return
	}

	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	var req struct {
		Title  string `json:"title"`
		Artist string `json:"artist"`
		Tags   string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data", http.StatusBadRequest)
		return
	}

	// TODO: Vérifier que l'utilisateur est propriétaire + mise à jour en BDD
	track := map[string]interface{}{
		"id":          trackID,
		"title":       req.Title,
		"artist":      req.Artist,
		"tags":        req.Tags,
		"uploader_id": userID,
		"updated_at":  "2025-01-01T00:00:00Z",
	}

	response.SuccessJSON(c.Writer, track, "Track updated successfully")
}

// DeleteTrack supprime une piste
func (h *Handler) DeleteTrack(c *gin.Context) {
	idStr := c.Param("id")
	trackID, err := strconv.Atoi(idStr)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid track ID", http.StatusBadRequest)
		return
	}

	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	// TODO: Vérifier propriétaire + suppression BDD + fichier
	_ = trackID
	_ = userID

	response.SuccessJSON(c.Writer, nil, "Track deleted successfully")
}