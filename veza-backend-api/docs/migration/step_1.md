# Étape 1 - Consolidation de l'architecture

## 🎯 Objectif
Éliminer la double architecture et choisir un système unifié (architecture modulaire).

## ⏱️ Durée estimée : 30-45 minutes

## 🚨 Problèmes à résoudre
- Double système de routing (legacy + modulaire)
- Conflits entre `internal/routes/` et `internal/api/*/`
- `main.go` complexe avec routes hybrides
- Handlers non importés mais documentés

## 📋 Actions détaillées

### Action 1.1 : Backup et analyse initiale
```bash
# Créer un backup
git add . && git commit -m "Backup avant consolidation architecture"

# Analyser l'état actuel
find internal/ -name "*.go" | grep -E "(route|handler)" | sort
ls -la internal/routes/
ls -la internal/api/*/
```

### Action 1.2 : Supprimer l'ancien système de routing
```bash
# Fichiers à supprimer (après backup)
rm -rf internal/routes/
```

**Fichiers concernés à supprimer :**
- `internal/routes/router.go`
- `internal/routes/auth.go`
- `internal/routes/user.go`
- `internal/routes/chat.go`
- `internal/routes/search.go`
- `internal/routes/shared_resources.go`
- `internal/routes/tracks.go`
- `internal/routes/listings.go`
- `internal/routes/products.go`
- `internal/routes/admin.go`
- `internal/routes/direct.go`

### Action 1.3 : Inventaire des modules API existants

**Modules déjà présents dans `internal/api/` :**
```
✅ auth/          # Non utilisé mais à structure correcte
✅ user/          # Service implémenté
✅ room/          # Handler basique
✅ listing/       # Handler basique  
✅ message/       # Handler basique
✅ offer/         # Handler basique
✅ search/        # Handler basique
✅ shared_ressources/ # Handler basique
✅ suggestions/   # Handler basique
✅ tag/           # Handler basique
✅ track/         # Handler basique
```

**Modules manquants à créer :**
```
❌ admin/         # Besoin handler + service
❌ file/          # Besoin handler + service
❌ product/       # Besoin handler + service
❌ analytics/     # Optionnel
```

### Action 1.4 : Créer les modules manquants

#### Créer `internal/api/admin/`
```bash
mkdir -p internal/api/admin
```

**Fichier : `internal/api/admin/handler.go`**
```go
package admin

import (
	"net/http"
	"github.com/okinrev/veza-web-app/internal/utils/response"
	"github.com/okinrev/veza-web-app/internal/common"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Dashboard retourne les statistiques admin
func (h *Handler) Dashboard(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// TODO: Vérifier rôle admin
	stats, err := h.service.GetDashboardStats()
	if err != nil {
		response.ErrorJSON(c.Writer, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(c.Writer, stats, "Dashboard stats retrieved")
}

// GetUsers retourne la liste des utilisateurs (admin)
func (h *Handler) GetUsers(c *gin.Context) {
	// TODO: Implémenter basé sur doc_admin_handler.md
	response.SuccessJSON(c.Writer, []interface{}{}, "Users retrieved")
}

// TODO: Autres méthodes basées sur doc_admin_handler.md
```

**Fichier : `internal/api/admin/service.go`**
```go
package admin

import (
	"github.com/okinrev/veza-web-app/internal/database"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

type DashboardStats struct {
	TotalUsers    int `json:"total_users"`
	TotalTracks   int `json:"total_tracks"`
	TotalListings int `json:"total_listings"`
	// TODO: Autres stats basées sur doc_admin_handler.md
}

func (s *Service) GetDashboardStats() (*DashboardStats, error) {
	stats := &DashboardStats{}
	
	// TODO: Requêtes SQL basées sur doc_admin_handler.md
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
```

**Fichier : `internal/api/admin/routes.go`**
```go
package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	admin := router.Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware(jwtSecret))
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/dashboard", handler.Dashboard)
		admin.GET("/users", handler.GetUsers)
		// TODO: Autres routes basées sur doc_admin_handler.md
	}
}
```

#### Créer `internal/api/file/` (similaire)
```bash
mkdir -p internal/api/file
```

**Pattern identique avec :**
- `handler.go` (basé sur `doc_file_handler.md`)
- `service.go` 
- `routes.go`

#### Créer `internal/api/product/` (similaire)
```bash
mkdir -p internal/api/product
```

### Action 1.5 : Simplifier main.go

**Fichier : `cmd/server/main.go` (version simplifiée)**
```go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/joho/godotenv"
	
	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/api/auth"
	"github.com/okinrev/veza-web-app/internal/api/user"
	"github.com/okinrev/veza-web-app/internal/api/admin"
	// TODO: Autres imports API
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	Environment string
	Debug       bool
}

func loadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	config := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Port:        os.Getenv("PORT"),
		Environment: os.Getenv("ENVIRONMENT"),
	}

	// Defaults
	if config.Port == "" {
		config.Port = "8080"
	}
	if config.Environment == "" {
		config.Environment = "development"
	}
	config.Debug = config.Environment != "production"

	// Validation
	if config.DatabaseURL == "" {
		log.Fatal("DATABASE_URL required")
	}
	if config.JWTSecret == "" {
		log.Fatal("JWT_SECRET required")
	}

	return config
}

func main() {
	config := loadConfig()

	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// Database
	db, err := database.NewConnection(config.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	// Gin router
	router := gin.New()
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

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"service":     "talas-backend",
			"version":     "1.0.0",
			"timestamp":   time.Now().Unix(),
			"environment": config.Environment,
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Setup module routes
		setupAPIRoutes(api, db, config.JWTSecret)
	}

	// Server
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("🚀 Talas backend starting on port %s", config.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}

func setupAPIRoutes(api *gin.RouterGroup, db *database.DB, jwtSecret string) {
	// Auth routes
	authService := user.NewService(db) // Temporaire, auth sera séparé
	authHandler := auth.NewHandler(authService)
	auth.SetupRoutes(api, authHandler, jwtSecret)

	// User routes  
	userService := user.NewService(db)
	userHandler := user.NewHandler(userService)
	user.SetupRoutes(api, userHandler, jwtSecret)

	// Admin routes
	adminService := admin.NewService(db)
	adminHandler := admin.NewHandler(adminService)
	admin.SetupRoutes(api, adminHandler, jwtSecret)

	// TODO: Autres modules
}
```

### Action 1.6 : Mise à jour des structures existantes

**Corriger `internal/api/user/routes.go` :**
```go
package user

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	users := router.Group("/users")
	
	// Routes publiques
	users.GET("", handler.GetUsers)
	users.GET("/:id", handler.GetUserByID) // TODO: Implémenter
	
	// Routes protégées
	protected := users.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		protected.GET("/me", handler.GetMe)
		protected.PUT("/me", handler.UpdateMe)
		protected.PUT("/password", handler.ChangePassword)
		protected.GET("/except-me", handler.GetUsersExceptMe)
		// TODO: Avatar routes
	}
}
```

## ✅ Checklist de validation

Après cette étape, vérifier :

```bash
# 1. Structure des modules
ls -la internal/api/*/
# Attendu : admin/, auth/, file/, product/, user/, [autres]/

# 2. Fichiers créés
find internal/api/ -name "*.go" | sort
# Attendu : handler.go, service.go, routes.go dans chaque module

# 3. main.go simplifié
wc -l cmd/server/main.go
# Attendu : ~150 lignes (vs ~300+ avant)

# 4. Suppression legacy
ls internal/routes/ 2>/dev/null || echo "✅ Legacy routes supprimé"

# 5. Compilation basique (peut échouer sur imports)
go mod tidy
go build ./cmd/server 2>&1 | head -5
```

## 🚨 Points d'attention

1. **Ne pas implémenter la logique métier** dans cette étape
2. **Garder les TODO** pour l'étape 3
3. **Vérifier que tous les modules ont la même structure**
4. **Ne pas corriger les imports** (étape 2)

## ⏭️ Étape suivante
Une fois cette étape terminée → `02_correction_imports.md`

---

**💾 IMPORTANT** : Commit après cette étape
```bash
git add .
git commit -m "Étape 1: Consolidation architecture - modules API unifiés"
```