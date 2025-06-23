import { useState } from 'react';
import { useProductStore } from '../store/productStore';
import { ProductList } from './ProductList';
import { ProductForm } from './ProductForm';
import { ProductFilters } from './ProductFilters';
import { Button } from '@/components/ui/button';
import { Plus } from 'lucide-react';
import { Product } from '../types';

export const ProductContainer = () => {
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState<Product | undefined>();
  const { addProduct, updateProduct } = useProductStore();

  const handleSubmit = (product: Partial<Product>) => {
    if (selectedProduct) {
      updateProduct(selectedProduct.id, product);
    } else {
      addProduct(product as Product);
    }
    setIsFormOpen(false);
    setSelectedProduct(undefined);
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-semibold">Produits</h1>
        <Button onClick={() => setIsFormOpen(true)}>
          <Plus size={16} className="mr-2" />
          Ajouter un produit
        </Button>
      </div>

      <ProductFilters />

      {isFormOpen ? (
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">
            {selectedProduct ? 'Modifier le produit' : 'Ajouter un produit'}
          </h2>
          <ProductForm
            product={selectedProduct}
            onSubmit={handleSubmit}
            onCancel={() => {
              setIsFormOpen(false);
              setSelectedProduct(undefined);
            }}
          />
        </div>
      ) : (
        <ProductList />
      )}
    </div>
  );
}; 