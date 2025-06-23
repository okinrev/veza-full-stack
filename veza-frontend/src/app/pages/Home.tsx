import { Link } from 'react-router-dom';
import { useAuth } from '../providers/AuthProvider';

export default function Home() {
  const { isAuthenticated, user } = useAuth();

  return (
    <div className="min-h-screen bg-background">
      <main className="container mx-auto px-4 py-8">
        <h1 className="text-4xl font-bold mb-8">Bienvenue sur Talas</h1>
        
        {isAuthenticated ? (
          <div className="space-y-4">
            <p className="text-lg">Bonjour, {user?.email}</p>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <Link
                to="/chat"
                className="p-4 bg-card rounded-lg shadow hover:shadow-md transition-shadow"
              >
                <h2 className="text-xl font-semibold mb-2">Chat</h2>
                <p className="text-muted-foreground">Discutez avec la communauté</p>
              </Link>
              
              <Link
                to="/products"
                className="p-4 bg-card rounded-lg shadow hover:shadow-md transition-shadow"
              >
                <h2 className="text-xl font-semibold mb-2">Produits</h2>
                <p className="text-muted-foreground">Découvrez nos produits</p>
              </Link>
              
              <Link
                to="/tracks"
                className="p-4 bg-card rounded-lg shadow hover:shadow-md transition-shadow"
              >
                <h2 className="text-xl font-semibold mb-2">Pistes Audio</h2>
                <p className="text-muted-foreground">Écoutez et partagez de la musique</p>
              </Link>
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            <p className="text-lg">Connectez-vous pour accéder à toutes les fonctionnalités</p>
            <div className="flex gap-4">
              <Link
                to="/auth/login"
                className="px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
              >
                Connexion
              </Link>
              <Link
                to="/auth/register"
                className="px-4 py-2 bg-secondary text-secondary-foreground rounded-md hover:bg-secondary/90"
              >
                Inscription
              </Link>
            </div>
          </div>
        )}
      </main>
    </div>
  );
} 