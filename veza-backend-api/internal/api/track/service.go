package track

import (
	"github.com/okinrev/veza-web-app/internal/database"
)

type Service struct {
	db        *database.DB
	jwtSecret string
}

func NewService(db *database.DB, jwtSecret string) *Service {
	return &Service{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// TODO: Implement methods based on doc_track_handler.md