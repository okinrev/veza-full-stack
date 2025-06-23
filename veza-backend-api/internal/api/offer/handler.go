// internal/api/offer/handler.go
package offer

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

// CreateOffer crée une nouvelle offre
func (h *Handler) CreateOffer(c *gin.Context) {
	listingIDStr := c.Param("id")
	listingID, err := strconv.Atoi(listingIDStr)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid listing ID", http.StatusBadRequest)
		return
	}

	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	var req struct {
		ProposedProductID int    `json:"proposed_product_id" binding:"required"`
		Message           string `json:"message"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data", http.StatusBadRequest)
		return
	}

	// TODO: Sauvegarder en BDD
	offer := map[string]interface{}{
		"id":                  1,
		"listing_id":          listingID,
		"from_user_id":        userID,
		"proposed_product_id": req.ProposedProductID,
		"message":             req.Message,
		"status":              "pending",
		"created_at":          "2025-01-01T00:00:00Z",
	}

	response.SuccessJSON(c.Writer, offer, "Offer created successfully")
}

// AcceptOffer accepte une offre
func (h *Handler) AcceptOffer(c *gin.Context) {
	offerIDStr := c.Param("id")
	offerID, err := strconv.Atoi(offerIDStr)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid offer ID", http.StatusBadRequest)
		return
	}

	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User ID not found", http.StatusUnauthorized)
		return
	}

	// TODO: Vérifier que l'utilisateur est propriétaire du listing + accepter l'offre
	_ = offerID
	_ = userID

	response.SuccessJSON(c.Writer, nil, "Offer accepted successfully")
}