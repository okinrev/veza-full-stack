package track

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

// RouteGroup représente un groupe de routes pour le module track
type RouteGroup struct {
	handler *Handler
	secret  string
}

// NewRouteGroup crée une nouvelle instance de RouteGroup
func NewRouteGroup(handler *Handler, jwtSecret string) *RouteGroup {
	return &RouteGroup{
		handler: handler,
		secret:  jwtSecret,
	}
}

// Register enregistre toutes les routes du module track
func (rg *RouteGroup) Register(router *gin.RouterGroup) {
	// Groupe principal des tracks
	tracks := router.Group("/tracks")
	{
		// Routes publiques
		rg.registerPublicRoutes(tracks)

		// Routes protégées
		rg.registerProtectedRoutes(tracks)
	}
}

// registerPublicRoutes enregistre les routes publiques
func (rg *RouteGroup) registerPublicRoutes(router *gin.RouterGroup) {
	// GET /api/v1/tracks - Liste des tracks
	router.GET("", rg.handler.ListTracks)

	// GET /api/v1/tracks/:id - Détails d'un track
	router.GET("/:id", rg.handler.GetTrack)
}

// registerProtectedRoutes enregistre les routes protégées
func (rg *RouteGroup) registerProtectedRoutes(router *gin.RouterGroup) {
	protected := router.Group("")
	protected.Use(middleware.JWTAuthMiddleware(rg.secret))
	{
		// POST /api/v1/tracks - Ajout d'un nouveau track avec upload
		protected.POST("", rg.handler.AddTrackWithUpload)

		// PUT /api/v1/tracks/:id - Mise à jour d'un track
		protected.PUT("/:id", rg.handler.UpdateTrack)

		// DELETE /api/v1/tracks/:id - Suppression d'un track
		protected.DELETE("/:id", rg.handler.DeleteTrack)
	}
}

// SetupRoutes configure les routes du module track (pour la compatibilité)
func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	rg := NewRouteGroup(handler, jwtSecret)
	rg.Register(router)
}
