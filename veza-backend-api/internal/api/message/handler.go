// internal/api/message/handler.go
package message

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

// GetDmHandler récupère les messages directs avec un utilisateur
func (h *Handler) GetDmHandler(c *gin.Context) {
	userIDStr := c.Param("user_id")
	targetUserID, err := strconv.Atoi(userIDStr)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid user ID", http.StatusBadRequest)
		return
	}

	currentUserID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	// TODO: Récupérer les messages depuis la BDD
	messages := []map[string]interface{}{
		{
			"id":        1,
			"from_user": currentUserID,
			"to_user":   targetUserID,
			"content":   "Hello!",
			"timestamp": "2025-01-01T00:00:00Z",
		},
	}

	response.SuccessJSON(c.Writer, messages, "Direct messages retrieved")
}