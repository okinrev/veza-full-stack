package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/common"
	"github.com/okinrev/veza-web-app/internal/utils/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Dashboard(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if !h.service.IsAdmin(userID) {
		response.ErrorJSON(c.Writer, "Admin access required", http.StatusForbidden)
		return
	}

	stats, err := h.service.GetDashboardStats()
	if err != nil {
		response.ErrorJSON(c.Writer, "Failed to get dashboard stats", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(c.Writer, stats, "Dashboard stats retrieved")
}

func (h *Handler) GetUsers(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if !h.service.IsAdmin(userID) {
		response.ErrorJSON(c.Writer, "Admin access required", http.StatusForbidden)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	role := c.Query("role")

	users, total, err := h.service.GetUsers(page, limit, search, role)
	if err != nil {
		response.ErrorJSON(c.Writer, "Failed to get users", http.StatusInternalServerError)
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

func (h *Handler) GetAnalytics(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if !h.service.IsAdmin(userID) {
		response.ErrorJSON(c.Writer, "Admin access required", http.StatusForbidden)
		return
	}

	analytics, err := h.service.GetAnalytics()
	if err != nil {
		response.ErrorJSON(c.Writer, "Failed to get analytics", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(c.Writer, analytics, "Analytics retrieved successfully")
}

func (h *Handler) GetCategories(c *gin.Context) {
	categories, err := h.service.GetCategories()
	if err != nil {
		response.ErrorJSON(c.Writer, "Failed to get categories", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(c.Writer, categories, "Categories retrieved")
}

// TODO: CreateCategory, UpdateCategory, DeleteCategory
