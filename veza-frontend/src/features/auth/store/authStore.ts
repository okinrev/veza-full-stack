import { create } from 'zustand';
import { persist } from 'zustand/middleware';

// Types
interface User {
  id: number;
  email: string;
  username: string;
  first_name?: string;
  last_name?: string;
  role?: string;
  is_active?: boolean;
  is_verified?: boolean;
  created_at: string;
  updated_at?: string;
  last_login_at?: string;
}

interface AuthResponse {
  success: boolean;
  message?: string;
  data?: {
    access_token: string;
    token?: string; // Compatibilité
    refresh_token: string;
    expires_in: number;
    user: User;
  };
  // Compatibilité avec l'ancien format
  access_token?: string;
  refresh_token?: string;
  user?: User;
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isLoading: boolean;
  error: string | null;
  isAuthenticated: boolean;
  
  // Actions
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, username: string, password: string, firstName?: string, lastName?: string) => Promise<void>;
  logout: () => void;
  refreshAuth: () => Promise<boolean>;
  checkExistingAuth: () => Promise<boolean>;
  clearError: () => void;
  setTokens: (accessToken: string, refreshToken?: string) => void;
}

// Configuration API
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

// Utilitaires JWT
const isTokenValid = (token: string): boolean => {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    return !payload.exp || payload.exp > Date.now() / 1000;
  } catch {
    return false;
  }
};

const getUserFromToken = (token: string): User | null => {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    return {
      id: payload.user_id || payload.id || payload.sub,
      email: payload.email,
      username: payload.username || payload.email.split('@')[0],
      first_name: payload.first_name,
      last_name: payload.last_name,
      role: payload.role || 'user',
      created_at: payload.iat ? new Date(payload.iat * 1000).toISOString() : new Date().toISOString()
    };
  } catch {
    return null;
  }
};

// Store d'authentification unifié
export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isLoading: false,
      error: null,
      isAuthenticated: false,

      // Connexion - Compatible avec l'ancien endpoint
      login: async (email: string, password: string) => {
        set({ isLoading: true, error: null });

        try {
          const response = await fetch(`${API_BASE_URL}/auth/login`, {
            method: 'POST',
            headers: { 
              'Content-Type': 'application/json' 
            },
            body: JSON.stringify({ email, password })
          });

          if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || errorData.message || 'Erreur de connexion');
          }

          const data: AuthResponse = await response.json();
          
          // Support pour les 2 formats de réponse
          const token = data.data?.access_token || data.data?.token || data.access_token;
          const refreshToken = data.data?.refresh_token || data.refresh_token;
          const user = data.data?.user || data.user;
          
          if (!token) {
            throw new Error('Token d\'accès manquant dans la réponse');
          }

          // Stockage des tokens
          localStorage.setItem('access_token', token);
          if (refreshToken) {
            localStorage.setItem('refresh_token', refreshToken);
          }

          // Extraction des infos utilisateur depuis le token si pas fourni
          const userInfo = user || getUserFromToken(token);
          
          if (!userInfo) {
            throw new Error('Impossible d\'extraire les informations utilisateur');
          }

          set({ 
            user: userInfo,
            accessToken: token,
            refreshToken,
            isLoading: false, 
            error: null,
            isAuthenticated: true
          });

        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Erreur de connexion';
          set({ 
            user: null, 
            accessToken: null,
            refreshToken: null,
            isLoading: false, 
            error: errorMessage,
            isAuthenticated: false
          });
          throw error;
        }
      },

      // Inscription - Compatible avec l'ancien endpoint
      register: async (email: string, username: string, password: string, firstName?: string, lastName?: string) => {
        set({ isLoading: true, error: null });

        try {
          const payload: any = { email, username, password };
          if (firstName) payload.first_name = firstName;
          if (lastName) payload.last_name = lastName;

          const response = await fetch(`${API_BASE_URL}/auth/signup`, {
            method: 'POST',
            headers: { 
              'Content-Type': 'application/json' 
            },
            body: JSON.stringify(payload)
          });

          if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.error || errorData.message || 'Erreur lors de l\'inscription');
          }

          const data: AuthResponse = await response.json();
          
          // Support pour les 2 formats de réponse
          const token = data.data?.access_token || data.data?.token || data.access_token;
          const refreshToken = data.data?.refresh_token || data.refresh_token;
          const user = data.data?.user || data.user;

          if (!token) {
            throw new Error('Token d\'accès manquant dans la réponse');
          }

          // Stockage des tokens
          localStorage.setItem('access_token', token);
          if (refreshToken) {
            localStorage.setItem('refresh_token', refreshToken);
          }

          // Extraction des infos utilisateur
          const userInfo = user || getUserFromToken(token);
          
          if (!userInfo) {
            throw new Error('Impossible d\'extraire les informations utilisateur');
          }

          set({ 
            user: userInfo,
            accessToken: token,
            refreshToken,
            isLoading: false, 
            error: null,
            isAuthenticated: true
          });

        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Erreur lors de l\'inscription';
          set({ 
            user: null, 
            accessToken: null,
            refreshToken: null,
            isLoading: false, 
            error: errorMessage,
            isAuthenticated: false
          });
          throw error;
        }
      },

      // Déconnexion
      logout: () => {
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        set({ 
          user: null, 
          accessToken: null,
          refreshToken: null,
          isLoading: false, 
          error: null,
          isAuthenticated: false
        });
      },

      // Rafraîchissement du token
      refreshAuth: async () => {
        const { refreshToken } = get();
        if (!refreshToken) return false;

        try {
          const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
            method: 'POST',
            headers: { 
              'Content-Type': 'application/json' 
            },
            body: JSON.stringify({ refresh_token: refreshToken })
          });

          if (!response.ok) throw new Error('Impossible de rafraîchir le token');

          const data: AuthResponse = await response.json();
          const newToken = data.data?.token || data.access_token;
          const newRefreshToken = data.data?.refresh_token || data.refresh_token;

          if (!newToken) throw new Error('Nouveau token manquant');

          localStorage.setItem('access_token', newToken);
          if (newRefreshToken) {
            localStorage.setItem('refresh_token', newRefreshToken);
          }

          const userInfo = getUserFromToken(newToken);
          
          set({
            accessToken: newToken,
            refreshToken: newRefreshToken || refreshToken,
            user: userInfo,
            isAuthenticated: true,
            error: null
          });

          return true;
        } catch (error) {
          // Token refresh a échoué, déconnexion
          get().logout();
          return false;
        }
      },

      // Vérifier l'authentification existante
      checkExistingAuth: async () => {
        const token = localStorage.getItem('access_token');
        const refreshTokenStored = localStorage.getItem('refresh_token');
        
        if (!token) {
          set({ user: null, isAuthenticated: false });
          return false;
        }

        // Vérifier la validité du token
        if (!isTokenValid(token)) {
          // Essayer de rafraîchir le token
          set({ refreshToken: refreshTokenStored });
          const refreshed = await get().refreshAuth();
          if (!refreshed) {
          set({ 
            user: null, 
              accessToken: null,
              refreshToken: null,
              isAuthenticated: false,
            error: 'Session expirée, veuillez vous reconnecter' 
          });
            return false;
          }
          return true;
        }

        // Token valide, récupérer les infos utilisateur
        const user = getUserFromToken(token);
        if (user) {
          set({ 
            user,
            accessToken: token,
            refreshToken: refreshTokenStored,
            isAuthenticated: true,
            error: null 
          });
          return true;
        } else {
          // Token invalide
          localStorage.removeItem('access_token');
          localStorage.removeItem('refresh_token');
          set({ 
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false,
            error: 'Token invalide' 
          });
          return false;
        }
      },

      // Définir les tokens manuellement (utile pour les tests)
      setTokens: (accessToken: string, refreshToken?: string) => {
        localStorage.setItem('access_token', accessToken);
        if (refreshToken) {
          localStorage.setItem('refresh_token', refreshToken);
        }

        const user = getUserFromToken(accessToken);
        set({
          accessToken,
          refreshToken,
          user,
          isAuthenticated: !!user,
          error: null
        });
      },

      clearError: () => {
        set({ error: null });
      }
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ 
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated
      })
    }
  )
); 