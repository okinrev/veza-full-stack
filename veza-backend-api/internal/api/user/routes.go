package user

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

// RouteGroup représente un groupe de routes pour le module utilisateur
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

// Register enregistre toutes les routes du module utilisateur
func (rg *RouteGroup) Register(router *gin.RouterGroup) {
	// Groupe principal des utilisateurs
	users := router.Group("/users")
	{
		// Routes publiques
		rg.registerPublicRoutes(users)
		
		// Routes protégées
		rg.registerProtectedRoutes(users)
	}
}

// registerPublicRoutes enregistre les routes publiques
func (rg *RouteGroup) registerPublicRoutes(router *gin.RouterGroup) {
	// GET /api/v1/users - Liste des utilisateurs
	router.GET("", rg.handler.GetUsers)
	
	// GET /api/v1/users/:id/avatar - Avatar d'un utilisateur
	router.GET("/:id/avatar", rg.handler.GetUserAvatar)
}

// registerProtectedRoutes enregistre les routes protégées
func (rg *RouteGroup) registerProtectedRoutes(router *gin.RouterGroup) {
	protected := router.Group("")
	protected.Use(middleware.JWTAuthMiddleware(rg.secret))
	{
		// GET /api/v1/users/me - Informations de l'utilisateur connecté
		protected.GET("/me", rg.handler.GetMe)
		
		// PUT /api/v1/users/me - Mise à jour des informations de l'utilisateur
		protected.PUT("/me", rg.handler.UpdateMe)
		
		// PUT /api/v1/users/me/password - Changement de mot de passe
		protected.PUT("/me/password", rg.handler.ChangePassword)
		
		// GET /api/v1/users/except-me - Liste des utilisateurs sauf l'utilisateur connecté
		protected.GET("/except-me", rg.handler.GetUsersExceptMe)
		
		// GET /api/v1/users/search - Recherche d'utilisateurs
		protected.GET("/search", rg.handler.SearchUsers)
	}
}

// SetupRoutes configure les routes du module utilisateur (pour la compatibilité)
func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	rg := NewRouteGroup(handler, jwtSecret)
	rg.Register(router)
}