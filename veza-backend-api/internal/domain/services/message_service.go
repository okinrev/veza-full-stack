package services

import (
	"context"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
	"github.com/okinrev/veza-web-app/internal/domain/repositories"
)

// MessageService définit les opérations métier pour les messages
type MessageService interface {
	// CRUD basique
	Create(ctx context.Context, message *entities.Message) error
	GetByID(ctx context.Context, id int64) (*entities.Message, error)
	Update(ctx context.Context, message *entities.Message) error
	Delete(ctx context.Context, id int64) error

	// Recherche et listing
	List(ctx context.Context, filters repositories.MessageFilters) ([]*entities.Message, error)
	Count(ctx context.Context, filters repositories.MessageFilters) (int64, error)
	Search(ctx context.Context, query string, limit int) ([]*entities.Message, error)

	// Gestion des réactions
	AddReaction(ctx context.Context, messageID, userID int64, reaction string) error
	RemoveReaction(ctx context.Context, messageID, userID int64, reaction string) error
	GetReactions(ctx context.Context, messageID int64) ([]*entities.Reaction, error)

	// Statistiques
	GetMessageStats(ctx context.Context, messageID int64) (*repositories.MessageStats, error)
	GetTotalMessages(ctx context.Context) (int64, error)
	GetMessagesToday(ctx context.Context) (int64, error)
}
