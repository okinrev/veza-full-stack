// internal/api/user/handler.go
package user

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

// GetMe récupère le profil de l'utilisateur connecté
func (h *Handler) GetMe(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		response.ErrorJSON(c.Writer, "User not found", http.StatusNotFound)
		return
	}

	response.SuccessJSON(c.Writer, user, "User profile retrieved")
}

// UpdateMe met à jour le profil de l'utilisateur connecté
func (h *Handler) UpdateMe(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data", http.StatusBadRequest)
		return
	}

	user, err := h.service.UpdateUser(userID, req)
	if err != nil {
		response.ErrorJSON(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	response.SuccessJSON(c.Writer, user, "Profile updated successfully")
}

// ChangePassword change le mot de passe de l'utilisateur
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data", http.StatusBadRequest)
		return
	}

	err := h.service.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		response.ErrorJSON(c.Writer, err.Error(), http.StatusBadRequest)
		return
	}

	response.SuccessJSON(c.Writer, nil, "Password changed successfully")
}

// GetUsers liste tous les utilisateurs
func (h *Handler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	users, total, err := h.service.GetUsers(page, limit, search)
	if err != nil {
		response.ErrorJSON(c.Writer, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	meta := &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}

	response.PaginatedJSON(c.Writer, users, meta, "Users retrieved successfully")
}

// GetUsersExceptMe liste tous les utilisateurs sauf l'utilisateur connecté
func (h *Handler) GetUsersExceptMe(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	// Ajouter le filtre pour exclure l'utilisateur actuel
	users, total, err := h.service.GetUsers(page, limit, search)
	if err != nil {
		response.ErrorJSON(c.Writer, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	// Filtrer l'utilisateur connecté
	filteredUsers := []UserResponse{}
	for _, user := range users {
		if user.ID != userID {
			filteredUsers = append(filteredUsers, user)
		}
	}

	meta := &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total - 1, // -1 car on exclut l'utilisateur connecté
		TotalPages: (total + limit - 2) / limit,
	}

	response.PaginatedJSON(c.Writer, filteredUsers, meta, "Users retrieved successfully")
}

// SearchUsers recherche des utilisateurs
func (h *Handler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.ErrorJSON(c.Writer, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	users, total, err := h.service.GetUsers(page, limit, query)
	if err != nil {
		response.ErrorJSON(c.Writer, "Failed to search users", http.StatusInternalServerError)
		return
	}

	meta := &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}

	response.PaginatedJSON(c.Writer, users, meta, "Search results")
}

func (h *Handler) GetUserAvatar(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.Atoi(idStr)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.service.GetUserByID(userID)
	if err != nil {
		response.ErrorJSON(c.Writer, "User not found", http.StatusNotFound)
		return
	}

	// ✅ Correct way to handle sql.NullString
	if !user.Avatar.Valid || user.Avatar.String == "" {
		response.ErrorJSON(c.Writer, "No avatar found", http.StatusNotFound)
		return
	}

	// Rediriger vers l'URL de l'avatar ou servir le fichier
	c.Redirect(http.StatusFound, user.Avatar.String)
}
