import { useEffect } from 'react';
import type { ReactNode } from 'react';
import { Navbar } from '@/components/layout/Navbar';
import { useAuthStore } from '@/features/auth/store/authStore';

interface LayoutProps {
  children: ReactNode;
  showNavbar?: boolean;
}

export function Layout({ children, showNavbar = true }: LayoutProps) {
  const { checkExistingAuth } = useAuthStore();

  useEffect(() => {
    // Vérifier l'authentification au chargement
    checkExistingAuth();
  }, [checkExistingAuth]);

  return (
    <div className="min-h-screen bg-gray-50">
      {showNavbar && <Navbar />}
      <main className="flex-1">
        {children}
      </main>
    </div>
  );
} 