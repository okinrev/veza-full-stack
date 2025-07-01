package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

// RouteGroup représente un groupe de routes pour le module d'authentification
type RouteGroup struct {
	handler      *Handler
	oauthHandler *OAuthHandler
	secret       string
}

// NewRouteGroup crée une nouvelle instance de RouteGroup
func NewRouteGroup(handler *Handler, jwtSecret string) *RouteGroup {
	service := handler.service
	return &RouteGroup{
		handler:      handler,
		oauthHandler: NewOAuthHandler(service),
		secret:       jwtSecret,
	}
}

// Register enregistre toutes les routes du module d'authentification
func (rg *RouteGroup) Register(router *gin.RouterGroup) {
	// Groupe principal d'authentification
	auth := router.Group("/auth")
	{
		// Routes publiques
		rg.registerPublicRoutes(auth)

		// Routes OAuth2
		rg.registerOAuthRoutes(auth)

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

// registerOAuthRoutes enregistre les routes OAuth2
func (rg *RouteGroup) registerOAuthRoutes(router *gin.RouterGroup) {
	oauth := router.Group("/oauth")
	{
		// Google OAuth2
		google := oauth.Group("/google")
		{
			// GET /api/v1/auth/oauth/google - URL d'authentification Google
			google.GET("", rg.oauthHandler.GoogleOAuthURL)
			// GET /api/v1/auth/oauth/google/callback - Callback Google
			google.GET("/callback", rg.oauthHandler.GoogleOAuthCallback)
		}

		// GitHub OAuth2
		github := oauth.Group("/github")
		{
			// GET /api/v1/auth/oauth/github - URL d'authentification GitHub
			github.GET("", rg.oauthHandler.GitHubOAuthURL)
			// GET /api/v1/auth/oauth/github/callback - Callback GitHub
			github.GET("/callback", rg.oauthHandler.GitHubOAuthCallback)
		}

		// Discord OAuth2
		discord := oauth.Group("/discord")
		{
			// GET /api/v1/auth/oauth/discord - URL d'authentification Discord
			discord.GET("", rg.oauthHandler.DiscordOAuthURL)
			// GET /api/v1/auth/oauth/discord/callback - Callback Discord
			discord.GET("/callback", rg.oauthHandler.DiscordOAuthCallback)
		}
	}
}

// registerProtectedRoutes enregistre les routes protégées
func (rg *RouteGroup) registerProtectedRoutes(router *gin.RouterGroup) {
	protected := router.Group("")
	protected.Use(middleware.JWTAuthMiddleware(rg.secret))
	{
		// GET /api/v1/auth/me - Informations de l'utilisateur connecté
		protected.GET("/me", rg.handler.GetMe)

		// GET /api/v1/auth/test - Test de validation JWT pour tous les services
		protected.GET("/test", rg.handler.TestAuthEndpoint)
	}
}

// SetupRoutes configure les routes du module d'authentification (pour la compatibilité)
func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	rg := NewRouteGroup(handler, jwtSecret)
	rg.Register(router)
}
