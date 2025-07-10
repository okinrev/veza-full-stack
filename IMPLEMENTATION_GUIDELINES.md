# 🔧 GUIDELINES D'IMPLÉMENTATION - VEZA BACKEND

## 📋 CONVENTIONS DE DÉVELOPPEMENT

### 1. **Structure de Code**

#### Go (Backend API)
```
veza-backend-api/
├── cmd/                    # Points d'entrée
│   ├── server/            # Serveur principal
│   └── production-server/ # Serveur production
├── internal/              # Code privé
│   ├── domain/           # Entités et règles métier
│   │   ├── entities/     # Entités du domaine
│   │   ├── repositories/ # Interfaces des repositories
│   │   └── services/     # Services du domaine
│   ├── application/      # Cas d'usage
│   │   ├── commands/     # Commandes CQRS
│   │   ├── queries/      # Requêtes CQRS
│   │   └── handlers/     # Gestionnaires
│   ├── infrastructure/   # Implémentations techniques
│   │   ├── database/     # PostgreSQL
│   │   ├── cache/        # Redis
│   │   ├── messaging/    # NATS/Kafka
│   │   └── external/     # APIs externes
│   └── interfaces/       # Controllers, présentateurs
│       ├── http/         # Handlers HTTP
│       ├── grpc/         # gRPC services
│       └── websocket/    # WebSocket handlers
├── pkg/                   # Code public réutilisable
├── proto/                 # Définitions protobuf
├── docs/                  # Documentation
├── scripts/               # Scripts utilitaires
├── tests/                 # Tests d'intégration
├── go.mod
├── go.sum
└── Dockerfile
```

#### Rust (Chat/Stream Server)
```
veza-chat-server/
├── src/
│   ├── core/             # Logique métier
│   ├── hub/              # WebSocket hub
│   ├── auth/             # Authentification
│   ├── cache/            # Cache Redis
│   ├── database/         # Base de données
│   ├── monitoring/       # Métriques et logs
│   └── main.rs
├── proto/                 # Définitions protobuf
├── migrations/            # Migrations SQL
├── scripts/               # Scripts utilitaires
├── Cargo.toml
└── Dockerfile
```

### 2. **Conventions de Nommage**

#### Go
```go
// Packages : lowercase, single word
package user

// Types : PascalCase
type UserService struct {}

// Méthodes : PascalCase
func (s *UserService) CreateUser() {}

// Variables : camelCase
var userID int64

// Constantes : PascalCase
const MaxRetries = 3

// Interfaces : PascalCase + "er"
type UserRepository interface {}

// Erreurs : camelCase
var ErrUserNotFound = errors.New("user not found")
```

#### Rust
```rust
// Modules : snake_case
mod user_service;

// Types : PascalCase
struct UserService {}

// Fonctions : snake_case
fn create_user() -> Result<User, Error> {}

// Variables : snake_case
let user_id: i64 = 123;

// Constantes : SCREAMING_SNAKE_CASE
const MAX_RETRIES: u32 = 3;

// Traits : PascalCase
trait UserRepository {}

// Erreurs : PascalCase
#[derive(Debug, thiserror::Error)]
pub enum UserError {
    #[error("User not found")]
    NotFound,
}
```

### 3. **Standards de Code**

#### Go
```go
// Imports organisés
import (
    // Standard library
    "context"
    "time"
    
    // Third party
    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
    
    // Internal
    "github.com/okinrev/veza-web-app/internal/domain/entities"
    "github.com/okinrev/veza-web-app/internal/infrastructure/database"
)

// Documentation des fonctions
// CreateUser crée un nouvel utilisateur avec validation
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*entities.User, error) {
    // Validation
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // Logique métier
    user, err := entities.NewUser(req.Username, req.Email, req.Password)
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    // Persistence
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to save user: %w", err)
    }
    
    return user, nil
}
```

#### Rust
```rust
// Documentation des fonctions
/// Crée un nouvel utilisateur avec validation
pub async fn create_user(
    &self,
    req: CreateUserRequest,
) -> Result<User, UserError> {
    // Validation
    req.validate()?;
    
    // Logique métier
    let user = User::new(req.username, req.email, req.password)?;
    
    // Persistence
    self.repo.create(&user).await?;
    
    Ok(user)
}

// Gestion d'erreurs avec thiserror
#[derive(Debug, thiserror::Error)]
pub enum UserError {
    #[error("Invalid request: {0}")]
    ValidationError(String),
    
    #[error("Database error: {0}")]
    DatabaseError(#[from] sqlx::Error),
    
    #[error("User not found")]
    NotFound,
}
```

## 🧪 TESTS

### 1. **Tests Unitaires**

#### Go
```go
// user_service_test.go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{}
    service := NewUserService(mockRepo)
    
    req := CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "SecurePass123!",
    }
    
    // Act
    user, err := service.CreateUser(context.Background(), req)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, req.Username, user.Username)
    assert.Equal(t, req.Email, user.Email)
}
```

#### Rust
```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[tokio::test]
    async fn test_create_user() {
        // Arrange
        let mock_repo = MockUserRepository::new();
        let service = UserService::new(mock_repo);
        
        let req = CreateUserRequest {
            username: "testuser".to_string(),
            email: "test@example.com".to_string(),
            password: "SecurePass123!".to_string(),
        };
        
        // Act
        let result = service.create_user(req).await;
        
        // Assert
        assert!(result.is_ok());
        let user = result.unwrap();
        assert_eq!(user.username, "testuser");
    }
}
```

### 2. **Tests d'Intégration**

```go
// integration_test.go
func TestUserAPI_Integration(t *testing.T) {
    // Setup
    db := setupTestDatabase(t)
    defer cleanupTestDatabase(t, db)
    
    router := setupTestRouter(t, db)
    
    // Test
    req := CreateUserRequest{
        Username: "integrationuser",
        Email:    "integration@example.com",
        Password: "SecurePass123!",
    }
    
    body, _ := json.Marshal(req)
    resp := httptest.NewRecorder()
    request := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(body))
    
    router.ServeHTTP(resp, request)
    
    assert.Equal(t, http.StatusCreated, resp.Code)
}
```

### 3. **Tests de Performance**

```go
// benchmark_test.go
func BenchmarkUserService_CreateUser(b *testing.B) {
    service := setupBenchmarkService(b)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        req := CreateUserRequest{
            Username: fmt.Sprintf("user%d", i),
            Email:    fmt.Sprintf("user%d@example.com", i),
            Password: "SecurePass123!",
        }
        
        _, err := service.CreateUser(context.Background(), req)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## 🔒 SÉCURITÉ

### 1. **Validation des Entrées**

```go
// validation.go
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,containsany=!@#$%^&*"`
}

func (r *CreateUserRequest) Validate() error {
    validate := validator.New()
    return validate.Struct(r)
}
```

### 2. **Authentification JWT**

```go
// middleware/auth.go
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
            c.Abort()
            return
        }
        
        // Remove "Bearer " prefix
        token = strings.TrimPrefix(token, "Bearer ")
        
        claims, err := utils.VerifyToken(token, jwtSecret)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("role", claims.Role)
        
        c.Next()
    }
}
```

### 3. **Rate Limiting**

```go
// middleware/rate_limiter.go
func RateLimiter(limit int, window time.Duration) gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Every(window/time.Duration(limit)), limit)
    
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## 📊 MONITORING

### 1. **Métriques Prometheus**

```go
// monitoring/metrics.go
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
)

func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        
        httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
        httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
    }
}
```

### 2. **Logging Structuré**

```go
// utils/logger.go
func SetupLogger(level string) *zap.Logger {
    config := zap.NewProductionConfig()
    config.Level = zap.NewAtomicLevelAt(getLogLevel(level))
    
    logger, err := config.Build()
    if err != nil {
        log.Fatal("failed to build logger:", err)
    }
    
    return logger
}

func LogRequest(c *gin.Context, logger *zap.Logger) {
    logger.Info("HTTP request",
        zap.String("method", c.Request.Method),
        zap.String("path", c.Request.URL.Path),
        zap.String("ip", c.ClientIP()),
        zap.String("user_agent", c.Request.UserAgent()),
        zap.Int("status", c.Writer.Status()),
        zap.Duration("duration", time.Since(c.GetTime("start_time"))),
    )
}
```

## 🚀 WORKFLOW PR

### 1. **Création de Branche**

```bash
# Convention de nommage
git checkout -b feat/user-service-tests
git checkout -b refactor/clean-architecture
git checkout -b fix/auth-middleware-bug
git checkout -b docs/api-documentation
```

### 2. **Développement**

```bash
# Commits atomiques (≤150 LOC)
git add .
git commit -m "feat: add user service unit tests

- Add tests for CreateUser method
- Add tests for ValidateUser method
- Add mock repository for testing
- Achieve 95% test coverage

Closes #123"
```

### 3. **Tests Locaux**

```bash
# Go
go test -v ./...
go test -cover ./...
go vet ./...
golangci-lint run

# Rust
cargo test
cargo clippy
cargo fmt --check
```

### 4. **Pull Request**

```markdown
## 🎯 Description
Ajout de tests unitaires complets pour le UserService avec 95% de couverture.

## 🔧 Changements
- [x] Tests unitaires pour CreateUser
- [x] Tests unitaires pour ValidateUser  
- [x] Mock repository pour isolation
- [x] Configuration coverage reporting

## 🧪 Tests
- [x] Tests unitaires passent
- [x] Tests d'intégration passent
- [x] Couverture > 90%
- [x] Linting OK

## 📊 Métriques
- Couverture : 95% (+65%)
- Complexité cyclomatique : 8 (-2)
- Duplication : 2% (-3%)

## 🔗 Issues
Closes #123
Relates to #456

## ✅ Checklist
- [ ] Code review effectuée
- [ ] Tests passent
- [ ] Documentation mise à jour
- [ ] Breaking changes documentées
- [ ] Performance testé
- [ ] Sécurité vérifiée
```

### 5. **Code Review**

#### Critères de Review
- ✅ Code lisible et maintenable
- ✅ Tests complets et pertinents
- ✅ Documentation à jour
- ✅ Performance acceptable
- ✅ Sécurité respectée
- ✅ Conventions respectées

#### Template de Review
```markdown
## 👀 Code Review

### ✅ Points Positifs
- Tests bien structurés
- Documentation claire
- Respect des conventions

### ⚠️ Points d'Amélioration
- [ ] Ajouter test pour cas d'erreur
- [ ] Optimiser requête SQL
- [ ] Ajouter validation supplémentaire

### 🚨 Problèmes Critiques
- [ ] Gestion d'erreur manquante ligne 45
- [ ] Race condition possible ligne 78

### 📝 Suggestions
- Utiliser context.WithTimeout pour la DB
- Ajouter métrique pour les erreurs
- Considérer circuit breaker pattern

## 🎯 Décision
- [ ] ✅ Approuvé
- [ ] ⚠️ Approuvé avec modifications
- [ ] ❌ Rejeté
```

## 📚 DOCUMENTATION

### 1. **Documentation API**

```go
// @Summary Créer un utilisateur
// @Description Crée un nouvel utilisateur avec validation
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "Données utilisateur"
// @Success 201 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
    // Implementation
}
```

### 2. **README de Service**

```markdown
# User Service

## 🎯 Description
Service de gestion des utilisateurs avec authentification et profils.

## 🏗️ Architecture
- Domain Layer : Entités User, UserRepository
- Application Layer : UserService, CreateUserCommand
- Infrastructure Layer : PostgreSQL, Redis
- Interface Layer : HTTP handlers, gRPC

## 🚀 Démarrage
```bash
go run cmd/server/main.go
```

## 🧪 Tests
```bash
go test -v ./internal/domain/services
go test -cover ./...
```

## 📊 Métriques
- Latence moyenne : 15ms
- Throughput : 1000 req/s
- Couverture tests : 95%
```

## 🔄 CI/CD

### 1. **GitHub Actions**

```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.23'
    
    - name: Run tests
      run: |
        go test -v -cover ./...
        go vet ./...
        golangci-lint run
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
    - name: Build Docker image
      run: |
        docker build -t veza-backend-api .
    
    - name: Security scan
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: 'veza-backend-api:latest'
        format: 'sarif'
        output: 'trivy-results.sarif'
```

### 2. **Docker**

```dockerfile
# Dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/config.example.env .env

EXPOSE 8080
CMD ["./main"]
```

## 🎯 BONNES PRATIQUES

### 1. **Performance**
- Utiliser des goroutines pour les opérations I/O
- Implémenter du caching (Redis)
- Optimiser les requêtes SQL
- Utiliser des pools de connexions

### 2. **Sécurité**
- Valider toutes les entrées
- Utiliser HTTPS en production
- Implémenter rate limiting
- Logger les événements de sécurité

### 3. **Maintenabilité**
- Code lisible et documenté
- Tests complets
- Séparation des responsabilités
- Injection de dépendances

### 4. **Observabilité**
- Logging structuré
- Métriques Prometheus
- Distributed tracing
- Health checks

---

*Guidelines créées par le Lead Backend Engineer & Refactor Bot*  
*Prochaine étape : Implémentation de la Phase 1 - Foundation* 