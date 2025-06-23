//file: internal/models/admin/analytics.go

package handlers

import "time"

// DashboardStats statistiques pour le dashboard admin
type DashboardStats struct {
	TotalProducts         int     `json:"total_products" db:"total_products"`
	ActiveProducts        int     `json:"active_products" db:"active_products"`
	TotalCategories       int     `json:"total_categories" db:"total_categories"`
	TotalUsers            int     `json:"total_users" db:"total_users"`
	TotalUserProducts     int     `json:"total_user_products" db:"total_user_products"`
	ProductsWithDocs      int     `json:"products_with_docs" db:"products_with_docs"`
	ProductsWithWarranty  int     `json:"products_with_warranty" db:"products_with_warranty"`
	AvgWarrantyMonths     float64 `json:"avg_warranty_months" db:"avg_warranty_months"`
	LastUpdated           time.Time `json:"last_updated"`
}

// ProductAnalytics analytics détaillés par produit
type ProductAnalytics struct {
	ProductID        int64     `json:"product_id" db:"product_id"`
	ProductName      string    `json:"product_name" db:"product_name"`
	UserCount        int       `json:"user_count" db:"user_count"`
	DocumentCount    int       `json:"document_count" db:"document_count"`
	AvgRating        *float64  `json:"avg_rating" db:"avg_rating"`
	TotalDownloads   int       `json:"total_downloads" db:"total_downloads"`
	LastActivity     *time.Time `json:"last_activity" db:"last_activity"`
}

// UserAnalytics analytics par utilisateur
type UserAnalytics struct {
	UserID           int64     `json:"user_id" db:"user_id"`
	Username         string    `json:"username" db:"username"`
	ProductCount     int       `json:"product_count" db:"product_count"`
	DocumentCount    int       `json:"document_count" db:"document_count"`
	RegistrationDate time.Time `json:"registration_date" db:"registration_date"`
	LastActivity     *time.Time `json:"last_activity" db:"last_activity"`
	IsActive         bool      `json:"is_active" db:"is_active"`
}