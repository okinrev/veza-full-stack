import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useSessionTimeout } from '../useSessionTimeout';
import { useAuthStore } from '../../store/authStore';
import { resetMocks } from '@/test/utils';
import { useNotifications } from '@/shared/utils/notifications';

// Mock des notifications
vi.mock('@/shared/utils/notifications', () => ({
  useNotifications: () => ({
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn(),
  }),
}));

describe('useSessionTimeout', () => {
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

    // Mock des événements du navigateur
    vi.spyOn(window, 'addEventListener');
    vi.spyOn(window, 'removeEventListener');
  });

  it('devrait mettre à jour l\'activité sur les événements utilisateur', () => {
    const { result } = renderHook(() => useSessionTimeout());

    // Simuler un événement de souris
    act(() => {
      window.dispatchEvent(new MouseEvent('mousemove'));
    });

    expect(useAuthStore.getState().lastActivity).toBeGreaterThan(0);
  });

  it('devrait nettoyer les écouteurs d\'événements', () => {
    const { unmount } = renderHook(() => useSessionTimeout());

    unmount();

    expect(window.removeEventListener).toHaveBeenCalledWith('mousemove', expect.any(Function));
    expect(window.removeEventListener).toHaveBeenCalledWith('keydown', expect.any(Function));
    expect(window.removeEventListener).toHaveBeenCalledWith('click', expect.any(Function));
    expect(window.removeEventListener).toHaveBeenCalledWith('scroll', expect.any(Function));
  });

  it('devrait avertir l\'utilisateur avant l\'expiration de la session', async () => {
    const mockNotifications = {
      success: vi.fn(),
      error: vi.fn(),
      warning: vi.fn(),
      info: vi.fn(),
    };

    vi.mocked(useNotifications).mockReturnValue(mockNotifications);

    // Simuler une session proche de l'expiration
    const oldTimestamp = Date.now() - 25 * 60 * 1000; // 25 minutes
    useAuthStore.setState({
      isAuthenticated: true,
      lastActivity: oldTimestamp,
    });

    renderHook(() => useSessionTimeout());

    // Attendre que l'intervalle de vérification s'exécute
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 100));
    });

    expect(mockNotifications.warning).toHaveBeenCalledWith(
      'Session expirant',
      expect.stringContaining('minute')
    );
  });

  it('devrait déconnecter l\'utilisateur après l\'expiration de la session', async () => {
    const mockNotifications = {
      success: vi.fn(),
      error: vi.fn(),
      warning: vi.fn(),
      info: vi.fn(),
    };

    vi.mocked(useNotifications).mockReturnValue(mockNotifications);

    // Simuler une session expirée
    const oldTimestamp = Date.now() - 31 * 60 * 1000; // 31 minutes
    useAuthStore.setState({
      isAuthenticated: true,
      lastActivity: oldTimestamp,
    });

    renderHook(() => useSessionTimeout());

    // Attendre que l'intervalle de vérification s'exécute
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 100));
    });

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(mockNotifications.error).toHaveBeenCalledWith(
      'Session expirée',
      'Vous avez été déconnecté pour inactivité.'
    );
  });
}); 