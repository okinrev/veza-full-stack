# ProductHandler Documentation

This file refer to : backend/internal/handlers/product.go

## Package & Dependencies
```go
package handlers
imports: gin, database, middleware, models, strconv, strings, net/http
```

## Struct
```go
type ProductHandler struct {
    db *database.DB
}
```

## Constructor
```go
func NewProductHandler(db *database.DB) *ProductHandler
```

## Types/Models
```go
type ProductResponse struct {
    ID int
    Name, CategoryName, Status string
    CategoryID *int
    Brand, Model, Description *string
    Price, WarrantyMonths *int
    WarrantyConditions, ManufacturerWebsite, Specifications *string
    DocumentationCount, UserCount int  // Computed fields
    CreatedAt, UpdatedAt string
}

type CreateProductRequest struct {
    Name string `binding:"required,min=1,max=200"`
    CategoryID *int
    Brand, Model, Description *string
    Price, WarrantyMonths *int
    WarrantyConditions, ManufacturerWebsite, Specifications *string
    Status string
}

type UpdateProductRequest struct {
    // All fields optional pointers for partial updates
    Name *string
    CategoryID *int
    Brand, Model, Description *string
    Price, WarrantyMonths *int
    WarrantyConditions, ManufacturerWebsite, Specifications *string
    Status *string
}
```

## Methods

### Admin Product Management
```go
func (h *ProductHandler) GetProducts(c *gin.Context)
func (h *ProductHandler) GetProduct(c *gin.Context)
func (h *ProductHandler) CreateProduct(c *gin.Context)
func (h *ProductHandler) UpdateProduct(c *gin.Context)
func (h *ProductHandler) DeleteProduct(c *gin.Context)
```
- **Auth**: Admin required (isAdmin check)
- **Features**: Full CRUD with search/filter capabilities

### Public Access
```go
func (h *ProductHandler) SearchProducts(c *gin.Context)
```
- **Auth**: None (public)
- **Process**: Search active products only

### GetProducts (Admin)
- **Params**: page, limit, search, category, status, brand
- **Process**: Complex query with category JOINs, document/user counts
- **Response**: Paginated products with metadata

### CreateProduct (Admin)
- **Validation**: Required name, default status 'active'
- **Process**: Insert with all product attributes

### UpdateProduct (Admin)
- **Process**: Dynamic UPDATE with optional fields
- **Validation**: Product existence

### DeleteProduct (Admin)
- **Validation**: Check if product is in use (user_products)
- **Process**: Hard delete if not in use

### SearchProducts (Public)
- **Params**: q (query), limit
- **Process**: Search name, brand, model, description
- **Filter**: Only active products, no admin data

## Helper Methods
```go
func (h *ProductHandler) getProductByID(productID int) (*ProductResponse, error)
func (h *ProductHandler) isAdmin(userID int) bool
```

## Route Mapping Expectations
- `GET /admin/products` → GetProducts
- `GET /admin/products/:id` → GetProduct
- `POST /admin/products` → CreateProduct
- `PUT /admin/products/:id` → UpdateProduct
- `DELETE /admin/products/:id` → DeleteProduct
- `GET /products/search` → SearchProducts

## Middleware Dependencies
- Authentication middleware for admin routes
- Admin role validation (handled in methods)

## Database Tables
- products (main product catalog)
- categories (product categorization)
- product_documents (documentation count)
- user_products (usage count, delete validation)

## Key Relationships
- `products.category_id` → `categories.id`
- `user_products.product_id` → `products.id`
- `product_documents.product_id` → `products.id`

## Admin Features
- **Role Check**: isAdmin() validates "admin" OR "super_admin"
- **Full CRUD**: Complete product catalog management
- **Search/Filter**: Multi-field search with category/status filters
- **Usage Validation**: Prevents deletion of products in use

## Computed Fields
- **DocumentationCount**: COUNT from product_documents table
- **UserCount**: COUNT from user_products table (ownership)

## Public vs Admin Access
- **Public**: SearchProducts (active products only, basic info)
- **Admin**: Full access with statistics and management

## Error Patterns
- 401: User not authenticated (admin routes)
- 403: Admin access required
- 404: Product not found
- 409: Product in use (cannot delete)
- 500: Database errors

## Database Relationships
- Products → Categories (optional foreign key)
- Products ← UserProducts (usage tracking)
- Products ← ProductDocuments (documentation count)