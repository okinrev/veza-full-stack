# Documentation - internal/middleware/auth.go

## Vue d'ensemble

Le package `middleware` contient les middlewares d'authentification et d'autorisation pour sécuriser les endpoints de l'API. Il utilise JWT (JSON Web Tokens) pour l'authentification et gère les rôles utilisateur.

## Middlewares d'Authentification

### `JWTAuthMiddleware()`

**Description** : Middleware principal qui valide les tokens JWT et établit le contexte utilisateur.

**Signature** :
```go
func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc
```

**Paramètres** :
- `jwtSecret` : Clé secrète pour valider les tokens JWT

**Comportement** :
1. Extrait le token du header `Authorization`
2. Vérifie le format `Bearer <token>`
3. Valide le token JWT avec la clé secrète
4. Extrait les claims (userID, username, role)
5. Injecte les informations dans le contexte Gin
6. Permet la continuation de la requête ou retourne une erreur 401

**Code d'exemple** :
```go
func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   "Authorization header required",
            })
            c.Abort()
            return
        }

        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   "Authorization header must start with 'Bearer '",
            })
            c.Abort()
            return
        }

        tokenString := authHeader[7:]
        claims, err := utils.ValidateJWT(tokenString, jwtSecret)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   "Invalid token: " + err.Error(),
            })
            c.Abort()
            return
        }

        // Injection dans le contexte
        common.SetUserIDInContext(c, claims.UserID)
        common.SetUsernameInContext(c, claims.Username)
        common.SetUserRoleInContext(c, claims.Role)
        c.Next()
    }
}
```

**Utilisation** :
```go
// Protection d'un groupe de routes
protected := router.Group("/api/v1/protected")
protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
{
    protected.GET("/profile", handlers.GetProfile)
    protected.PUT("/profile", handlers.UpdateProfile)
}
```

### `OptionalJWTAuthMiddleware()`

**Description** : Middleware qui valide les tokens JWT s'ils sont présents, mais n'exige pas d'authentification.

**Signature** :
```go
func OptionalJWTAuthMiddleware(jwtSecret string) gin.HandlerFunc
```

**Utilisation** :
- Endpoints publics qui peuvent bénéficier d'informations utilisateur
- APIs qui retournent des données différentes selon l'authentification
- Endpoints de recherche avec personnalisation

**Code d'exemple** :
```go
func OptionalJWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
            tokenString := authHeader[7:]
            if claims, err := utils.ValidateJWT(tokenString, jwtSecret); err == nil {
                common.SetUserIDInContext(c, claims.UserID)
                common.SetUsernameInContext(c, claims.Username)
                common.SetUserRoleInContext(c, claims.Role)
            }
        }
        c.Next()
    }
}
```

**Exemple d'utilisation** :
```go
// Endpoint public avec personnalisation optionnelle
router.GET("/api/v1/public/listings", 
    middleware.OptionalJWTAuthMiddleware(jwtSecret),
    handlers.GetListings) // Peut retourner des données personnalisées si authentifié
```

## Middlewares d'Autorisation

### `RequireRole()`

**Description** : Middleware qui vérifie si l'utilisateur authentifié possède un rôle spécifique.

**Signature** :
```go
func RequireRole(roles ...string) gin.HandlerFunc
```

**Paramètres** :
- `roles` : Liste des rôles autorisés (variadic)

**Comportement** :
1. Récupère le rôle utilisateur du contexte
2. Vérifie si le rôle correspond à un des rôles autorisés
3. Retourne 403 Forbidden si le rôle n'est pas autorisé
4. Permet la continuation si le rôle est valide

**Code d'exemple** :
```go
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("user_role")
        if !exists {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   "User role not found in context",
            })
            c.Abort()
            return
        }

        role, ok := userRole.(string)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{
                "success": false,
                "error":   "Invalid user role",
            })
            c.Abort()
            return
        }

        hasRole := false
        for _, requiredRole := range roles {
            if role == requiredRole {
                hasRole = true
                break
            }
        }

        if !hasRole {
            c.JSON(http.StatusForbidden, gin.H{
                "success": false,
                "error":   "Insufficient permissions",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

**Utilisation** :
```go
// Endpoint nécessitant le rôle admin ou super_admin
router.DELETE("/api/v1/admin/users/:id",
    middleware.JWTAuthMiddleware(jwtSecret),
    middleware.RequireRole("admin", "super_admin"),
    handlers.DeleteUser)
```

### `AdminMiddleware()`

**Description** : Middleware de convenance qui vérifie si l'utilisateur a des privilèges administrateur.

**Signature** :
```go
func AdminMiddleware() gin.HandlerFunc
```

**Équivalent à** : `RequireRole("admin", "super_admin")`

**Utilisation** :
```go
// Groupe de routes administrateur
admin := router.Group("/api/v1/admin")
admin.Use(middleware.JWTAuthMiddleware(jwtSecret))
admin.Use(middleware.AdminMiddleware())
{
    admin.GET("/dashboard", handlers.GetAdminDashboard)
    admin.GET("/users", handlers.GetAllUsers)
    admin.POST("/users/:id/ban", handlers.BanUser)
}
```

## Gestion du Contexte

### Injection des Données Utilisateur

Le middleware utilise le package `common` pour injecter les données utilisateur dans le contexte Gin :

```go
// Injection dans le contexte
common.SetUserIDInContext(c, claims.UserID)
common.SetUsernameInContext(c, claims.Username)
common.SetUserRoleInContext(c, claims.Role)
```

### Récupération des Données

Dans les handlers, récupérez les données utilisateur :

```go
func GetProfile(c *gin.Context) {
    userID := common.GetUserIDFromContext(c)
    username := common.GetUsernameFromContext(c)
    role := common.GetUserRoleFromContext(c)
    
    // Utilisation des données...
}
```

## Structure des Claims JWT

### Claims Personnalisés

```go
type Claims struct {
    UserID   int    `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}
```

### Validation des Claims

```go
func ValidateJWT(tokenString, jwtSecret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(jwtSecret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}
```

## Gestion des Erreurs

### Types d'Erreurs

1. **401 Unauthorized** :
   - Header Authorization manquant
   - Format Bearer incorrect
   - Token invalide ou expiré
   - Rôle manquant dans le contexte

2. **403 Forbidden** :
   - Rôle insuffisant pour l'opération
   - Permissions non accordées

### Réponses d'Erreur

```json
{
    "success": false,
    "error": "Authorization header required"
}
```

```json
{
    "success": false,
    "error": "Insufficient permissions"
}
```

## Sécurité

### Bonnes Pratiques

1. **Secrets JWT** :
   ```go
   // Utiliser des secrets forts en production
   jwtSecret := os.Getenv("JWT_SECRET")
   if jwtSecret == "" {
       log.Fatal("JWT_SECRET must be set")
   }
   ```

2. **Expiration des Tokens** :
   ```go
   claims := &Claims{
       UserID:   userID,
       Username: username,
       Role:     role,
       RegisteredClaims: jwt.RegisteredClaims{
           ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
           IssuedAt:  jwt.NewNumericDate(time.Now()),
       },
   }
   ```

3. **Validation Stricte** :
   - Vérifier l'algorithme de signature
   - Valider les claims obligatoires
   - Vérifier l'expiration

### Protection CSRF

```go
// Vérification de l'origine pour les requêtes sensibles
func CSRFProtection() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method != "GET" {
            origin := c.GetHeader("Origin")
            if origin == "" {
                c.JSON(http.StatusForbidden, gin.H{
                    "error": "Origin header required",
                })
                c.Abort()
                return
            }
            // Vérifier l'origine autorisée...
        }
        c.Next()
    }
}
```

## Patterns d'Utilisation

### Middleware Chain

```go
// Chaîne de middlewares pour différents niveaux de sécurité
public := router.Group("/api/v1/public")
{
    public.GET("/health", handlers.HealthCheck)
}

authenticated := router.Group("/api/v1/auth")
authenticated.Use(middleware.JWTAuthMiddleware(jwtSecret))
{
    authenticated.GET("/profile", handlers.GetProfile)
    authenticated.PUT("/profile", handlers.UpdateProfile)
}

admin := router.Group("/api/v1/admin")
admin.Use(middleware.JWTAuthMiddleware(jwtSecret))
admin.Use(middleware.AdminMiddleware())
{
    admin.GET("/dashboard", handlers.AdminDashboard)
    admin.GET("/users", handlers.GetAllUsers)
}
```

### Middleware Conditionnel

```go
func ConditionalAuth(requireAuth bool, jwtSecret string) gin.HandlerFunc {
    if requireAuth {
        return middleware.JWTAuthMiddleware(jwtSecret)
    }
    return middleware.OptionalJWTAuthMiddleware(jwtSecret)
}
```

## Intégration Frontend

### Envoi du Token

```javascript
// Stockage du token (localStorage ou sessionStorage)
localStorage.setItem('authToken', token);

// Envoi avec les requêtes
const apiCall = async (url, options = {}) => {
    const token = localStorage.getItem('authToken');
    
    const response = await fetch(url, {
        ...options,
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
            ...options.headers,
        },
    });
    
    if (response.status === 401) {
        // Token expiré, rediriger vers login
        window.location.href = '/login';
    }
    
    return response;
};
```

### Gestion des Erreurs Frontend

```javascript
const handleAuthError = (error) => {
    if (error.status === 401) {
        // Token invalide ou expiré
        localStorage.removeItem('authToken');
        window.location.href = '/login';
    } else if (error.status === 403) {
        // Permissions insuffisantes
        alert('Vous n\'avez pas les permissions nécessaires');
    }
};
```

## Tests

### Tests Unitaires

```go
func TestJWTAuthMiddleware(t *testing.T) {
    // Setup
    jwtSecret := "test-secret"
    token := generateTestToken(jwtSecret)
    
    // Test avec token valide
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/test", nil)
    c.Request.Header.Set("Authorization", "Bearer "+token)
    
    middleware := JWTAuthMiddleware(jwtSecret)
    middleware(c)
    
    assert.Equal(t, http.StatusOK, w.Code)
    
    // Test sans token
    w = httptest.NewRecorder()
    c, _ = gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/test", nil)
    
    middleware(c)
    
    assert.Equal(t, http.StatusUnauthorized, w.Code)
}
```

### Tests d'Intégration

```go
func TestProtectedEndpoint(t *testing.T) {
    router := setupTestRouter()
    
    // Test avec authentification
    token := generateTestToken("test-secret")
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/api/v1/protected/profile", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

## Monitoring et Logging

### Logging des Authentifications

```go
func LoggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        userID := common.GetUserIDFromContext(c)
        username := common.GetUsernameFromContext(c)
        
        log.Printf("AUTH: User %d (%s) accessed %s %s - Status: %d - Duration: %v",
            userID, username, c.Request.Method, c.Request.URL.Path, 
            c.Writer.Status(), time.Since(start))
    }
}
```

### Métriques de Sécurité

```go
type SecurityMetrics struct {
    LoginAttempts   int64
    FailedLogins    int64
    TokenValidations int64
    AccessDenied    int64
}

func (m *SecurityMetrics) RecordFailedLogin() {
    atomic.AddInt64(&m.FailedLogins, 1)
}
```

## Extensions Possibles

1. **Rate Limiting** : Limiter les tentatives d'authentification
2. **Session Management** : Gestion des sessions utilisateur
3. **Multi-factor Authentication** : Support 2FA
4. **Refresh Tokens** : Tokens de rafraîchissement automatique
5. **Audit Logging** : Journalisation des actions sensibles 