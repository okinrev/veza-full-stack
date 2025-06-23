# AuthHandler Documentation

This file refer to : backend/internal/handlers/auth.go

## Package & Dependencies
```go
package handlers
imports: gin, database, middleware, models, utils, net/http, strings
```

## Struct
```go
type AuthHandler struct {
    db        *database.DB
    jwtSecret string
}
```

## Constructor
```go
func NewAuthHandler(db *database.DB, jwtSecret string) *AuthHandler
```

## Types/Models
```go
type SignupRequest struct {
    Username, Email, Password string
    // Validations: username(min=3,max=50), email(required,email), password(min=8)
}

type LoginRequest struct {
    Email, Password string
    // Validations: email(required,email), password(required)
}

type LoginResponse struct {
    AccessToken, RefreshToken string
    User models.User
    ExpiresIn int64
}

type RefreshRequest struct {
    RefreshToken string `binding:"required"`
}

type RefreshResponse struct {
    AccessToken string
    ExpiresIn   int64
}
```

## Methods

### Register
```go
func (h *AuthHandler) Register(c *gin.Context)
```
- **Input**: SignupRequest (JSON binding)
- **Process**: Email/username normalization, password hashing
- **Database**: INSERT into users table with default role 'user'
- **Response**: Success with user_id or conflict error
- **Dependencies**: utils.HashPassword

### Login
```go
func (h *AuthHandler) Login(c *gin.Context)
```
- **Input**: LoginRequest (JSON binding)
- **Process**: User lookup, password verification, token generation
- **Database**: SELECT user + INSERT/UPDATE refresh_tokens
- **Response**: LoginResponse with tokens and user data
- **Dependencies**: utils.CheckPasswordHash, utils.GenerateTokenPair

### RefreshToken
```go
func (h *AuthHandler) RefreshToken(c *gin.Context)
```
- **Input**: RefreshRequest (JSON binding)
- **Process**: Validate refresh token, generate new access token
- **Database**: JOIN refresh_tokens with users
- **Response**: RefreshResponse with new access token
- **Dependencies**: utils.GenerateAccessToken

### Logout
```go
func (h *AuthHandler) Logout(c *gin.Context)
```
- **Input**: RefreshRequest (JSON binding)
- **Process**: Invalidate refresh token
- **Database**: DELETE from refresh_tokens

### GetMe
```go
func (h *AuthHandler) GetMe(c *gin.Context)
```
- **Auth**: common.GetUserIDFromContext required
- **Process**: Fetch current user profile
- **Database**: SELECT user by ID (excluding password)
- **Response**: Full user profile

## Route Mapping Expectations
- `POST /auth/register` → Register
- `POST /auth/login` → Login
- `POST /auth/refresh` → RefreshToken
- `POST /auth/logout` → Logout
- `GET /auth/me` → GetMe

## Middleware Dependencies
- Authentication middleware for GetMe only
- JSON binding middleware for all endpoints

## Database Tables
- users (main user data)
- refresh_tokens (session management)

## Utils Dependencies
```go
utils.HashPassword(password string) (string, error)
utils.CheckPasswordHash(password, hash string) error
utils.GenerateTokenPair(userID int, username, role, secret string) (access, refresh string, expiresIn int64, error)
utils.GenerateAccessToken(userID int, username, role, secret string) (string, int64, error)
```

## Models Dependencies
```go
models.User struct {
    ID, Username, Email, PasswordHash, Role string
    FirstName, LastName, Bio, Avatar *string
    CreatedAt, UpdatedAt time.Time
}
```

## Error Patterns
- 400: Invalid request data, validation errors
- 401: Invalid credentials, expired tokens
- 409: Email/username already exists
- 500: Password hashing, token generation, database errors