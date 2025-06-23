import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { apiClient } from '../api/client';

// Types compatibles avec le backend Go
export interface User {
  id: number;
  username: string;
  email: string;
  first_name?: {
    String: string;
    Valid: boolean;
  } | string;
  last_name?: {
    String: string;
    Valid: boolean;
  } | string;
  bio?: {
    String: string;
    Valid: boolean;
  } | string;
  avatar?: {
    String: string;
    Valid: boolean;
  } | string;
  role: string;
  is_active: boolean;
  is_verified?: boolean;
  last_login_at?: {
    Time: string;
    Valid: boolean;
  };
  created_at: string;
  updated_at: string;
}

export interface LoginCredentials {
  email: string; // Backend utilise email, pas username
  password: string;
}

export interface RegisterData {
  username: string;
  email: string;
  password: string;
  first_name?: string;
  last_name?: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}

interface AuthState {
  // État
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  login: (credentials: LoginCredentials) => Promise<void>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => void;
  refreshAuth: () => Promise<void>;
  updateProfile: (data: Partial<User>) => Promise<void>;
  checkAuth: () => Promise<void>;
  clearError: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      // État initial
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      // Actions
      login: async (credentials: LoginCredentials) => {
        try {
          set({ isLoading: true, error: null });
          
          const response = await apiClient.post<LoginResponse>('/api/v1/auth/login', credentials);
          
          set({
            user: response.user,
            accessToken: response.access_token,
            refreshToken: response.refresh_token,
            isAuthenticated: true,
            isLoading: false
          });
          
        } catch (error: any) {
          const errorMessage = error.response?.data?.error || 'Erreur de connexion';
          set({ 
            error: errorMessage, 
            isLoading: false,
            isAuthenticated: false,
            user: null,
            accessToken: null,
            refreshToken: null
          });
          throw new Error(errorMessage);
        }
      },

      register: async (data: RegisterData) => {
        try {
          set({ isLoading: true, error: null });
          
          const response = await apiClient.post<{user: User}>('/api/v1/auth/register', data);
          
          // Après l'inscription, on connecte automatiquement l'utilisateur
          await get().login({ email: data.email, password: data.password });
          
        } catch (error: any) {
          const errorMessage = error.response?.data?.error || 'Erreur lors de l\'inscription';
          set({ 
            error: errorMessage, 
            isLoading: false 
          });
          throw new Error(errorMessage);
        }
      },

      logout: async () => {
        try {
          const { refreshToken } = get();
          if (refreshToken) {
            await apiClient.post('/api/v1/auth/logout', { refresh_token: refreshToken });
          }
        } catch (error) {
          console.error('Erreur lors de la déconnexion:', error);
        } finally {
          set({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false,
            error: null
          });
        }
      },

      refreshAuth: async () => {
        try {
          const { refreshToken } = get();
          if (!refreshToken) {
            throw new Error('Aucun refresh token disponible');
          }

          const response = await apiClient.post<{access_token: string; expires_in: number}>(
            '/api/v1/auth/refresh', 
            { refresh_token: refreshToken }
          );

          set({
            accessToken: response.access_token,
          });

        } catch (error) {
          console.error('Erreur lors du refresh du token:', error);
          get().logout();
          throw error;
        }
      },

      updateProfile: async (data: Partial<User>) => {
        try {
          set({ isLoading: true, error: null });
          
          const response = await apiClient.put<{user: User}>('/api/v1/users/me', data);
          
          set({
            user: response.user,
            isLoading: false
          });
          
        } catch (error: any) {
          const errorMessage = error.response?.data?.error || 'Erreur lors de la mise à jour';
          set({ 
            error: errorMessage, 
            isLoading: false 
          });
          throw new Error(errorMessage);
        }
      },

      checkAuth: async () => {
        try {
          const { accessToken } = get();
          if (!accessToken) {
            return;
          }

          const user = await apiClient.get<User>('/api/v1/auth/me');
          
          set({
            user,
            isAuthenticated: true
          });
          
        } catch (error) {
          console.error('Erreur lors de la vérification d\'authentification:', error);
          get().logout();
        }
      },

      clearError: () => set({ error: null })
    }),
    {
      name: 'auth-store',
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated
      })
    }
  )
); 