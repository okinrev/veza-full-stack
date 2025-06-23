# Étape 5 - Refactoring du système de routes

## 🎯 Objectif
Créer un système de routing simple, cohérent et maintenable en supprimant la complexité actuelle.

## ⏱️ Durée estimée : 45-60 minutes

## 🚨 Problèmes à résoudre
- `main.go` complexe avec double système de routes
- Routes hybrides (legacy + v1)
- Mounting complexe avec `gin.WrapH(apiEngine)`
- Pas de séparation claire frontend/API

## 📊 État actuel du routing

### Problèmes dans `cmd/server/main.go`
```go
// ❌ Système hybride complexe
v1 := router.Group("/api/v1")
{
    apiEngine := apiRouter.GetEngine()
    v1.Any("/*path", gin.WrapH(apiEngine))  // Problématique
}

// ❌ Legacy routes avec NoRoute
router.NoRoute(func(c *gin.Context) {
    // Logique complexe de fallback
    legacyEngine.ServeHTTP(c.Writer, c.Request)
})
```

### Structure actuelle
```
main.go (300+ lignes)
├── Frontend routes (HTML files)
├── Health check
├── API v1 (hybride)
├── Legacy compatibility
└── NoRoute fallback
```

## 📋 Plan de refactoring

### Phase 5.1 : Simplification du main.go
1. Séparer les responsabilités
2. Éliminer le double routing
3. Routes directes sans proxy

### Phase 5.2 : Création d'un router centralisé
4. Router API unifié
5. Routes frontend séparées
6. Configuration CORS centralisée

### Phase 5.3 : Suppression du legacy
7. Éliminer les routes de compatibilité
8. Nettoyer le code mort

## 🔧 Implémentation détaillée

### Phase 5.1 : Nouveau main.go simplifié

#### `cmd/server/main.go` (version simplifiée)
```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	
	"github.com/okinrev/veza-web-app/internal/config"
	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/api"
)

func main() {
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

	// Router
	router := setupRouter(cfg)
	
	// Routes
	setupRoutes(router, db, cfg)

	// Server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("🚀 Talas backend starting on port %s", cfg.Server.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server startup failed:", err)
	}
}

func setupRouter(cfg *config.Config) *gin.Engine {
	router := gin.New()
	
	// Middleware globaux
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// CORS
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))
	
	return router
}

func setupRoutes(router *gin.Engine, db *database.DB, cfg *config.Config) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"service":     "talas-backend",
			"version":     "1.0.0",
			"timestamp":   time.Now().Unix(),
			"environment": cfg.Server.Environment,
		})
	})

	// API routes
	api.SetupRoutes(router, db, cfg)
	
	// Frontend routes (optionnel pour SPA)
	setupFrontendRoutes(router)
}

func setupFrontendRoutes(router *gin.Engine) {
	// Servir les fichiers statiques
	router.Static("/static", "./static")
	router.Static("/public", "../frontend/public")
	
	// SPA fallback (optionnel)
	router.NoRoute(func(c *gin.Context) {
		// Si c'est une route API qui n'existe pas
		if gin.IsDebugging() || len(c.Request.URL.Path) > 4 && c.Request.URL.Path[:5] == "/api/" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Route not found",
				"path":  c.Request.URL.Path,
			})
			return
		}
		
		// Sinon, servir l'index.html (SPA)
		c.File("../frontend/public/index.html")
	})
}
```

### Phase 5.2 : Router API centralisé

#### `internal/api/router.go` (nouveau fichier)
```go
package api

import (
	"github.com/gin-gonic/gin"
	
	"github.com/okinrev/veza-web-app/internal/config"
	"github.com/okinrev/veza-web-app/internal/database"
	
	// Import de tous les modules API
	"github.com/okinrev/veza-web-app/internal/api/auth"
	"github.com/okinrev/veza-web-app/internal/api/user"
	"github.com/okinrev/veza-web-app/internal/api/admin"
	"github.com/okinrev/veza-web-app/internal/api/track"
	"github.com/okinrev/veza-web-app/internal/api/file"
	"github.com/okinrev/veza-web-app/internal/api/listing"
	"github.com/okinrev/veza-web-app/internal/api/offer"
	"github.com/okinrev/veza-web-app/internal/api/message"
	"github.com/okinrev/veza-web-app/internal/api/room"
	"github.com/okinrev/veza-web-app/internal/api/search"
	"github.com/okinrev/veza-web-app/internal/api/tag"
	"github.com/okinrev/veza-web-app/internal/api/shared_resources"
	"github.com/okinrev/veza-web-app/internal/api/product"
)

// SetupRoutes configure toutes les routes API
func SetupRoutes(router *gin.Engine, db *database.DB, cfg *config.Config) {
	// Groupe API v1
	v1 := router.Group("/api/v1")
	
	// Auth (prioritaire)
	setupAuthRoutes(v1, db, cfg.JWT.Secret)
	
	// Core modules
	setupUserRoutes(v1, db, cfg.JWT.Secret)
	setupAdminRoutes(v1, db, cfg.JWT.Secret)
	
	// Content modules
	setupTrackRoutes(v1, db, cfg.JWT.Secret)
	setupFileRoutes(v1, db, cfg.JWT.Secret)
	setupProductRoutes(v1, db, cfg.JWT.Secret)
	setupSharedResourceRoutes(v1, db, cfg.JWT.Secret)
	
	// Community modules
	setupListingRoutes(v1, db, cfg.JWT.Secret)
	setupOfferRoutes(v1, db, cfg.JWT.Secret)
	setupMessageRoutes(v1, db, cfg.JWT.Secret)
	setupRoomRoutes(v1, db, cfg.JWT.Secret)
	
	// Utility modules
	setupSearchRoutes(v1, db, cfg.JWT.Secret)
	setupTagRoutes(v1, db, cfg.JWT.Secret)
}

// setupAuthRoutes configure les routes d'authentification
func setupAuthRoutes(v1 *gin.RouterGroup, db *database.DB, jwtSecret string) {
	service := auth.NewService(db, jwtSecret)
	handler := auth.NewHandler(service)
	auth.SetupRoutes(v1, handler, jwtSecret)
}

// setupUserRoutes configure les routes utilisateur
func setupUserRoutes(v1 *gin.RouterGroup, db *database.DB, jwtSecret string) {
	service := user.NewService(db)
	handler := user.NewHandler(service)
	user.SetupRoutes(v1, handler, jwtSecret)
}

// setupAdminRoutes configure les routes admin
func setupAdminRoutes(v1 *gin.RouterGroup, db *database.DB, jwtSecret string) {
	service := admin.NewService(db)
	handler := admin.NewHandler(service)
	admin.SetupRoutes(v1, handler, jwtSecret)
}

// setupTrackRoutes configure les routes audio
func setupTrackRoutes(v1 *gin.RouterGroup, db *database.DB, jwtSecret string) {
	service := track.NewService(db, jwtSecret)
	handler := track.NewHandler(service)
	track.SetupRoutes(v1, handler, jwtSecret)
}

// setupFileRoutes configure les routes de fichiers
func setupFileRoutes(v1 *gin.RouterGroup, db *database.DB, jwtSecret string) {
	service := file.NewService(db)
	handler := file.NewHandler(service)
	file.SetupRoutes(v1, handler, jwtSecret)
}

// setupProductRoutes configure les routes produits
func setupProductRoutes(v1 *gin.RouterGroup, db *database.DB, jwtSecret string) {
	service := product.NewService(db)
	handler := product.NewHandler(service)
	product.SetupRoutes(v1, handler, jwtSecret)
}

// setupSharedResourceRoutes configure les routes ressources partagées
func setupSharedResourceRoutes(v1 *gin.RouterGroup, db *database.DB, jwtSecret string) {
	service := shared_resources.NewService(db)
	handler := shared_resources.NewHandler(service)
	shared_resources.SetupRoutes(v1, handler, jwtSecret)
}

// setupListingRoutes configure les routes marketplace
func setupListingRoutes(v1 *gin.RouterGroup, db *database.DB, jwtSecret string) {
	service := listing.NewService(db)
	handler := listing.NewHandler(service)
	listing.SetupRoutes(v1, handler, jwtSecret)
}

// setupOfferRoutes configure les routes d'offres
func setupOfferRoutes(v1 *gin.RouterGroup, db *database.DB, jwtSecret string) {
	service := offer.NewService(db)
	handler := offer.NewHandler(service)
	offer.SetupRoutes(v1, handler, jwtSecret)
}

// setupMessageRoutes configure les routes de messages
func setupMessageRoutes(v1 *gin.RouterGroup, db *database.DB, jwtSecret string) {
	service := message.NewService(db)
	handler := message.NewHandler(service)
	message.SetupRoutes(v1, handler, jwtSecret)
}

// setupRoomRoutes configure les routes de salons
func setupRoomRoutes(v1 *gin.RouterGroup, db *database.DB, jwtSecret string) {
	service := room.NewService(db)
	handler := room.NewHandler(service)
	room.SetupRoutes(v1, handler, jwtSecret)
}

// setupSearchRoutes configure les routes de recherche
func setupSearchRoutes(v1 *gin.RouterGroup, db *database.DB, jwtSecret string) {
	service := search.NewService(db)
	handler := search.NewHandler(service)
	search.SetupRoutes(v1, handler, jwtSecret)
}

// setupTagRoutes configure les routes de tags
func setupTagRoutes(v1 *gin.RouterGroup, db *database.DB, jwtSecret string) {
	service := tag.NewService(db)
	handler := tag.NewHandler(service)
	tag.SetupRoutes(v1, handler, jwtSecret)
}
```

### Phase 5.3 : Standardisation des routes par module

#### Template de `routes.go` standardisé

**Pour chaque module, créer `internal/api/[module]/routes.go` :**
```go
package [module]

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

// SetupRoutes configure les routes du module [module]
func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	[module] := router.Group("/[module]")
	
	// Routes publiques (si applicable)
	[module].GET("", handler.List[Module])
	[module].GET("/:id", handler.Get[Module])
	
	// Routes protégées
	protected := [module].Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		protected.POST("", handler.Create[Module])
		protected.PUT("/:id", handler.Update[Module])
		protected.DELETE("/:id", handler.Delete[Module])
	}
	
	// Routes admin (si applicable)
	admin := [module].Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware(jwtSecret))
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/stats", handler.Get[Module]Stats)
	}
}
```

#### Exemple concret pour `internal/api/track/routes.go`
```go
package track

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

// SetupRoutes configure les routes audio
func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	tracks := router.Group("/tracks")
	
	// Routes publiques
	tracks.GET("", handler.ListTracks)
	tracks.GET("/:id", handler.GetTrack)
	tracks.GET("/:id/stats", handler.GetTrackStats)
	
	// Routes protégées
	protected := tracks.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		protected.POST("", handler.AddTrackWithUpload)
		protected.PUT("/:id", handler.UpdateTrack)
		protected.DELETE("/:id", handler.DeleteTrack)
	}
	
	// Routes de streaming
	streaming := router.Group("/stream")
	{
		streaming.GET("/:filename", handler.StreamAudio)
		streaming.GET("/signed/:filename", handler.StreamAudioSigned)
		
		protected := streaming.Group("")
		protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
		{
			protected.GET("/generate-url", handler.GenerateStreamURL)
		}
	}
}
```

### Phase 5.4 : Configuration centralisée

#### Améliorer `internal/config/config.go`
```go
package config

import (
    "os"
    "strconv"
    "time"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    JWT      JWTConfig
    Files    FilesConfig
}

type ServerConfig struct {
    Port            string
    ReadTimeout     time.Duration
    WriteTimeout    time.Duration
    ShutdownTimeout time.Duration
    Environment     string
    Debug           bool
}

type DatabaseConfig struct {
    URL          string
    Host         string
    Port         string
    Username     string
    Password     string
    Database     string
    SSLMode      string
    MaxOpenConns int
    MaxIdleConns int
    MaxLifetime  time.Duration
}

type JWTConfig struct {
    Secret         string
    ExpirationTime time.Duration
    RefreshTime    time.Duration
}

type FilesConfig struct {
    UploadDir     string
    MaxSize       int64
    AllowedTypes  []string
}

func New() *Config {
    return &Config{
        Server: ServerConfig{
            Port:            getEnv("PORT", "8080"),
            ReadTimeout:     getDurationEnv("READ_TIMEOUT", 10*time.Second),
            WriteTimeout:    getDurationEnv("WRITE_TIMEOUT", 10*time.Second),
            ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 30*time.Second),
            Environment:     getEnv("ENVIRONMENT", "development"),
            Debug:           getEnv("ENVIRONMENT", "development") == "development",
        },
        Database: DatabaseConfig{
            URL:          getEnv("DATABASE_URL", ""),
            Host:         getEnv("DB_HOST", "localhost"),
            Port:         getEnv("DB_PORT", "5432"),
            Username:     getEnv("DB_USERNAME", "postgres"),
            Password:     getEnv("DB_PASSWORD", ""),
            Database:     getEnv("DB_NAME", "talas"),
            SSLMode:      getEnv("DB_SSLMODE", "disable"),
            MaxOpenConns: getIntEnv("DB_MAX_OPEN_CONNS", 25),
            MaxIdleConns: getIntEnv("DB_MAX_IDLE_CONNS", 25),
            MaxLifetime:  getDurationEnv("DB_MAX_LIFETIME", 5*time.Minute),
        },
        JWT: JWTConfig{
            Secret:         getEnv("JWT_SECRET", "your-secret-key"),
            ExpirationTime: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
            RefreshTime:    getDurationEnv("JWT_REFRESH_TIME", 7*24*time.Hour),
        },
        Files: FilesConfig{
            UploadDir:    getEnv("UPLOAD_DIR", "./static"),
            MaxSize:      getIntEnv64("MAX_FILE_SIZE", 100<<20), // 100MB
            AllowedTypes: []string{".jpg", ".png", ".pdf", ".mp3", ".wav"},
        },
    }
}

// Fonctions utilitaires
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getIntEnv64(key string, defaultValue int64) int64 {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}
```

### Phase 5.5 : Suppression des routes legacy

```bash
# Supprimer les fichiers de routes complexes
rm -f cmd/server/main.go.bak  # Backup de l'ancien main.go

# Nettoyer les imports inutiles
# Les anciens internal/routes/ ont déjà été supprimés dans l'étape 1
```

## ✅ Structure finale

### Arborescence des routes
```
cmd/server/main.go                    # Simple, ~80 lignes
internal/api/router.go                # Router centralisé
internal/api/*/routes.go              # Routes par module
internal/config/config.go             # Configuration centralisée
```

### Flux des routes
```
main.go
├── setupRouter() → CORS, middleware
├── setupRoutes() → api.SetupRoutes()
└── server.ListenAndServe()

api.SetupRoutes()
├── /api/v1/auth/*     → auth.SetupRoutes()
├── /api/v1/users/*    → user.SetupRoutes()
├── /api/v1/tracks/*   → track.SetupRoutes()
└── [autres modules]
```

## ✅ Checklist de validation

```bash
# 1. Compilation réussie
go build ./cmd/server
echo $?
# Attendu : 0

# 2. Taille du main.go réduite
wc -l cmd/server/main.go
# Attendu : ~80 lignes (vs 300+ avant)

# 3. Routes standardisées
find internal/api/ -name "routes.go" | wc -l
# Attendu : nombre égal au nombre de modules

# 4. Configuration centralisée
grep -c "getEnv\|config\." internal/config/config.go
# Attendu : >20 (configuration riche)

# 5. Test des endpoints
curl -s http://localhost:8080/health | jq .status
# Attendu : "ok"

curl -s http://localhost:8080/api/v1/auth/register -X POST \
  -H "Content-Type: application/json" \
  -d '{}' | jq .success
# Attendu : false (validation)
```

## 🎯 Bénéfices obtenus

1. **Simplicité** : main.go lisible et maintenable
2. **Modularité** : Chaque module gère ses propres routes
3. **Cohérence** : Pattern uniforme pour tous les modules
4. **Performance** : Pas de proxy/wrapper complexe
5. **Debugging** : Routes claires et traçables

## ⏭️ Étape suivante
Une fois le routing simplifié → `06_tests_validation.md`

---

**💾 IMPORTANT** : Commit après cette étape
```bash
git add .
git commit -m "Étape 5: Refactoring routes - système simplifié et modulaire"
```