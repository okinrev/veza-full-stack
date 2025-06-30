package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	CSRFTokenHeader = "X-CSRF-Token"
	CSRFTokenField  = "csrf_token"
	CSRFCookieName  = "csrf_token"
	CSRFTokenLength = 32
)

// CSRFConfig représente la configuration CSRF
type CSRFConfig struct {
	TokenLength  int
	CookieName   string
	HeaderName   string
	FieldName    string
	CookiePath   string
	CookieDomain string
	Secure       bool
	HTTPOnly     bool
	SameSite     http.SameSite
}

// DefaultCSRFConfig retourne la configuration CSRF par défaut
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength:  CSRFTokenLength,
		CookieName:   CSRFCookieName,
		HeaderName:   CSRFTokenHeader,
		FieldName:    CSRFTokenField,
		CookiePath:   "/",
		CookieDomain: "",
		Secure:       true,
		HTTPOnly:     true,
		SameSite:     http.SameSiteStrictMode,
	}
}

// CSRFProtection retourne le middleware de protection CSRF
func CSRFProtection() gin.HandlerFunc {
	return CSRFWithConfig(DefaultCSRFConfig())
}

// CSRFWithConfig retourne le middleware CSRF avec configuration personnalisée
func CSRFWithConfig(config CSRFConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF pour les requêtes GET, HEAD, OPTIONS
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Skip CSRF pour les endpoints publics (auth/login, auth/register)
		if isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Générer un token CSRF si nécessaire
		token := getCSRFToken(c, config)
		if token == "" {
			token = generateCSRFToken()
			setCSRFToken(c, token, config)
		}

		// Valider le token CSRF pour les requêtes POST, PUT, DELETE, PATCH
		if !validateCSRFToken(c, token, config) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token validation failed",
				"code":  "CSRF_INVALID",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isPublicEndpoint vérifie si l'endpoint est public (pas de protection CSRF)
func isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/api/auth/login",
		"/api/auth/register",
		"/api/auth/refresh",
		"/health",
		"/hexagonal/status",
	}

	for _, endpoint := range publicEndpoints {
		if path == endpoint {
			return true
		}
	}
	return false
}

// generateCSRFToken génère un token CSRF aléatoire
func generateCSRFToken() string {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback avec timestamp si erreur
		return fmt.Sprintf("csrf_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// getCSRFToken récupère le token CSRF depuis le cookie
func getCSRFToken(c *gin.Context, config CSRFConfig) string {
	cookie, err := c.Cookie(config.CookieName)
	if err != nil {
		return ""
	}
	return cookie
}

// setCSRFToken définit le token CSRF dans le cookie
func setCSRFToken(c *gin.Context, token string, config CSRFConfig) {
	c.SetCookie(
		config.CookieName,   // name
		token,               // value
		3600,                // maxAge (1 heure)
		config.CookiePath,   // path
		config.CookieDomain, // domain
		config.Secure,       // secure
		config.HTTPOnly,     // httpOnly
	)

	// Exposer le token dans les headers pour les clients JavaScript
	c.Header(config.HeaderName, token)
}

// validateCSRFToken valide le token CSRF
func validateCSRFToken(c *gin.Context, expectedToken string, config CSRFConfig) bool {
	// Récupérer le token depuis le header
	headerToken := c.GetHeader(config.HeaderName)
	if headerToken != "" && headerToken == expectedToken {
		return true
	}

	// Récupérer le token depuis le formulaire (pour les requêtes form-data)
	formToken := c.PostForm(config.FieldName)
	if formToken != "" && formToken == expectedToken {
		return true
	}

	// Récupérer le token depuis le JSON body
	var jsonData map[string]interface{}
	if err := c.ShouldBindJSON(&jsonData); err == nil {
		if jsonToken, exists := jsonData[config.FieldName]; exists {
			if tokenStr, ok := jsonToken.(string); ok && tokenStr == expectedToken {
				return true
			}
		}
	}

	return false
}

// GetCSRFToken est un handler pour obtenir un token CSRF
func GetCSRFToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		config := DefaultCSRFConfig()
		token := generateCSRFToken()
		setCSRFToken(c, token, config)

		c.JSON(http.StatusOK, gin.H{
			"csrf_token": token,
			"expires_in": 3600,
			"usage": gin.H{
				"header": config.HeaderName,
				"field":  config.FieldName,
			},
		})
	}
}
