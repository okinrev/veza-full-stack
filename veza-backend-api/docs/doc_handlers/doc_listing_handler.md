# ListingHandler Documentation

This file refer to : backend/internal/handlers/listing.go

## Package & Dependencies
```go
package handlers
imports: gin, database, middleware, pq, strconv, strings, net/http
```

## Struct
```go
type ListingHandler struct {
    db *database.DB
}
```

## Constructor
```go
func NewListingHandler(db *database.DB) *ListingHandler
```

## Types/Models
```go
type ListingResponse struct {
    ID, UserID, ProductID int
    Username, ProductName, Description, State, Status string
    Price *int
    ExchangeFor *string
    Images []string
    CreatedAt, UpdatedAt string
}

type CreateListingRequest struct {
    ProductID int `binding:"required"`  // user_products.id
    Description, State string `binding:"required"`
    Price *int
    ExchangeFor *string
    Images []string
}

type UpdateListingRequest struct {
    Description, State, Status *string
    Price *int
    ExchangeFor *string
    Images *[]string
}
```

## Methods

### Listing CRUD
```go
func (h *ListingHandler) CreateListing(c *gin.Context)
func (h *ListingHandler) GetAllListings(c *gin.Context)
func (h *ListingHandler) GetListingByID(c *gin.Context)
func (h *ListingHandler) UpdateListing(c *gin.Context)
func (h *ListingHandler) DeleteListing(c *gin.Context)
```

### CreateListing
- **Auth**: common.GetUserIDFromContext required
- **Validation**: Product ownership via user_products table
- **Process**: Creates listing with status 'open'
- **Database**: INSERT into listings with pq.Array for images

### GetAllListings
- **Auth**: None (public)
- **Params**: page, limit, status (default 'open')
- **Process**: Paginated listings with user/product JOINs
- **Response**: Listings with metadata

### GetListingByID
- **Auth**: None (public)
- **Process**: Single listing with full details
- **Database**: Complex JOIN (listings + users + user_products + products)

### UpdateListing
- **Auth**: Required + ownership validation
- **Process**: Dynamic UPDATE with multiple optional fields
- **Validation**: User must own the listing

### DeleteListing
- **Auth**: Required + ownership validation
- **Process**: Hard delete from database
- **Note**: Cascade handling for related offers

## Helper Methods
```go
func (h *ListingHandler) getListingByID(listingID int) (*ListingResponse, error)
```
- **Process**: Reusable method for listing retrieval with JOINs

## Route Mapping Expectations
- `POST /listings` → CreateListing
- `GET /listings` → GetAllListings
- `GET /listings/:id` → GetListingByID
- `PUT /listings/:id` → UpdateListing
- `DELETE /listings/:id` → DeleteListing

## Middleware Dependencies
- Authentication middleware for create/update/delete
- JSON binding for POST/PUT requests

## Database Tables
- listings (main listing data)
- users (username for display)
- user_products (ownership validation)
- products (product details)

## Key Relationships
- `listings.user_id` → `users.id`
- `listings.product_id` → `user_products.id` (NOT products.id)
- `user_products.product_id` → `products.id`

## Array Handling
- Uses `pq.Array` for images field
- Images stored as PostgreSQL text array

## Ownership Model
- Users create listings for their owned products (user_products)
- Validation ensures user owns the user_product before listing

## Error Patterns
- 401: User not authenticated
- 403: Not authorized (ownership)
- 404: Listing/product not found
- 409: Conflicts (if implemented)
- 500: Database errors

## Status Flow
- Created: 'open'
- Can be updated to: 'closed', 'sold', etc.
- Public listings show only specified status (default 'open')