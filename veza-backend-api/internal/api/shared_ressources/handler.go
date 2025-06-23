// internal/api/shared_ressources/handler.go
package shared_ressources

import (
	"net/http"
	"strconv"
	"github.com/okinrev/veza-web-app/internal/middleware"
	"github.com/okinrev/veza-web-app/internal/utils/response"  // ADD THIS
    "github.com/okinrev/veza-web-app/internal/common"

	"github.com/gin-gonic/gin"
)

// Dans search/handler.go, tag/handler.go
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// UploadSharedResource upload une nouvelle ressource
func (h *Handler) UploadSharedResource(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	title := c.PostForm("title")
	resourceType := c.PostForm("type")
	description := c.PostForm("description")
	tags := c.PostForm("tags")

	if title == "" {
		response.ErrorJSON(c.Writer, "Title is required", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		response.ErrorJSON(c.Writer, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// TODO: Sauvegarder le fichier et en BDD
	resource := map[string]interface{}{
		"id":          1,
		"title":       title,
		"type":        resourceType,
		"description": description,
		"tags":        tags,
		"filename":    fileHeader.Filename,
		"uploader_id": userID,
		"is_public":   true,
		"uploaded_at": "2025-01-01T00:00:00Z",
	}

	response.SuccessJSON(c.Writer, resource, "Resource uploaded successfully")
}

// ListSharedResources liste les ressources
func (h *Handler) ListSharedResources(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	resources := []map[string]interface{}{
		{
			"id":       1,
			"title":    "Sample Resource",
			"type":     "sample",
			"filename": "sample.zip",
		},
	}

	meta := &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      len(resources),
		TotalPages: 1,
	}

	response.PaginatedJSON(c.Writer, resources, meta, "Resources retrieved successfully")
}

// SearchSharedResources recherche dans les ressources
func (h *Handler) SearchSharedResources(c *gin.Context) {
	query := c.Query("q")
	resourceType := c.Query("type")
	tags := c.Query("tags")

	// TODO: Implémenter la recherche
	results := []map[string]interface{}{
		{
			"id":    1,
			"title": "Found Resource",
			"type":  resourceType,
		},
	}

	response.SuccessJSON(c.Writer, results, "Search completed")
}

// UpdateSharedResource met à jour une ressource
func (h *Handler) UpdateSharedResource(c *gin.Context) {
	idStr := c.Param("id")
	resourceID, err := strconv.Atoi(idStr)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid resource ID", http.StatusBadRequest)
		return
	}

	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Tags        string `json:"tags"`
		IsPublic    bool   `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data", http.StatusBadRequest)
		return
	}

	// TODO: Vérifier propriétaire + mise à jour
	resource := map[string]interface{}{
		"id":          resourceID,
		"title":       req.Title,
		"description": req.Description,
		"uploader_id": userID,
		"is_public":   req.IsPublic,
	}

	response.SuccessJSON(c.Writer, resource, "Resource updated successfully")
}

// DeleteSharedResource supprime une ressource
func (h *Handler) DeleteSharedResource(c *gin.Context) {
	idStr := c.Param("id")
	resourceID, err := strconv.Atoi(idStr)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid resource ID", http.StatusBadRequest)
		return
	}

	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	// TODO: Vérifier propriétaire + suppression
	_ = resourceID
	_ = userID

	response.SuccessJSON(c.Writer, nil, "Resource deleted successfully")
}

// ServeSharedFile sert un fichier partagé
func (h *Handler) ServeSharedFile(c *gin.Context) {
	filename := c.Param("filename")
	
	// TODO: Vérifier que le fichier existe et les permissions
	// Pour l'instant, retourner une erreur
	response.ErrorJSON(c.Writer, "File serving not implemented yet", http.StatusNotImplemented)
}