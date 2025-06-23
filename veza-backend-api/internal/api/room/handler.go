// internal/api/room/handler.go
package room

import (
	"net/http"
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

// GetPublicRoomsHandler liste les salons publics
func (h *Handler) GetPublicRoomsHandler(c *gin.Context) {
	// TODO: Récupérer depuis la BDD
	rooms := []map[string]interface{}{
		{
			"id":         1,
			"name":       "general",
			"is_private": false,
			"created_at": "2025-01-01T00:00:00Z",
		},
	}

	response.SuccessJSON(c.Writer, rooms, "Public rooms retrieved")
}

// CreateRoomHandler crée un nouveau salon
func (h *Handler) CreateRoomHandler(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IsPrivate   bool   `json:"is_private"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data", http.StatusBadRequest)
		return
	}

	// TODO: Sauvegarder en BDD
	room := map[string]interface{}{
		"id":          1,
		"name":        req.Name,
		"description": req.Description,
		"is_private":  req.IsPrivate,
		"creator_id":  userID,
		"created_at":  "2025-01-01T00:00:00Z",
	}

	response.SuccessJSON(c.Writer, room, "Room created successfully")
}

// GetRoomMessagesHandler récupère les messages d'un salon
func (h *Handler) GetRoomMessagesHandler(c *gin.Context) {
	roomName := c.Param("room")

	// TODO: Récupérer les messages depuis la BDD
	messages := []map[string]interface{}{
		{
			"id":        1,
			"from_user": 1,
			"room":      roomName,
			"content":   "Hello room!",
			"timestamp": "2025-01-01T00:00:00Z",
		},
	}

	response.SuccessJSON(c.Writer, messages, "Room messages retrieved")
}