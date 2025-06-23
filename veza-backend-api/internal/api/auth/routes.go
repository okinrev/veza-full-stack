package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

// RouteGroup représente un groupe de routes pour le module d'authentification
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

// Register enregistre toutes les routes du module d'authentification
func (rg *RouteGroup) Register(router *gin.RouterGroup) {
	// Groupe principal d'authentification
	auth := router.Group("/auth")
	{
		// Routes publiques
		rg.registerPublicRoutes(auth)
		
		// Routes protégées
		rg.registerProtectedRoutes(auth)
	}
}

// registerPublicRoutes enregistre les routes publiques
func (rg *RouteGroup) registerPublicRoutes(router *gin.RouterGroup) {
	// POST /api/v1/auth/register - Inscription d'un nouvel utilisateur
	router.POST("/register", rg.handler.Register)
	
	// POST /api/v1/auth/signup - Alias pour l'inscription
	router.POST("/signup", rg.handler.Register)
	
	// POST /api/v1/auth/login - Connexion
	router.POST("/login", rg.handler.Login)
	
	// POST /api/v1/auth/refresh - Rafraîchissement du token
	router.POST("/refresh", rg.handler.RefreshToken)
	
	// POST /api/v1/auth/logout - Déconnexion
	router.POST("/logout", rg.handler.Logout)
}

// registerProtectedRoutes enregistre les routes protégées
func (rg *RouteGroup) registerProtectedRoutes(router *gin.RouterGroup) {
	protected := router.Group("")
	protected.Use(middleware.JWTAuthMiddleware(rg.secret))
	{
		// GET /api/v1/auth/me - Informations de l'utilisateur connecté
		protected.GET("/me", rg.handler.GetMe)
	}
}

// SetupRoutes configure les routes du module d'authentification (pour la compatibilité)
func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	rg := NewRouteGroup(handler, jwtSecret)
	rg.Register(router)
}
