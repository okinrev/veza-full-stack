package api

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/okinrev/veza-web-app/internal/config"
	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/middleware"

	"github.com/okinrev/veza-web-app/internal/api/admin"
	"github.com/okinrev/veza-web-app/internal/api/auth"
	"github.com/okinrev/veza-web-app/internal/api/chat"
	"github.com/okinrev/veza-web-app/internal/api/listing"
	"github.com/okinrev/veza-web-app/internal/api/message"
	"github.com/okinrev/veza-web-app/internal/api/offer"
	"github.com/okinrev/veza-web-app/internal/api/room"
	"github.com/okinrev/veza-web-app/internal/api/search"
	"github.com/okinrev/veza-web-app/internal/api/shared_resources"
	"github.com/okinrev/veza-web-app/internal/api/tag"
	"github.com/okinrev/veza-web-app/internal/api/track"
	"github.com/okinrev/veza-web-app/internal/api/user"
)

// APIRouter gère la configuration des routes de l'API
type APIRouter struct {
	db     *database.DB
	config *config.Config
	engine *gin.Engine
}

// NewAPIRouter crée une nouvelle instance de APIRouter
func NewAPIRouter(db *database.DB, cfg *config.Config) *APIRouter {
	return &APIRouter{
		db:     db,
		config: cfg,
	}
}

// Setup configure toutes les routes de l'API
func (r *APIRouter) Setup(router *gin.Engine) {
	r.engine = router

	// Middlewares globaux
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.RateLimiter(100, time.Minute)) // 100 requêtes par minute par IP

	// Groupe API v1 (nouveau frontend React)
	v1 := router.Group("/api/v1")
	{
		r.setupAuthRoutes(v1)
		r.setupUserRoutes(v1)
		r.setupAdminRoutes(v1)
		r.setupTrackRoutes(v1)
		r.setupListingRoutes(v1)
		r.setupOfferRoutes(v1)
		r.setupMessageRoutes(v1)
		r.setupRoomRoutes(v1)
		r.setupSearchRoutes(v1)
		r.setupTagRoutes(v1)
		r.setupSharedResourcesRoutes(v1)
		r.setupChatRoutes(v1)
	}

	// Note: Frontend basique modifié pour utiliser /api/v1/ au lieu de routes legacy
}

// Méthodes de configuration des routes par module
func (r *APIRouter) setupAuthRoutes(router *gin.RouterGroup) {
	authService := auth.NewService(r.db, r.config.JWT.Secret)
	authHandler := auth.NewHandler(authService)
	auth.SetupRoutes(router, authHandler, r.config.JWT.Secret)
}

func (r *APIRouter) setupUserRoutes(router *gin.RouterGroup) {
	userService := user.NewService(r.db)
	userHandler := user.NewHandler(userService)
	user.SetupRoutes(router, userHandler, r.config.JWT.Secret)
}

func (r *APIRouter) setupAdminRoutes(router *gin.RouterGroup) {
	adminService := admin.NewService(r.db)
	adminHandler := admin.NewHandler(adminService)
	admin.SetupRoutes(router, adminHandler, r.config.JWT.Secret)
}

func (r *APIRouter) setupTrackRoutes(router *gin.RouterGroup) {
	trackService := track.NewService(r.db, r.config.JWT.Secret)
	trackHandler := track.NewHandler(trackService)
	track.SetupRoutes(router, trackHandler, r.config.JWT.Secret)
}

func (r *APIRouter) setupListingRoutes(router *gin.RouterGroup) {
	listingService := listing.NewService(r.db)
	listingHandler := listing.NewHandler(listingService)
	listing.SetupRoutes(router, listingHandler, r.config.JWT.Secret)
}

func (r *APIRouter) setupOfferRoutes(router *gin.RouterGroup) {
	offerService := offer.NewService(r.db)
	offerHandler := offer.NewHandler(offerService)
	offer.SetupRoutes(router, offerHandler, r.config.JWT.Secret)
}

func (r *APIRouter) setupMessageRoutes(router *gin.RouterGroup) {
	messageService := message.NewService(r.db)
	messageHandler := message.NewHandler(messageService)
	message.SetupRoutes(router, messageHandler, r.config.JWT.Secret)
}

func (r *APIRouter) setupRoomRoutes(router *gin.RouterGroup) {
	roomService := room.NewService(r.db)
	roomHandler := room.NewHandler(roomService)
	room.SetupRoutes(router, roomHandler, r.config.JWT.Secret)
}

func (r *APIRouter) setupSearchRoutes(router *gin.RouterGroup) {
	searchService := search.NewService(r.db)
	searchHandler := search.NewHandler(searchService)
	search.SetupRoutes(router, searchHandler, r.config.JWT.Secret)
}

func (r *APIRouter) setupTagRoutes(router *gin.RouterGroup) {
	tagService := tag.NewService(r.db)
	tagHandler := tag.NewHandler(tagService)
	tag.SetupRoutes(router, tagHandler, r.config.JWT.Secret)
}

func (r *APIRouter) setupSharedResourcesRoutes(router *gin.RouterGroup) {
	sharedResourcesService := shared_resources.NewService(r.db)
	sharedResourcesHandler := shared_resources.NewHandler(sharedResourcesService)
	shared_resources.SetupRoutes(router, sharedResourcesHandler, r.config.JWT.Secret)
}

func (r *APIRouter) setupChatRoutes(router *gin.RouterGroup) {
	chatHandler := chat.NewHandler(r.db)
	chat.SetupRoutes(router, chatHandler, r.config.JWT.Secret)
}

// SetupRoutes configure toutes les routes API (pour la compatibilité)
func SetupRoutes(router *gin.Engine, db *database.DB, cfg *config.Config) {
	apiRouter := NewAPIRouter(db, cfg)
	apiRouter.Setup(router)
}
