---
id: coding-standards
title: Standards de Code
sidebar_label: Standards de Code
description: Standards et conventions de code pour le projet Veza
---

# Standards de Code - Veza Platform

Ce document définit les standards et conventions de code pour maintenir la cohérence et la qualité dans le projet Veza.

## Principes généraux

### 1. Lisibilité
- **Code auto-documenté** : Les noms doivent être explicites
- **Structure claire** : Organisation logique du code
- **Commentaires** : Quand le code ne peut pas être auto-documenté
- **Formatage** : Cohérence dans le style

### 2. Maintenabilité
- **DRY** : Don't Repeat Yourself
- **SOLID** : Principes de conception
- **Modularité** : Séparation des responsabilités
- **Testabilité** : Code facilement testable

### 3. Performance
- **Efficacité** : Optimisation appropriée
- **Ressources** : Gestion mémoire et CPU
- **Scalabilité** : Évolutivité du code
- **Monitoring** : Métriques de performance

## Standards par langage

### Go

#### Structure des fichiers
```go
// Package declaration
package user

// Imports (standard library first, then third-party)
import (
    "context"
    "fmt"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

// Constants
const (
    DefaultTimeout = 30 * time.Second
    MaxRetries     = 3
)

// Types
type User struct {
    ID        uuid.UUID `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    Username  string    `json:"username" db:"username"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Interfaces
type UserRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id uuid.UUID) error
}

// Functions
func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}

// Methods
func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
    if id == uuid.Nil {
        return nil, ErrInvalidUserID
    }
    
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    return user, nil
}
```

#### Conventions de nommage
```go
// Variables et fonctions
var (
    userCount    int    // camelCase pour les variables
    maxRetries   = 3    // camelCase pour les constantes
    defaultPort  = 8080 // camelCase pour les valeurs par défaut
)

// Types et interfaces
type UserService struct{} // PascalCase pour les types
type UserRepository interface{} // PascalCase pour les interfaces

// Fonctions publiques
func GetUserByID(id string) (*User, error) {} // PascalCase pour les exports

// Fonctions privées
func validateEmail(email string) error {} // camelCase pour les fonctions internes

// Constantes
const (
    StatusActive   = "active"   // PascalCase pour les constantes exportées
    StatusInactive = "inactive"
)
```

#### Gestion d'erreurs
```go
// ✅ Bon - Gestion d'erreur appropriée
func (s *UserService) CreateUser(ctx context.Context, user *User) error {
    if user == nil {
        return ErrInvalidUser
    }
    
    if err := s.validateUser(user); err != nil {
        return fmt.Errorf("invalid user data: %w", err)
    }
    
    if err := s.repo.Create(ctx, user); err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    
    return nil
}

// ❌ Mauvais - Pas de gestion d'erreur
func CreateUser(user *User) {
    repo.Create(user)
}
```

#### Tests
```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        user    *User
        wantErr bool
    }{
        {
            name: "valid user",
            user: &User{
                Email:    "test@example.com",
                Username: "testuser",
            },
            wantErr: false,
        },
        {
            name:    "nil user",
            user:    nil,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &MockUserRepository{}
            service := NewUserService(mockRepo)
            
            err := service.CreateUser(context.Background(), tt.user)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Rust

#### Structure des fichiers
```rust
// Module declaration
pub mod user;

// Imports
use std::collections::HashMap;
use tokio::sync::RwLock;
use uuid::Uuid;

// Types
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: Uuid,
    pub email: String,
    pub username: String,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

// Traits
pub trait UserRepository {
    async fn get_by_id(&self, id: Uuid) -> Result<User, Error>;
    async fn create(&self, user: User) -> Result<User, Error>;
    async fn update(&self, user: User) -> Result<User, Error>;
    async fn delete(&self, id: Uuid) -> Result<(), Error>;
}

// Structs
pub struct UserService {
    repo: Box<dyn UserRepository>,
}

// Implementations
impl UserService {
    pub fn new(repo: Box<dyn UserRepository>) -> Self {
        Self { repo }
    }
    
    pub async fn get_user_by_id(&self, id: Uuid) -> Result<User, Error> {
        if id.is_nil() {
            return Err(Error::InvalidUserId);
        }
        
        let user = self.repo.get_by_id(id).await?;
        Ok(user)
    }
}
```

#### Conventions de nommage
```rust
// Variables et fonctions
let user_count: i32 = 0; // snake_case pour les variables
let max_retries = 3; // snake_case pour les constantes

// Types et traits
struct UserService {} // PascalCase pour les types
trait UserRepository {} // PascalCase pour les traits

// Fonctions publiques
pub fn get_user_by_id(id: Uuid) -> Result<User, Error> {} // snake_case pour les exports

// Fonctions privées
fn validate_email(email: &str) -> Result<(), Error> {} // snake_case pour les fonctions internes

// Constantes
pub const STATUS_ACTIVE: &str = "active"; // SCREAMING_SNAKE_CASE pour les constantes
pub const STATUS_INACTIVE: &str = "inactive";
```

#### Gestion d'erreurs
```rust
// ✅ Bon - Gestion d'erreur appropriée
pub async fn create_user(&self, user: User) -> Result<User, Error> {
    if user.email.is_empty() {
        return Err(Error::InvalidEmail);
    }
    
    if let Err(e) = self.validate_user(&user) {
        return Err(Error::ValidationFailed(e.to_string()));
    }
    
    let created_user = self.repo.create(user).await?;
    Ok(created_user)
}

// ❌ Mauvais - Pas de gestion d'erreur
pub async fn create_user(user: User) -> User {
    repo.create(user).await.unwrap()
}
```

#### Tests
```rust
#[cfg(test)]
mod tests {
    use super::*;
    
    #[tokio::test]
    async fn test_create_user_success() {
        let mock_repo = MockUserRepository::new();
        let service = UserService::new(Box::new(mock_repo));
        
        let user = User {
            id: Uuid::new_v4(),
            email: "test@example.com".to_string(),
            username: "testuser".to_string(),
            created_at: Utc::now(),
            updated_at: Utc::now(),
        };
        
        let result = service.create_user(user).await;
        assert!(result.is_ok());
    }
    
    #[tokio::test]
    async fn test_create_user_invalid_email() {
        let mock_repo = MockUserRepository::new();
        let service = UserService::new(Box::new(mock_repo));
        
        let user = User {
            id: Uuid::new_v4(),
            email: "".to_string(), // Email invalide
            username: "testuser".to_string(),
            created_at: Utc::now(),
            updated_at: Utc::now(),
        };
        
        let result = service.create_user(user).await;
        assert!(result.is_err());
    }
}
```

### TypeScript/JavaScript

#### Structure des fichiers
```typescript
// Imports
import { Request, Response } from 'express';
import { User, UserRepository } from '../types';
import { validateUser } from '../utils/validation';

// Types
export interface User {
    id: string;
    email: string;
    username: string;
    createdAt: Date;
    updatedAt: Date;
}

export interface UserRepository {
    getById(id: string): Promise<User | null>;
    create(user: Omit<User, 'id' | 'createdAt' | 'updatedAt'>): Promise<User>;
    update(id: string, user: Partial<User>): Promise<User>;
    delete(id: string): Promise<void>;
}

// Classes
export class UserService {
    constructor(private repo: UserRepository) {}
    
    async getUserById(id: string): Promise<User> {
        if (!id) {
            throw new Error('Invalid user ID');
        }
        
        const user = await this.repo.getById(id);
        if (!user) {
            throw new Error('User not found');
        }
        
        return user;
    }
    
    async createUser(userData: Omit<User, 'id' | 'createdAt' | 'updatedAt'>): Promise<User> {
        const validationError = validateUser(userData);
        if (validationError) {
            throw new Error(`Invalid user data: ${validationError}`);
        }
        
        return await this.repo.create(userData);
    }
}
```

#### Conventions de nommage
```typescript
// Variables et fonctions
const userCount: number = 0; // camelCase pour les variables
const maxRetries = 3; // camelCase pour les constantes

// Types et interfaces
interface UserService {} // PascalCase pour les interfaces
class UserRepository {} // PascalCase pour les classes

// Fonctions publiques
export function getUserById(id: string): Promise<User> {} // camelCase pour les exports

// Fonctions privées
function validateEmail(email: string): boolean {} // camelCase pour les fonctions internes

// Constantes
export const STATUS_ACTIVE = 'active'; // SCREAMING_SNAKE_CASE pour les constantes
export const STATUS_INACTIVE = 'inactive';
```

#### Gestion d'erreurs
```typescript
// ✅ Bon - Gestion d'erreur appropriée
async function createUser(userData: UserData): Promise<User> {
    try {
        if (!userData.email) {
            throw new Error('Email is required');
        }
        
        const validationError = validateUser(userData);
        if (validationError) {
            throw new Error(`Validation failed: ${validationError}`);
        }
        
        const user = await userRepository.create(userData);
        return user;
    } catch (error) {
        logger.error('Failed to create user', { error, userData });
        throw error;
    }
}

// ❌ Mauvais - Pas de gestion d'erreur
async function createUser(userData: UserData): Promise<User> {
    return await userRepository.create(userData);
}
```

#### Tests
```typescript
describe('UserService', () => {
    let userService: UserService;
    let mockRepo: jest.Mocked<UserRepository>;
    
    beforeEach(() => {
        mockRepo = {
            getById: jest.fn(),
            create: jest.fn(),
            update: jest.fn(),
            delete: jest.fn(),
        };
        userService = new UserService(mockRepo);
    });
    
    describe('getUserById', () => {
        it('should return user when valid ID provided', async () => {
            const mockUser: User = {
                id: '123',
                email: 'test@example.com',
                username: 'testuser',
                createdAt: new Date(),
                updatedAt: new Date(),
            };
            
            mockRepo.getById.mockResolvedValue(mockUser);
            
            const result = await userService.getUserById('123');
            
            expect(result).toEqual(mockUser);
            expect(mockRepo.getById).toHaveBeenCalledWith('123');
        });
        
        it('should throw error when user not found', async () => {
            mockRepo.getById.mockResolvedValue(null);
            
            await expect(userService.getUserById('123')).rejects.toThrow('User not found');
        });
    });
});
```

## Standards de documentation

### Commentaires
```go
// Package user provides user management functionality
package user

// UserService handles business logic for user operations
type UserService struct {
    repo UserRepository
}

// GetUserByID retrieves a user by their unique identifier.
// Returns ErrUserNotFound if the user doesn't exist.
func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
    // Implementation...
}
```

### Documentation API
```go
// @Summary Get user by ID
// @Description Retrieve a user by their unique identifier
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} User
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
    // Implementation...
}
```

## Standards de sécurité

### Validation des entrées
```go
// ✅ Bon - Validation appropriée
func (s *UserService) CreateUser(ctx context.Context, user *User) error {
    if user == nil {
        return ErrInvalidUser
    }
    
    if err := s.validateEmail(user.Email); err != nil {
        return fmt.Errorf("invalid email: %w", err)
    }
    
    if err := s.validateUsername(user.Username); err != nil {
        return fmt.Errorf("invalid username: %w", err)
    }
    
    return s.repo.Create(ctx, user)
}

// ❌ Mauvais - Pas de validation
func CreateUser(user *User) error {
    return repo.Create(user)
}
```

### Gestion des secrets
```go
// ✅ Bon - Utilisation d'environnement
type Config struct {
    DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://localhost:5432/veza"`
    JWTSecret   string `env:"JWT_SECRET" envRequired:"true"`
    APIKey      string `env:"API_KEY" envRequired:"true"`
}

// ❌ Mauvais - Secrets en dur
type Config struct {
    DatabaseURL string
    JWTSecret   string
    APIKey      string
}

func init() {
    config.DatabaseURL = "postgres://user:pass@localhost:5432/veza"
    config.JWTSecret = "my-secret-key"
    config.APIKey = "api-key-123"
}
```

## Standards de performance

### Optimisations Go
```go
// ✅ Bon - Pool de connexions
func NewDatabase() (*sql.DB, error) {
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        return nil, err
    }
    
    // Configuration du pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    return db, nil
}

// ✅ Bon - Cache approprié
type UserService struct {
    repo  UserRepository
    cache Cache
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
    // Vérifier le cache d'abord
    if user, found := s.cache.Get(id.String()); found {
        return user.(*User), nil
    }
    
    // Récupérer de la base de données
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Mettre en cache
    s.cache.Set(id.String(), user, 5*time.Minute)
    
    return user, nil
}
```

### Optimisations Rust
```rust
// ✅ Bon - Utilisation de Arc pour le partage
use std::sync::Arc;
use tokio::sync::RwLock;

pub struct UserService {
    repo: Arc<dyn UserRepository>,
    cache: Arc<RwLock<HashMap<Uuid, User>>>,
}

impl UserService {
    pub async fn get_user_by_id(&self, id: Uuid) -> Result<User, Error> {
        // Vérifier le cache d'abord
        {
            let cache = self.cache.read().await;
            if let Some(user) = cache.get(&id) {
                return Ok(user.clone());
            }
        }
        
        // Récupérer de la base de données
        let user = self.repo.get_by_id(id).await?;
        
        // Mettre en cache
        {
            let mut cache = self.cache.write().await;
            cache.insert(id, user.clone());
        }
        
        Ok(user)
    }
}
```

## Outils et linting

### Go
```bash
# Linting
golangci-lint run

# Formatting
go fmt ./...

# Security
gosec ./...

# Tests
go test -v ./...
go test -race ./...
go test -cover ./...
```

### Rust
```bash
# Linting
cargo clippy

# Formatting
cargo fmt

# Security
cargo audit

# Tests
cargo test
cargo test --release
```

### TypeScript
```bash
# Linting
npm run lint

# Formatting
npm run format

# Security
npm audit

# Tests
npm test
npm run test:coverage
```

## Conclusion

Ces standards de code garantissent la cohérence, la maintenabilité et la qualité du projet Veza. Ils doivent être respectés par tous les développeurs et revus régulièrement pour s'adapter aux évolutions du projet.

### Ressources supplémentaires
- [Guide de développement](./development-environment.md)
- [Guide de code review](./code-review.md)
- [Architecture du projet](../architecture/backend-architecture.md) 