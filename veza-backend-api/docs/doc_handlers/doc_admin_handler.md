# AdminHandler Documentation

This file refer to : backend/internal/handlers/admin.go

## Package & Dependencies
```go
package handlers
imports: gin, database, middleware, models, pq, strconv, strings, context, time
```

## Struct
```go
type AdminHandler struct {
    db *database.DB
}
```

## Constructor
```go
func NewAdminHandler(db *database.DB) *AdminHandler
```

## Types/Models
```go
type DashboardStats struct {
    TotalUsers, ActiveUsers, TotalTracks, PublicTracks int
    TotalSharedResources, TotalListings, ActiveListings int
    TotalOffers, PendingOffers, TotalMessages, TotalRooms int
    LastUpdated string
}

type UserAnalytics struct {
    UserID, TracksCount, ResourcesCount, ListingsCount, MessagesCount int
    Username, Email, Role, RegistrationDate string
    LastActivity *string
    IsActive bool
}

type ContentAnalytics struct {
    TracksByMonth, ResourcesByMonth, UsersByMonth []MonthlyCount
    PopularTags []TagCount
    TopUploaders []UploaderStats
}

type CategoryResponse struct {
    ID, SortOrder, ProductCount int
    Name string
    Description, Icon, Color *string
    IsActive bool
    CreatedAt, UpdatedAt string
}

type CreateCategoryRequest struct {
    Name string `binding:"required,min=2,max=100"`
    Description, Icon, Color *string
    SortOrder int
    IsActive bool
}
```

## Constants
```go
const (
    RoleAdmin, RoleSuperAdmin = "admin", "super_admin"
    ListingStatusOpen, OfferStatusPending = "open", "pending"
    DefaultPage, DefaultLimit, MaxLimit = 1, 20, 100
)
```

## Methods

### Dashboard
```go
func (h *AdminHandler) Dashboard(c *gin.Context)
```
- **Auth**: common.GetUserIDFromContext + isAdmin check
- **Response**: DashboardStats with database counts
- **Queries**: Multiple COUNT queries on users, tracks, listings, offers, messages, rooms
- **Context**: Uses c.Request.Context() for DB calls

### GetUsers
```go
func (h *AdminHandler) GetUsers(c *gin.Context)
```
- **Auth**: Admin required
- **Params**: page, limit, search, role (query params)
- **Response**: Paginated UserAnalytics with meta
- **Query**: Complex JOIN with activity counts, dynamic WHERE clause

### GetAnalytics
```go
func (h *AdminHandler) GetAnalytics(c *gin.Context)
```
- **Auth**: Admin required
- **Response**: ContentAnalytics with 12-month data
- **Queries**: Monthly aggregations, tag analysis, top uploaders

### Category CRUD
```go
func (h *AdminHandler) GetCategories(c *gin.Context)
func (h *AdminHandler) CreateCategory(c *gin.Context)
func (h *AdminHandler) UpdateCategory(c *gin.Context)
func (h *AdminHandler) DeleteCategory(c *gin.Context)
```
- **Auth**: Admin required for all
- **Validation**: Name uniqueness, product usage check for delete
- **Response**: CategoryResponse with product counts

### Helper
```go
func (h *AdminHandler) isAdmin(userID int) bool
```
- **Logic**: Checks role = "admin" OR "super_admin"

## Route Mapping Expectations
- `GET /admin/dashboard` → Dashboard
- `GET /admin/users` → GetUsers
- `GET /admin/analytics` → GetAnalytics
- `GET /admin/categories` → GetCategories
- `POST /admin/categories` → CreateCategory
- `PUT /admin/categories/:id` → UpdateCategory
- `DELETE /admin/categories/:id` → DeleteCategory

## Middleware Dependencies
- Authentication middleware (provides user context)
- Admin role middleware (optional, handled in methods)

## Database Tables
- users (role-based access)
- tracks, shared_resources, listings, offers, messages, rooms (stats)
- categories, products (category management)

## Error Patterns
- 401: User not authenticated
- 403: Admin access required
- 409: Conflicts (name exists, category in use)
- 500: Database errors