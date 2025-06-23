import React, { useEffect } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/shared/stores/authStore';
import { Loader2 } from 'lucide-react';

interface AuthGuardProps {
  children: React.ReactNode;
  requireAuth?: boolean;
}

export const AuthGuard: React.FC<AuthGuardProps> = ({ 
  children, 
  requireAuth = true 
}) => {
  const { isAuthenticated, isLoading, checkAuth } = useAuthStore();
  const location = useLocation();
  
  useEffect(() => {
    checkAuth();
  }, [checkAuth]);
  
  // Loading state
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="flex flex-col items-center space-y-4">
          <Loader2 className="h-8 w-8 animate-spin text-blue-600" />
          <p className="text-gray-600">Vérification de la session...</p>
        </div>
      </div>
    );
  }
  
  // Redirection si auth requise mais pas connecté
  if (requireAuth && !isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  
  // Redirection si déjà connecté sur pages publiques
  if (!requireAuth && isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }
  
  return <>{children}</>;
}; 