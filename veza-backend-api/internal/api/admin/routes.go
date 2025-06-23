package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

// RouteGroup représente un groupe de routes pour le module admin
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

// Register enregistre toutes les routes du module admin
func (rg *RouteGroup) Register(router *gin.RouterGroup) {
	// Groupe principal admin
	admin := router.Group("/admin")
	
	// Application des middlewares communs
	admin.Use(middleware.JWTAuthMiddleware(rg.secret))
	admin.Use(middleware.AdminMiddleware())
	
	// Enregistrement des routes
	rg.registerAdminRoutes(admin)
}

// registerAdminRoutes enregistre les routes d'administration
func (rg *RouteGroup) registerAdminRoutes(router *gin.RouterGroup) {
	// GET /api/v1/admin/dashboard - Tableau de bord administrateur
	router.GET("/dashboard", rg.handler.Dashboard)
	
	// GET /api/v1/admin/users - Liste des utilisateurs
	router.GET("/users", rg.handler.GetUsers)
	
	// GET /api/v1/admin/analytics - Données analytiques
	router.GET("/analytics", rg.handler.GetAnalytics)
	
	// GET /api/v1/admin/categories - Liste des catégories
	router.GET("/categories", rg.handler.GetCategories)
}

// SetupRoutes configure les routes du module admin (pour la compatibilité)
func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	rg := NewRouteGroup(handler, jwtSecret)
	rg.Register(router)
}
