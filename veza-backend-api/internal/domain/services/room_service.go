package services

import (
	"context"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
	"github.com/okinrev/veza-web-app/internal/domain/repositories"
)

// RoomService définit les opérations métier pour les salles de chat
type RoomService interface {
	// CRUD basique
	Create(ctx context.Context, room *entities.Room) error
	GetByID(ctx context.Context, id int64) (*entities.Room, error)
	GetByUUID(ctx context.Context, uuid string) (*entities.Room, error)
	Update(ctx context.Context, room *entities.Room) error
	Delete(ctx context.Context, id int64) error

	// Recherche et listing
	List(ctx context.Context, filters repositories.RoomFilters) ([]*entities.Room, error)
	Count(ctx context.Context, filters repositories.RoomFilters) (int64, error)
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
	GetRoomStats(ctx context.Context, roomID int64) (*repositories.RoomStats, error)
	GetTotalRooms(ctx context.Context) (int64, error)
	GetActiveRooms(ctx context.Context) (int64, error)
}
