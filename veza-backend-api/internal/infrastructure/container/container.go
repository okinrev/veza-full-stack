package container

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/okinrev/veza-web-app/internal/adapters/postgres"
	"github.com/okinrev/veza-web-app/internal/adapters/redis_cache"
	"github.com/okinrev/veza-web-app/internal/domain/repositories"
	"github.com/okinrev/veza-web-app/internal/domain/services"
	"github.com/okinrev/veza-web-app/internal/infrastructure/config"
	"github.com/okinrev/veza-web-app/internal/ports/http"
)

// Container contient toutes les dépendances de l'application
type Container struct {
	// Configuration
	Config *config.AppConfig

	// Infrastructure
	Logger      *zap.Logger
	Database    *sql.DB
	RedisClient *redis.Client
	Metrics     *prometheus.Registry

	// Clients externes (optionnels pour maintenant)
	// ChatClient   grpc.ChatClient
	// StreamClient grpc.StreamClient

	// Repositories (utilisant les interfaces)
	UserRepository    repositories.UserRepository
	RoomRepository    repositories.RoomRepository
	MessageRepository repositories.MessageRepository
	TrackRepository   repositories.TrackRepository
	ListingRepository repositories.ListingRepository

	// Cache
	CacheService repositories.CacheRepository

	// Services (domain)
	AuthService    services.AuthService
	UserService    services.UserService
	RoomService    services.RoomService
	MessageService services.MessageService
	TrackService   services.TrackService
	ListingService services.ListingService
	SearchService  services.SearchService

	// Handlers (ports)
	AuthHandler    http.AuthHandler
	UserHandler    http.UserHandler
	RoomHandler    http.RoomHandler
	MessageHandler http.MessageHandler
	TrackHandler   http.TrackHandler
	ListingHandler http.ListingHandler
	SearchHandler  http.SearchHandler
	AdminHandler   http.AdminHandler
}

// New crée une nouvelle instance du conteneur avec toutes les dépendances
func New(cfg *config.AppConfig) (*Container, error) {
	container := &Container{
		Config: cfg,
	}

	// Initialisation dans l'ordre des dépendances
	if err := container.initializeInfrastructure(); err != nil {
		return nil, fmt.Errorf("erreur initialisation infrastructure: %w", err)
	}

	if err := container.initializeRepositories(); err != nil {
		return nil, fmt.Errorf("erreur initialisation repositories: %w", err)
	}

	if err := container.initializeServices(); err != nil {
		return nil, fmt.Errorf("erreur initialisation services: %w", err)
	}

	if err := container.initializeHandlers(); err != nil {
		return nil, fmt.Errorf("erreur initialisation handlers: %w", err)
	}

	log.Printf("✅ Container initialisé avec succès")
	return container, nil
}

// initializeInfrastructure initialise les composants d'infrastructure
func (c *Container) initializeInfrastructure() error {
	// Logger
	var err error
	if c.Config.IsDevelopment() {
		c.Logger, err = zap.NewDevelopment()
	} else {
		c.Logger, err = zap.NewProduction()
	}
	if err != nil {
		return fmt.Errorf("erreur initialisation logger: %w", err)
	}

	// Base de données PostgreSQL
	c.Database, err = postgres.NewConnection(c.Config.Database)
	if err != nil {
		return fmt.Errorf("erreur connexion base de données: %w", err)
	}

	// Métriques Prometheus
	c.Metrics = prometheus.NewRegistry()

	// Redis (optionnel)
	if c.Config.Redis.Enabled {
		c.RedisClient, err = redis_cache.NewClient(c.Config.Redis)
		if err != nil {
			c.Logger.Warn("Redis non disponible, fonctionnement sans cache", zap.Error(err))
		} else {
			// Test de connexion Redis
			ctx := context.Background()
			if err := c.RedisClient.Ping(ctx).Err(); err != nil {
				c.Logger.Warn("Test connexion Redis échoué", zap.Error(err))
			} else {
				c.Logger.Info("✅ Redis connecté avec succès")
			}
		}
	}

	// Clients gRPC (optionnels)
	if c.Config.GRPC.Enabled {
		// c.ChatClient, err = grpc.NewChatClient(c.Config.GRPC.ChatServer)
		// if err != nil {
		// 	c.Logger.Warn("Chat gRPC client non disponible", zap.Error(err))
		// }

		// c.StreamClient, err = grpc.NewStreamClient(c.Config.GRPC.StreamServer)
		// if err != nil {
		// 	c.Logger.Warn("Stream gRPC client non disponible", zap.Error(err))
		// }
	}

	c.Logger.Info("✅ Infrastructure initialisée")
	return nil
}

// initializeRepositories initialise les repositories
func (c *Container) initializeRepositories() error {
	var err error

	// Cache service
	if c.RedisClient != nil {
		c.CacheService = redis_cache.NewCacheService(c.RedisClient, c.Config.Redis.DefaultTTL)
	}

	// Repositories
	c.UserRepository, err = postgres.NewUserRepository(c.Database, c.Logger)
	if err != nil {
		return fmt.Errorf("erreur création user repository: %w", err)
	}

	c.RoomRepository, err = postgres.NewRoomRepository(c.Database, c.Logger)
	if err != nil {
		return fmt.Errorf("erreur création room repository: %w", err)
	}

	c.MessageRepository, err = postgres.NewMessageRepository(c.Database, c.Logger)
	if err != nil {
		return fmt.Errorf("erreur création message repository: %w", err)
	}

	c.TrackRepository, err = postgres.NewTrackRepository(c.Database, c.Logger)
	if err != nil {
		return fmt.Errorf("erreur création track repository: %w", err)
	}

	c.ListingRepository, err = postgres.NewListingRepository(c.Database, c.Logger)
	if err != nil {
		return fmt.Errorf("erreur création listing repository: %w", err)
	}

	c.Logger.Info("✅ Repositories initialisés")
	return nil
}

// initializeServices initialise les services métier
func (c *Container) initializeServices() error {
	var err error

	// Services
	c.AuthService, err = services.NewAuthService(
		c.UserRepository,
		c.CacheService,
		c.Config.JWT,
		c.Logger,
	)
	if err != nil {
		return fmt.Errorf("erreur création auth service: %w", err)
	}

	c.UserService, err = services.NewUserService(
		c.UserRepository,
		c.CacheService,
		c.Logger,
	)
	if err != nil {
		return fmt.Errorf("erreur création user service: %w", err)
	}

	c.RoomService, err = services.NewRoomService(
		c.RoomRepository,
		c.UserRepository,
		c.CacheService,
		// c.ChatClient,
		c.Logger,
	)
	if err != nil {
		return fmt.Errorf("erreur création room service: %w", err)
	}

	c.MessageService, err = services.NewMessageService(
		c.MessageRepository,
		c.RoomRepository,
		c.UserRepository,
		c.CacheService,
		// c.ChatClient,
		c.Logger,
	)
	if err != nil {
		return fmt.Errorf("erreur création message service: %w", err)
	}

	c.TrackService, err = services.NewTrackService(
		c.TrackRepository,
		c.UserRepository,
		c.CacheService,
		// c.StreamClient,
		c.Logger,
	)
	if err != nil {
		return fmt.Errorf("erreur création track service: %w", err)
	}

	c.ListingService, err = services.NewListingService(
		c.ListingRepository,
		c.UserRepository,
		c.CacheService,
		c.Logger,
	)
	if err != nil {
		return fmt.Errorf("erreur création listing service: %w", err)
	}

	c.SearchService, err = services.NewSearchService(
		c.UserRepository,
		c.TrackRepository,
		c.ListingRepository,
		c.CacheService,
		c.Logger,
	)
	if err != nil {
		return fmt.Errorf("erreur création search service: %w", err)
	}

	c.Logger.Info("✅ Services initialisés")
	return nil
}

// initializeHandlers initialise les handlers HTTP
func (c *Container) initializeHandlers() error {
	var err error

	// Handlers
	c.AuthHandler, err = http.NewAuthHandler(c.AuthService, c.Config, c.Logger, c.Metrics)
	if err != nil {
		return fmt.Errorf("erreur création auth handler: %w", err)
	}

	c.UserHandler, err = http.NewUserHandler(c.UserService, c.Config, c.Logger, c.Metrics)
	if err != nil {
		return fmt.Errorf("erreur création user handler: %w", err)
	}

	c.RoomHandler, err = http.NewRoomHandler(c.RoomService, c.Config, c.Logger, c.Metrics)
	if err != nil {
		return fmt.Errorf("erreur création room handler: %w", err)
	}

	c.MessageHandler, err = http.NewMessageHandler(c.MessageService, c.Config, c.Logger, c.Metrics)
	if err != nil {
		return fmt.Errorf("erreur création message handler: %w", err)
	}

	c.TrackHandler, err = http.NewTrackHandler(c.TrackService, c.Config, c.Logger, c.Metrics)
	if err != nil {
		return fmt.Errorf("erreur création track handler: %w", err)
	}

	c.ListingHandler, err = http.NewListingHandler(c.ListingService, c.Config, c.Logger, c.Metrics)
	if err != nil {
		return fmt.Errorf("erreur création listing handler: %w", err)
	}

	c.SearchHandler, err = http.NewSearchHandler(c.SearchService, c.Config, c.Logger, c.Metrics)
	if err != nil {
		return fmt.Errorf("erreur création search handler: %w", err)
	}

	c.AdminHandler, err = http.NewAdminHandler(
		c.UserService,
		c.RoomService,
		c.TrackService,
		c.Config,
		c.Logger,
		c.Metrics,
	)
	if err != nil {
		return fmt.Errorf("erreur création admin handler: %w", err)
	}

	c.Logger.Info("✅ Handlers initialisés")
	return nil
}

// Close ferme toutes les connexions
func (c *Container) Close() {
	if c.Database != nil {
		c.Database.Close()
	}
	if c.RedisClient != nil {
		c.RedisClient.Close()
	}
	if c.Logger != nil {
		c.Logger.Sync()
	}
}
