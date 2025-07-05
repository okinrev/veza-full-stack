package repository

import (
	"context"
	"database/sql"

	"github.com/okinrev/veza-web-app/internal/models"
)

type ProductRepository interface {
	Create(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error)
	GetByID(ctx context.Context, id int64) (*models.Product, error)
	GetAllWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.Product, error)
	CountWithFilters(ctx context.Context, filters map[string]interface{}) (int, error)
	Update(ctx context.Context, id int64, req *models.UpdateProductRequest) (*models.Product, error)
	Delete(ctx context.Context, id int64) error
	Exists(ctx context.Context, id int64) (bool, error)
	CategoryExists(ctx context.Context, categoryID int64) (bool, error)
	IsInUse(ctx context.Context, id int64) (bool, error)
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, req *models.CreateProductRequest) (*models.Product, error) {
	query := `
        INSERT INTO products (name, description, price, category_id, brand, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
        RETURNING id
    `

	product := &models.Product{}
	err := r.db.QueryRowContext(ctx, query,
		req.Name,
		req.Description,
		req.Price,
		req.CategoryID,
		req.Brand,
		req.Status,
	).Scan(&product.ID)

	return product, err
}

func (r *productRepository) GetByID(ctx context.Context, id int64) (*models.Product, error) {
	query := `
        SELECT id, name, description, price, category_id, brand, model, status, created_at, updated_at
        FROM products WHERE id = $1
    `

	product := &models.Product{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.CategoryID,
		&product.Brand,
		&product.Model,
		&product.Status,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *productRepository) GetAllWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.Product, error) {
	query := `
        SELECT id, name, description, price, category_id, brand, model, status, created_at, updated_at
        FROM products ORDER BY created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.CategoryID,
			&product.Brand,
			&product.Model,
			&product.Status,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

func (r *productRepository) CountWithFilters(ctx context.Context, filters map[string]interface{}) (int, error) {
	query := `
        SELECT COUNT(*) FROM products
    `

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (r *productRepository) Update(ctx context.Context, id int64, req *models.UpdateProductRequest) (*models.Product, error) {
	query := `
        UPDATE products 
        SET name = $2, description = $3, price = $4, category_id = $5, brand = $6, status = $7, updated_at = NOW()
        WHERE id = $1
    `

	product := &models.Product{}
	err := r.db.QueryRowContext(ctx, query,
		id,
		req.Name,
		req.Description,
		req.Price,
		req.CategoryID,
		req.Brand,
		req.Status,
	).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.CategoryID,
		&product.Brand,
		&product.Status,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	return product, err
}

func (r *productRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *productRepository) Exists(ctx context.Context, id int64) (bool, error) {
	query := `
        SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)
    `

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}

func (r *productRepository) CategoryExists(ctx context.Context, categoryID int64) (bool, error) {
	query := `
        SELECT EXISTS (SELECT 1 FROM products WHERE category_id = $1)
    `

	var exists bool
	err := r.db.QueryRowContext(ctx, query, categoryID).Scan(&exists)
	return exists, err
}

func (r *productRepository) IsInUse(ctx context.Context, id int64) (bool, error) {
	query := `
        SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)
    `

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}
