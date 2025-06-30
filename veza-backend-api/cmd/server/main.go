package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/okinrev/veza-web-app/internal/api"
	"github.com/okinrev/veza-web-app/internal/config"
	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/websocket"
)

// getProjectRoot retourne le chemin absolu vers la racine du projet
func getProjectRoot() string {
	// Obtenir le chemin du répertoire de travail actuel
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Erreur lors de la récupération du répertoire de travail:", err)
	}
	// Remonter d'un niveau depuis le dossier backend
	return filepath.Dir(wd)
}

// serveReactApp sert l'application React en mode SPA
func serveReactApp(frontendPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Si c'est une route API ou WebSocket, ne pas intercepter
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/ws/") {
			c.Next()
			return
		}

		// Pour toutes les autres routes, servir index.html (SPA routing)
		indexPath := filepath.Join(frontendPath, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
		} else {
			c.Status(http.StatusNotFound)
		}
	}
}

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Configuration
	cfg := config.New()

	// Mode Gin
	if cfg.Server.Environment != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Database
	db, err := database.NewConnection(cfg.Database.URL)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Migrations
	if err := database.RunMigrations(db); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	// Obtenir le chemin absolu du projet
	projectRoot := getProjectRoot()
	reactFrontendPath := filepath.Join(projectRoot, "veza-frontend", "dist")

	log.Printf("🚀 Chemin du frontend React: %s", reactFrontendPath)

	// Vérifier que le dossier dist existe
	if _, err := os.Stat(reactFrontendPath); os.IsNotExist(err) {
		log.Printf("⚠️  Le dossier frontend React n'existe pas: %s", reactFrontendPath)
		log.Printf("💡 Veuillez exécuter 'npm run build' dans le dossier veza-frontend")
	}

	// Configurer les routes
	router := gin.Default()

	// Middleware pour servir les assets statiques React (JS, CSS, images)
	router.Static("/assets", filepath.Join(reactFrontendPath, "assets"))
	router.StaticFile("/favicon.svg", filepath.Join(reactFrontendPath, "favicon.svg"))
	router.StaticFile("/favicon.ico", filepath.Join(reactFrontendPath, "favicon.ico"))
	router.StaticFile("/vite.svg", filepath.Join(reactFrontendPath, "vite.svg"))

	// Initialiser le gestionnaire WebSocket
	chatManager := websocket.NewChatManager(cfg.JWT.Secret)
	go chatManager.Run()

	// Route WebSocket pour le chat
	router.GET("/ws/chat", func(c *gin.Context) {
		chatManager.HandleWebSocket(c)
	})

	// Endpoint de santé avec test DB pour HAProxy
	router.GET("/api/health", func(c *gin.Context) {
		// Test database connection
		dbStatus := "ok"
		dbConnected := true
		if err := db.Ping(); err != nil {
			dbStatus = "error: " + err.Error()
			dbConnected = false
		}

		// Overall health status
		overallStatus := "healthy"
		httpStatus := http.StatusOK
		if !dbConnected {
			overallStatus = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, gin.H{
			"status":    overallStatus,
			"service":   "veza-backend-dev",
			"version":   "2.0.0",
			"frontend":  "react",
			"timestamp": time.Now().Unix(),
			"database":  dbStatus,
			"components": gin.H{
				"database_connected": dbConnected,
				"websocket_active":   true,
				"frontend_path":      reactFrontendPath,
			},
		})
	})

	// Configurer les routes API
	api.SetupRoutes(router, db, cfg)

	// Middleware pour servir l'application React (SPA) pour toutes les autres routes
	router.Use(serveReactApp(reactFrontendPath))

	// Démarrer le serveur
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("🎯 Serveur démarré sur le port %s avec frontend React", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
	}
}
