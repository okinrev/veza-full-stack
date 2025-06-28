package repository

import (
	"database/sql"

	"github.com/okinrev/veza-web-app/internal/models"
)

type ProductRepository interface {
	Create(product *models.Product) error
	GetByID(id int) (*models.Product, error)
	GetAll() ([]*models.Product, error)
	Update(product *models.Product) error
	Delete(id int) error
	GetByCategory(categoryID int) ([]*models.Product, error)
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *models.Product) error {
	query := `
        INSERT INTO products (name, description, price, category_id, brand, model, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
        RETURNING id
    `

	err := r.db.QueryRow(
		query,
		product.Name,
		product.Description,
		product.Price,
		product.CategoryID,
		product.Brand,
		product.Model,
		product.Status,
	).Scan(&product.ID)

	return err
}

func (r *productRepository) GetByID(id int) (*models.Product, error) {
	query := `
        SELECT id, name, description, price, category_id, brand, model, status, created_at, updated_at
        FROM products WHERE id = $1
    `

	product := &models.Product{}
	err := r.db.QueryRow(query, id).Scan(
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

func (r *productRepository) GetAll() ([]*models.Product, error) {
	query := `
        SELECT id, name, description, price, category_id, brand, model, status, created_at, updated_at
        FROM products ORDER BY created_at DESC
    `

	rows, err := r.db.Query(query)
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

func (r *productRepository) Update(product *models.Product) error {
	query := `
        UPDATE products 
        SET name = $2, description = $3, price = $4, category_id = $5, brand = $6, model = $7, status = $8, updated_at = NOW()
        WHERE id = $1
    `

	_, err := r.db.Exec(
		query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.CategoryID,
		product.Brand,
		product.Model,
		product.Status,
	)

	return err
}

func (r *productRepository) Delete(id int) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *productRepository) GetByCategory(categoryID int) ([]*models.Product, error) {
	query := `
        SELECT id, name, description, price, category_id, brand, model, status, created_at, updated_at
        FROM products WHERE category_id = $1 ORDER BY created_at DESC
    `

	rows, err := r.db.Query(query, categoryID)
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
