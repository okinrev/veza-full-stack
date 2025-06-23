import { motion, AnimatePresence } from 'framer-motion';
import { ProductCard } from './ProductCard';

interface Product {
  id: number;
  name: string;
  description: string;
  price: number;
  image: string;
  category: string;
  stock: number;
}

interface ProductGridProps {
  products: Product[];
  onEdit?: (id: number) => void;
  onDelete?: (id: number) => void;
}

export function ProductGrid({ products, onEdit, onDelete }: ProductGridProps) {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <AnimatePresence>
        {products.map((product) => (
          <ProductCard
            key={product.id}
            product={product}
            onEdit={onEdit}
            onDelete={onDelete}
          />
        ))}
      </AnimatePresence>
    </div>
  );
} 