// internal/models/product.go
package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// ProductSpecifications represents JSON specifications for a product
type ProductSpecifications map[string]interface{}

// Value implements the driver.Valuer interface for JSON marshaling
func (ps ProductSpecifications) Value() (driver.Value, error) {
	return json.Marshal(ps)
}

// Scan implements the sql.Scanner interface for JSON unmarshaling
func (ps *ProductSpecifications) Scan(value interface{}) error {
	if value == nil {
		*ps = make(ProductSpecifications)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ps)
	case string:
		return json.Unmarshal([]byte(v), ps)
	default:
		return errors.New("cannot scan ProductSpecifications")
	}
}

// Product represents a product in the catalog
type Product struct {
	ID                   int                    `db:"id" json:"id"`
	Name                 string                 `db:"name" json:"name"`
	CategoryID           sql.NullInt32          `db:"category_id" json:"category_id,omitempty"`
	Brand                sql.NullString         `db:"brand" json:"brand,omitempty"`
	Model                sql.NullString         `db:"model" json:"model,omitempty"`
	Description          sql.NullString         `db:"description" json:"description,omitempty"`
	Price                sql.NullFloat64        `db:"price" json:"price,omitempty"`
	WarrantyMonths       sql.NullInt32          `db:"warranty_months" json:"warranty_months,omitempty"`
	WarrantyConditions   sql.NullString         `db:"warranty_conditions" json:"warranty_conditions,omitempty"`
	ManufacturerWebsite  sql.NullString         `db:"manufacturer_website" json:"manufacturer_website,omitempty"`
	Specifications       ProductSpecifications  `db:"specifications" json:"specifications,omitempty"`
	Status               string                 `db:"status" json:"status"` // active, discontinued, draft
	CreatedAt            time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time              `db:"updated_at" json:"updated_at"`
}

// ProductWithCategory represents a product with category information
type ProductWithCategory struct {
	Product
	CategoryName       sql.NullString `db:"category_name" json:"category_name,omitempty"`
	DocumentationCount int            `db:"documentation_count" json:"documentation_count,omitempty"`
	UserCount          int            `db:"user_count" json:"user_count,omitempty"`
}

// Category represents a product category
type Category struct {
	ID           int            `db:"id" json:"id"`
	Name         string         `db:"name" json:"name"`
	Description  sql.NullString `db:"description" json:"description,omitempty"`
	Icon         sql.NullString `db:"icon" json:"icon,omitempty"`
	Color        sql.NullString `db:"color" json:"color,omitempty"`
	SortOrder    int            `db:"sort_order" json:"sort_order"`
	IsActive     bool           `db:"is_active" json:"is_active"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
}

// CategoryWithCount represents a category with product count
type CategoryWithCount struct {
	Category
	ProductCount int `db:"product_count" json:"product_count,omitempty"`
}

// UserProduct represents a product owned by a user
type UserProduct struct {
	ID              int            `db:"id" json:"id"`
	UserID          int            `db:"user_id" json:"user_id"`
	ProductID       int            `db:"product_id" json:"product_id"`
	Version         sql.NullString `db:"version" json:"version,omitempty"`
	PurchaseDate    sql.NullTime   `db:"purchase_date" json:"purchase_date,omitempty"`
	WarrantyExpires sql.NullTime   `db:"warranty_expires" json:"warranty_expires,omitempty"`
	PurchasePrice   sql.NullInt32  `db:"purchase_price" json:"purchase_price,omitempty"`
	SerialNumber    sql.NullString `db:"serial_number" json:"serial_number,omitempty"`
	Notes           sql.NullString `db:"notes" json:"notes,omitempty"`
	Status          string         `db:"status" json:"status"` // active, sold, broken, lost
	CreatedAt       time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at" json:"updated_at"`
}

// UserProductWithDetails represents a user product with full product details
type UserProductWithDetails struct {
	UserProduct
	ProductName    string         `db:"product_name" json:"product_name,omitempty"`
	CategoryName   sql.NullString `db:"category_name" json:"category_name,omitempty"`
	Brand          sql.NullString `db:"brand" json:"brand,omitempty"`
	Model          sql.NullString `db:"model" json:"model,omitempty"`
	FilesCount     int            `db:"files_count" json:"files_count,omitempty"`
	DocsCount      int            `db:"docs_count" json:"docs_count,omitempty"`
	IsUnderWarranty bool          `json:"is_under_warranty,omitempty"` // Computed field
}