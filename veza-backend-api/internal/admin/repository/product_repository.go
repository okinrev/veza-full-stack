package repository

import (
	"database/sql"

	"github.com/google/uuid"
)

type ProductRepository interface {
	Create(product *admin.Product) error
	GetByID(id uuid.UUID) (*admin.Product, error)
	GetAll() ([]*admin.Product, error)
	Update(product *admin.Product) error
	Delete(id uuid.UUID) error
	GetByCategory(category string) ([]*admin.Product, error)
}

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *admin.Product) error {
	query := `
        INSERT INTO products (id, name, description, price, category, user_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
    `

	_, err := r.db.Exec(
		query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.Category,
		product.UserID,
	)

	return err
}

func (r *productRepository) GetByID(id uuid.UUID) (*admin.Product, error) {
	query := `
        SELECT id, name, description, price, category, user_id, created_at, updated_at
        FROM products WHERE id = $1
    `

	product := &admin.Product{}
	err := r.db.QueryRow(query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Category,
		&product.UserID,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return product, nil
}

func (r *productRepository) GetAll() ([]*admin.Product, error) {
	query := `
        SELECT id, name, description, price, category, user_id, created_at, updated_at
        FROM products ORDER BY created_at DESC
    `

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*admin.Product
	for rows.Next() {
		product := &admin.Product{}
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Category,
			&product.UserID,
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

func (r *productRepository) Update(product *admin.Product) error {
	query := `
        UPDATE products 
        SET name = $2, description = $3, price = $4, category = $5, updated_at = NOW()
        WHERE id = $1
    `

	_, err := r.db.Exec(
		query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.Category,
	)

	return err
}

func (r *productRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *productRepository) GetByCategory(category string) ([]*admin.Product, error) {
	query := `
        SELECT id, name, description, price, category, user_id, created_at, updated_at
        FROM products WHERE category = $1 ORDER BY created_at DESC
    `

	rows, err := r.db.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*admin.Product
	for rows.Next() {
		product := &admin.Product{}
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Category,
			&product.UserID,
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
