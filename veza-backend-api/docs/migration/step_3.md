# Étape 3 - Implémentation des handlers

## 🎯 Objectif
Remplacer tous les TODO par la vraie logique métier en s'appuyant sur la documentation existante.

## ⏱️ Durée estimée : 60-90 minutes

## 📚 Sources de documentation
- `backend/docs/doc_*_handler.md` (documentation détaillée de chaque handler)
- `internal/models/*.go` (structures de données)
- `internal/database/migrations/*.sql` (schéma de base)

## 📋 Plan d'implémentation

### Phase 3.1 : Handlers prioritaires (authentification)
1. **auth** - Authentication essentielle
2. **user** - Gestion utilisateurs
3. **admin** - Administration

### Phase 3.2 : Handlers métier  
4. **track** - Gestion audio
5. **file** - Gestion fichiers
6. **product** - Catalogue produits

### Phase 3.3 : Handlers communautaires
7. **listing** - Marketplace
8. **offer** - Offres d'échange  
9. **message** - Chat/messages
10. **room** - Salons de discussion

### Phase 3.4 : Handlers utilitaires
11. **search** - Recherche globale
12. **tag** - Gestion des tags
13. **shared_ressources** - Ressources partagées

## 🔧 Implementation détaillée

### Phase 3.1 : Auth Handler

#### `internal/api/auth/handler.go`
```go
package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/common"
	"github.com/okinrev/veza-web-app/internal/utils/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         interface{} `json:"user"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Register crée un nouveau compte utilisateur
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.service.Register(req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			response.ErrorJSON(c.Writer, err.Error(), http.StatusConflict)
			return
		}
		response.ErrorJSON(c.Writer, "Registration failed", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(c.Writer, map[string]interface{}{
		"user_id": user.ID,
		"username": user.Username,
		"email": user.Email,
	}, "User registered successfully")
}

// Login authentifie un utilisateur
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data: "+err.Error(), http.StatusBadRequest)
		return
	}

	loginResp, err := h.service.Login(req)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	response.SuccessJSON(c.Writer, loginResp, "Login successful")
}

// RefreshToken génère un nouveau token d'accès
func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data", http.StatusBadRequest)
		return
	}

	tokenResp, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		response.ErrorJSON(c.Writer, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	response.SuccessJSON(c.Writer, tokenResp, "Token refreshed")
}

// Logout invalide le token de rafraîchissement
func (h *Handler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c.Writer, "Invalid request data", http.StatusBadRequest)
		return
	}

	err := h.service.Logout(req.RefreshToken)
	if err != nil {
		response.ErrorJSON(c.Writer, "Logout failed", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(c.Writer, nil, "Logged out successfully")
}

// GetMe retourne le profil de l'utilisateur connecté
func (h *Handler) GetMe(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User not authenticated", http.StatusUnauthorized)
		return
	}

	user, err := h.service.GetMe(userID)
	if err != nil {
		response.ErrorJSON(c.Writer, "User not found", http.StatusNotFound)
		return
	}

	response.SuccessJSON(c.Writer, user, "User profile retrieved")
}
```

#### `internal/api/auth/service.go`
```go
package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/models"
	"github.com/okinrev/veza-web-app/internal/utils"
)

type Service struct {
	db        *database.DB
	jwtSecret string
}

func NewService(db *database.DB, jwtSecret string) *Service {
	return &Service{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// Register crée un nouveau compte utilisateur
func (s *Service) Register(req RegisterRequest) (*models.User, error) {
	// Normalisation
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)

	// Vérifier email unique
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("email already exists")
	}

	// Vérifier username unique
	err = s.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", req.Username).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("username already exists")
	}

	// Hasher le mot de passe
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Créer l'utilisateur
	var user models.User
	err = s.db.QueryRow(`
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, 'user', NOW(), NOW()) 
		RETURNING id, username, email, role, created_at, updated_at
	`, req.Username, req.Email, hashedPassword).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role, 
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// Login authentifie un utilisateur
func (s *Service) Login(req LoginRequest) (*LoginResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// Récupérer l'utilisateur
	var user models.User
	var passwordHash string
	err := s.db.QueryRow(`
		SELECT id, username, email, password_hash, role, created_at, updated_at 
		FROM users WHERE email = $1 AND is_active = true
	`, req.Email).Scan(
		&user.ID, &user.Username, &user.Email, &passwordHash, 
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Vérifier le mot de passe
	if err := utils.CheckPasswordHash(req.Password, passwordHash); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Générer les tokens
	accessToken, refreshToken, expiresIn, err := utils.GenerateTokenPair(
		user.ID, user.Username, user.Role, s.jwtSecret,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Stocker le refresh token
	_, err = s.db.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, NOW() + INTERVAL '7 days', NOW())
		ON CONFLICT (user_id) DO UPDATE SET 
			token = EXCLUDED.token, 
			expires_at = EXCLUDED.expires_at,
			created_at = EXCLUDED.created_at
	`, user.ID, refreshToken)

	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
		ExpiresIn:    expiresIn,
	}, nil
}

// RefreshToken génère un nouveau token d'accès
func (s *Service) RefreshToken(refreshToken string) (*TokenResponse, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT u.id, u.username, u.email, u.role, u.created_at, u.updated_at
		FROM refresh_tokens rt
		JOIN users u ON u.id = rt.user_id
		WHERE rt.token = $1 AND rt.expires_at > NOW() AND u.is_active = true
	`, refreshToken).Scan(
		&user.ID, &user.Username, &user.Email, &user.Role, 
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Générer nouveau token d'accès
	accessToken, expiresIn, err := utils.GenerateAccessToken(
		user.ID, user.Username, user.Role, s.jwtSecret,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &TokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	}, nil
}

// Logout invalide le refresh token
func (s *Service) Logout(refreshToken string) error {
	_, err := s.db.Exec("DELETE FROM refresh_tokens WHERE token = $1", refreshToken)
	return err
}

// GetMe récupère le profil utilisateur
func (s *Service) GetMe(userID int) (*models.UserResponse, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, username, email, first_name, last_name, bio, avatar, 
		       role, is_active, is_verified, last_login_at, created_at, updated_at
		FROM users WHERE id = $1 AND is_active = true
	`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, 
		&user.LastName, &user.Bio, &user.Avatar, &user.Role,
		&user.IsActive, &user.IsVerified, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user.ToResponse(), nil
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}
```

#### `internal/api/auth/routes.go`
```go
package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	auth := router.Group("/auth")
	{
		// Routes publiques
		auth.POST("/register", handler.Register)
		auth.POST("/signup", handler.Register) // Alias
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.RefreshToken)
		auth.POST("/logout", handler.Logout)

		// Routes protégées
		protected := auth.Group("")
		protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
		{
			protected.GET("/me", handler.GetMe)
		}
	}
}
```

### Phase 3.2 : Admin Handler

#### `internal/api/admin/handler.go` (complet)
```go
package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/common"
	"github.com/okinrev/veza-web-app/internal/utils/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Dashboard retourne les statistiques admin (basé sur doc_admin_handler.md)
func (h *Handler) Dashboard(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if !h.service.IsAdmin(userID) {
		response.ErrorJSON(c.Writer, "Admin access required", http.StatusForbidden)
		return
	}

	stats, err := h.service.GetDashboardStats()
	if err != nil {
		response.ErrorJSON(c.Writer, "Failed to get dashboard stats", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(c.Writer, stats, "Dashboard stats retrieved")
}

// GetUsers retourne la liste des utilisateurs avec pagination
func (h *Handler) GetUsers(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if !h.service.IsAdmin(userID) {
		response.ErrorJSON(c.Writer, "Admin access required", http.StatusForbidden)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	role := c.Query("role")

	users, total, err := h.service.GetUsers(page, limit, search, role)
	if err != nil {
		response.ErrorJSON(c.Writer, "Failed to get users", http.StatusInternalServerError)
		return
	}

	meta := &response.Meta{
		Page:       page,
		PerPage:    limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}

	response.PaginatedJSON(c.Writer, users, meta, "Users retrieved successfully")
}

// GetAnalytics retourne les analytics de contenu
func (h *Handler) GetAnalytics(c *gin.Context) {
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User not authenticated", http.StatusUnauthorized)
		return
	}

	if !h.service.IsAdmin(userID) {
		response.ErrorJSON(c.Writer, "Admin access required", http.StatusForbidden)
		return
	}

	analytics, err := h.service.GetAnalytics()
	if err != nil {
		response.ErrorJSON(c.Writer, "Failed to get analytics", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(c.Writer, analytics, "Analytics retrieved successfully")
}

// TODO: Implémenter GetCategories, CreateCategory, UpdateCategory, DeleteCategory
// Basé sur doc_admin_handler.md sections Category CRUD
```

#### `internal/api/admin/service.go` (complet)
```go
package admin

import (
	"fmt"

	"github.com/okinrev/veza-web-app/internal/database"
	"github.com/okinrev/veza-web-app/internal/models"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

// IsAdmin vérifie si l'utilisateur a les droits admin
func (s *Service) IsAdmin(userID int) bool {
	var role string
	err := s.db.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&role)
	if err != nil {
		return false
	}
	return role == "admin" || role == "super_admin"
}

// GetDashboardStats récupère les statistiques du dashboard
func (s *Service) GetDashboardStats() (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}

	// Requêtes parallèles pour les stats (basé sur doc_admin_handler.md)
	queries := map[string]string{
		"total_users":           "SELECT COUNT(*) FROM users WHERE is_active = true",
		"active_users":          "SELECT COUNT(*) FROM users WHERE is_active = true AND last_login_at > NOW() - INTERVAL '30 days'",
		"total_tracks":          "SELECT COUNT(*) FROM tracks",
		"public_tracks":         "SELECT COUNT(*) FROM tracks WHERE is_public = true",
		"total_shared_resources": "SELECT COUNT(*) FROM shared_resources",
		"total_listings":        "SELECT COUNT(*) FROM listings",
		"active_listings":       "SELECT COUNT(*) FROM listings WHERE status = 'open'",
		"total_offers":          "SELECT COUNT(*) FROM offers",
		"pending_offers":        "SELECT COUNT(*) FROM offers WHERE status = 'pending'",
		"total_messages":        "SELECT COUNT(*) FROM messages",
		"total_rooms":           "SELECT COUNT(*) FROM rooms",
	}

	// Exécuter les requêtes
	err := s.db.QueryRow(queries["total_users"]).Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(queries["active_users"]).Scan(&stats.ActiveUsers)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(queries["total_tracks"]).Scan(&stats.TotalTracks)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(queries["public_tracks"]).Scan(&stats.PublicTracks)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(queries["total_shared_resources"]).Scan(&stats.TotalSharedResources)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(queries["total_listings"]).Scan(&stats.TotalListings)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(queries["active_listings"]).Scan(&stats.ActiveListings)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(queries["total_offers"]).Scan(&stats.TotalOffers)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(queries["pending_offers"]).Scan(&stats.PendingOffers)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(queries["total_messages"]).Scan(&stats.TotalMessages)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(queries["total_rooms"]).Scan(&stats.TotalRooms)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetUsers récupère les utilisateurs avec pagination et filtres
func (s *Service) GetUsers(page, limit int, search, role string) ([]models.UserAnalytics, int, error) {
	// Validation
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Construction de la requête dynamique
	baseQuery := `
		SELECT u.id, u.username, u.email, u.role, 
		       COALESCE(t.tracks_count, 0) as tracks_count,
		       COALESCE(sr.resources_count, 0) as resources_count,
		       COALESCE(l.listings_count, 0) as listings_count,
		       COALESCE(m.messages_count, 0) as messages_count,
		       u.created_at as registration_date,
		       u.last_login_at as last_activity,
		       u.is_active
		FROM users u
		LEFT JOIN (SELECT uploader_id, COUNT(*) as tracks_count FROM tracks GROUP BY uploader_id) t ON u.id = t.uploader_id
		LEFT JOIN (SELECT uploader_id, COUNT(*) as resources_count FROM shared_resources GROUP BY uploader_id) sr ON u.id = sr.uploader_id
		LEFT JOIN (SELECT user_id, COUNT(*) as listings_count FROM listings GROUP BY user_id) l ON u.id = l.user_id
		LEFT JOIN (SELECT from_user, COUNT(*) as messages_count FROM messages GROUP BY from_user) m ON u.id = m.from_user
		WHERE u.is_active = true
	`

	countQuery := "SELECT COUNT(*) FROM users WHERE is_active = true"

	args := []interface{}{}
	argIndex := 1

	// Filtres
	if search != "" {
		baseQuery += fmt.Sprintf(" AND (u.username ILIKE $%d OR u.email ILIKE $%d)", argIndex, argIndex)
		countQuery += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	if role != "" {
		baseQuery += fmt.Sprintf(" AND u.role = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND role = $%d", argIndex)
		args = append(args, role)
		argIndex++
	}

	// Total
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Données
	baseQuery += fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.UserAnalytics
	for rows.Next() {
		var user models.UserAnalytics
		err := rows.Scan(
			&user.UserID, &user.Username, &user.Email, &user.Role,
			&user.TracksCount, &user.ResourcesCount, &user.ListingsCount, &user.MessagesCount,
			&user.RegistrationDate, &user.LastActivity, &user.IsActive,
		)
		if err != nil {
			continue
		}
		users = append(users, user)
	}

	return users, total, nil
}

// GetAnalytics récupère les analytics de contenu
func (s *Service) GetAnalytics() (*models.ContentAnalytics, error) {
	analytics := &models.ContentAnalytics{}

	// TODO: Implémenter basé sur doc_admin_handler.md
	// - TracksByMonth (12 derniers mois)
	// - ResourcesByMonth 
	// - UsersByMonth
	// - PopularTags
	// - TopUploaders

	return analytics, nil
}
```

## ✅ Template pour les autres handlers

Pour les phases suivantes, utiliser ce template :

### Template de handler
```go
package [module]

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/common"
	"github.com/okinrev/veza-web-app/internal/utils/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// [Method] description basée sur doc_[module]_handler.md
func (h *Handler) [Method](c *gin.Context) {
	// 1. Authentification si nécessaire
	userID, exists := common.GetUserIDFromContext(c)
	if !exists {
		response.ErrorJSON(c.Writer, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// 2. Parsing des paramètres
	// 3. Validation métier
	// 4. Appel service
	// 5. Réponse JSON
}
```

## 📝 Checklist par module

### ✅ Auth Handler
- [x] Register (avec validation email/username unique)
- [x] Login (avec génération tokens)
- [x] RefreshToken
- [x] Logout
- [x] GetMe

### ✅ Admin Handler  
- [x] Dashboard (stats complètes)
- [x] GetUsers (avec pagination/filtres)
- [x] IsAdmin (vérification rôle)
- [ ] GetAnalytics (TODO: requêtes complexes)
- [ ] Category CRUD (TODO)

### 🔄 User Handler (à compléter)
- [ ] Terminer les méthodes incomplètes
- [ ] Avatar management
- [ ] Password change
- [ ] Search users

### 🔄 Track Handler (à implémenter)
- [ ] Upload avec validation audio
- [ ] Stream avec signed URLs
- [ ] CRUD tracks
- [ ] Permissions public/private

## 🚨 Points d'attention

1. **Validation des entrées** : Toujours valider avec `c.ShouldBindJSON()`
2. **Authentification** : Vérifier le contexte utilisateur
3. **Autorisations** : Vérifier les permissions (admin, propriétaire)
4. **Gestion d'erreurs** : Messages d'erreur clairs et codes HTTP appropriés
5. **Réponses cohérentes** : Utiliser `response.SuccessJSON/ErrorJSON`

## ⏭️ Étape suivante
Une fois les handlers prioritaires implémentés → `04_consolidation_services.md`

---

**💾 IMPORTANT** : Commit après chaque handler implémenté
```bash
git add .
git commit -m "Étape 3: Implémentation handlers auth + admin"
```