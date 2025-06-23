package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	basicFrontendPath := filepath.Join(projectRoot, "veza-basic-frontend")

	log.Printf("📂 Chemin du frontend HTML/JS basique: %s", basicFrontendPath)

	// Configurer les routes
	router := gin.Default()

	// Middleware pour servir les fichiers statiques du frontend HTML/JS basique
	router.Static("/css", filepath.Join(basicFrontendPath, "css"))
	router.Static("/js", filepath.Join(basicFrontendPath, "js"))
	router.StaticFile("/favicon.ico", filepath.Join(basicFrontendPath, "favicon.ico"))

	// Servir toutes les pages HTML individuelles
	router.StaticFile("/", filepath.Join(basicFrontendPath, "login.html"))
	router.StaticFile("/login", filepath.Join(basicFrontendPath, "login.html"))
	router.StaticFile("/register", filepath.Join(basicFrontendPath, "register.html"))
	router.StaticFile("/dashboard", filepath.Join(basicFrontendPath, "dashboard.html"))
	router.StaticFile("/main", filepath.Join(basicFrontendPath, "main.html"))
	router.StaticFile("/hub", filepath.Join(basicFrontendPath, "hub.html"))
	router.StaticFile("/hub_v2", filepath.Join(basicFrontendPath, "hub_v2.html"))
	router.StaticFile("/gg", filepath.Join(basicFrontendPath, "gg.html"))
	router.StaticFile("/chat", filepath.Join(basicFrontendPath, "chat.html"))
	router.StaticFile("/room", filepath.Join(basicFrontendPath, "room.html"))
	router.StaticFile("/message", filepath.Join(basicFrontendPath, "message.html"))
	router.StaticFile("/dm", filepath.Join(basicFrontendPath, "dm.html"))
	router.StaticFile("/users", filepath.Join(basicFrontendPath, "users.html"))
	router.StaticFile("/produits", filepath.Join(basicFrontendPath, "user_products.html"))
	router.StaticFile("/admin_products", filepath.Join(basicFrontendPath, "admin_products.html"))
	router.StaticFile("/track", filepath.Join(basicFrontendPath, "track.html"))
	router.StaticFile("/shared_ressources", filepath.Join(basicFrontendPath, "shared_ressources.html"))
	router.StaticFile("/listings", filepath.Join(basicFrontendPath, "listings.html"))
	router.StaticFile("/search", filepath.Join(basicFrontendPath, "search.html"))
	router.StaticFile("/search_v2", filepath.Join(basicFrontendPath, "search_v2.html"))
	router.StaticFile("/api", filepath.Join(basicFrontendPath, "api.html"))
	router.StaticFile("/test", filepath.Join(basicFrontendPath, "test.html"))

	// Initialiser le gestionnaire WebSocket
	chatManager := websocket.NewChatManager(cfg.JWT.Secret)
	go chatManager.Run()

	// Route WebSocket pour le chat
	router.GET("/ws/chat", func(c *gin.Context) {
		chatManager.HandleWebSocket(c)
	})

	// Endpoint de santé simple pour HAProxy
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "veza-backend",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
		})
	})

	// Configurer les routes API
	api.SetupRoutes(router, db, cfg)

	// Démarrer le serveur
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Serveur démarré sur le port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
	}
}
