# Documentation - API Authentication

## Vue d'ensemble

L'API d'authentification gère l'inscription, la connexion, et la gestion des sessions utilisateur via JWT (JSON Web Tokens). Elle fournit les endpoints nécessaires pour sécuriser l'application.

## Base URL

```
/api/v1/auth
```

## Endpoints

### 1. Inscription

**Endpoint** : `POST /api/v1/auth/register`

**Description** : Crée un nouveau compte utilisateur.

**Headers** :
```
Content-Type: application/json
```

**Body** :
```json
{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "motdepasse123",
    "first_name": "John",
    "last_name": "Doe"
}
```

**Champs obligatoires** :
- `username` : Nom d'utilisateur unique (3-50 caractères, alphanumérique)
- `email` : Adresse email valide et unique
- `password` : Mot de passe (minimum 8 caractères)

**Champs optionnels** :
- `first_name` : Prénom
- `last_name` : Nom de famille

**Réponse Succès (201)** :
```json
{
    "success": true,
    "message": "User created successfully",
    "data": {
        "user": {
            "id": 1,
            "username": "johndoe",
            "email": "john@example.com",
            "first_name": "John",
            "last_name": "Doe",
            "role": "user",
            "is_active": true,
            "is_verified": false,
            "created_at": "2023-11-20T10:00:00Z",
            "updated_at": "2023-11-20T10:00:00Z"
        },
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expires_at": "2023-11-21T10:00:00Z"
    }
}
```

**Erreurs possibles** :
- `400 Bad Request` : Données invalides
- `409 Conflict` : Username ou email déjà utilisé

**Exemple d'erreur (400)** :
```json
{
    "success": false,
    "error": "Validation failed",
    "details": {
        "username": "Username must be between 3 and 50 characters",
        "email": "Invalid email format",
        "password": "Password must be at least 8 characters"
    }
}
```

### 2. Connexion

**Endpoint** : `POST /api/v1/auth/login`

**Description** : Authentifie un utilisateur et retourne un token JWT.

**Headers** :
```
Content-Type: application/json
```

**Body** :
```json
{
    "email": "john@example.com",
    "password": "motdepasse123"
}
```

**Champs obligatoires** :
- `email` : Adresse email du compte
- `password` : Mot de passe du compte

**Réponse Succès (200)** :
```json
{
    "success": true,
    "message": "Login successful",
    "data": {
        "user": {
            "id": 1,
            "username": "johndoe",
            "email": "john@example.com",
            "first_name": "John",
            "last_name": "Doe",
            "role": "user",
            "is_active": true,
            "is_verified": true,
            "last_login_at": "2023-11-20T10:30:00Z",
            "created_at": "2023-11-20T10:00:00Z",
            "updated_at": "2023-11-20T10:30:00Z"
        },
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expires_at": "2023-11-21T10:30:00Z"
    }
}
```

**Erreurs possibles** :
- `400 Bad Request` : Données manquantes
- `401 Unauthorized` : Credentials invalides
- `423 Locked` : Compte désactivé

**Exemple d'erreur (401)** :
```json
{
    "success": false,
    "error": "Invalid credentials"
}
```

### 3. Rafraîchissement de Token

**Endpoint** : `POST /api/v1/auth/refresh`

**Description** : Rafraîchit un token JWT expiré.

**Headers** :
```
Content-Type: application/json
```

**Body** :
```json
{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Réponse Succès (200)** :
```json
{
    "success": true,
    "message": "Token refreshed successfully",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "expires_at": "2023-11-21T11:00:00Z"
    }
}
```

**Erreurs possibles** :
- `400 Bad Request` : Token manquant
- `401 Unauthorized` : Token invalide ou expiré

### 4. Déconnexion

**Endpoint** : `POST /api/v1/auth/logout`

**Description** : Déconnecte l'utilisateur et invalide le refresh token.

**Headers** :
```
Content-Type: application/json
Authorization: Bearer <jwt_token>
```

**Body** :
```json
{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Réponse Succès (200)** :
```json
{
    "success": true,
    "message": "Logout successful"
}
```

### 5. Vérification Email

**Endpoint** : `POST /api/v1/auth/verify-email`

**Description** : Vérifie l'adresse email d'un utilisateur.

**Headers** :
```
Content-Type: application/json
```

**Body** :
```json
{
    "token": "verification_token_here"
}
```

**Réponse Succès (200)** :
```json
{
    "success": true,
    "message": "Email verified successfully"
}
```

### 6. Demande de Réinitialisation Mot de Passe

**Endpoint** : `POST /api/v1/auth/forgot-password`

**Description** : Demande un lien de réinitialisation de mot de passe.

**Headers** :
```
Content-Type: application/json
```

**Body** :
```json
{
    "email": "john@example.com"
}
```

**Réponse Succès (200)** :
```json
{
    "success": true,
    "message": "Password reset email sent"
}
```

### 7. Réinitialisation Mot de Passe

**Endpoint** : `POST /api/v1/auth/reset-password`

**Description** : Réinitialise le mot de passe avec un token de réinitialisation.

**Headers** :
```
Content-Type: application/json
```

**Body** :
```json
{
    "token": "reset_token_here",
    "new_password": "nouveaumotdepasse123"
}
```

**Réponse Succès (200)** :
```json
{
    "success": true,
    "message": "Password reset successful"
}
```

## Structure des Tokens JWT

### Access Token

**Durée de vie** : 24 heures (configurable)

**Claims** :
```json
{
    "user_id": 1,
    "username": "johndoe",
    "role": "user",
    "iat": 1699612800,
    "exp": 1699699200
}
```

### Refresh Token

**Durée de vie** : 7 jours (configurable)

**Claims** :
```json
{
    "user_id": 1,
    "type": "refresh",
    "iat": 1699612800,
    "exp": 1700217600
}
```

## Validation des Données

### Règles de Validation

#### Username
- Longueur : 3-50 caractères
- Caractères autorisés : lettres, chiffres, underscore
- Unique dans la base de données

#### Email
- Format email valide
- Longueur maximum : 100 caractères
- Unique dans la base de données

#### Password
- Longueur minimum : 8 caractères
- Recommandé : majuscules, minuscules, chiffres, caractères spéciaux

### Exemples de Validation

```go
type RegisterRequest struct {
    Username  string `json:"username" validate:"required,min=3,max=50,alphanum"`
    Email     string `json:"email" validate:"required,email,max=100"`
    Password  string `json:"password" validate:"required,min=8"`
    FirstName string `json:"first_name" validate:"max=50"`
    LastName  string `json:"last_name" validate:"max=50"`
}
```

## Sécurité

### Hachage des Mots de Passe

```go
import "golang.org/x/crypto/bcrypt"

// Hachage lors de l'inscription
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// Vérification lors de la connexion
err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
```

### Protection contre les Attaques

#### Rate Limiting
- Limite : 5 tentatives de connexion par minute par IP
- Blocage temporaire après 5 échecs consécutifs

#### Validation des Entrées
- Sanitisation des données d'entrée
- Validation stricte des formats
- Protection contre l'injection SQL

#### Sécurité des Tokens
- Signature HMAC-SHA256
- Expiration automatique
- Rotation des refresh tokens

## Intégration Frontend

### Authentification React

```javascript
// Service d'authentification
class AuthService {
    constructor() {
        this.baseURL = '/api/v1/auth';
    }

    async register(userData) {
        const response = await fetch(`${this.baseURL}/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(userData),
        });

        const data = await response.json();
        
        if (data.success) {
            this.storeTokens(data.data.token, data.data.refresh_token);
            return data.data.user;
        } else {
            throw new Error(data.error);
        }
    }

    async login(email, password) {
        const response = await fetch(`${this.baseURL}/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password }),
        });

        const data = await response.json();
        
        if (data.success) {
            this.storeTokens(data.data.token, data.data.refresh_token);
            return data.data.user;
        } else {
            throw new Error(data.error);
        }
    }

    storeTokens(token, refreshToken) {
        localStorage.setItem('authToken', token);
        localStorage.setItem('refreshToken', refreshToken);
    }

    getToken() {
        return localStorage.getItem('authToken');
    }

    async refreshToken() {
        const refreshToken = localStorage.getItem('refreshToken');
        
        const response = await fetch(`${this.baseURL}/refresh`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ refresh_token: refreshToken }),
        });

        const data = await response.json();
        
        if (data.success) {
            this.storeTokens(data.data.token, data.data.refresh_token);
            return data.data.token;
        } else {
            this.logout();
            throw new Error('Token refresh failed');
        }
    }

    logout() {
        localStorage.removeItem('authToken');
        localStorage.removeItem('refreshToken');
    }
}
```

### Hook React pour l'Authentification

```javascript
import { useState, useEffect, useContext, createContext } from 'react';

const AuthContext = createContext();

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);
    const authService = new AuthService();

    useEffect(() => {
        // Vérifier si l'utilisateur est connecté au chargement
        const token = authService.getToken();
        if (token) {
            // Vérifier la validité du token
            fetchUserProfile();
        } else {
            setLoading(false);
        }
    }, []);

    const fetchUserProfile = async () => {
        try {
            const response = await fetch('/api/v1/users/profile', {
                headers: {
                    'Authorization': `Bearer ${authService.getToken()}`,
                },
            });

            if (response.ok) {
                const data = await response.json();
                setUser(data.data);
            } else {
                authService.logout();
            }
        } catch (error) {
            authService.logout();
        } finally {
            setLoading(false);
        }
    };

    const login = async (email, password) => {
        const userData = await authService.login(email, password);
        setUser(userData);
        return userData;
    };

    const register = async (userData) => {
        const user = await authService.register(userData);
        setUser(user);
        return user;
    };

    const logout = () => {
        authService.logout();
        setUser(null);
    };

    const value = {
        user,
        login,
        register,
        logout,
        loading,
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
};
```

## Codes d'Erreur

### Erreurs Communes

| Code | Description | Solution |
|------|-------------|----------|
| 400 | Bad Request | Vérifier les données envoyées |
| 401 | Unauthorized | Vérifier les credentials ou le token |
| 403 | Forbidden | Permissions insuffisantes |
| 409 | Conflict | Username ou email déjà utilisé |
| 422 | Unprocessable Entity | Données invalides |
| 429 | Too Many Requests | Attendre avant de réessayer |
| 500 | Internal Server Error | Erreur serveur |

### Messages d'Erreur Spécifiques

```json
{
    "success": false,
    "error": "Username already exists",
    "code": "USERNAME_EXISTS"
}
```

```json
{
    "success": false,
    "error": "Email already exists",
    "code": "EMAIL_EXISTS"
}
```

```json
{
    "success": false,
    "error": "Invalid credentials",
    "code": "INVALID_CREDENTIALS"
}
```

## Tests

### Tests d'Intégration

```go
func TestAuthEndpoints(t *testing.T) {
    router := setupTestRouter()

    // Test inscription
    t.Run("Register", func(t *testing.T) {
        reqBody := RegisterRequest{
            Username: "testuser",
            Email:    "test@example.com",
            Password: "password123",
        }
        
        body, _ := json.Marshal(reqBody)
        req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusCreated, w.Code)
        
        var response AuthResponse
        json.Unmarshal(w.Body.Bytes(), &response)
        assert.True(t, response.Success)
        assert.NotEmpty(t, response.Data.Token)
    })

    // Test connexion
    t.Run("Login", func(t *testing.T) {
        reqBody := LoginRequest{
            Email:    "test@example.com",
            Password: "password123",
        }
        
        body, _ := json.Marshal(reqBody)
        req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusOK, w.Code)
    })
}
```

## Monitoring

### Métriques à Surveiller

- Nombre d'inscriptions par jour
- Taux de tentatives de connexion échouées
- Temps de réponse des endpoints
- Utilisation des refresh tokens

### Logs de Sécurité

```go
// Log des tentatives de connexion
log.Printf("LOGIN_ATTEMPT: Email=%s, IP=%s, Success=%t", 
    email, clientIP, success)

// Log des inscriptions
log.Printf("REGISTRATION: Username=%s, Email=%s, IP=%s", 
    username, email, clientIP)
```

## Configuration

### Variables d'Environnement

```env
# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRATION=24h
JWT_REFRESH_TIME=168h

# Email Configuration (pour vérification)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-email-password

# Rate Limiting
AUTH_RATE_LIMIT=5
AUTH_RATE_WINDOW=1m
```

Cette documentation couvre tous les aspects de l'API d'authentification nécessaires pour une intégration complète avec le frontend React et les modules Rust. 