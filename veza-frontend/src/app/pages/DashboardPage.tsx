import React from 'react';
import { Link } from 'react-router-dom';
import { useAuthStore } from '@/shared/stores/authStore';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { 
  MessageCircle, 
  Package, 
  Music, 
  FileText, 
  User, 
  LogOut 
} from 'lucide-react';

// Fonction utilitaire pour extraire les valeurs sql.NullString de Go
const extractNullString = (value: any): string => {
  if (value && typeof value === 'object' && 'String' in value) {
    return value.Valid ? value.String : '';
  }
  return value || '';
};

export function DashboardPage() {
  const { user, logout } = useAuthStore();

  const handleLogout = () => {
    logout();
  };

  // Extraction sécurisée des valeurs utilisateur
  const safeFirstName = extractNullString(user?.first_name) || user?.username || 'Utilisateur';
  const safeLastName = extractNullString(user?.last_name);
  const safeEmail = user?.email || '';
  const displayName = safeFirstName + (safeLastName ? ` ${safeLastName}` : '');

  const modules = [
    {
      title: 'Chat',
      description: 'Communiquez avec les autres utilisateurs',
      icon: MessageCircle,
      path: '/chat',
      color: 'bg-blue-500'
    },
    {
      title: 'Produits',
      description: 'Gérez vos produits et services',
      icon: Package,
      path: '/products',
      color: 'bg-green-500'
    },
    {
      title: 'Pistes',
      description: 'Explorez les pistes musicales',
      icon: Music,
      path: '/tracks',
      color: 'bg-purple-500'
    },
    {
      title: 'Ressources',
      description: 'Accédez aux ressources partagées',
      icon: FileText,
      path: '/resources',
      color: 'bg-orange-500'
    }
  ];

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">
            Bienvenue, {safeFirstName} !
          </h1>
          <p className="text-gray-600 mt-2">
            Choisissez un module pour commencer
          </p>
        </div>
        <div className="flex gap-4">
          <Link to="/profile">
            <Button variant="outline" className="flex items-center gap-2">
              <User className="w-4 h-4" />
              Profil
            </Button>
          </Link>
          <Button 
            variant="destructive" 
            onClick={handleLogout}
            className="flex items-center gap-2"
          >
            <LogOut className="w-4 h-4" />
            Déconnexion
          </Button>
        </div>
      </div>

      {/* Modules Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {modules.map((module) => (
          <Link key={module.path} to={module.path}>
            <Card className="hover:shadow-lg transition-shadow duration-200 cursor-pointer">
              <CardHeader className="pb-3">
                <div className={`w-12 h-12 rounded-lg ${module.color} flex items-center justify-center mb-3`}>
                  <module.icon className="w-6 h-6 text-white" />
                </div>
                <CardTitle className="text-lg">{module.title}</CardTitle>
                <CardDescription>{module.description}</CardDescription>
              </CardHeader>
              <CardContent>
                <Button variant="ghost" className="w-full justify-start">
                  Accéder →
                </Button>
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>

      {/* User Info */}
      {user && (
        <div className="mt-12 bg-gray-50 rounded-lg p-6">
          <h2 className="text-xl font-semibold mb-4">Informations du compte</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <p className="text-sm text-gray-600">Email</p>
              <p className="font-medium">{safeEmail}</p>
            </div>
            <div>
              <p className="text-sm text-gray-600">Nom complet</p>
              <p className="font-medium">
                {displayName || safeFirstName}
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// Export par défaut pour la compatibilité avec le lazy loading
export default DashboardPage; 