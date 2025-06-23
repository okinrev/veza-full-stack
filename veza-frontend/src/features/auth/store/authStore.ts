import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';
import { apiClient } from '@/shared/api/client';
import axios from 'axios';
import type { User } from '@/shared/api/types';

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  lastActivity: number | null;
  
  // Actions
  login: (email: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<void>;
  logout: () => void;
  refreshAuth: () => Promise<void>;
  clearError: () => void;
  updateLastActivity: () => void;
  checkSessionTimeout: () => boolean;
}

const SESSION_TIMEOUT = 30 * 60 * 1000; // 30 minutes

export const useAuthStore = create<AuthState>()(
  devtools(
    persist(
      immer((set, get) => ({
        // Initial state
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
        lastActivity: null,
        
        // Actions
        login: async (email: string, password: string) => {
          set((state) => {
            state.isLoading = true;
            state.error = null;
          });
          
          try {
            const response = await apiClient.post<{ access_token: string; user: User }>('/auth/login', { email, password });
            const { access_token, user } = response;
            set((state) => {
              state.user = user;
              state.accessToken = access_token;
              state.isAuthenticated = true;
              state.isLoading = false;
              state.lastActivity = Date.now();
            });
          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Une erreur est survenue';
              state.isLoading = false;
            });
          }
        },
        
        register: async (username: string, email: string, password: string) => {
          set((state) => {
            state.isLoading = true;
            state.error = null;
          });
          
          try {
            await apiClient.post('/auth/register', { username, email, password });
            set((state) => {
              state.isLoading = false;
            });
          } catch (error) {
            set((state) => {
              state.error = error instanceof Error ? error.message : 'Une erreur est survenue';
              state.isLoading = false;
            });
          }
        },
        
        logout: () => {
          set((state) => {
            state.user = null;
            state.accessToken = null;
            state.refreshToken = null;
            state.isAuthenticated = false;
            state.error = null;
            state.lastActivity = null;
          });
        },
        
        refreshAuth: async () => {
          const { refreshToken } = get();
          if (!refreshToken) return;
          
          try {
            const response = await axios.post('/api/auth/refresh', {
              refreshToken,
            });
            const { accessToken, newRefreshToken } = response.data;
            set((state) => {
              state.accessToken = accessToken;
              state.refreshToken = newRefreshToken;
              state.lastActivity = Date.now();
            });
          } catch (error) {
            get().logout();
          }
        },
        
        clearError: () =>
          set((state) => {
            state.error = null;
          }),
        
        updateLastActivity: () => {
          set((state) => {
            state.lastActivity = Date.now();
          });
        },
        
        checkSessionTimeout: () => {
          const { lastActivity, isAuthenticated } = get();
          if (!isAuthenticated || !lastActivity) return false;
          
          const now = Date.now();
          const timeSinceLastActivity = now - lastActivity;
          
          if (timeSinceLastActivity > SESSION_TIMEOUT) {
            get().logout();
            return true;
          }
          
          return false;
        },
      })),
      {
        name: 'auth-storage',
        partialize: (state) => ({
          user: state.user,
          accessToken: state.accessToken,
          refreshToken: state.refreshToken,
          isAuthenticated: state.isAuthenticated,
          lastActivity: state.lastActivity,
        }),
      }
    )
  )
); 