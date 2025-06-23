# Documentation - internal/models/user.go

## Vue d'ensemble

Le package `models` définit les structures de données qui représentent les entités de l'application. Le modèle `User` est central au système d'authentification et de gestion des utilisateurs.

## Structure User

### Définition

```go
type User struct {
    ID           int            `db:"id" json:"id"`
    Username     string         `db:"username" json:"username"`
    Email        string         `db:"email" json:"email"`
    PasswordHash string         `db:"password_hash" json:"-"` // Never serialize password
    FirstName    sql.NullString `db:"first_name" json:"first_name,omitempty"`
    LastName     sql.NullString `db:"last_name" json:"last_name,omitempty"`
    Bio          sql.NullString `db:"bio" json:"bio,omitempty"`
    Avatar       sql.NullString `db:"avatar" json:"avatar,omitempty"`
    Role         string         `db:"role" json:"role"` // user, admin, super_admin
    IsActive     bool           `db:"is_active" json:"is_active"`
    IsVerified   bool           `db:"is_verified" json:"is_verified"`
    LastLoginAt  sql.NullTime   `db:"last_login_at" json:"last_login_at,omitempty"`
    CreatedAt    time.Time      `db:"created_at" json:"created_at"`
    UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
}
```

### Champs Obligatoires

- **ID** : Identifiant unique auto-incrémenté
- **Username** : Nom d'utilisateur unique (3-50 caractères)
- **Email** : Adresse email unique et valide
- **PasswordHash** : Hash du mot de passe (bcrypt recommandé)
- **Role** : Rôle de l'utilisateur (user, admin, super_admin)
- **IsActive** : Statut actif/inactif du compte
- **IsVerified** : Statut de vérification email
- **CreatedAt** : Date de création du compte
- **UpdatedAt** : Date de dernière modification

### Champs Optionnels

- **FirstName** : Prénom (nullable)
- **LastName** : Nom de famille (nullable)
- **Bio** : Biographie/description (nullable)
- **Avatar** : URL ou chemin vers l'avatar (nullable)
- **LastLoginAt** : Date de dernière connexion (nullable)

### Tags de Structure

#### Tags de Base de Données (`db`)
- Mappent les champs avec les colonnes de la table PostgreSQL
- Utilisés par les drivers SQL et ORM

#### Tags JSON (`json`)
- Contrôlent la sérialisation/désérialisation JSON
- `json:"-"` : Exclut le champ de la sérialisation (mot de passe)
- `json:"field,omitempty"` : Omet le champ si vide

## Structure UserResponse

### Définition

```go
type UserResponse struct {
    ID          int            `json:"id"`
    Username    string         `json:"username"`
    Email       string         `json:"email"`
    FirstName   sql.NullString `json:"first_name,omitempty"`
    LastName    sql.NullString `json:"last_name,omitempty"`
    Bio         sql.NullString `json:"bio,omitempty"`
    Avatar      sql.NullString `json:"avatar,omitempty"`
    Role        string         `json:"role"`
    IsActive    bool           `json:"is_active"`
    IsVerified  bool           `json:"is_verified"`
    LastLoginAt sql.NullTime   `json:"last_login_at,omitempty"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
}
```

### Objectif

La structure `UserResponse` représente les données utilisateur **sans informations sensibles** pour les réponses API. Elle exclut notamment :
- `PasswordHash` : Ne doit jamais être exposé
- Autres données sensibles futures

## Méthodes

### `ToResponse()`

**Description** : Convertit un `User` en `UserResponse` en excluant les données sensibles.

**Signature** :
```go
func (u *User) ToResponse() *UserResponse
```

**Utilisation** :
```go
user := &User{
    ID:       1,
    Username: "johndoe",
    Email:    "john@example.com",
    // ... autres champs
}

response := user.ToResponse()
// response ne contient pas PasswordHash
```

**Comportement** :
1. Copie tous les champs non sensibles
2. Exclut automatiquement `PasswordHash`
3. Conserve la structure des champs nullable
4. Ajoute des logs de debug pour le développement

## Structure RefreshToken

### Définition

```go
type RefreshToken struct {
    ID        int       `db:"id" json:"id"`
    UserID    int       `db:"user_id" json:"user_id"`
    Token     string    `db:"token" json:"token"`
    ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
}
```

### Objectif

Gère les tokens de rafraîchissement JWT pour maintenir les sessions utilisateur sans redemander les credentials.

### Champs

- **ID** : Identifiant unique du token
- **UserID** : Référence vers l'utilisateur propriétaire
- **Token** : Token de rafraîchissement cryptographique
- **ExpiresAt** : Date d'expiration du token
- **CreatedAt** : Date de création du token

## Utilisation dans l'Application

### Création d'Utilisateur

```go
func CreateUser(db *database.DB, username, email, password string) (*User, error) {
    // Hash du mot de passe
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    user := &User{
        Username:     username,
        Email:        email,
        PasswordHash: string(hashedPassword),
        Role:         "user",
        IsActive:     true,
        IsVerified:   false,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }

    // Insertion en base
    err = db.QueryRow(`
        INSERT INTO users (username, email, password_hash, role, is_active, is_verified, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `, user.Username, user.Email, user.PasswordHash, user.Role, user.IsActive, user.IsVerified, user.CreatedAt, user.UpdatedAt).Scan(&user.ID)

    if err != nil {
        return nil, err
    }

    return user, nil
}
```

### Récupération d'Utilisateur

```go
func GetUserByEmail(db *database.DB, email string) (*User, error) {
    user := &User{}
    err := db.QueryRow(`
        SELECT id, username, email, password_hash, first_name, last_name, bio, avatar, 
               role, is_active, is_verified, last_login_at, created_at, updated_at
        FROM users 
        WHERE email = $1 AND is_active = true
    `, email).Scan(
        &user.ID, &user.Username, &user.Email, &user.PasswordHash,
        &user.FirstName, &user.LastName, &user.Bio, &user.Avatar,
        &user.Role, &user.IsActive, &user.IsVerified, &user.LastLoginAt,
        &user.CreatedAt, &user.UpdatedAt,
    )

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrUserNotFound
        }
        return nil, err
    }

    return user, nil
}
```

### Réponse API Sécurisée

```go
func (h *UserHandler) GetProfile(c *gin.Context) {
    userID := getUserIDFromContext(c)
    
    user, err := h.service.GetUserByID(userID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }

    // Utilisation de ToResponse() pour exclure les données sensibles
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    user.ToResponse(),
    })
}
```

## Validation

### Validation des Champs

```go
type UserValidation struct {
    Username string `validate:"required,min=3,max=50,alphanum"`
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=8"`
}

func ValidateUser(username, email, password string) error {
    validation := UserValidation{
        Username: username,
        Email:    email,
        Password: password,
    }
    
    validate := validator.New()
    return validate.Struct(validation)
}
```

### Règles de Validation

- **Username** : 3-50 caractères alphanumériques
- **Email** : Format email valide
- **Password** : Minimum 8 caractères
- **Role** : Valeurs autorisées (user, admin, super_admin)

## Sécurité

### Hachage des Mots de Passe

```go
import "golang.org/x/crypto/bcrypt"

// Hachage lors de la création
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// Vérification lors de la connexion
err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
if err != nil {
    return ErrInvalidCredentials
}
```

### Protection des Données

1. **PasswordHash** : Jamais sérialisé en JSON (`json:"-"`)
2. **UserResponse** : Structure séparée pour les réponses API
3. **Logs** : Éviter de logger les mots de passe en clair

## Rôles et Permissions

### Hiérarchie des Rôles

1. **user** : Utilisateur standard
2. **admin** : Administrateur avec privilèges étendus
3. **super_admin** : Super administrateur (tous pouvoirs)

### Vérification des Permissions

```go
func (u *User) HasRole(role string) bool {
    switch role {
    case "user":
        return true // Tous les utilisateurs ont ce rôle
    case "admin":
        return u.Role == "admin" || u.Role == "super_admin"
    case "super_admin":
        return u.Role == "super_admin"
    default:
        return false
    }
}

func (u *User) CanAccessAdmin() bool {
    return u.HasRole("admin")
}
```

## Intégration Frontend

### Structure JSON de Réponse

```json
{
    "success": true,
    "data": {
        "id": 1,
        "username": "johndoe",
        "email": "john@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "bio": "Software developer",
        "avatar": "/uploads/avatars/johndoe.jpg",
        "role": "user",
        "is_active": true,
        "is_verified": true,
        "last_login_at": "2023-11-20T10:30:00Z",
        "created_at": "2023-01-15T08:00:00Z",
        "updated_at": "2023-11-20T10:30:00Z"
    }
}
```

### Utilisation React

```javascript
// Hook pour récupérer le profil utilisateur
const useUserProfile = () => {
    const [user, setUser] = useState(null);
    
    useEffect(() => {
        const fetchProfile = async () => {
            const response = await fetch('/api/v1/users/profile', {
                headers: {
                    'Authorization': `Bearer ${getAuthToken()}`
                }
            });
            
            if (response.ok) {
                const data = await response.json();
                setUser(data.data);
            }
        };
        
        fetchProfile();
    }, []);
    
    return user;
};
```

## Migration Base de Données

### Table users

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    bio TEXT,
    avatar VARCHAR(255),
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin', 'super_admin')),
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes pour les performances
CREATE UNIQUE INDEX idx_users_username ON users(username);
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_active ON users(is_active);
```

### Table refresh_tokens

```sql
CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index pour les performances
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
```

## Tests

### Tests Unitaires

```go
func TestUserToResponse(t *testing.T) {
    user := &User{
        ID:           1,
        Username:     "testuser",
        Email:        "test@example.com",
        PasswordHash: "hashed_password",
        Role:         "user",
        IsActive:     true,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }

    response := user.ToResponse()
    
    assert.Equal(t, user.ID, response.ID)
    assert.Equal(t, user.Username, response.Username)
    assert.Equal(t, user.Email, response.Email)
    // PasswordHash ne doit pas être présent dans UserResponse
}
```

## Bonnes Pratiques

1. **Toujours utiliser `ToResponse()`** pour les réponses API
2. **Valider les données avant sauvegarde**
3. **Hasher les mots de passe avec bcrypt**
4. **Utiliser des transactions pour les opérations critiques**
5. **Logger les opérations sensibles** (connexions, modifications de profil)
6. **Implémenter des limites de débit** pour les opérations de connexion 