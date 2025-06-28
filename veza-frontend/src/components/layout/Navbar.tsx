import { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/features/auth/store/authStore';
import { Button } from '@/components/ui/button';

export function Navbar() {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, isAuthenticated, logout } = useAuthStore();
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const handleLogout = () => {
    logout();
    navigate('/login');
    setIsMenuOpen(false);
  };

  const navItems = [
    {
      name: 'Dashboard',
      path: '/dashboard',
      icon: '🏠',
      requireAuth: true
    },
    {
      name: 'Chat',
      path: '/chat',
      icon: '💬',
      requireAuth: true
    },
    {
      name: 'Produits',
      path: '/products',
      icon: '📦',
      requireAuth: true
    },
    {
      name: 'Ressources',
      path: '/resources',
      icon: '📁',
      requireAuth: true
    },
    {
      name: 'Tracks',
      path: '/tracks',
      icon: '🎵',
      requireAuth: true
    }
  ];

  const isActivePath = (path: string) => {
    return location.pathname === path || location.pathname.startsWith(path + '/');
  };

  if (!isAuthenticated) {
    return (
      <nav className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            {/* Logo */}
            <div className="flex items-center">
              <Link to="/" className="text-2xl font-bold text-blue-600">
                🎵 Veza
              </Link>
            </div>

            {/* Auth Links */}
            <div className="flex items-center space-x-4">
              <Link
                to="/login"
                className="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium transition-colors"
              >
                Connexion
              </Link>
              <Link
                to="/register"
                className="bg-blue-600 text-white hover:bg-blue-700 px-4 py-2 rounded-md text-sm font-medium transition-colors"
              >
                Inscription
              </Link>
            </div>
          </div>
        </div>
      </nav>
    );
  }

  return (
    <nav className="bg-white shadow-sm border-b border-gray-200 sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Logo */}
          <div className="flex items-center">
            <Link to="/dashboard" className="text-2xl font-bold text-blue-600">
              🎵 Veza
            </Link>
          </div>

          {/* Navigation Desktop */}
          <div className="hidden md:flex items-center space-x-4">
            {navItems.map((item) => (
              <Link
                key={item.path}
                to={item.path}
                className={`flex items-center px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                  isActivePath(item.path)
                    ? 'bg-blue-100 text-blue-700'
                    : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
                }`}
              >
                <span className="mr-2">{item.icon}</span>
                {item.name}
              </Link>
            ))}
          </div>

          {/* User Menu Desktop */}
          <div className="hidden md:flex items-center space-x-4">
            {/* Notifications */}
            <Button
              variant="ghost"
              size="sm"
              className="relative"
              title="Notifications"
            >
              🔔
              {/* Badge de notification (exemple) */}
              <span className="absolute -top-1 -right-1 h-4 w-4 bg-red-500 text-white text-xs rounded-full flex items-center justify-center">
                3
              </span>
            </Button>

            {/* Profil utilisateur */}
            <div className="flex items-center space-x-3">
              <div className="text-sm">
                <div className="font-medium text-gray-900">
                  {user?.username || user?.email}
                </div>
                <div className="text-gray-500 text-xs">
                  {user?.role || 'Utilisateur'}
                </div>
              </div>
              
              <div className="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full flex items-center justify-center text-white text-sm font-semibold">
                {(user?.username || user?.email || 'U').charAt(0).toUpperCase()}
              </div>
            </div>

            {/* Bouton de déconnexion */}
            <Button
              onClick={handleLogout}
              variant="outline"
              size="sm"
              className="text-red-600 border-red-200 hover:bg-red-50"
            >
              <span className="mr-1">🚪</span>
              Déconnexion
            </Button>
          </div>

          {/* Menu mobile */}
          <div className="md:hidden">
            <Button
              onClick={() => setIsMenuOpen(!isMenuOpen)}
              variant="ghost"
              size="sm"
              className="text-gray-600"
            >
              {isMenuOpen ? '✕' : '☰'}
            </Button>
          </div>
        </div>

        {/* Navigation Mobile */}
        {isMenuOpen && (
          <div className="md:hidden border-t border-gray-200 bg-white">
            <div className="px-2 pt-2 pb-3 space-y-1">
              {navItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  onClick={() => setIsMenuOpen(false)}
                  className={`flex items-center px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                    isActivePath(item.path)
                      ? 'bg-blue-100 text-blue-700'
                      : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
                  }`}
                >
                  <span className="mr-3">{item.icon}</span>
                  {item.name}
                </Link>
              ))}
              
              {/* Divider */}
              <div className="border-t border-gray-200 my-2"></div>
              
              {/* User Info Mobile */}
              <div className="px-3 py-2">
                <div className="text-sm font-medium text-gray-900">
                  {user?.username || user?.email}
                </div>
                <div className="text-xs text-gray-500">
                  {user?.role || 'Utilisateur'}
                </div>
              </div>
              
              {/* Actions Mobile */}
              <div className="px-3 py-2 space-y-2">
                <Button
                  variant="outline"
                  size="sm"
                  className="w-full justify-start"
                  title="Notifications"
                >
                  <span className="mr-2">🔔</span>
                  Notifications (3)
                </Button>
                
                <Button
                  onClick={handleLogout}
                  variant="outline"
                  size="sm"
                  className="w-full justify-start text-red-600 border-red-200 hover:bg-red-50"
                >
                  <span className="mr-2">🚪</span>
                  Déconnexion
                </Button>
              </div>
            </div>
          </div>
        )}
      </div>
    </nav>
  );
} 