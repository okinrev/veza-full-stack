---
id: architecture-patterns
title: Patterns d'architecture
sidebar_label: Patterns d'architecture
description: Patterns et principes d'architecture utilisés dans le projet Veza
---

# Patterns d'architecture

Ce document décrit les patterns d'architecture et les principes de conception utilisés dans le projet Veza pour maintenir la qualité, la maintenabilité et l'évolutivité du code.

## Principes fondamentaux

### 1. Clean Architecture
Le projet Veza suit les principes de Clean Architecture pour séparer les préoccupations et maintenir l'indépendance des couches.

```go
// Structure Clean Architecture
veza-backend-api/
├── internal/
│   ├── domain/           # Entités et règles métier
│   ├── usecases/         # Cas d'usage applicatifs
│   ├── interfaces/       # Ports (interfaces)
│   └── adapters/         # Adaptateurs (implémentations)
├── cmd/                  # Points d'entrée
└── pkg/                  # Code public réutilisable
```

### 2. Domain-Driven Design (DDD)
Organisation du code autour des domaines métier plutôt que des aspects techniques.

```go
// Exemple de domaine User
type User struct {
    ID        int64     `json:"id"`
    Email     string    `json:"email"`
    Username  string    `json:"username"`
    Profile   *Profile  `json:"profile"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type UserRepository interface {
    GetByID(id int64) (*User, error)
    GetByEmail(email string) (*User, error)
    Create(user *User) error
    Update(user *User) error
    Delete(id int64) error
}

type UserService interface {
    RegisterUser(ctx context.Context, input RegisterUserInput) (*User, error)
    AuthenticateUser(ctx context.Context, email, password string) (*User, error)
    UpdateProfile(ctx context.Context, userID int64, profile *Profile) error
}
```

### 3. SOLID Principles
Application des principes SOLID pour la conception des composants.

#### Single Responsibility Principle (SRP)
```go
// ✅ Bon - Chaque service a une responsabilité unique
type UserService struct {
    repo UserRepository
}

type AuthService struct {
    jwtService JWTService
    userService UserService
}

type NotificationService struct {
    emailService EmailService
    pushService PushService
}
```

#### Open/Closed Principle (OCP)
```go
// ✅ Bon - Extensible sans modification
type PaymentProcessor interface {
    ProcessPayment(amount decimal.Decimal, currency string) error
}

type StripeProcessor struct{}
type PayPalProcessor struct{}
type CryptoProcessor struct{}
```

#### Liskov Substitution Principle (LSP)
```go
// ✅ Bon - Les implémentations sont interchangeables
type Storage interface {
    Store(key string, data []byte) error
    Retrieve(key string) ([]byte, error)
    Delete(key string) error
}

type FileStorage struct{}
type S3Storage struct{}
type RedisStorage struct{}
```

#### Interface Segregation Principle (ISP)
```go
// ✅ Bon - Interfaces spécifiques
type UserReader interface {
    GetByID(id int64) (*User, error)
    GetByEmail(email string) (*User, error)
}

type UserWriter interface {
    Create(user *User) error
    Update(user *User) error
    Delete(id int64) error
}

type UserRepository interface {
    UserReader
    UserWriter
}
```

#### Dependency Inversion Principle (DIP)
```go
// ✅ Bon - Dépendance vers les abstractions
type UserService struct {
    repo UserRepository
    auth AuthService
    notif NotificationService
}

func NewUserService(
    repo UserRepository,
    auth AuthService,
    notif NotificationService,
) *UserService {
    return &UserService{
        repo:  repo,
        auth:  auth,
        notif: notif,
    }
}
```

## Patterns de conception

### 1. Repository Pattern
Abstraction de l'accès aux données pour isoler la logique métier.

```go
// Interface du repository
type UserRepository interface {
    GetByID(ctx context.Context, id int64) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, filter UserFilter) ([]*User, error)
}

// Implémentation PostgreSQL
type PostgresUserRepository struct {
    db *sql.DB
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    query := `SELECT id, email, username, created_at, updated_at 
              FROM users WHERE id = $1`
    
    var user User
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &user.ID, &user.Email, &user.Username, &user.CreatedAt, &user.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}
```

### 2. Service Layer Pattern
Couche de logique métier entre les contrôleurs et les repositories.

```go
type UserService struct {
    repo        UserRepository
    authService AuthService
    emailService EmailService
}

func (s *UserService) RegisterUser(ctx context.Context, input RegisterUserInput) (*User, error) {
    // Validation
    if err := input.Validate(); err != nil {
        return nil, err
    }
    
    // Vérification unicité email
    existing, err := s.repo.GetByEmail(ctx, input.Email)
    if err == nil && existing != nil {
        return nil, ErrEmailAlreadyExists
    }
    
    // Création utilisateur
    user := &User{
        Email:    input.Email,
        Username: input.Username,
        Password: input.Password, // Sera hashé
    }
    
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    // Envoi email de confirmation
    go s.emailService.SendWelcomeEmail(user.Email)
    
    return user, nil
}
```

### 3. Factory Pattern
Création d'objets complexes avec configuration.

```go
type DatabaseFactory struct {
    config DatabaseConfig
}

func (f *DatabaseFactory) CreateConnection() (*sql.DB, error) {
    db, err := sql.Open("postgres", f.config.URL)
    if err != nil {
        return nil, err
    }
    
    db.SetMaxOpenConns(f.config.MaxConnections)
    db.SetMaxIdleConns(f.config.MaxIdleConnections)
    db.SetConnMaxLifetime(f.config.ConnMaxLifetime)
    
    return db, nil
}

type ServiceFactory struct {
    db *sql.DB
    redis *redis.Client
}

func (f *ServiceFactory) CreateUserService() *UserService {
    repo := NewPostgresUserRepository(f.db)
    authService := NewAuthService(f.redis)
    emailService := NewEmailService()
    
    return NewUserService(repo, authService, emailService)
}
```

### 4. Observer Pattern
Communication découplée entre composants.

```go
type EventBus interface {
    Publish(event Event) error
    Subscribe(eventType string, handler EventHandler) error
    Unsubscribe(eventType string, handler EventHandler) error
}

type UserEventHandler struct {
    emailService EmailService
    notificationService NotificationService
}

func (h *UserEventHandler) HandleUserRegistered(event UserRegisteredEvent) {
    // Envoi email de bienvenue
    h.emailService.SendWelcomeEmail(event.User.Email)
    
    // Notification push
    h.notificationService.SendWelcomeNotification(event.User.ID)
}
```

### 5. Strategy Pattern
Algorithme interchangeable selon le contexte.

```go
type PaymentStrategy interface {
    ProcessPayment(amount decimal.Decimal, currency string) error
}

type StripeStrategy struct {
    client *stripe.Client
}

func (s *StripeStrategy) ProcessPayment(amount decimal.Decimal, currency string) error {
    // Implémentation Stripe
    return nil
}

type PayPalStrategy struct {
    client *paypal.Client
}

func (p *PayPalStrategy) ProcessPayment(amount decimal.Decimal, currency string) error {
    // Implémentation PayPal
    return nil
}

type PaymentService struct {
    strategies map[string]PaymentStrategy
}

func (ps *PaymentService) ProcessPayment(method string, amount decimal.Decimal, currency string) error {
    strategy, exists := ps.strategies[method]
    if !exists {
        return ErrUnsupportedPaymentMethod
    }
    
    return strategy.ProcessPayment(amount, currency)
}
```

## Patterns d'architecture distribuée

### 1. Microservices Pattern
Séparation des services par domaine métier.

```yaml
# docker-compose.yml
services:
  api:
    build: ./veza-backend-api
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgresql://db:5432/veza
      - REDIS_URL=redis://redis:6379
  
  chat:
    build: ./veza-chat-server
    ports:
      - "8081:8081"
    environment:
      - DATABASE_URL=postgresql://db:5432/veza
      - REDIS_URL=redis://redis:6379
  
  stream:
    build: ./veza-stream-server
    ports:
      - "8082:8082"
    environment:
      - STORAGE_BUCKET=veza-audio
      - CDN_BASE_URL=https://cdn.veza.app
```

### 2. Event-Driven Architecture
Communication asynchrone entre services.

```go
// Event definitions
type UserRegisteredEvent struct {
    UserID    int64     `json:"user_id"`
    Email     string    `json:"email"`
    Username  string    `json:"username"`
    Timestamp time.Time `json:"timestamp"`
}

type TrackUploadedEvent struct {
    TrackID   int64     `json:"track_id"`
    UserID    int64     `json:"user_id"`
    FilePath  string    `json:"file_path"`
    Timestamp time.Time `json:"timestamp"`
}

// Event handlers
type EventHandler interface {
    Handle(event Event) error
}

type UserEventHandler struct {
    emailService EmailService
    analyticsService AnalyticsService
}

func (h *UserEventHandler) Handle(event Event) error {
    switch e := event.(type) {
    case *UserRegisteredEvent:
        return h.handleUserRegistered(e)
    case *TrackUploadedEvent:
        return h.handleTrackUploaded(e)
    default:
        return ErrUnsupportedEvent
    }
}
```

### 3. CQRS Pattern
Séparation des commandes et des requêtes.

```go
// Commands
type CreateUserCommand struct {
    Email    string `json:"email" validate:"required,email"`
    Username string `json:"username" validate:"required,min=3"`
    Password string `json:"password" validate:"required,min=8"`
}

type UpdateUserCommand struct {
    UserID   int64   `json:"user_id" validate:"required"`
    Username *string `json:"username,omitempty"`
    Profile  *Profile `json:"profile,omitempty"`
}

// Queries
type GetUserQuery struct {
    UserID int64 `json:"user_id" validate:"required"`
}

type ListUsersQuery struct {
    Page     int    `json:"page" validate:"min=1"`
    PageSize int    `json:"page_size" validate:"min=1,max=100"`
    Filter   string `json:"filter,omitempty"`
    SortBy   string `json:"sort_by,omitempty"`
    SortDir  string `json:"sort_dir,omitempty"`
}

// Handlers
type CommandHandler interface {
    Handle(ctx context.Context, command Command) error
}

type QueryHandler interface {
    Handle(ctx context.Context, query Query) (interface{}, error)
}
```

## Patterns de résilience

### 1. Circuit Breaker Pattern
Protection contre les défaillances en cascade.

```go
type CircuitBreaker struct {
    state     State
    failures  int
    threshold int
    timeout   time.Duration
    lastFailure time.Time
    mu        sync.RWMutex
}

func (cb *CircuitBreaker) Execute(operation func() error) error {
    if !cb.canExecute() {
        return ErrCircuitBreakerOpen
    }
    
    err := operation()
    cb.recordResult(err)
    return err
}

func (cb *CircuitBreaker) canExecute() bool {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    
    switch cb.state {
    case StateClosed:
        return true
    case StateOpen:
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = StateHalfOpen
            return true
        }
        return false
    case StateHalfOpen:
        return true
    default:
        return false
    }
}
```

### 2. Retry Pattern
Nouvel essai automatique en cas d'échec.

```go
type RetryConfig struct {
    MaxAttempts int
    Backoff     BackoffStrategy
    ShouldRetry func(error) bool
}

type ExponentialBackoff struct {
    initialDelay time.Duration
    maxDelay     time.Duration
    multiplier   float64
}

func (eb *ExponentialBackoff) NextDelay(attempt int) time.Duration {
    delay := time.Duration(float64(eb.initialDelay) * math.Pow(eb.multiplier, float64(attempt)))
    if delay > eb.maxDelay {
        return eb.maxDelay
    }
    return delay
}

func Retry(operation func() error, config RetryConfig) error {
    var lastErr error
    
    for attempt := 0; attempt < config.MaxAttempts; attempt++ {
        if err := operation(); err == nil {
            return nil
        } else {
            lastErr = err
            if !config.ShouldRetry(err) {
                return err
            }
        }
        
        if attempt < config.MaxAttempts-1 {
            delay := config.Backoff.NextDelay(attempt)
            time.Sleep(delay)
        }
    }
    
    return lastErr
}
```

### 3. Bulkhead Pattern
Isolation des ressources pour éviter les défaillances en cascade.

```go
type Bulkhead struct {
    maxConcurrency int
    semaphore      chan struct{}
}

func NewBulkhead(maxConcurrency int) *Bulkhead {
    return &Bulkhead{
        maxConcurrency: maxConcurrency,
        semaphore:      make(chan struct{}, maxConcurrency),
    }
}

func (b *Bulkhead) Execute(operation func() error) error {
    select {
    case b.semaphore <- struct{}{}:
        defer func() { <-b.semaphore }()
        return operation()
    default:
        return ErrBulkheadFull
    }
}
```

## Patterns de performance

### 1. Caching Pattern
Mise en cache des données fréquemment accédées.

```go
type Cache interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}, ttl time.Duration) error
    Delete(key string) error
    Clear() error
}

type CachedUserRepository struct {
    repo  UserRepository
    cache Cache
}

func (r *CachedUserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", id)
    
    // Essayer le cache d'abord
    if cached, err := r.cache.Get(cacheKey); err == nil {
        if user, ok := cached.(*User); ok {
            return user, nil
        }
    }
    
    // Cache miss, aller en base
    user, err := r.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Mettre en cache
    r.cache.Set(cacheKey, user, 5*time.Minute)
    
    return user, nil
}
```

### 2. Connection Pooling
Gestion efficace des connexions de base de données.

```go
type ConnectionPool struct {
    pool *sql.DB
    config PoolConfig
}

type PoolConfig struct {
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}

func NewConnectionPool(dsn string, config PoolConfig) (*ConnectionPool, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    
    db.SetMaxOpenConns(config.MaxOpenConns)
    db.SetMaxIdleConns(config.MaxIdleConns)
    db.SetConnMaxLifetime(config.ConnMaxLifetime)
    db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
    
    return &ConnectionPool{pool: db, config: config}, nil
}
```

### 3. Lazy Loading
Chargement à la demande des données.

```go
type LazyUser struct {
    id       int64
    repo     UserRepository
    user     *User
    loaded   bool
    mu       sync.RWMutex
}

func (lu *LazyUser) GetUser() (*User, error) {
    lu.mu.RLock()
    if lu.loaded {
        user := lu.user
        lu.mu.RUnlock()
        return user, nil
    }
    lu.mu.RUnlock()
    
    lu.mu.Lock()
    defer lu.mu.Unlock()
    
    // Double-check locking
    if lu.loaded {
        return lu.user, nil
    }
    
    user, err := lu.repo.GetByID(context.Background(), lu.id)
    if err != nil {
        return nil, err
    }
    
    lu.user = user
    lu.loaded = true
    
    return user, nil
}
```

## Patterns de sécurité

### 1. Authentication Middleware
Vérification d'authentification centralisée.

```go
type AuthMiddleware struct {
    jwtService JWTService
    userService UserService
}

func (am *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := extractToken(r)
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        
        claims, err := am.jwtService.ValidateToken(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        user, err := am.userService.GetByID(r.Context(), claims.UserID)
        if err != nil {
            http.Error(w, "User not found", http.StatusUnauthorized)
            return
        }
        
        ctx := context.WithValue(r.Context(), "user", user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 2. Rate Limiting
Protection contre les abus.

```go
type RateLimiter struct {
    store  RateLimitStore
    config RateLimitConfig
}

type RateLimitConfig struct {
    MaxRequests int
    Window      time.Duration
}

func (rl *RateLimiter) Allow(key string) (bool, error) {
    now := time.Now()
    windowStart := now.Add(-rl.config.Window)
    
    count, err := rl.store.GetCount(key, windowStart)
    if err != nil {
        return false, err
    }
    
    if count >= rl.config.MaxRequests {
        return false, nil
    }
    
    err = rl.store.Increment(key, now)
    if err != nil {
        return false, err
    }
    
    return true, nil
}
```

## Conclusion

Ces patterns d'architecture permettent de construire un système robuste, maintenable et évolutif. Ils doivent être appliqués de manière cohérente à travers le projet pour maximiser leurs bénéfices.

### Ressources supplémentaires
- [Guide de développement](./development-environment.md)
- [Standards de code](./code-review.md)
- [Architecture backend](../architecture/backend-architecture.md) 