# Auth Store - Documentation

Le store d'authentification gère l'état global avec Zustand.

## Interface

```typescript
interface AuthState {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => void;
  register: (data: RegisterData) => Promise<void>;
  checkAuth: () => Promise<void>;
}
```

## Utilisation

```typescript
const { login, logout, user, isAuthenticated } = useAuthStore();

// Connexion
await login({ email: 'user@example.com', password: 'password' });

// Déconnexion
logout();
```

## Backend Go Requis

### Endpoints
- POST /auth/login
- POST /auth/register  
- POST /auth/refresh
- GET /auth/me

### Structure Response

```go
type LoginResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
    User         User   `json:"user"`
}
```

### Middleware JWT

```go
func AuthMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "Authorization required"})
            c.Abort()
            return
        }
        
        if strings.HasPrefix(token, "Bearer ") {
            token = token[7:]
        }
        
        claims, err := validateJWT(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Next()
    })
}
```

