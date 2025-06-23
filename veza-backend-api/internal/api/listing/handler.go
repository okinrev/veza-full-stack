// internal/api/listing/handler.go
package listing

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

// CreateListing crée un nouveau listing
func (h *Handler) CreateListing(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	var req struct {
		ProductID   int    `json:"product_id" binding:"required"`
		Description string `json:"description" binding:"required"`
		State       string `json:"state" binding:"required"`
		Price       *int   `json:"price"`
		ExchangeFor string `json:"exchange_for"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data", http.StatusBadRequest)
		return
	}

	// TODO: Sauvegarder en BDD
	listing := map[string]interface{}{
		"id":          1,
		"user_id":     userID,
		"product_id":  req.ProductID,
		"description": req.Description,
		"state":       req.State,
		"price":       req.Price,
		"exchange_for": req.ExchangeFor,
		"status":      "open",
		"created_at":  "2025-01-01T00:00:00Z",
	}

	response.SuccessJSON(c.Writer, listing, "Listing created successfully")
}

// GetAllListings récupère tous les listings
func (h *Handler) GetAllListings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	listings := []map[string]interface{}{
		{
			"id":          1,
			"description": "Sample listing",
			"state":       "excellent",
			"status":      "open",
		},
	}

	meta := &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      len(listings),
		TotalPages: 1,
	}

	response.PaginatedJSON(c.Writer, listings, meta, "Listings retrieved successfully")
}

// GetListingByID récupère un listing spécifique
func (h *Handler) GetListingByID(c *gin.Context) {
	idStr := c.Param("id")
	listingID, err := strconv.Atoi(idStr)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid listing ID", http.StatusBadRequest)
		return
	}

	// TODO: Récupérer depuis BDD
	listing := map[string]interface{}{
		"id":          listingID,
		"description": "Sample listing",
		"state":       "excellent",
		"status":      "open",
	}

	response.SuccessJSON(c.Writer, listing, "Listing retrieved successfully")
}

// DeleteListing supprime un listing
func (h *Handler) DeleteListing(c *gin.Context) {
	idStr := c.Param("id")
	listingID, err := strconv.Atoi(idStr)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid listing ID", http.StatusBadRequest)
		return
	}

	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	// TODO: Vérifier propriétaire + suppression
	_ = listingID
	_ = userID

	response.SuccessJSON(c.Writer, nil, "Listing deleted successfully")
}