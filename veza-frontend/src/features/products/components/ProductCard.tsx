import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { formatPrice } from '@/shared/lib/utils';
import { getOptimizedImageUrl, getImagePlaceholder, preloadImage } from '@/lib/imageOptimization';
import { cn } from '@/shared/lib/utils';

interface ProductCardProps {
  product: {
    id: number;
    name: string;
    description: string;
    price: number;
    image: string;
    category: string;
    stock: number;
  };
  onEdit?: (id: number) => void;
  onDelete?: (id: number) => void;
}

export function ProductCard({ product, onEdit, onDelete }: ProductCardProps) {
  const [imageLoaded, setImageLoaded] = useState(false);
  const [imageError, setImageError] = useState(false);

  const optimizedImageUrl = getOptimizedImageUrl(product.image, {
    width: 400,
    height: 400,
    quality: 80,
    format: 'webp',
  });

  useEffect(() => {
    preloadImage(optimizedImageUrl)
      .then(() => setImageLoaded(true))
      .catch(() => setImageError(true));
  }, [optimizedImageUrl]);

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -20 }}
      whileHover={{ scale: 1.02 }}
      transition={{ duration: 0.2 }}
    >
      <Card className="overflow-hidden">
        <div className="relative aspect-square">
          {!imageLoaded && !imageError && (
            <img
              src={getImagePlaceholder(400, 400)}
              alt=""
              className="absolute inset-0 w-full h-full object-cover"
            />
          )}
          <img
            src={imageError ? getImagePlaceholder(400, 400) : optimizedImageUrl}
            alt={product.name}
            className={cn(
              "object-cover w-full h-full transition-opacity duration-300",
              imageLoaded ? "opacity-100" : "opacity-0"
            )}
            onLoad={() => setImageLoaded(true)}
            onError={() => setImageError(true)}
          />
          <Badge 
            variant={product.stock > 0 ? "default" : "destructive"}
            className="absolute top-2 right-2"
          >
            {product.stock > 0 ? `${product.stock} en stock` : 'Rupture de stock'}
          </Badge>
        </div>
        <CardHeader>
          <CardTitle className="line-clamp-1">{product.name}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground line-clamp-2">
            {product.description}
          </p>
          <div className="mt-2 flex items-center justify-between">
            <span className="text-lg font-semibold">
              {formatPrice(product.price)}
            </span>
            <Badge variant="outline">{product.category}</Badge>
          </div>
        </CardContent>
        <CardFooter className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            className="flex-1"
            onClick={() => onEdit?.(product.id)}
          >
            Modifier
          </Button>
          <Button
            variant="destructive"
            size="sm"
            className="flex-1"
            onClick={() => onDelete?.(product.id)}
          >
            Supprimer
          </Button>
        </CardFooter>
      </Card>
    </motion.div>
  );
} 