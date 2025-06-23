import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useAuthStore } from '../authStore';
import { resetMocks, mockApiError } from '@/test/utils';

describe('AuthStore', () => {
  beforeEach(() => {
    resetMocks();
    useAuthStore.setState({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
      lastActivity: null,
    });
  });

  describe('login', () => {
    it('devrait mettre à jour l\'état après une connexion réussie', async () => {
      const mockUser = {
        id: '1',
        email: 'test@example.com',
        username: 'testuser',
      };
      const mockTokens = {
        accessToken: 'access-token',
        refreshToken: 'refresh-token',
      };

      vi.spyOn(global, 'fetch').mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ user: mockUser, ...mockTokens }),
      } as Response);

      await useAuthStore.getState().login({
        email: 'test@example.com',
        password: 'password123',
      });

      const state = useAuthStore.getState();
      expect(state.user).toEqual(mockUser);
      expect(state.accessToken).toBe(mockTokens.accessToken);
      expect(state.refreshToken).toBe(mockTokens.refreshToken);
      expect(state.isAuthenticated).toBe(true);
      expect(state.error).toBeNull();
    });

    it('devrait gérer les erreurs de connexion', async () => {
      const errorMessage = 'Identifiants invalides';
      vi.spyOn(global, 'fetch').mockRejectedValueOnce(mockApiError(401, errorMessage));

      await useAuthStore.getState().login({
        email: 'test@example.com',
        password: 'wrongpassword',
      });

      const state = useAuthStore.getState();
      expect(state.error).toBe(errorMessage);
      expect(state.isAuthenticated).toBe(false);
      expect(state.user).toBeNull();
    });
  });

  describe('logout', () => {
    it('devrait réinitialiser l\'état lors de la déconnexion', () => {
      // Préparer un état connecté
      useAuthStore.setState({
        user: { id: '1', email: 'test@example.com', username: 'testuser' },
        accessToken: 'access-token',
        refreshToken: 'refresh-token',
        isAuthenticated: true,
        lastActivity: Date.now(),
      });

      useAuthStore.getState().logout();

      const state = useAuthStore.getState();
      expect(state.user).toBeNull();
      expect(state.accessToken).toBeNull();
      expect(state.refreshToken).toBeNull();
      expect(state.isAuthenticated).toBe(false);
      expect(state.lastActivity).toBeNull();
    });
  });

  describe('session timeout', () => {
    it('devrait détecter l\'expiration de la session', () => {
      const oldTimestamp = Date.now() - 31 * 60 * 1000; // 31 minutes
      useAuthStore.setState({
        isAuthenticated: true,
        lastActivity: oldTimestamp,
      });

      const hasTimedOut = useAuthStore.getState().checkSessionTimeout();
      expect(hasTimedOut).toBe(true);
    });

    it('ne devrait pas expirer une session active', () => {
      useAuthStore.setState({
        isAuthenticated: true,
        lastActivity: Date.now(),
      });

      const hasTimedOut = useAuthStore.getState().checkSessionTimeout();
      expect(hasTimedOut).toBe(false);
    });
  });
}); 