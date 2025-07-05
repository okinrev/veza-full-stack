package repositories

import (
	"context"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
)

// MessageRepository définit les opérations de persistance pour les messages
type MessageRepository interface {
	// CRUD basique
	Create(ctx context.Context, message *entities.Message) error
	GetByID(ctx context.Context, id int64) (*entities.Message, error)
	Update(ctx context.Context, message *entities.Message) error
	Delete(ctx context.Context, id int64) error

	// Recherche et listing
	List(ctx context.Context, filters MessageFilters) ([]*entities.Message, error)
	Count(ctx context.Context, filters MessageFilters) (int64, error)
	Search(ctx context.Context, query string, limit int) ([]*entities.Message, error)

	// Gestion des réactions
	AddReaction(ctx context.Context, messageID, userID int64, reaction string) error
	RemoveReaction(ctx context.Context, messageID, userID int64, reaction string) error
	GetReactions(ctx context.Context, messageID int64) ([]*entities.Reaction, error)

	// Statistiques
	GetMessageStats(ctx context.Context, messageID int64) (*MessageStats, error)
	GetTotalMessages(ctx context.Context) (int64, error)
	GetMessagesToday(ctx context.Context) (int64, error)
}

// MessageFilters filtres pour la recherche de messages
type MessageFilters struct {
	RoomID        *int64                  `json:"room_id,omitempty"`
	UserID        *int64                  `json:"user_id,omitempty"`
	Type          *entities.MessageType   `json:"type,omitempty"`
	Status        *entities.MessageStatus `json:"status,omitempty"`
	CreatedAfter  *int64                  `json:"created_after,omitempty"`
	CreatedBefore *int64                  `json:"created_before,omitempty"`
	Search        string                  `json:"search,omitempty"`

	// Pagination
	Limit  int `json:"limit"`
	Offset int `json:"offset"`

	// Tri
	OrderBy string `json:"order_by"` // id, created_at, updated_at
	Order   string `json:"order"`    // asc, desc
}

// MessageStats statistiques d'un message
type MessageStats struct {
	MessageID     int64 `json:"message_id"`
	ReactionCount int64 `json:"reaction_count"`
	ReplyCount    int64 `json:"reply_count"`
	ViewCount     int64 `json:"view_count"`
}

// Reaction représente une réaction à un message
type Reaction struct {
	ID        int64  `json:"id"`
	MessageID int64  `json:"message_id"`
	UserID    int64  `json:"user_id"`
	Reaction  string `json:"reaction"`
	CreatedAt int64  `json:"created_at"`
}
