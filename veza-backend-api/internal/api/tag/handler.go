// internal/api/tag/handler.go
package tag

import (
	"net/http"
	"github.com/okinrev/veza-web-app/internal/utils/response"  // ADD THIS

	"github.com/gin-gonic/gin"
)

// Dans search/handler.go, tag/handler.go
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetAllTags récupère tous les tags
func (h *Handler) GetAllTags(c *gin.Context) {
	// TODO: Récupérer depuis la BDD
	tags := []map[string]interface{}{
		{"id": 1, "name": "electronic"},
		{"id": 2, "name": "ambient"},
		{"id": 3, "name": "techno"},
	}

	response.SuccessJSON(c.Writer, tags, "Tags retrieved")
}

// SearchTags recherche des tags
func (h *Handler) SearchTags(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.ErrorJSON(c.Writer, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// TODO: Recherche de tags
	tags := []map[string]interface{}{
		{"id": 1, "name": query + "tronic"},
	}

	response.SuccessJSON(c.Writer, tags, "Tags search completed")
}