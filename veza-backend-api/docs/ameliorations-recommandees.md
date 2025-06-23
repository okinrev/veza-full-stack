# Améliorations Recommandées - Veza Backend API

## Vue d'ensemble

Après analyse du code existant, voici les améliorations recommandées pour rendre l'implémentation plus propre, plus maintenable et plus professionnelle.

## 1. Structure et Organisation

### 1.1 Nettoyage des Répertoires

#### Problème identifié
- Duplication de répertoires : `shared_resources` et `shared_ressources`
- Structure incohérente dans certains modules

#### Solution recommandée
```bash
# Fusionner les répertoires dupliqués
mv internal/api/shared_ressources/* internal/api/shared_resources/
rm -rf internal/api/shared_ressources/

# Standardiser la structure
mkdir -p internal/api/shared_resources/{handlers,services,models}
```

#### Implémentation
```go
// Créer un fichier de migration pour la structure
// scripts/clean_structure.go
package main

import (
    "fmt"
    "os"
    "path/filepath"
)

func main() {
    // Script pour nettoyer la structure des répertoires
    fmt.Println("Nettoyage de la structure des répertoires...")
    
    // Fusionner shared_ressources vers shared_resources
    if err := mergeDirectories("internal/api/shared_ressources", "internal/api/shared_resources"); err != nil {
        fmt.Printf("Erreur lors de la fusion: %v\n", err)
    }
    
    fmt.Println("Structure nettoyée avec succès!")
}
```

### 1.2 Standardisation des Patterns

#### Problème identifié
- Incohérence dans la structure des handlers/services
- Patterns d'injection de dépendance non uniformes

#### Solution recommandée
```go
// internal/common/patterns.go
package common

// BaseHandler structure commune pour tous les handlers
type BaseHandler struct {
    service BaseService
    logger  Logger
}

// BaseService interface commune pour tous les services
type BaseService interface {
    Validate() error
    Execute() error
}

// Standardiser l'injection de dépendances
type ServiceContainer struct {
    DB     *database.DB
    Config *config.Config
    Logger Logger
}

func NewServiceContainer(db *database.DB, cfg *config.Config) *ServiceContainer {
    return &ServiceContainer{
        DB:     db,
        Config: cfg,
        Logger: NewLogger(cfg.Server.Environment),
    }
}
```

## 2. Implémentation des TODOs

### 2.1 Services de Listing

#### Fichier : `internal/api/listing/service.go`

```go
// internal/api/listing/service.go
package listing

import (
    "database/sql"
    "fmt"
    "time"
    
    "github.com/okinrev/veza-web-app/internal/database"
    "github.com/okinrev/veza-web-app/internal/models"
)

type Service struct {
    db *database.DB
}

func NewService(db *database.DB) *Service {
    return &Service{db: db}
}

// CreateListing implémente la création d'une annonce
func (s *Service) CreateListing(userID int, listing *models.CreateListingRequest) (*models.Listing, error) {
    query := `
        INSERT INTO listings (user_id, title, description, category, price, location, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, 'active', NOW(), NOW())
        RETURNING id, created_at, updated_at
    `
    
    var listingModel models.Listing
    err := s.db.QueryRow(query, userID, listing.Title, listing.Description, 
        listing.Category, listing.Price, listing.Location).Scan(
        &listingModel.ID, &listingModel.CreatedAt, &listingModel.UpdatedAt)
    
    if err != nil {
        return nil, fmt.Errorf("failed to create listing: %w", err)
    }
    
    // Copier les données de la requête
    listingModel.UserID = userID
    listingModel.Title = listing.Title
    listingModel.Description = listing.Description
    listingModel.Category = listing.Category
    listingModel.Price = listing.Price
    listingModel.Location = listing.Location
    listingModel.Status = "active"
    
    return &listingModel, nil
}

// GetListingsByUser récupère les annonces d'un utilisateur
func (s *Service) GetListingsByUser(userID int, page, limit int) ([]*models.Listing, int, error) {
    offset := (page - 1) * limit
    
    // Compter le total
    var total int
    err := s.db.QueryRow("SELECT COUNT(*) FROM listings WHERE user_id = $1", userID).Scan(&total)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to count listings: %w", err)
    }
    
    // Récupérer les annonces
    query := `
        SELECT id, user_id, title, description, category, price, location, status, created_at, updated_at
        FROM listings 
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
    
    rows, err := s.db.Query(query, userID, limit, offset)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to get listings: %w", err)
    }
    defer rows.Close()
    
    var listings []*models.Listing
    for rows.Next() {
        var listing models.Listing
        err := rows.Scan(&listing.ID, &listing.UserID, &listing.Title, 
            &listing.Description, &listing.Category, &listing.Price, 
            &listing.Location, &listing.Status, &listing.CreatedAt, &listing.UpdatedAt)
        if err != nil {
            return nil, 0, fmt.Errorf("failed to scan listing: %w", err)
        }
        listings = append(listings, &listing)
    }
    
    return listings, total, nil
}

// SearchListings implémente la recherche d'annonces
func (s *Service) SearchListings(query string, category string, minPrice, maxPrice float64, location string, page, limit int) ([]*models.Listing, int, error) {
    // Construction de la requête SQL dynamique
    var conditions []string
    var args []interface{}
    argIndex := 1
    
    baseQuery := `
        SELECT id, user_id, title, description, category, price, location, status, created_at, updated_at
        FROM listings 
        WHERE status = 'active'
    `
    
    if query != "" {
        conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
        args = append(args, "%"+query+"%")
        argIndex++
    }
    
    if category != "" {
        conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
        args = append(args, category)
        argIndex++
    }
    
    if minPrice > 0 {
        conditions = append(conditions, fmt.Sprintf("price >= $%d", argIndex))
        args = append(args, minPrice)
        argIndex++
    }
    
    if maxPrice > 0 {
        conditions = append(conditions, fmt.Sprintf("price <= $%d", argIndex))
        args = append(args, maxPrice)
        argIndex++
    }
    
    if location != "" {
        conditions = append(conditions, fmt.Sprintf("location ILIKE $%d", argIndex))
        args = append(args, "%"+location+"%")
        argIndex++
    }
    
    if len(conditions) > 0 {
        baseQuery += " AND " + strings.Join(conditions, " AND ")
    }
    
    // Compter le total
    countQuery := strings.Replace(baseQuery, "SELECT id, user_id, title, description, category, price, location, status, created_at, updated_at", "SELECT COUNT(*)", 1)
    var total int
    err := s.db.QueryRow(countQuery, args...).Scan(&total)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to count search results: %w", err)
    }
    
    // Ajouter pagination
    offset := (page - 1) * limit
    baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
    args = append(args, limit, offset)
    
    rows, err := s.db.Query(baseQuery, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to search listings: %w", err)
    }
    defer rows.Close()
    
    var listings []*models.Listing
    for rows.Next() {
        var listing models.Listing
        err := rows.Scan(&listing.ID, &listing.UserID, &listing.Title, 
            &listing.Description, &listing.Category, &listing.Price, 
            &listing.Location, &listing.Status, &listing.CreatedAt, &listing.UpdatedAt)
        if err != nil {
            return nil, 0, fmt.Errorf("failed to scan listing: %w", err)
        }
        listings = append(listings, &listing)
    }
    
    return listings, total, nil
}

// UpdateListing met à jour une annonce
func (s *Service) UpdateListing(listingID, userID int, updateData *models.UpdateListingRequest) (*models.Listing, error) {
    // Vérifier que l'utilisateur est propriétaire
    var ownerID int
    err := s.db.QueryRow("SELECT user_id FROM listings WHERE id = $1", listingID).Scan(&ownerID)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("listing not found")
        }
        return nil, fmt.Errorf("failed to check ownership: %w", err)
    }
    
    if ownerID != userID {
        return nil, fmt.Errorf("unauthorized: not listing owner")
    }
    
    // Mettre à jour
    query := `
        UPDATE listings 
        SET title = $1, description = $2, category = $3, price = $4, location = $5, updated_at = NOW()
        WHERE id = $6
        RETURNING id, user_id, title, description, category, price, location, status, created_at, updated_at
    `
    
    var listing models.Listing
    err = s.db.QueryRow(query, updateData.Title, updateData.Description, 
        updateData.Category, updateData.Price, updateData.Location, listingID).Scan(
        &listing.ID, &listing.UserID, &listing.Title, &listing.Description,
        &listing.Category, &listing.Price, &listing.Location, &listing.Status,
        &listing.CreatedAt, &listing.UpdatedAt)
    
    if err != nil {
        return nil, fmt.Errorf("failed to update listing: %w", err)
    }
    
    return &listing, nil
}

// DeleteListing supprime une annonce
func (s *Service) DeleteListing(listingID, userID int) error {
    // Vérifier que l'utilisateur est propriétaire
    var ownerID int
    err := s.db.QueryRow("SELECT user_id FROM listings WHERE id = $1", listingID).Scan(&ownerID)
    if err != nil {
        if err == sql.ErrNoRows {
            return fmt.Errorf("listing not found")
        }
        return fmt.Errorf("failed to check ownership: %w", err)
    }
    
    if ownerID != userID {
        return fmt.Errorf("unauthorized: not listing owner")
    }
    
    // Supprimer (soft delete)
    _, err = s.db.Exec("UPDATE listings SET status = 'deleted', updated_at = NOW() WHERE id = $1", listingID)
    if err != nil {
        return fmt.Errorf("failed to delete listing: %w", err)
    }
    
    return nil
}
```

### 2.2 Modèles de Données Manquants

```go
// internal/models/listing.go
package models

import (
    "database/sql"
    "time"
)

type Listing struct {
    ID          int            `db:"id" json:"id"`
    UserID      int            `db:"user_id" json:"user_id"`
    Title       string         `db:"title" json:"title"`
    Description string         `db:"description" json:"description"`
    Category    string         `db:"category" json:"category"`
    Price       float64        `db:"price" json:"price"`
    Location    string         `db:"location" json:"location"`
    Status      string         `db:"status" json:"status"` // active, sold, deleted
    Images      []string       `json:"images,omitempty"`
    CreatedAt   time.Time      `db:"created_at" json:"created_at"`
    UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
}

type CreateListingRequest struct {
    Title       string  `json:"title" validate:"required,min=3,max=100"`
    Description string  `json:"description" validate:"required,min=10,max=1000"`
    Category    string  `json:"category" validate:"required"`
    Price       float64 `json:"price" validate:"required,gt=0"`
    Location    string  `json:"location" validate:"required"`
}

type UpdateListingRequest struct {
    Title       string  `json:"title" validate:"required,min=3,max=100"`
    Description string  `json:"description" validate:"required,min=10,max=1000"`
    Category    string  `json:"category" validate:"required"`
    Price       float64 `json:"price" validate:"required,gt=0"`
    Location    string  `json:"location" validate:"required"`
}

// internal/models/offer.go
package models

type Offer struct {
    ID        int       `db:"id" json:"id"`
    ListingID int       `db:"listing_id" json:"listing_id"`
    BuyerID   int       `db:"buyer_id" json:"buyer_id"`
    Amount    float64   `db:"amount" json:"amount"`
    Message   string    `db:"message" json:"message"`
    Status    string    `db:"status" json:"status"` // pending, accepted, rejected
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CreateOfferRequest struct {
    ListingID int     `json:"listing_id" validate:"required"`
    Amount    float64 `json:"amount" validate:"required,gt=0"`
    Message   string  `json:"message" validate:"max=500"`
}
```

## 3. Gestion d'Erreurs Améliorée

### 3.1 Erreurs Personnalisées

```go
// internal/utils/errors/errors.go
package errors

import (
    "fmt"
    "net/http"
)

// Error types
type ErrorType string

const (
    ErrorTypeValidation     ErrorType = "VALIDATION_ERROR"
    ErrorTypeNotFound      ErrorType = "NOT_FOUND"
    ErrorTypeUnauthorized  ErrorType = "UNAUTHORIZED"
    ErrorTypeForbidden     ErrorType = "FORBIDDEN"
    ErrorTypeConflict      ErrorType = "CONFLICT"
    ErrorTypeInternal      ErrorType = "INTERNAL_ERROR"
    ErrorTypeRateLimit     ErrorType = "RATE_LIMIT"
)

// AppError structure d'erreur personnalisée
type AppError struct {
    Type    ErrorType `json:"type"`
    Message string    `json:"message"`
    Code    int       `json:"code"`
    Details map[string]interface{} `json:"details,omitempty"`
}

func (e *AppError) Error() string {
    return e.Message
}

// Constructors d'erreurs
func NewValidationError(message string, details map[string]interface{}) *AppError {
    return &AppError{
        Type:    ErrorTypeValidation,
        Message: message,
        Code:    http.StatusBadRequest,
        Details: details,
    }
}

func NewNotFoundError(resource string) *AppError {
    return &AppError{
        Type:    ErrorTypeNotFound,
        Message: fmt.Sprintf("%s not found", resource),
        Code:    http.StatusNotFound,
    }
}

func NewUnauthorizedError(message string) *AppError {
    return &AppError{
        Type:    ErrorTypeUnauthorized,
        Message: message,
        Code:    http.StatusUnauthorized,
    }
}

func NewForbiddenError(message string) *AppError {
    return &AppError{
        Type:    ErrorTypeForbidden,
        Message: message,
        Code:    http.StatusForbidden,
    }
}

func NewConflictError(message string) *AppError {
    return &AppError{
        Type:    ErrorTypeConflict,
        Message: message,
        Code:    http.StatusConflict,
    }
}

func NewInternalError(message string) *AppError {
    return &AppError{
        Type:    ErrorTypeInternal,
        Message: message,
        Code:    http.StatusInternalServerError,
    }
}

// ErrorHandler middleware pour gérer les erreurs
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            var appErr *AppError
            if errors.As(err, &appErr) {
                c.JSON(appErr.Code, gin.H{
                    "success": false,
                    "error":   appErr.Message,
                    "type":    appErr.Type,
                    "details": appErr.Details,
                })
            } else {
                // Erreur non gérée
                c.JSON(http.StatusInternalServerError, gin.H{
                    "success": false,
                    "error":   "Internal server error",
                    "type":    ErrorTypeInternal,
                })
            }
        }
    }
}
```

### 3.2 Validation Améliorée

```go
// internal/utils/validation/validator.go
package validation

import (
    "reflect"
    "strings"
    
    "github.com/go-playground/validator/v10"
    "github.com/okinrev/veza-web-app/internal/utils/errors"
)

type Validator struct {
    validator *validator.Validate
}

func NewValidator() *Validator {
    v := validator.New()
    
    // Utiliser les noms JSON pour les erreurs
    v.RegisterTagNameFunc(func(fld reflect.StructField) string {
        name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
        if name == "-" {
            return ""
        }
        return name
    })
    
    // Ajouter des validations personnalisées
    v.RegisterValidation("password", validatePassword)
    v.RegisterValidation("username", validateUsername)
    
    return &Validator{validator: v}
}

func (v *Validator) Validate(s interface{}) error {
    err := v.validator.Struct(s)
    if err == nil {
        return nil
    }
    
    validationErrors := err.(validator.ValidationErrors)
    errorDetails := make(map[string]interface{})
    
    for _, fieldError := range validationErrors {
        field := fieldError.Field()
        tag := fieldError.Tag()
        
        switch tag {
        case "required":
            errorDetails[field] = "This field is required"
        case "email":
            errorDetails[field] = "Invalid email format"
        case "min":
            errorDetails[field] = fmt.Sprintf("Minimum length is %s", fieldError.Param())
        case "max":
            errorDetails[field] = fmt.Sprintf("Maximum length is %s", fieldError.Param())
        case "password":
            errorDetails[field] = "Password must be at least 8 characters with uppercase, lowercase, and number"
        case "username":
            errorDetails[field] = "Username must be 3-50 characters, alphanumeric only"
        default:
            errorDetails[field] = fmt.Sprintf("Validation failed for %s", tag)
        }
    }
    
    return errors.NewValidationError("Validation failed", errorDetails)
}

// Validations personnalisées
func validatePassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    if len(password) < 8 {
        return false
    }
    
    hasUpper := false
    hasLower := false
    hasNumber := false
    
    for _, char := range password {
        switch {
        case char >= 'A' && char <= 'Z':
            hasUpper = true
        case char >= 'a' && char <= 'z':
            hasLower = true
        case char >= '0' && char <= '9':
            hasNumber = true
        }
    }
    
    return hasUpper && hasLower && hasNumber
}

func validateUsername(fl validator.FieldLevel) bool {
    username := fl.Field().String()
    if len(username) < 3 || len(username) > 50 {
        return false
    }
    
    for _, char := range username {
        if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_') {
            return false
        }
    }
    
    return true
}
```

## 4. Logging Amélioré

### 4.1 Logger Structuré

```go
// internal/utils/logger/logger.go
package logger

import (
    "context"
    "log/slog"
    "os"
    "time"
)

type Logger struct {
    *slog.Logger
}

func NewLogger(env string) *Logger {
    var handler slog.Handler
    
    if env == "production" {
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelInfo,
        })
    } else {
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        })
    }
    
    logger := slog.New(handler)
    return &Logger{Logger: logger}
}

// Méthodes de logging spécialisées
func (l *Logger) LogRequest(ctx context.Context, method, path string, userID int, duration time.Duration, status int) {
    l.InfoContext(ctx, "HTTP Request",
        slog.String("method", method),
        slog.String("path", path),
        slog.Int("user_id", userID),
        slog.Duration("duration", duration),
        slog.Int("status", status),
    )
}

func (l *Logger) LogAuth(ctx context.Context, event string, userID int, email string, success bool) {
    l.InfoContext(ctx, "Auth Event",
        slog.String("event", event),
        slog.Int("user_id", userID),
        slog.String("email", email),
        slog.Bool("success", success),
    )
}

func (l *Logger) LogError(ctx context.Context, err error, details map[string]interface{}) {
    attrs := []slog.Attr{
        slog.String("error", err.Error()),
    }
    
    for key, value := range details {
        attrs = append(attrs, slog.Any(key, value))
    }
    
    l.ErrorContext(ctx, "Application Error", attrs...)
}
```

## 5. Tests Améliorés

### 5.1 Tests d'Intégration

```go
// internal/tests/integration_test.go
package tests

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "github.com/okinrev/veza-web-app/internal/api"
    "github.com/okinrev/veza-web-app/internal/config"
    "github.com/okinrev/veza-web-app/internal/database"
)

type IntegrationTestSuite struct {
    suite.Suite
    db     *database.DB
    router *gin.Engine
    config *config.Config
}

func (suite *IntegrationTestSuite) SetupSuite() {
    // Configuration de test
    suite.config = &config.Config{
        Database: config.DatabaseConfig{
            URL: "postgres://postgres:password@localhost:5432/veza_test?sslmode=disable",
        },
        JWT: config.JWTConfig{
            Secret: "test-secret",
        },
    }
    
    // Connexion base de données de test
    db, err := database.NewConnection(suite.config.Database.URL)
    suite.Require().NoError(err)
    suite.db = db
    
    // Migrations
    err = database.RunMigrations(db)
    suite.Require().NoError(err)
    
    // Router
    suite.router = gin.New()
    api.SetupRoutes(suite.router, db, suite.config)
}

func (suite *IntegrationTestSuite) TearDownSuite() {
    suite.db.Close()
}

func (suite *IntegrationTestSuite) SetupTest() {
    // Nettoyer les tables avant chaque test
    suite.db.Exec("TRUNCATE TABLE users, listings, offers, messages CASCADE")
}

func (suite *IntegrationTestSuite) TestUserRegistrationAndLogin() {
    // Test d'inscription
    registerData := map[string]interface{}{
        "username": "testuser",
        "email":    "test@example.com",
        "password": "Password123",
    }
    
    body, _ := json.Marshal(registerData)
    req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    suite.router.ServeHTTP(w, req)
    
    assert.Equal(suite.T(), http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.True(suite.T(), response["success"].(bool))
    
    // Test de connexion
    loginData := map[string]interface{}{
        "email":    "test@example.com",
        "password": "Password123",
    }
    
    body, _ = json.Marshal(loginData)
    req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    w = httptest.NewRecorder()
    suite.router.ServeHTTP(w, req)
    
    assert.Equal(suite.T(), http.StatusOK, w.Code)
    
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.True(suite.T(), response["success"].(bool))
    assert.NotEmpty(suite.T(), response["data"].(map[string]interface{})["token"])
}

func TestIntegrationSuite(t *testing.T) {
    suite.Run(t, new(IntegrationTestSuite))
}
```

## 6. Monitoring et Métriques

### 6.1 Middleware de Métriques

```go
// internal/middleware/metrics.go
package middleware

import (
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "Duration of HTTP requests",
        },
        []string{"method", "path"},
    )
    
    activeConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "websocket_connections_active",
            Help: "Number of active WebSocket connections",
        },
    )
)

func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        status := strconv.Itoa(c.Writer.Status())
        
        httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
        httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration.Seconds())
    }
}

// WebSocket metrics
func IncrementActiveConnections() {
    activeConnections.Inc()
}

func DecrementActiveConnections() {
    activeConnections.Dec()
}
```

## 7. Configuration Avancée

### 7.1 Configuration par Environnement

```go
// internal/config/environments.go
package config

import (
    "fmt"
    "os"
)

type Environment string

const (
    Development Environment = "development"
    Testing     Environment = "testing"
    Production  Environment = "production"
)

// LoadConfig charge la configuration selon l'environnement
func LoadConfig() (*Config, error) {
    env := Environment(os.Getenv("ENVIRONMENT"))
    if env == "" {
        env = Development
    }
    
    switch env {
    case Development:
        return loadDevelopmentConfig()
    case Testing:
        return loadTestingConfig()
    case Production:
        return loadProductionConfig()
    default:
        return nil, fmt.Errorf("unknown environment: %s", env)
    }
}

func loadDevelopmentConfig() (*Config, error) {
    config := New()
    
    // Surcharges pour le développement
    config.Server.ReadTimeout = 10 * time.Second
    config.Server.WriteTimeout = 10 * time.Second
    config.Database.MaxOpenConns = 10
    
    return config, nil
}

func loadTestingConfig() (*Config, error) {
    config := New()
    
    // Configuration pour les tests
    config.Database.URL = "postgres://postgres:password@localhost:5432/veza_test?sslmode=disable"
    config.JWT.Secret = "test-secret-key"
    
    return config, nil
}

func loadProductionConfig() (*Config, error) {
    config := New()
    
    // Validations strictes pour la production
    if config.JWT.Secret == "your-super-secret-key-change-in-production" {
        return nil, fmt.Errorf("JWT secret must be changed in production")
    }
    
    if config.Database.SSLMode == "disable" {
        return nil, fmt.Errorf("SSL must be enabled in production")
    }
    
    // Configuration optimisée pour la production
    config.Server.ReadTimeout = 30 * time.Second
    config.Server.WriteTimeout = 30 * time.Second
    config.Database.MaxOpenConns = 50
    config.Database.MaxIdleConns = 10
    
    return config, nil
}
```

## Plan d'Implémentation

### Phase 1 : Nettoyage et Structure (Semaine 1)
1. Nettoyer la structure des répertoires
2. Standardiser les patterns
3. Implémenter le système d'erreurs amélioré

### Phase 2 : Implémentation des TODOs (Semaine 2-3)
1. Implémenter tous les services manquants
2. Créer les modèles de données manquants
3. Ajouter la validation complète

### Phase 3 : Tests et Qualité (Semaine 4)
1. Écrire les tests d'intégration
2. Ajouter le monitoring
3. Améliorer le logging

### Phase 4 : Optimisation (Semaine 5)
1. Configuration par environnement
2. Performance optimizations
3. Documentation finale

Ces améliorations rendront le projet plus professionnel, maintenable et prêt pour la production. 