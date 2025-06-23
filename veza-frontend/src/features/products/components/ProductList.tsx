import { useProductStore } from '../store/productStore';
import { Card } from '@/components/ui/card';
import { LoadingSpinner } from '@/shared/components/common/LoadingSpinner';
import { EmptyState } from '@/shared/components/common/EmptyState';
import { Package } from 'lucide-react';

export const ProductList = () => {
  const { products, isLoading } = useProductStore();

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (!products.length) {
    return (
      <EmptyState
        icon={Package}
        title="Aucun produit trouvé"
        description="Commencez par ajouter des produits à votre catalogue"
      />
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {products.map((product) => (
        <Card key={product.id} className="p-4">
          <div className="aspect-square bg-gray-100 rounded-lg mb-4" />
          <h3 className="font-semibold mb-2">{product.name}</h3>
          <p className="text-gray-500 text-sm mb-2">{product.description}</p>
          <div className="flex justify-between items-center">
            <span className="font-medium">{product.price} €</span>
            <span className="text-sm text-gray-500">
              Stock: {product.stock}
            </span>
          </div>
        </Card>
      ))}
    </div>
  );
}; 