// internal/models/search.go
package models

import (
	"database/sql"
	"time"
)

// Tag represents a tag used for categorizing content
type Tag struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	UsageCount int      `db:"usage_count" json:"usage_count,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// GlobalSearchResult represents the result of a global search
type GlobalSearchResult struct {
	Users           []UserSearchResult           `json:"users"`
	Tracks          []TrackSearchResult          `json:"tracks"`
	SharedResources []SharedResourceSearchResult `json:"shared_resources"`
	Products        []ProductSearchResult        `json:"products,omitempty"`
	TotalResults    int                          `json:"total_results"`
	Query           string                       `json:"query"`
	SearchTime      float64                      `json:"search_time,omitempty"`
}

// UserSearchResult represents a user in search results
type UserSearchResult struct {
	ID        int            `json:"id"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	FirstName sql.NullString `json:"first_name,omitempty"`
	LastName  sql.NullString `json:"last_name,omitempty"`
	Avatar    sql.NullString `json:"avatar,omitempty"`
	Role      string         `json:"role"`
}

// TrackSearchResult represents a track in search results
type TrackSearchResult struct {
	ID              int      `json:"id"`
	Title           string   `json:"title"`
	Artist          string   `json:"artist"`
	Filename        string   `json:"filename"`
	Tags            []string `json:"tags"`
	UploaderID      int      `json:"uploader_id"`
	UploaderName    string   `json:"uploader_name"`
	CreatedAt       string   `json:"created_at"`
	StreamURL       string   `json:"stream_url"`
	DurationSeconds int      `json:"duration_seconds,omitempty"`
}

// SharedResourceSearchResult represents a shared resource in search results
type SharedResourceSearchResult struct {
	ID               int      `json:"id"`
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	Filename         string   `json:"filename"`
	Type             string   `json:"type"`
	Tags             []string `json:"tags"`
	UploaderID       int      `json:"uploader_id"`
	UploaderUsername string   `json:"uploader_username"`
	UploadedAt       string   `json:"uploaded_at"`
	DownloadURL      string   `json:"download_url"`
	DownloadCount    int      `json:"download_count"`
}

// ProductSearchResult represents a product in search results
type ProductSearchResult struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Brand        sql.NullString `json:"brand,omitempty"`
	Model        sql.NullString `json:"model,omitempty"`
	Description  sql.NullString `json:"description,omitempty"`
	CategoryName sql.NullString `json:"category_name,omitempty"`
	Price        sql.NullFloat64 `json:"price,omitempty"`
}

// AutocompleteResult represents autocomplete suggestions
type AutocompleteResult struct {
	Tags     []string `json:"tags"`
	Artists  []string `json:"artists"`
	Users    []string `json:"users"`
	Types    []string `json:"types"`
	Products []string `json:"products,omitempty"`
}

// SuggestionResponse represents a suggestion with metadata
type SuggestionResponse struct {
	Type      string  `json:"type"`
	Value     string  `json:"value"`
	Count     int     `json:"count,omitempty"`
	Score     float64 `json:"score,omitempty"`
	Trending  bool    `json:"trending,omitempty"`
	Context   map[string]string `json:"context,omitempty"`
	Related   []string `json:"related,omitempty"`
}

// SearchQuery represents a search query for analytics
type SearchQuery struct {
	ID        int            `db:"id" json:"id"`
	UserID    sql.NullInt32  `db:"user_id" json:"user_id,omitempty"`
	Query     string         `db:"query" json:"query"`
	Type      string         `db:"type" json:"type"` // global, tracks, resources, users
	Results   int            `db:"results" json:"results"`
	IPAddress sql.NullString `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent sql.NullString `db:"user_agent" json:"user_agent,omitempty"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
}

// PopularSearch represents popular search terms
type PopularSearch struct {
	Query      string    `db:"query" json:"query"`
	SearchCount int      `db:"search_count" json:"search_count"`
	LastSearched time.Time `db:"last_searched" json:"last_searched"`
}

// SearchFilter represents search filters
type SearchFilter struct {
	Type        string         `json:"type,omitempty"`         // tracks, resources, users, products
	Tags        []string       `json:"tags,omitempty"`
	Category    sql.NullString `json:"category,omitempty"`
	DateFrom    sql.NullTime   `json:"date_from,omitempty"`
	DateTo      sql.NullTime   `json:"date_to,omitempty"`
	IsPublic    sql.NullBool   `json:"is_public,omitempty"`
	UploaderID  sql.NullInt32  `json:"uploader_id,omitempty"`
	SortBy      string         `json:"sort_by,omitempty"`      // created_at, updated_at, name, popularity
	SortOrder   string         `json:"sort_order,omitempty"`   // asc, desc
	Limit       int            `json:"limit,omitempty"`
	Offset      int            `json:"offset,omitempty"`
}