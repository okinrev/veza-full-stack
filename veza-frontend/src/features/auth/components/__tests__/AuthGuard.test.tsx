import { describe, it, expect, beforeEach, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import { render } from '@/test/utils';
import { AuthGuard } from '../AuthGuard';
import { useAuthStore } from '../../store/authStore';
import { resetMocks } from '@/test/utils';

// Mock de react-router-dom
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useNavigate: () => vi.fn(),
  };
});

describe('AuthGuard', () => {
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

  it('devrait afficher le contenu protégé quand l\'utilisateur est authentifié', () => {
    useAuthStore.setState({
      isAuthenticated: true,
      user: { id: '1', email: 'test@example.com', username: 'testuser' },
    });

    render(
      <AuthGuard>
        <div>Contenu protégé</div>
      </AuthGuard>
    );

    expect(screen.getByText('Contenu protégé')).toBeInTheDocument();
  });

  it('devrait rediriger vers la page de connexion quand l\'utilisateur n\'est pas authentifié', async () => {
    const mockNavigate = vi.fn();
    vi.mocked(useNavigate).mockReturnValue(mockNavigate);

    render(
      <AuthGuard>
        <div>Contenu protégé</div>
      </AuthGuard>
    );

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/login');
    });
  });

  it('devrait afficher un indicateur de chargement pendant la vérification de l\'authentification', () => {
    useAuthStore.setState({
      isLoading: true,
    });

    render(
      <AuthGuard>
        <div>Contenu protégé</div>
      </AuthGuard>
    );

    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('devrait gérer la déconnexion automatique', async () => {
    const oldTimestamp = Date.now() - 31 * 60 * 1000; // 31 minutes
    useAuthStore.setState({
      isAuthenticated: true,
      lastActivity: oldTimestamp,
    });

    const mockNavigate = vi.fn();
    vi.mocked(useNavigate).mockReturnValue(mockNavigate);

    render(
      <AuthGuard>
        <div>Contenu protégé</div>
      </AuthGuard>
    );

    await waitFor(() => {
      expect(useAuthStore.getState().isAuthenticated).toBe(false);
      expect(mockNavigate).toHaveBeenCalledWith('/login');
    });
  });
}); 