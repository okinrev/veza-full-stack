import { BrowserRouter } from 'react-router-dom';
import { Router } from './app/Router';
import { Notifications } from './shared/components/Notifications';
import { DebugPanel } from '@/components/dev/DebugPanel';
import { productService } from '@/features/products/services/productService';

// Données de test
const mockProducts = [
  {
    id: 1,
    name: 'Produit Test 1',
    description: 'Description du produit test 1',
    price: 99.99,
    image: 'https://picsum.photos/400/400?random=1',
    category: 'Électronique',
    stock: 10,
  },
  {
    id: 2,
    name: 'Produit Test 2',
    description: 'Description du produit test 2',
    price: 149.99,
    image: 'https://picsum.photos/400/400?random=2',
    category: 'Vêtements',
    stock: 5,
  },
];

function App() {
  const handleMockData = async () => {
    try {
      // Simuler le chargement des données
      await Promise.all(
        mockProducts.map(product => productService.createProduct(product))
      );
      window.location.reload();
    } catch (error) {
      console.error('Erreur lors du chargement des données de test:', error);
    }
  };

  return (
    <BrowserRouter>
      <Notifications />
      <Router />
      <DebugPanel
        onClearCache={() => window.location.reload()}
        onMockData={handleMockData}
      />
    </BrowserRouter>
  );
}

export default App; 