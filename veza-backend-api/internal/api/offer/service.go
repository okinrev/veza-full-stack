package offer

import (
	"github.com/okinrev/veza-web-app/internal/database"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// TODO: Implement methods based on doc_offer_handler.md
