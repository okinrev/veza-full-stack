---
id: naming-conventions
title: Conventions de nommage
sidebar_label: Conventions de nommage
description: Standards et conventions de nommage pour le projet Veza
---

# Conventions de nommage

Ce document définit les standards de nommage utilisés dans le projet Veza pour maintenir la cohérence et la lisibilité du code.

## Principes généraux

### Clarté et expressivité
- **Noms descriptifs** : Choisir des noms qui expliquent clairement l'intention
- **Éviter les abréviations** : Préférer les noms complets aux abréviations
- **Contexte approprié** : Adapter le niveau de détail au contexte

### Cohérence
- **Suivre les conventions établies** : Respecter les patterns existants
- **Uniformité par langage** : Adapter aux conventions du langage utilisé
- **Évolution contrôlée** : Documenter les changements de conventions

## Conventions par langage

### Go

#### Variables et fonctions
```go
// Variables en camelCase
var userName string
var isActive bool
var maxRetryCount int

// Fonctions en PascalCase (exportées) ou camelCase (privées)
func GetUserByID(id int64) (*User, error) { }
func validateEmail(email string) bool { }

// Constantes en PascalCase
const MaxConnections = 100
const DefaultTimeout = 30 * time.Second
```

#### Types et interfaces
```go
// Types en PascalCase
type UserProfile struct {
    ID        int64     `json:"id"`
    Username  string    `json:"username"`
    CreatedAt time.Time `json:"created_at"`
}

// Interfaces en PascalCase avec suffixe "er" si approprié
type UserRepository interface {
    GetByID(id int64) (*User, error)
    Create(user *User) error
}
```

#### Packages
```go
// Packages en minuscules, un seul mot
package auth
package database
package middleware
```

### Rust

#### Variables et fonctions
```rust
// Variables en snake_case
let user_name: String = "john_doe".to_string();
let is_active: bool = true;
let max_retry_count: u32 = 3;

// Fonctions en snake_case
fn get_user_by_id(id: u64) -> Result<User, Error> { }
fn validate_email(email: &str) -> bool { }
```

#### Types et traits
```rust
// Types en PascalCase
struct UserProfile {
    id: u64,
    username: String,
    created_at: DateTime<Utc>,
}

// Traits en PascalCase
trait UserRepository {
    fn get_by_id(&self, id: u64) -> Result<User, Error>;
    fn create(&self, user: &User) -> Result<(), Error>;
}
```

### JavaScript/TypeScript

#### Variables et fonctions
```typescript
// Variables en camelCase
const userName: string = 'john_doe';
const isActive: boolean = true;
const maxRetryCount: number = 3;

// Fonctions en camelCase
function getUserById(id: number): Promise<User> { }
const validateEmail = (email: string): boolean => { }
```

#### Classes et interfaces
```typescript
// Classes en PascalCase
class UserService {
    private readonly repository: UserRepository;
    
    constructor(repository: UserRepository) {
        this.repository = repository;
    }
}

// Interfaces en PascalCase
interface UserProfile {
    id: number;
    username: string;
    createdAt: Date;
}
```

## Conventions API

### Endpoints REST
```typescript
// Endpoints en kebab-case
GET /api/v1/user-profiles
POST /api/v1/chat-rooms
PUT /api/v1/stream-sessions
DELETE /api/v1/audio-tracks

// Paramètres de requête en snake_case
GET /api/v1/tracks?sort_by=created_at&limit=20
```

### Paramètres JSON
```json
{
  "user_id": 123,
  "created_at": "2024-01-15T10:30:00Z",
  "is_active": true,
  "profile_data": {
    "display_name": "John Doe",
    "avatar_url": "https://example.com/avatar.jpg"
  }
}
```

### Headers HTTP
```http
Authorization: Bearer <token>
Content-Type: application/json
X-Request-ID: abc123-def456
X-User-Agent: veza-mobile/1.0.0
```

## Conventions base de données

### Tables
```sql
-- Tables en snake_case, pluriel
CREATE TABLE user_profiles (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Tables de liaison avec préfixe
CREATE TABLE chat_room_participants (
    room_id BIGINT REFERENCES chat_rooms(id),
    user_id BIGINT REFERENCES users(id),
    joined_at TIMESTAMP DEFAULT NOW()
);
```

### Colonnes
```sql
-- Colonnes en snake_case
SELECT 
    user_id,
    created_at,
    is_active,
    profile_data
FROM user_profiles
WHERE last_login_at > NOW() - INTERVAL '30 days';
```

### Index
```sql
-- Index avec préfixe descriptif
CREATE INDEX idx_user_profiles_username ON user_profiles(username);
CREATE INDEX idx_chat_messages_room_created ON chat_messages(room_id, created_at);
CREATE INDEX idx_tracks_user_public ON tracks(user_id, is_public) WHERE is_public = true;
```

## Conventions fichiers

### Structure des dossiers
```
veza-backend-api/
├── cmd/                    # Points d'entrée
├── internal/              # Code privé
│   ├── api/              # Handlers HTTP
│   ├── domain/           # Logique métier
│   └── infrastructure/   # Adaptateurs externes
├── pkg/                  # Code public réutilisable
└── proto/                # Définitions protobuf
```

### Noms de fichiers
```go
// Fichiers Go en snake_case
user_repository.go
chat_service.go
auth_middleware.go

// Fichiers de test avec suffixe _test
user_repository_test.go
chat_service_test.go
```

## Conventions variables d'environnement

```bash
# Variables en MAJUSCULES avec underscores
DATABASE_URL=postgresql://localhost:5432/veza
REDIS_HOST=localhost
REDIS_PORT=6379
JWT_SECRET=your-secret-key
API_PORT=8080
LOG_LEVEL=info
```

## Conventions documentation

### Fichiers Markdown
```markdown
# Titres en PascalCase
## Sous-sections en PascalCase

### Code blocks avec langage spécifié
```go
func example() {
    // Code exemple
}
```

### Liens internes
```markdown
[Guide d'authentification](./auth-guide.md)
[API Reference](../api/README.md)
```

## Validation et outils

### Linters configurés
- **Go** : `golangci-lint` avec règles personnalisées
- **Rust** : `clippy` avec configuration stricte
- **TypeScript** : `eslint` avec règles de nommage
- **SQL** : `sqlfluff` pour conventions base de données

### Scripts de validation
```bash
# Vérification des conventions
./scripts/check-naming.sh

# Correction automatique
./scripts/fix-naming.sh
```

## Évolution des conventions

### Processus de changement
1. **Proposer** : Créer une issue avec justification
2. **Discuter** : Review par l'équipe technique
3. **Approuver** : Validation par le tech lead
4. **Implémenter** : Mise à jour progressive
5. **Documenter** : Mise à jour de ce guide

### Migration
- **Période de transition** : 2 semaines minimum
- **Compatibilité** : Maintenir l'ancien format pendant la transition
- **Communication** : Notifier l'équipe des changements

## Exemples pratiques

### Bon vs Mauvais

#### Variables
```go
// ✅ Bon
var userProfile *UserProfile
var isAuthenticated bool
var maxRetryAttempts int

// ❌ Mauvais
var up *UserProfile
var auth bool
var mra int
```

#### Fonctions
```go
// ✅ Bon
func GetUserByEmail(email string) (*User, error)
func validateUserInput(input *UserInput) error
func processPaymentTransaction(tx *Transaction) error

// ❌ Mauvais
func GetUser(email string) (*User, error)
func validate(input *UserInput) error
func process(tx *Transaction) error
```

#### Endpoints
```typescript
// ✅ Bon
GET /api/v1/user-profiles/{id}
POST /api/v1/chat-rooms/{roomId}/messages
PUT /api/v1/stream-sessions/{sessionId}/status

// ❌ Mauvais
GET /api/v1/users/{id}
POST /api/v1/chat/{roomId}/msg
PUT /api/v1/stream/{sessionId}/s
```

## Conclusion

Ces conventions de nommage sont essentielles pour maintenir la qualité et la maintenabilité du code. Elles doivent être respectées par tous les développeurs et régulièrement mises à jour selon l'évolution du projet.

Pour toute question ou suggestion d'amélioration, n'hésitez pas à créer une issue ou à contacter l'équipe technique. 