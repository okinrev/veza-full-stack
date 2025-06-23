package message

import (
	"github.com/okinrev/veza-web-app/internal/database"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// TODO: Implement methods
