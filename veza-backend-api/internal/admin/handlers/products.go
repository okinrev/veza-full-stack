//file: internal/admin/handlers/products.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/okinrev/veza-web-app/internal/admin/services"
	"github.com/okinrev/veza-web-app/internal/models"
	"github.com/okinrev/veza-web-app/internal/utils/response"
	"github.com/okinrev/veza-web-app/pkg/logger"
)

type ProductHandler struct {
	service *services.ProductService
	logger  *logger.Logger
}

func NewProductHandler(service *services.ProductService, logger *logger.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		logger:  logger,
	}
}

// GetProducts récupère tous les produits avec filtres et pagination
func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	// Filtres
	filters := make(map[string]interface{})
	if category := query.Get("category"); category != "" {
		categoryID, err := strconv.ParseInt(category, 10, 64)
		if err == nil {
			filters["category_id"] = categoryID
		}
	}
	if status := query.Get("status"); status != "" {
		filters["status"] = status
	}
	if search := query.Get("search"); search != "" {
		filters["search"] = search
	}
	if brand := query.Get("brand"); brand != "" {
		filters["brand"] = brand
	}

	products, total, err := h.service.GetAllProducts(ctx, filters, page, limit)
	if err != nil {
		h.logger.Error("Failed to get products", "error", err)
		response.ErrorJSON(w, "Failed to retrieve products", http.StatusInternalServerError)
		return
	}

	// Métadonnées de pagination
	if limit == 0 {
		limit = 20
	}
	totalPages := (total + limit - 1) / limit

	meta := &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: totalPages,
	}

	response.PaginatedJSON(w, products, meta, "Products retrieved successfully")
}

// GetProduct récupère un produit par ID
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		response.ErrorJSON(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := h.service.GetProductByID(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get product", "id", id, "error", err)
		response.ErrorJSON(w, "Product not found", http.StatusNotFound)
		return
	}

	response.SuccessJSON(w, product, "Product retrieved successfully")
}

// CreateProduct crée un nouveau produit
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	product, err := h.service.CreateProduct(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create product", "error", err)
		response.ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("Product created successfully", "id", product.ID, "name", product.Name)
	response.SuccessJSON(w, product, "Product created successfully")
}

// UpdateProduct met à jour un produit
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		response.ErrorJSON(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	product, err := h.service.UpdateProduct(ctx, id, &req)
	if err != nil {
		h.logger.Error("Failed to update product", "id", id, "error", err)
		response.ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("Product updated successfully", "id", id)
	response.SuccessJSON(w, product, "Product updated successfully")
}

// DeleteProduct supprime un produit
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		response.ErrorJSON(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteProduct(ctx, id)
	if err != nil {
		h.logger.Error("Failed to delete product", "id", id, "error", err)
		response.ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("Product deleted successfully", "id", id)
	response.SuccessJSON(w, nil, "Product deleted successfully")
}

// DuplicateProduct duplique un produit existant
func (h *ProductHandler) DuplicateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		response.ErrorJSON(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := h.service.DuplicateProduct(ctx, id)
	if err != nil {
		h.logger.Error("Failed to duplicate product", "id", id, "error", err)
		response.ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("Product duplicated successfully", "original_id", id, "new_id", product.ID)
	response.SuccessJSON(w, product, "Product duplicated successfully")
}

// GetProductAnalytics récupère les analytics d'un produit
func (h *ProductHandler) GetProductAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		response.ErrorJSON(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	analytics, err := h.service.GetProductAnalytics(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get product analytics", "id", id, "error", err)
		response.ErrorJSON(w, "Failed to retrieve analytics", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(w, analytics, "Analytics retrieved successfully")
}

// BulkUpdateProducts met à jour plusieurs produits à la fois
func (h *ProductHandler) BulkUpdateProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		ProductIDs []int64                     `json:"product_ids" validate:"required,min=1"`
		Updates    models.UpdateProductRequest `json:"updates"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ErrorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.ProductIDs) == 0 {
		response.ErrorJSON(w, "Product IDs are required", http.StatusBadRequest)
		return
	}

	results, err := h.service.BulkUpdateProducts(ctx, req.ProductIDs, &req.Updates)
	if err != nil {
		h.logger.Error("Failed to bulk update products", "error", err)
		response.ErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Info("Bulk update completed", "updated_count", len(results))
	response.SuccessJSON(w, results, "Products updated successfully")
}

// ExportProducts exporte les produits en CSV
func (h *ProductHandler) ExportProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters pour les filtres
	query := r.URL.Query()
	filters := make(map[string]interface{})

	if category := query.Get("category"); category != "" {
		categoryID, err := strconv.ParseInt(category, 10, 64)
		if err == nil {
			filters["category_id"] = categoryID
		}
	}
	if status := query.Get("status"); status != "" {
		filters["status"] = status
	}

	csvData, err := h.service.ExportProducts(ctx, filters)
	if err != nil {
		h.logger.Error("Failed to export products", "error", err)
		response.ErrorJSON(w, "Failed to export products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=products.csv")
	w.Write(csvData)
}
