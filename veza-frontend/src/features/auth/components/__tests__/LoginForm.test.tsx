import { describe, it, expect, beforeEach, vi } from 'vitest';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import { render } from '@/test/utils';
import { LoginForm } from '../LoginForm';
import { useAuthStore } from '../../store/authStore';
import { resetMocks } from '@/test/utils';

describe('LoginForm', () => {
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

  it('devrait afficher le formulaire de connexion', () => {
    render(<LoginForm />);

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/mot de passe/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /se connecter/i })).toBeInTheDocument();
    expect(screen.getByText(/pas encore de compte/i)).toBeInTheDocument();
  });

  it('devrait valider les champs requis', async () => {
    render(<LoginForm />);

    const submitButton = screen.getByRole('button', { name: /se connecter/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/l'email est requis/i)).toBeInTheDocument();
      expect(screen.getByText(/le mot de passe est requis/i)).toBeInTheDocument();
    });
  });

  it('devrait valider le format de l\'email', async () => {
    render(<LoginForm />);

    const emailInput = screen.getByLabelText(/email/i);
    fireEvent.change(emailInput, { target: { value: 'invalid-email' } });

    const submitButton = screen.getByRole('button', { name: /se connecter/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/format d'email invalide/i)).toBeInTheDocument();
    });
  });

  it('devrait gérer la connexion réussie', async () => {
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

    render(<LoginForm />);

    const emailInput = screen.getByLabelText(/email/i);
    const passwordInput = screen.getByLabelText(/mot de passe/i);
    const submitButton = screen.getByRole('button', { name: /se connecter/i });

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(useAuthStore.getState().isAuthenticated).toBe(true);
      expect(useAuthStore.getState().user).toEqual(mockUser);
    });
  });

  it('devrait afficher les erreurs de connexion', async () => {
    const errorMessage = 'Identifiants invalides';
    vi.spyOn(global, 'fetch').mockRejectedValueOnce({
      response: {
        status: 401,
        data: { message: errorMessage },
      },
    });

    render(<LoginForm />);

    const emailInput = screen.getByLabelText(/email/i);
    const passwordInput = screen.getByLabelText(/mot de passe/i);
    const submitButton = screen.getByRole('button', { name: /se connecter/i });

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
  });
}); 