// internal/api/search/handler.go
package search

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

// GlobalSearchHandler recherche globale
func (h *Handler) GlobalSearchHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.ErrorJSON(c.Writer, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// TODO: Recherche dans toutes les tables
	results := map[string]interface{}{
		"tracks":    []map[string]interface{}{},
		"resources": []map[string]interface{}{},
		"users":     []map[string]interface{}{},
		"listings":  []map[string]interface{}{},
	}

	response.SuccessJSON(c.Writer, results, "Search completed")
}

// AdvancedSearchHandler recherche avancée
func (h *Handler) AdvancedSearchHandler(c *gin.Context) {
	query := c.Query("q")
	searchType := c.Query("type")
	tags := c.Query("tags")

	// TODO: Recherche avancée avec filtres
	results := []map[string]interface{}{
		{
			"type":  searchType,
			"query": query,
			"tags":  tags,
		},
	}

	response.SuccessJSON(c.Writer, results, "Advanced search completed")
}

// AutocompleteHandler auto-complétion
func (h *Handler) AutocompleteHandler(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.ErrorJSON(c.Writer, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// TODO: Auto-complétion
	suggestions := []string{
		query + " suggestion 1",
		query + " suggestion 2",
	}

	response.SuccessJSON(c.Writer, suggestions, "Autocomplete results")
}