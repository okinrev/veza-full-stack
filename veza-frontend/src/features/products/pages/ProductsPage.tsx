import { useState, useEffect, Suspense, lazy } from 'react';
import { motion } from 'framer-motion';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { useToast } from '@/components/ui/use-toast';
import { ProductGridSkeleton } from '../components/ProductGridSkeleton';
import { productService, type CreateProductData } from '../services/productService';
import type { Product } from '../services/productService';

// Lazy load components
const ProductGrid = lazy(() => import('../components/ProductGrid').then(mod => ({ default: mod.ProductGrid })));
const ProductForm = lazy(() => import('../components/ProductForm').then(mod => ({ default: mod.ProductForm })));

export function ProductsPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<string[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const { toast } = useToast();

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setIsLoading(true);
      const [productsData, categoriesData] = await Promise.all([
        productService.getProducts(),
        productService.getCategories(),
      ]);
      setProducts(productsData);
      setCategories(categoriesData);
    } catch (error) {
      toast({
        title: 'Erreur',
        description: 'Impossible de charger les données',
        variant: 'destructive',
      });
    } finally {
      setIsLoading(false);
    }
  };

  const filteredProducts = products.filter((product) => {
    const matchesSearch = product.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      product.description.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCategory = !selectedCategory || product.category === selectedCategory;
    return matchesSearch && matchesCategory;
  });

  const handleCreateProduct = async (data: CreateProductData) => {
    try {
      const newProduct = await productService.createProduct(data);
      setProducts((prev) => [...prev, newProduct]);
      setIsFormOpen(false);
      toast({
        title: 'Succès',
        description: 'Produit créé avec succès',
      });
    } catch (error) {
      toast({
        title: 'Erreur',
        description: 'Impossible de créer le produit',
        variant: 'destructive',
      });
    }
  };

  const handleEditProduct = async (data: Product) => {
    try {
      const updatedProduct = await productService.updateProduct(data);
      setProducts((prev) =>
        prev.map((p) => (p.id === updatedProduct.id ? updatedProduct : p))
      );
      setIsFormOpen(false);
      setEditingProduct(null);
      toast({
        title: 'Succès',
        description: 'Produit mis à jour avec succès',
      });
    } catch (error) {
      toast({
        title: 'Erreur',
        description: 'Impossible de mettre à jour le produit',
        variant: 'destructive',
      });
    }
  };

  const handleDeleteProduct = async (id: number) => {
    try {
      await productService.deleteProduct(id);
      setProducts((prev) => prev.filter((p) => p.id !== id));
      toast({
        title: 'Succès',
        description: 'Produit supprimé avec succès',
      });
    } catch (error) {
      toast({
        title: 'Erreur',
        description: 'Impossible de supprimer le produit',
        variant: 'destructive',
      });
    }
  };

  return (
    <div className="container mx-auto py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Produits</h1>
        <Button onClick={() => {
          setEditingProduct(null);
          setIsFormOpen(true);
        }}>
          Ajouter un produit
        </Button>
      </div>

      <div className="flex gap-4 mb-8">
        <Input
          placeholder="Rechercher un produit..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="max-w-sm"
        />
        <Select
          value={selectedCategory}
          onValueChange={setSelectedCategory}
        >
          <SelectTrigger className="w-[200px]">
            <SelectValue placeholder="Catégorie" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">Toutes les catégories</SelectItem>
            {categories.map((category) => (
              <SelectItem key={category} value={category}>
                {category}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {isLoading ? (
        <ProductGridSkeleton />
      ) : (
        <Suspense fallback={<ProductGridSkeleton />}>
          <ProductGrid
            products={filteredProducts}
            onEdit={(id: number) => {
              const product = products.find((p) => p.id === id);
              if (product) {
                setEditingProduct(product);
                setIsFormOpen(true);
              }
            }}
            onDelete={handleDeleteProduct}
          />
        </Suspense>
      )}

      <Dialog open={isFormOpen} onOpenChange={setIsFormOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editingProduct ? 'Modifier le produit' : 'Ajouter un produit'}
            </DialogTitle>
          </DialogHeader>
          <Suspense fallback={<div>Chargement...</div>}>
            <ProductForm
              product={editingProduct}
              onSubmit={editingProduct ? handleEditProduct : handleCreateProduct}
              onCancel={() => {
                setIsFormOpen(false);
                setEditingProduct(null);
              }}
            />
          </Suspense>
        </DialogContent>
      </Dialog>
    </div>
  );
} 