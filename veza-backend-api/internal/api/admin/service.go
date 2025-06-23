package admin

import (
	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/models"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

func (s *Service) IsAdmin(userID int) bool {
	var role string
	err := s.db.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)
	if err != nil {
		return false
	}
	return role == "admin" || role == "super_admin"
}

func (s *Service) GetDashboardStats() (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// Basic stats
	s.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_active = true").Scan(&stats.TotalUsers)
	s.db.QueryRow("SELECT COUNT(*) FROM tracks").Scan(&stats.TotalTracks)
	s.db.QueryRow("SELECT COUNT(*) FROM listings WHERE status = 'open'").Scan(&stats.ActiveListings)

	return stats, nil
}

func (s *Service) GetUsers(page, limit int, search, role string) ([]models.UserAnalytics, int, error) {
	// TODO: Implement based on doc_admin_handler.md
	return []models.UserAnalytics{}, 0, nil
}

func (s *Service) GetAnalytics() (*models.ContentAnalytics, error) {
	// TODO: Implement based on doc_admin_handler.md
	return &models.ContentAnalytics{}, nil
}

func (s *Service) GetCategories() ([]interface{}, error) {
	// TODO: Implement categories
	return []interface{}{}, nil
}
