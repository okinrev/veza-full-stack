//file: internal/admin/services/product_service.go

package services

import (
	"context"
	"fmt"

	"github.com/okinrev/veza-web-app/internal/admin/repository"
	"github.com/okinrev/veza-web-app/internal/models"
	"github.com/okinrev/veza-web-app/pkg/validator"
)

type ProductService struct {
	repo      repository.ProductRepository
	validator *validator.Validator
}

func NewProductService(repo repository.ProductRepository, v *validator.Validator) *ProductService {
	return &ProductService{
		repo:      repo,
		validator: v,
	}
}

// GetAllProducts récupère tous les produits avec filtres et pagination
func (s *ProductService) GetAllProducts(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.Product, int, error) {
	// Validation des paramètres
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	products, err := s.repo.GetAllWithFilters(ctx, filters, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get products: %w", err)
	}

	total, err := s.repo.CountWithFilters(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	return products, total, nil
}

// GetProductByID récupère un produit par ID
func (s *ProductService) GetProductByID(ctx context.Context, id int64) (*models.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return product, nil
}

// CreateProduct crée un nouveau produit
func (s *ProductService) CreateProduct(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error) {
	// Validation
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Vérifier que la catégorie existe si spécifiée
	if req.CategoryID > 0 {
		exists, err := s.repo.CategoryExists(ctx, int64(req.CategoryID))
		if err != nil {
			return nil, fmt.Errorf("failed to check category: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("category with ID %d does not exist", req.CategoryID)
		}
	}

	// Créer le produit
	product, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// UpdateProduct met à jour un produit
func (s *ProductService) UpdateProduct(ctx context.Context, id int64, req *models.UpdateProductRequest) (*models.Product, error) {
	// Validation
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Vérifier que le produit existe
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check product existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("product with ID %d does not exist", id)
	}

	// Vérifier la catégorie si spécifiée
	if req.CategoryID != nil && *req.CategoryID > 0 {
		exists, err := s.repo.CategoryExists(ctx, int64(*req.CategoryID))
		if err != nil {
			return nil, fmt.Errorf("failed to check category: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("category with ID %d does not exist", *req.CategoryID)
		}
	}

	// Mettre à jour
	product, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

// DeleteProduct supprime un produit
func (s *ProductService) DeleteProduct(ctx context.Context, id int64) error {
	// Vérifier que le produit existe
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check product existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("product with ID %d does not exist", id)
	}

	// Vérifier si le produit est utilisé
	inUse, err := s.repo.IsInUse(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if product is in use: %w", err)
	}
	if inUse {
		return fmt.Errorf("cannot delete product: it is currently in use by users")
	}

	// Supprimer
	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// DuplicateProduct duplique un produit
func (s *ProductService) DuplicateProduct(ctx context.Context, id int64) (*models.Product, error) {
	// Récupérer le produit original
	original, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get original product: %w", err)
	}

	// Créer une copie
	req := &models.CreateProductRequest{
		Name:        original.Name + " (Copy)",
		Description: "",
		Price:       0,
		CategoryID:  0,
		Brand:       "",
		Status:      "inactive",
	}
	if original.Description.Valid {
		req.Description = original.Description.String
	}
	if original.Price.Valid {
		req.Price = original.Price.Float64
	}
	if original.CategoryID.Valid {
		req.CategoryID = int(original.CategoryID.Int32)
	}
	if original.Brand.Valid {
		req.Brand = original.Brand.String
	}

	return s.CreateProduct(ctx, req)
}

// GetProductAnalytics récupère les analytics d'un produit
func (s *ProductService) GetProductAnalytics(ctx context.Context, id int64) (map[string]interface{}, error) {
	// Stub pour les analytics
	analytics := map[string]interface{}{
		"views":     0,
		"purchases": 0,
		"rating":    0.0,
	}
	return analytics, nil
}

// BulkUpdateProducts met à jour plusieurs produits
func (s *ProductService) BulkUpdateProducts(ctx context.Context, productIDs []int64, updates *models.UpdateProductRequest) error {
	for _, id := range productIDs {
		_, err := s.UpdateProduct(ctx, id, updates)
		if err != nil {
			return fmt.Errorf("failed to update product %d: %w", id, err)
		}
	}
	return nil
}

// ExportProducts exporte les produits
func (s *ProductService) ExportProducts(ctx context.Context, filters map[string]interface{}) ([]byte, error) {
	// Stub pour l'export
	return []byte("CSV data would be here"), nil
}
