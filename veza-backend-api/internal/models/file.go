// internal/models/file.go
package models

import (
	"database/sql"
	"time"
)

// File represents a file attached to a user product (manuals, warranties, etc.)
type File struct {
	ID         int            `db:"id" json:"id"`
	ProductID  int            `db:"product_id" json:"product_id"` // References user_products.id
	Filename   string         `db:"filename" json:"filename"`
	URL        string         `db:"url" json:"url"`
	Type       string         `db:"type" json:"type"` // manual, warranty, invoice, image, document
	Size       sql.NullInt64  `db:"size" json:"size,omitempty"`
	MimeType   sql.NullString `db:"mime_type" json:"mime_type,omitempty"`
	UploadedAt time.Time      `db:"uploaded_at" json:"uploaded_at"`
	UpdatedAt  time.Time      `db:"updated_at" json:"updated_at"`
}

// InternalDocument represents internal documentation for a user product
type InternalDocument struct {
	ID         int            `db:"id" json:"id"`
	ProductID  int            `db:"product_id" json:"product_id"` // References user_products.id
	Title      string         `db:"title" json:"title"`
	Filename   string         `db:"filename" json:"filename"`
	URL        string         `db:"url" json:"url"`
	Type       string         `db:"type" json:"type"` // manual, warranty, invoice, notes
	Size       sql.NullInt64  `db:"size" json:"size,omitempty"`
	MimeType   sql.NullString `db:"mime_type" json:"mime_type,omitempty"`
	UploadedAt time.Time      `db:"uploaded_at" json:"uploaded_at"`
	UpdatedAt  time.Time      `db:"updated_at" json:"updated_at"`
}

// ProductDocument represents official documentation for products in the catalog
type ProductDocument struct {
	ID          int            `db:"id" json:"id"`
	ProductID   int            `db:"product_id" json:"product_id"` // References products.id
	Name        string         `db:"name" json:"name"`
	Description sql.NullString `db:"description" json:"description,omitempty"`
	FileType    string         `db:"file_type" json:"file_type"` // manual, datasheet, warranty, image
	FilePath    string         `db:"file_path" json:"file_path"`
	FileSize    int64          `db:"file_size" json:"file_size"`
	MimeType    sql.NullString `db:"mime_type" json:"mime_type,omitempty"`
	UploadedAt  time.Time      `db:"uploaded_at" json:"uploaded_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
}

// FileResponse represents file data for API responses
type FileResponse struct {
	ID         int            `json:"id"`
	ProductID  int            `json:"product_id"`
	Filename   string         `json:"filename"`
	URL        string         `json:"url"`
	Type       string         `json:"type"`
	Size       sql.NullInt64  `json:"size,omitempty"`
	MimeType   sql.NullString `json:"mime_type,omitempty"`
	UploadedAt time.Time      `json:"uploaded_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// InternalDocResponse represents internal document data for API responses
type InternalDocResponse struct {
	ID         int            `json:"id"`
	ProductID  int            `json:"product_id"`
	Title      string         `json:"title"`
	Filename   string         `json:"filename"`
	URL        string         `json:"url"`
	Type       string         `json:"type"`
	Size       sql.NullInt64  `json:"size,omitempty"`
	MimeType   sql.NullString `json:"mime_type,omitempty"`
	UploadedAt time.Time      `json:"uploaded_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}