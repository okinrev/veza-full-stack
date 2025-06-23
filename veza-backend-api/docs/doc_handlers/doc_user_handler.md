# UserHandler Documentation

This file refer to : backend/internal/handlers/user.go

## Package & Dependencies
```go
package handlers
imports: gin, database, middleware, models, utils, strconv, strings, net/http, io, os, path/filepath, fmt, time
```

## Struct
```go
type UserHandler struct {
    db *database.DB
}
```

## Constructor
```go
func NewUserHandler(db *database.DB) *UserHandler
```

## Constants
```go
const (
    MaxAvatarSize = 2 << 20  // 2MB
    AvatarWidth = 256
    AvatarHeight = 256
)
```

## Types/Models
```go
type UserResponse struct {
    ID int
    Username, Email, Role, CreatedAt, UpdatedAt string
    FirstName, LastName, Bio, Avatar *string
}

type UpdateUserRequest struct {
    Username, Email, FirstName, LastName, Bio, Avatar *string
}

type ChangePasswordRequest struct {
    CurrentPassword, NewPassword string
    // Validations: current(required), new(required,min=8)
}
```

## Methods

### Profile Management
```go
func (h *UserHandler) UpdateMe(c *gin.Context)
func (h *UserHandler) ChangePassword(c *gin.Context)
```

### User Discovery
```go
func (h *UserHandler) GetUsers(c *gin.Context)
func (h *UserHandler) SearchUsers(c *gin.Context)
func (h *UserHandler) GetUserByID(c *gin.Context)
func (h *UserHandler) GetUsersExceptMe(c *gin.Context)
```

### Avatar Management
```go
func (h *UserHandler) UploadAvatar(c *gin.Context)
func (h *UserHandler) DeleteAvatar(c *gin.Context)
func (h *UserHandler) ServeAvatar(c *gin.Context)
func (h *UserHandler) GetUserAvatar(c *gin.Context)
```

### UpdateMe
- **Auth**: common.GetUserIDFromContext required
- **Process**: Dynamic UPDATE with optional fields
- **Validation**: Username/email uniqueness
- **Features**: Email normalization (lowercase, trim)

### ChangePassword
- **Auth**: Required
- **Process**: Current password verification, new password hashing
- **Dependencies**: utils.CheckPasswordHash, utils.HashPassword
- **Security**: Current password validation required

### GetUsers
- **Auth**: None
- **Params**: page, limit, search
- **Process**: Paginated user list with search across multiple fields
- **Filter**: Excludes deleted users (role != 'deleted')

### SearchUsers
- **Auth**: None
- **Params**: q (query), limit
- **Process**: ILIKE search across username, email, names
- **Response**: User list without pagination

### Avatar System
- **Upload**: Multipart form with image validation
- **Storage**: avatars/ directory with unique naming
- **Serving**: Direct file serving with cache headers
- **Fallback**: Default avatar if file missing
- **Validation**: Image content type, 2MB size limit

### GetUsersExceptMe
- **Auth**: Required
- **Purpose**: Chat user selection
- **Process**: All users except current user
- **Features**: Search capability, limit parameter

## Route Mapping Expectations
- `PUT /users/me` → UpdateMe
- `PUT /users/me/password` → ChangePassword
- `GET /users` → GetUsers
- `GET /users/search` → SearchUsers
- `GET /users/:id` → GetUserByID
- `GET /users/except-me` → GetUsersExceptMe
- `POST /users/me/avatar` → UploadAvatar
- `DELETE /users/me/avatar` → DeleteAvatar
- `GET /avatars/:filename` → ServeAvatar
- `GET /users/:id/avatar` → GetUserAvatar

## Middleware Dependencies
- Authentication middleware for profile operations
- Multipart form parsing for avatar upload

## Database Tables
- users (main user data)

## File System Structure
```
avatars/             # User avatar storage
  default.png        # Fallback avatar
  user_123_456.jpg   # User avatars
```

## Avatar Features
- **Validation**: Image content-type, size limits
- **Naming**: user_{id}_{timestamp}.{ext} format
- **Cleanup**: Old avatar deletion on update
- **Caching**: 24-hour cache headers for performance
- **Fallback**: Default avatar serving

## Search Capabilities
- **Fields**: username, email, first_name, last_name
- **Method**: PostgreSQL ILIKE for case-insensitive search
- **Exclusions**: Deleted users filtered out

## Security Features
- **Password Validation**: Current password required for changes
- **File Validation**: Image content-type verification
- **Path Security**: filepath.Base() prevents traversal
- **Size Limits**: 2MB avatar size restriction

## Error Patterns
- 401: User not authenticated
- 404: User not found
- 409: Username/email conflicts
- 413: Avatar file too large
- 415: Invalid image format
- 500: Database errors, file I/O errors

## Helper Methods
```go
func (h *UserHandler) getUserByID(userID int) (*UserResponse, error)
```

## Utils Dependencies
```go
utils.CheckPasswordHash(password, hash string) error
utils.HashPassword(password string) (string, error)
```