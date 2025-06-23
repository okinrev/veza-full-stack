import { useProductStore } from '../store/productStore';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Search, X } from 'lucide-react';

export const ProductFilters = () => {
  const { filters, setFilters } = useProductStore();

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    const { name, value, type } = e.target;
    setFilters({
      ...filters,
      [name]: type === 'number' ? Number(value) : value,
    });
  };

  const handleReset = () => {
    setFilters({});
  };

  return (
    <div className="space-y-4 p-4 bg-white rounded-lg shadow">
      <div className="flex items-center gap-2">
        <Input
          name="search"
          value={filters.search || ''}
          onChange={handleChange}
          placeholder="Rechercher un produit..."
          className="flex-1"
        />
        <Button variant="outline" onClick={handleReset}>
          <X size={16} />
        </Button>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Catégorie
          </label>
          <select
            name="category"
            value={filters.category || ''}
            onChange={handleChange}
            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">Toutes les catégories</option>
            <option value="electronics">Électronique</option>
            <option value="clothing">Vêtements</option>
            <option value="books">Livres</option>
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            En stock
          </label>
          <select
            name="inStock"
            value={filters.inStock ? 'true' : ''}
            onChange={handleChange}
            className="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">Tous</option>
            <option value="true">En stock</option>
            <option value="false">Rupture de stock</option>
          </select>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <Input
          name="minPrice"
          type="number"
          value={filters.minPrice || ''}
          onChange={handleChange}
          placeholder="Prix minimum"
          min={0}
        />
        
        <Input
          name="maxPrice"
          type="number"
          value={filters.maxPrice || ''}
          onChange={handleChange}
          placeholder="Prix maximum"
          min={0}
        />
      </div>
    </div>
  );
}; 