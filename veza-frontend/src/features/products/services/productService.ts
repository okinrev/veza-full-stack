import { apiClient } from '@/shared/api/client';
import type { Product, ProductFilters, PaginatedResponse } from '@/shared/api/types';

export interface ProductsResponse {
  products: Product[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface CreateProductData {
  name: string;
  description?: string;
  price: number;
  currency?: string;
  category?: string;
  status?: string;
}

export interface UpdateProductData extends Partial<CreateProductData> {
  id: number;
}

class ProductService {
  private static instance: ProductService;
  
  private constructor() {}

  public static getInstance(): ProductService {
    if (!ProductService.instance) {
      ProductService.instance = new ProductService();
    }
    return ProductService.instance;
  }

  // Récupérer tous les produits avec filtres
  async getProducts(filters?: ProductFilters): Promise<ProductsResponse> {
    const params = new URLSearchParams();
    
    if (filters?.page) params.append('page', filters.page.toString());
    if (filters?.limit) params.append('limit', filters.limit.toString());
    if (filters?.search) params.append('search', filters.search);
    if (filters?.category) params.append('category', filters.category);
    if (filters?.price_min) params.append('price_min', filters.price_min.toString());
    if (filters?.price_max) params.append('price_max', filters.price_max.toString());
    if (filters?.status) params.append('status', filters.status);

    const queryString = params.toString();
    const url = `/api/v1/products${queryString ? `?${queryString}` : ''}`;
    
    return apiClient.get<ProductsResponse>(url);
  }

  // Récupérer un produit spécifique
  async getProduct(id: number): Promise<Product> {
    return apiClient.get<Product>(`/api/v1/products/${id}`);
  }

  // Créer un nouveau produit
  async createProduct(data: CreateProductData): Promise<Product> {
    const productData = {
      name: data.name,
      description: data.description,
      price: data.price,
      currency: data.currency || 'EUR',
      category: data.category,
      status: data.status || 'available'
    };

    return apiClient.post<Product>('/api/v1/products', productData);
  }

  // Mettre à jour un produit
  async updateProduct(data: UpdateProductData): Promise<Product> {
    const { id, ...updateData } = data;
    return apiClient.put<Product>(`/api/v1/products/${id}`, updateData);
  }

  // Supprimer un produit
  async deleteProduct(id: number): Promise<void> {
    return apiClient.delete(`/api/v1/products/${id}`);
  }

  // Rechercher des produits
  async searchProducts(query: string, filters?: Omit<ProductFilters, 'search'>): Promise<ProductsResponse> {
    const params = new URLSearchParams();
    params.append('search', query);
    
    if (filters?.page) params.append('page', filters.page.toString());
    if (filters?.limit) params.append('limit', filters.limit.toString());
    if (filters?.category) params.append('category', filters.category);
    if (filters?.price_min) params.append('price_min', filters.price_min.toString());
    if (filters?.price_max) params.append('price_max', filters.price_max.toString());
    if (filters?.status) params.append('status', filters.status);

    return apiClient.get<ProductsResponse>(`/api/v1/products?${params.toString()}`);
  }

  // Obtenir les catégories disponibles
  async getCategories(): Promise<string[]> {
    return apiClient.get<string[]>('/api/v1/products/categories');
  }

  // Upload d'images pour un produit
  async uploadProductImage(productId: number, file: File): Promise<{ url: string }> {
    return apiClient.uploadFile<{ url: string }>(`/api/v1/products/${productId}/images`, file);
  }

  // Supprimer une image de produit
  async deleteProductImage(productId: number, imageUrl: string): Promise<void> {
    return apiClient.delete(`/api/v1/products/${productId}/images`, {
      data: { image_url: imageUrl }
    });
  }

  // Marquer un produit comme vendu
  async markAsSold(id: number): Promise<Product> {
    return this.updateProduct({ id, status: 'sold' });
  }

  // Marquer un produit comme disponible
  async markAsAvailable(id: number): Promise<Product> {
    return this.updateProduct({ id, status: 'available' });
  }

  // Obtenir les produits de l'utilisateur connecté
  async getMyProducts(filters?: Omit<ProductFilters, 'seller_id'>): Promise<ProductsResponse> {
    const params = new URLSearchParams();
    
    if (filters?.page) params.append('page', filters.page.toString());
    if (filters?.limit) params.append('limit', filters.limit.toString());
    if (filters?.search) params.append('search', filters.search);
    if (filters?.category) params.append('category', filters.category);
    if (filters?.status) params.append('status', filters.status);

    const queryString = params.toString();
    const url = `/api/v1/products/my${queryString ? `?${queryString}` : ''}`;
    
    return apiClient.get<ProductsResponse>(url);
  }

  // Obtenir les statistiques des produits
  async getProductStats(): Promise<{
    total: number;
    available: number;
    sold: number;
    categories: Array<{ name: string; count: number }>;
  }> {
    return apiClient.get('/api/v1/products/stats');
  }

  // Recherche globale incluant les produits
  async globalSearch(query: string): Promise<{
    products: Product[];
    total: number;
  }> {
    const response = await apiClient.get<{
      results: {
        products?: Product[];
      };
    }>(`/api/v1/search/global?q=${encodeURIComponent(query)}&type=products`);
    return {
      products: response.results.products || [],
      total: response.results.products?.length || 0
    };
  }
}

// Export de l'instance singleton
export const productService = ProductService.getInstance();
export default productService; 