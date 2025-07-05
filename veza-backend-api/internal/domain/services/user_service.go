package services

import (
	"context"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
	"github.com/okinrev/veza-web-app/internal/domain/repositories"
)

// UserService définit les opérations métier pour les utilisateurs
type UserService interface {
	// CRUD basique
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id int64) (*entities.User, error)
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id int64) error

	// Recherche et listing
	List(ctx context.Context, filters repositories.UserFilters) ([]*entities.User, error)
	Count(ctx context.Context, filters repositories.UserFilters) (int64, error)
	Search(ctx context.Context, query string, limit int) ([]*entities.User, error)

	// Vérifications d'unicité
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// Gestion des sessions et tokens
	SaveRefreshToken(ctx context.Context, userID int64, token string, expiresAt int64) error
	GetRefreshToken(ctx context.Context, token string) (*repositories.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllUserTokens(ctx context.Context, userID int64) error

	// Statistiques
	GetUserStats(ctx context.Context, userID int64) (*repositories.UserStats, error)
	GetTotalUsers(ctx context.Context) (int64, error)
	GetActiveUsers(ctx context.Context) (int64, error)
	GetNewUsersToday(ctx context.Context) (int64, error)
}
