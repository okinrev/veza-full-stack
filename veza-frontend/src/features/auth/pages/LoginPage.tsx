import { useState } from 'react';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
// import { Alert } from '@/components/ui/alert';

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, isLoading, error, clearError } = useAuthStore();
  
  const [formData, setFormData] = useState({
    email: '',
    password: ''
  });

  // Récupérer la destination après connexion
  const from = location.state?.from?.pathname || '/dashboard';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    clearError();

    if (!formData.email.trim() || !formData.password.trim()) {
      return;
    }

    try {
      await login(formData.email.trim(), formData.password);
      navigate(from, { replace: true });
    } catch (error) {
      // L'erreur est déjà gérée par le store
      console.error('Erreur de connexion:', error);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    
    // Effacer les erreurs lors de la saisie
    if (error) {
      clearError();
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-6">
      <div className="max-w-md w-full">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">🎵 Veza</h1>
          <h2 className="text-2xl font-semibold text-gray-800 mb-1">Connexion</h2>
          <p className="text-gray-600">Accédez à votre espace personnel</p>
        </div>

        {/* Formulaire */}
        <div className="bg-white rounded-lg shadow-lg p-8">
          {error && (
            <div className="mb-6 bg-red-50 border border-red-200 text-red-800 rounded-lg p-4">
              <div className="flex items-center">
                <span className="text-red-600 mr-2">⚠️</span>
                {error}
              </div>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <Label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-2">
                Adresse email *
              </Label>
              <Input
                id="email"
                name="email"
                type="email"
                value={formData.email}
                onChange={handleInputChange}
                placeholder="votre@email.com"
                required
                disabled={isLoading}
                className="w-full"
              />
            </div>

            <div>
              <Label htmlFor="password" className="block text-sm font-medium text-gray-700 mb-2">
                Mot de passe *
              </Label>
              <Input
                id="password"
                name="password"
                type="password"
                value={formData.password}
                onChange={handleInputChange}
                placeholder="Votre mot de passe"
                required
                disabled={isLoading}
                className="w-full"
              />
            </div>

            <Button
              type="submit"
              disabled={isLoading || !formData.email.trim() || !formData.password.trim()}
              className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-3 rounded-lg transition-colors"
            >
              {isLoading ? (
                <div className="flex items-center justify-center">
                  <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white mr-2"></div>
                  Connexion...
                </div>
              ) : (
                <>
                  <span className="mr-2">🔑</span>
                  Se connecter
                </>
              )}
            </Button>
          </form>

          {/* Liens utiles */}
          <div className="mt-6 text-center space-y-3">
            <div className="text-sm text-gray-600">
              Pas encore de compte ?{' '}
              <Link 
                to="/register" 
                className="text-blue-600 hover:text-blue-800 font-medium hover:underline"
              >
                Créer un compte
              </Link>
            </div>
            
            <div className="text-xs text-gray-500">
              <Link 
                to="/forgot-password" 
                className="hover:text-gray-700 hover:underline"
              >
                Mot de passe oublié ?
              </Link>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="mt-8 text-center text-xs text-gray-500">
          <p>Veza - Plateforme collaborative de création musicale</p>
          <p className="mt-1">
            <Link to="/terms" className="hover:underline">Conditions d'utilisation</Link>
            {' • '}
            <Link to="/privacy" className="hover:underline">Confidentialité</Link>
          </p>
        </div>
      </div>
    </div>
  );
} 