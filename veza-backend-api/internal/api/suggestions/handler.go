// internal/api/suggestions/handler.go
package suggestions

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/utils/response"
)

type Service struct{}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetSuggestions récupère des suggestions
func (h *Handler) GetSuggestions(c *gin.Context) {
	suggestionType := c.Query("type")
	query := c.Query("q")
	_ = query

	var suggestions []map[string]interface{}

	switch suggestionType {
	case "tag":
		suggestions = []map[string]interface{}{
			{"type": "tag", "value": "electronic"},
			{"type": "tag", "value": "ambient"},
		}
	case "user":
		suggestions = []map[string]interface{}{
			{"type": "user", "value": "john_doe"},
		}
	default:
		suggestions = []map[string]interface{}{
			{"type": "general", "value": "suggestion"},
		}
	}

	response.SuccessJSON(c.Writer, suggestions, "Suggestions retrieved")
}
