import { useEffect } from 'react';
import type { ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';

interface AuthGuardProps {
  children: ReactNode;
  requireAuth?: boolean;
}

export function AuthGuard({ children, requireAuth = true }: AuthGuardProps) {
  const location = useLocation();
  const { isAuthenticated, isLoading, checkExistingAuth, user } = useAuthStore();

  useEffect(() => {
    // Vérifier l'authentification au chargement
    checkExistingAuth();
  }, [checkExistingAuth]);

  // Attendre la vérification de l'authentification
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Vérification de l'authentification...</p>
        </div>
      </div>
    );
  }

  // Si l'authentification est requise mais l'utilisateur n'est pas connecté
  if (requireAuth && !isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // Si l'authentification n'est pas requise mais l'utilisateur est connecté
  // (pour les pages de login/register)
  if (!requireAuth && isAuthenticated) {
    const from = location.state?.from?.pathname || '/dashboard';
    return <Navigate to={from} replace />;
  }

  return <>{children}</>;
} 