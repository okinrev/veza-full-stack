import { create } from 'zustand';
import { createContext, useContext, ReactNode } from 'react';

interface AuthState {
  isAuthenticated: boolean;
  user: any | null;
  login: (credentials: { email: string; password: string }) => Promise<void>;
  logout: () => void;
  register: (userData: { email: string; password: string; name: string }) => Promise<void>;
}

const useAuthStore = create<AuthState>((set) => ({
  isAuthenticated: false,
  user: null,
  login: async (credentials) => {
    try {
      // TODO: Implémenter l'appel API
      set({ isAuthenticated: true, user: { email: credentials.email } });
    } catch (error) {
      console.error('Erreur de connexion:', error);
      throw error;
    }
  },
  logout: () => {
    set({ isAuthenticated: false, user: null });
  },
  register: async (userData) => {
    try {
      // TODO: Implémenter l'appel API
      set({ isAuthenticated: true, user: { email: userData.email } });
    } catch (error) {
      console.error('Erreur d\'inscription:', error);
      throw error;
    }
  },
}));

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const auth = useAuthStore();
  return <AuthContext.Provider value={auth}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth doit être utilisé à l\'intérieur d\'un AuthProvider');
  }
  return context;
} 