package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/utils"
)

// Logger middleware pour enregistrer les informations de la requête
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Début de la requête
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Traitement de la requête
		c.Next()

		// Calcul du temps d'exécution
		latency := time.Since(start)

		// Récupération des informations de la requête
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Construction du message de log
		if raw != "" {
			path = path + "?" + raw
		}

		// Log de la requête
		utils.LogInfo(fmt.Sprintf("[%s] %s %s %d %v %s %s",
			method,
			path,
			clientIP,
			statusCode,
			latency,
			errorMessage,
			c.GetString("user_id"),
		))
	}
}

// Recovery middleware pour gérer les panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log de l'erreur
				utils.LogError(fmt.Sprintf("Panic recovered: %v", err))

				// Réponse d'erreur
				c.JSON(500, gin.H{
					"success": false,
					"error":   "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// RateLimiter middleware pour limiter le nombre de requêtes
func RateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	limiter := utils.NewRateLimiter(limit, window)
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if !limiter.Allow(clientIP) {
			c.JSON(429, gin.H{
				"success": false,
				"error":   "Too many requests",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// CORS middleware pour gérer les en-têtes CORS
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestID middleware pour ajouter un ID unique à chaque requête
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = utils.GenerateUUID()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

// ValidateContentType middleware pour valider le type de contenu
func ValidateContentType(contentType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			if c.GetHeader("Content-Type") != contentType {
				c.JSON(400, gin.H{
					"success": false,
					"error":   fmt.Sprintf("Content-Type must be %s", contentType),
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
} 