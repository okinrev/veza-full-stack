package chat

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/database"
)

type Handler struct {
	db *database.DB
}

func NewHandler(db *database.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) GetDmHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	targetID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID utilisateur invalide"})
		return
	}

	messages, err := h.db.GetDMMessages(userID, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *Handler) GetPublicRoomsHandler(c *gin.Context) {
	rooms, err := h.db.GetPublicRooms()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des salons"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}

func (h *Handler) CreateRoomHandler(c *gin.Context) {
	var room struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IsPrivate   bool   `json:"is_private"`
	}

	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides"})
		return
	}

	userID := c.GetInt("user_id")
	roomID, err := h.db.CreateRoom(room.Name, room.Description, room.IsPrivate, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du salon"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"room_id": roomID})
}

func (h *Handler) GetRoomMessagesHandler(c *gin.Context) {
	roomID := c.Param("room")
	userID := c.GetInt("user_id")

	// Vérifier si l'utilisateur a accès au salon
	hasAccess, err := h.db.HasRoomAccess(userID, roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la vérification des droits d'accès"})
		return
	}
	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Accès non autorisé à ce salon"})
		return
	}

	messages, err := h.db.GetRoomMessages(roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}
