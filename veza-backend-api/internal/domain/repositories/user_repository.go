package repositories

import (
	"context"

	"github.com/okinrev/veza-web-app/internal/domain/entities"
)

// UserRepository définit les opérations de persistance pour les utilisateurs
type UserRepository interface {
	// CRUD basique
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id int64) (*entities.User, error)
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id int64) error

	// Recherche et listing
	List(ctx context.Context, filters UserFilters) ([]*entities.User, error)
	Count(ctx context.Context, filters UserFilters) (int64, error)
	Search(ctx context.Context, query string, limit int) ([]*entities.User, error)

	// Vérifications d'unicité
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// Gestion des sessions et tokens
	SaveRefreshToken(ctx context.Context, userID int64, token string, expiresAt int64) error
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllUserTokens(ctx context.Context, userID int64) error

	// Statistiques
	GetUserStats(ctx context.Context, userID int64) (*UserStats, error)
	GetTotalUsers(ctx context.Context) (int64, error)
	GetActiveUsers(ctx context.Context) (int64, error)
	GetNewUsersToday(ctx context.Context) (int64, error)
}

// UserFilters filtres pour la recherche d'utilisateurs
type UserFilters struct {
	Role          *entities.UserRole   `json:"role,omitempty"`
	Status        *entities.UserStatus `json:"status,omitempty"`
	IsActive      *bool                `json:"is_active,omitempty"`
	IsVerified    *bool                `json:"is_verified,omitempty"`
	CreatedAfter  *int64               `json:"created_after,omitempty"`
	CreatedBefore *int64               `json:"created_before,omitempty"`
	Search        string               `json:"search,omitempty"`

	// Pagination
	Limit  int `json:"limit"`
	Offset int `json:"offset"`

	// Tri
	OrderBy string `json:"order_by"` // id, username, email, created_at
	Order   string `json:"order"`    // asc, desc
}

// RefreshToken représente un token de rafraîchissement
type RefreshToken struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	CreatedAt int64  `json:"created_at"`
}

// UserStats statistiques utilisateur
type UserStats struct {
	UserID        int64 `json:"user_id"`
	TotalRooms    int64 `json:"total_rooms"`
	TotalMessages int64 `json:"total_messages"`
	TotalTracks   int64 `json:"total_tracks"`
	TotalListings int64 `json:"total_listings"`
	LastActivity  int64 `json:"last_activity"`
}
