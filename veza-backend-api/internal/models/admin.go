// internal/models/admin.go
package models

import (
	"database/sql"
	"time"
)

// DashboardStats represents admin dashboard statistics
type DashboardStats struct {
	TotalUsers           int       `db:"total_users" json:"total_users"`
	ActiveUsers          int       `db:"active_users" json:"active_users"`
	TotalTracks          int       `db:"total_tracks" json:"total_tracks"`
	PublicTracks         int       `db:"public_tracks" json:"public_tracks"`
	TotalSharedResources int       `db:"total_shared_resources" json:"total_shared_resources"`
	TotalListings        int       `db:"total_listings" json:"total_listings"`
	ActiveListings       int       `db:"active_listings" json:"active_listings"`
	TotalOffers          int       `db:"total_offers" json:"total_offers"`
	PendingOffers        int       `db:"pending_offers" json:"pending_offers"`
	TotalMessages        int       `db:"total_messages" json:"total_messages"`
	TotalRooms           int       `db:"total_rooms" json:"total_rooms"`
	TotalProducts        int       `db:"total_products" json:"total_products"`
	TotalCategories      int       `db:"total_categories" json:"total_categories"`
	LastUpdated          time.Time `json:"last_updated"`
}

// UserAnalytics represents detailed user analytics for admin
type UserAnalytics struct {
	UserID           int          `db:"user_id" json:"user_id"`
	Username         string       `db:"username" json:"username"`
	Email            string       `db:"email" json:"email"`
	Role             string       `db:"role" json:"role"`
	TracksCount      int          `db:"tracks_count" json:"tracks_count"`
	ResourcesCount   int          `db:"resources_count" json:"resources_count"`
	ListingsCount    int          `db:"listings_count" json:"listings_count"`
	MessagesCount    int          `db:"messages_count" json:"messages_count"`
	ProductsCount    int          `db:"products_count" json:"products_count"`
	RegistrationDate time.Time    `db:"registration_date" json:"registration_date"`
	LastActivity     sql.NullTime `db:"last_activity" json:"last_activity,omitempty"`
	IsActive         bool         `db:"is_active" json:"is_active"`
	StorageUsed      int64        `db:"storage_used" json:"storage_used,omitempty"`
}

// ContentAnalytics represents content analytics
type ContentAnalytics struct {
	TracksByMonth    []MonthlyCount  `json:"tracks_by_month"`
	ResourcesByMonth []MonthlyCount  `json:"resources_by_month"`
	UsersByMonth     []MonthlyCount  `json:"users_by_month"`
	PopularTags      []TagCount      `json:"popular_tags"`
	TopUploaders     []UploaderStats `json:"top_uploaders"`
	CategoryStats    []CategoryStats `json:"category_stats,omitempty"`
}

// MonthlyCount represents count data by month
type MonthlyCount struct {
	Month string `db:"month" json:"month"`
	Count int    `db:"count" json:"count"`
}

// TagCount represents tag usage statistics
type TagCount struct {
	Tag   string `db:"tag" json:"tag"`
	Count int    `db:"count" json:"count"`
}

// UploaderStats represents uploader statistics
type UploaderStats struct {
	UserID         int    `db:"user_id" json:"user_id"`
	Username       string `db:"username" json:"username"`
	TracksCount    int    `db:"tracks_count" json:"tracks_count"`
	ResourcesCount int    `db:"resources_count" json:"resources_count"`
	TotalUploads   int    `db:"total_uploads" json:"total_uploads"`
	TotalDownloads int    `db:"total_downloads" json:"total_downloads"`
}

// CategoryStats represents category statistics
type CategoryStats struct {
	CategoryID   int    `db:"category_id" json:"category_id"`
	CategoryName string `db:"category_name" json:"category_name"`
	ProductCount int    `db:"product_count" json:"product_count"`
	UserCount    int    `db:"user_count" json:"user_count"`
}

// SystemHealth represents system health metrics
type SystemHealth struct {
	DatabaseStatus    string        `json:"database_status"`
	StorageUsed       int64         `json:"storage_used"`
	StorageAvailable  int64         `json:"storage_available"`
	MemoryUsage       float64       `json:"memory_usage"`
	CPUUsage          float64       `json:"cpu_usage"`
	ActiveConnections int           `json:"active_connections"`
	Uptime            time.Duration `json:"uptime"`
	LastBackup        sql.NullTime  `json:"last_backup,omitempty"`
	ErrorCount        int           `json:"error_count"`
	LastChecked       time.Time     `json:"last_checked"`
}

// AuditLog represents admin audit log entries
type AuditLog struct {
	ID           int            `db:"id" json:"id"`
	UserID       int            `db:"user_id" json:"user_id"`
	Action       string         `db:"action" json:"action"`
	ResourceType string         `db:"resource_type" json:"resource_type"`
	ResourceID   sql.NullInt32  `db:"resource_id" json:"resource_id,omitempty"`
	Details      sql.NullString `db:"details" json:"details,omitempty"`
	IPAddress    sql.NullString `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent    sql.NullString `db:"user_agent" json:"user_agent,omitempty"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
}

// AuditLogWithUser represents audit log with user information
type AuditLogWithUser struct {
	AuditLog
	Username string `db:"username" json:"username,omitempty"`
	UserRole string `db:"user_role" json:"user_role,omitempty"`
}

// AdminSettings represents system settings manageable by admin
type AdminSettings struct {
	ID          int            `db:"id" json:"id"`
	Key         string         `db:"key" json:"key"`
	Value       string         `db:"value" json:"value"`
	Type        string         `db:"type" json:"type"` // string, int, bool, json
	Description sql.NullString `db:"description" json:"description,omitempty"`
	Category    string         `db:"category" json:"category"` // system, features, limits, etc.
	IsPublic    bool           `db:"is_public" json:"is_public"`
	UpdatedBy   sql.NullInt32  `db:"updated_by" json:"updated_by,omitempty"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
}

// ProductRequest types for admin operations
type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description string  `json:"description" validate:"max=500"`
	Price       float64 `json:"price" validate:"min=0"`
	CategoryID  int     `json:"category_id" validate:"required,min=1"`
	Brand       string  `json:"brand" validate:"max=50"`
	Status      string  `json:"status" validate:"required,oneof=active inactive"`
}

type UpdateProductRequest struct {
	Name        *string  `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=500"`
	Price       *float64 `json:"price,omitempty" validate:"omitempty,min=0"`
	CategoryID  *int     `json:"category_id,omitempty" validate:"omitempty,min=1"`
	Brand       *string  `json:"brand,omitempty" validate:"omitempty,max=50"`
	Status      *string  `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
}

type BulkUpdateRequest struct {
	ProductIDs []int                `json:"product_ids" validate:"required,min=1"`
	Updates    UpdateProductRequest `json:"updates"`
}

// Product est défini dans models/product.go
