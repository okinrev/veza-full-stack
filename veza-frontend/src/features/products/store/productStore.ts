import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import { Product, ProductFilters } from '../types';

interface ProductState {
  products: Product[];
  isLoading: boolean;
  error: string | null;
  filters: ProductFilters;
  
  // Actions
  setProducts: (products: Product[]) => void;
  setLoading: (isLoading: boolean) => void;
  setError: (error: string | null) => void;
  setFilters: (filters: ProductFilters) => void;
  addProduct: (product: Product) => void;
  updateProduct: (id: number, product: Partial<Product>) => void;
  deleteProduct: (id: number) => void;
}

export const useProductStore = create<ProductState>()(
  devtools(
    immer((set) => ({
      // Initial state
      products: [],
      isLoading: false,
      error: null,
      filters: {},
      
      // Actions
      setProducts: (products) =>
        set((state) => {
          state.products = products;
        }),
        
      setLoading: (isLoading) =>
        set((state) => {
          state.isLoading = isLoading;
        }),
        
      setError: (error) =>
        set((state) => {
          state.error = error;
        }),
        
      setFilters: (filters) =>
        set((state) => {
          state.filters = filters;
        }),
        
      addProduct: (product) =>
        set((state) => {
          state.products.push(product);
        }),
        
      updateProduct: (id, product) =>
        set((state) => {
          const index = state.products.findIndex((p) => p.id === id);
          if (index !== -1) {
            state.products[index] = { ...state.products[index], ...product };
          }
        }),
        
      deleteProduct: (id) =>
        set((state) => {
          state.products = state.products.filter((p) => p.id !== id);
        }),
    }))
  )
); 