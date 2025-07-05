package repositories

import (
	"context"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
)

// RoomRepository définit les opérations de persistance pour les salles de chat
type RoomRepository interface {
	// CRUD basique
	Create(ctx context.Context, room *entities.Room) error
	GetByID(ctx context.Context, id int64) (*entities.Room, error)
	GetByUUID(ctx context.Context, uuid string) (*entities.Room, error)
	Update(ctx context.Context, room *entities.Room) error
	Delete(ctx context.Context, id int64) error

	// Recherche et listing
	List(ctx context.Context, filters RoomFilters) ([]*entities.Room, error)
	Count(ctx context.Context, filters RoomFilters) (int64, error)
	Search(ctx context.Context, query string, limit int) ([]*entities.Room, error)

	// Gestion des membres
	AddMember(ctx context.Context, roomID, userID int64) error
	RemoveMember(ctx context.Context, roomID, userID int64) error
	GetMembers(ctx context.Context, roomID int64) ([]*entities.User, error)
	IsMember(ctx context.Context, roomID, userID int64) (bool, error)

	// Gestion des messages
	GetMessages(ctx context.Context, roomID int64, limit, offset int) ([]*entities.Message, error)
	GetMessageCount(ctx context.Context, roomID int64) (int64, error)

	// Statistiques
	GetRoomStats(ctx context.Context, roomID int64) (*RoomStats, error)
	GetTotalRooms(ctx context.Context) (int64, error)
	GetActiveRooms(ctx context.Context) (int64, error)
}

// RoomFilters filtres pour la recherche de salles
type RoomFilters struct {
	Type          *entities.RoomType    `json:"type,omitempty"`
	Privacy       *entities.RoomPrivacy `json:"privacy,omitempty"`
	Status        *entities.RoomStatus  `json:"status,omitempty"`
	CreatorID     *int64                `json:"creator_id,omitempty"`
	CreatedAfter  *int64                `json:"created_after,omitempty"`
	CreatedBefore *int64                `json:"created_before,omitempty"`
	Search        string                `json:"search,omitempty"`

	// Pagination
	Limit  int `json:"limit"`
	Offset int `json:"offset"`

	// Tri
	OrderBy string `json:"order_by"` // id, name, created_at, member_count
	Order   string `json:"order"`    // asc, desc
}

// RoomStats statistiques d'une salle
type RoomStats struct {
	RoomID       int64 `json:"room_id"`
	MemberCount  int64 `json:"member_count"`
	MessageCount int64 `json:"message_count"`
	LastActivity int64 `json:"last_activity"`
}
