//file: backend/services/product_service.go

package services

type ProductService interface {
    CreateProduct(req CreateProductRequest) (*models.Product, error)
    GetProducts(page, limit int, filters map[string]interface{}) ([]models.Product, int, error)
    UpdateProduct(id int, req UpdateProductRequest) (*models.Product, error)
    DeleteProduct(id int) error
    SearchProducts(query string, limit int) ([]models.Product, error)
}