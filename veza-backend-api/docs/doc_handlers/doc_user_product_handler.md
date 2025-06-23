# UserProductHandler Documentation

This file refer to : backend/internal/handlers/user_product.go

## Package & Dependencies
```go
package handlers
imports: gin, database, middleware, models, strconv, strings, time, net/http
```

## Struct
```go
type UserProductHandler struct {
    db *database.DB
}
```

## Constructor
```go
func NewUserProductHandler(db *database.DB) *UserProductHandler
```

## Types/Models
```go
type UserProductResponse struct {
    ID, UserID, ProductID, FilesCount, DocsCount int
    ProductName, CategoryName, Status, CreatedAt, UpdatedAt string
    Brand, Model, Version, SerialNumber, Notes *string
    PurchaseDate, WarrantyExpires *string
    PurchasePrice *int
    IsUnderWarranty bool
}

type CreateUserProductRequest struct {
    ProductID int `binding:"required"`
    Version, SerialNumber, Notes *string
    PurchaseDate, WarrantyExpires *string
    PurchasePrice *int
}

type UpdateUserProductRequest struct {
    Version, SerialNumber, Notes, Status *string
    PurchaseDate, WarrantyExpires *string
    PurchasePrice *int
}
```

## Methods

### Product Collection CRUD
```go
func (h *UserProductHandler) ListUserProducts(c *gin.Context)
func (h *UserProductHandler) GetUserProduct(c *gin.Context)
func (h *UserProductHandler) CreateUserProduct(c *gin.Context)
func (h *UserProductHandler) UpdateUserProduct(c *gin.Context)
func (h *UserProductHandler) DeleteUserProduct(c *gin.Context)
```

### Specialized Features
```go
func (h *UserProductHandler) GetWarrantyStatus(c *gin.Context)
func (h *UserProductHandler) SearchUserProducts(c *gin.Context)
```

### ListUserProducts
- **Auth**: common.GetUserIDFromContext required
- **Params**: page, limit, status
- **Process**: Paginated collection with file/doc counts
- **Joins**: products + categories + file counts

### CreateUserProduct
- **Auth**: Required
- **Validation**: Product exists in catalog, not already owned
- **Process**: Add product to user's collection
- **Default**: status 'active'

### UpdateUserProduct
- **Auth**: Required + ownership validation
- **Process**: Dynamic UPDATE with optional fields
- **Features**: Warranty tracking, purchase details

### DeleteUserProduct
- **Auth**: Required + ownership validation
- **Validation**: No associated files/documents
- **Process**: Hard delete with dependency check

### GetWarrantyStatus
- **Auth**: Required
- **Params**: filter (expiring/expired/active)
- **Process**: Warranty analysis with date calculations
- **Features**: Days remaining calculation

### SearchUserProducts
- **Auth**: Required
- **Params**: q (query), limit
- **Process**: Multi-field search within user's collection
- **Fields**: product name, brand, model, version, serial, notes

## Warranty Logic
```go
// Warranty status calculation
if warrantyExpires != nil {
    warrantyDate, err := time.Parse("2006-01-02", *warrantyExpires)
    if err == nil {
        product.IsUnderWarranty = warrantyDate.After(time.Now())
        daysRemaining = int(time.Until(warrantyDate).Hours() / 24)
    }
}
```

## Warranty Filters
- **expiring**: Within 30 days of expiration
- **expired**: Past warranty expiration date
- **active**: Currently under warranty

## Route Mapping Expectations
- `GET /user-products` → ListUserProducts
- `GET /user-products/:id` → GetUserProduct
- `POST /user-products` → CreateUserProduct
- `PUT /user-products/:id` → UpdateUserProduct
- `DELETE /user-products/:id` → DeleteUserProduct
- `GET /user-products/warranty` → GetWarrantyStatus
- `GET /user-products/search` → SearchUserProducts

## Middleware Dependencies
- Authentication middleware for all endpoints
- JSON binding for POST/PUT requests

## Database Tables
- user_products (main collection data)
- products (product catalog reference)
- categories (product categorization)
- files (file count aggregation)
- internal_documents (document count aggregation)

## Key Relationships
- `user_products.user_id` → `users.id`
- `user_products.product_id` → `products.id`
- `products.category_id` → `categories.id`
- `files.product_id` → `user_products.id`
- `internal_documents.product_id` → `user_products.id`

## Ownership Model
- Users own instances of products (user_products)
- Each user can own one instance per product
- Ownership required for all operations

## Status Values
- `active`: Currently owned and in use
- `sold`: No longer owned
- `broken`: Owned but non-functional
- `archived`: Stored/not in active use

## Computed Features
- **FilesCount**: Associated files for this product instance
- **DocsCount**: Associated internal documents
- **IsUnderWarranty**: Real-time warranty status calculation
- **DaysRemaining**: Warranty time calculation

## Validation Rules
- Product must exist in catalog and be active
- User can only own one instance per product
- Cannot delete if files/documents attached
- All updates require ownership validation

## Error Patterns
- 401: User not authenticated
- 403: Not authorized (ownership)
- 404: Product not found
- 409: Already owned, files/docs prevent deletion
- 500: Database errors

## Helper Methods
```go
func (h *UserProductHandler) getUserProductByID(userProductID, userID int) (*UserProductResponse, error)
```