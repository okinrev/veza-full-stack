import { describe, it, expect, beforeEach, vi } from 'vitest';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import { render } from '@/test/utils';
import { RegisterForm } from '../RegisterForm';
import { useAuthStore } from '../../store/authStore';
import { resetMocks } from '@/test/utils';

describe('RegisterForm', () => {
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

  it('devrait afficher le formulaire d\'inscription', () => {
    render(<RegisterForm />);

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/nom d'utilisateur/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/mot de passe/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/confirmer le mot de passe/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /s'inscrire/i })).toBeInTheDocument();
    expect(screen.getByText(/déjà un compte/i)).toBeInTheDocument();
  });

  it('devrait valider les champs requis', async () => {
    render(<RegisterForm />);

    const submitButton = screen.getByRole('button', { name: /s'inscrire/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/l'email est requis/i)).toBeInTheDocument();
      expect(screen.getByText(/le nom d'utilisateur est requis/i)).toBeInTheDocument();
      expect(screen.getByText(/le mot de passe est requis/i)).toBeInTheDocument();
    });
  });

  it('devrait valider le format de l\'email', async () => {
    render(<RegisterForm />);

    const emailInput = screen.getByLabelText(/email/i);
    fireEvent.change(emailInput, { target: { value: 'invalid-email' } });

    const submitButton = screen.getByRole('button', { name: /s'inscrire/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/format d'email invalide/i)).toBeInTheDocument();
    });
  });

  it('devrait valider la longueur du nom d\'utilisateur', async () => {
    render(<RegisterForm />);

    const usernameInput = screen.getByLabelText(/nom d'utilisateur/i);
    fireEvent.change(usernameInput, { target: { value: 'ab' } });

    const submitButton = screen.getByRole('button', { name: /s'inscrire/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/le nom d'utilisateur doit contenir au moins 3 caractères/i)).toBeInTheDocument();
    });
  });

  it('devrait valider la complexité du mot de passe', async () => {
    render(<RegisterForm />);

    const passwordInput = screen.getByLabelText(/mot de passe/i);
    fireEvent.change(passwordInput, { target: { value: 'simple' } });

    const submitButton = screen.getByRole('button', { name: /s'inscrire/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/le mot de passe doit contenir au moins une majuscule, une minuscule, un chiffre et un caractère spécial/i)).toBeInTheDocument();
    });
  });

  it('devrait valider la correspondance des mots de passe', async () => {
    render(<RegisterForm />);

    const passwordInput = screen.getByLabelText(/mot de passe/i);
    const confirmPasswordInput = screen.getByLabelText(/confirmer le mot de passe/i);

    fireEvent.change(passwordInput, { target: { value: 'Password123!' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'Password456!' } });

    const submitButton = screen.getByRole('button', { name: /s'inscrire/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/les mots de passe ne correspondent pas/i)).toBeInTheDocument();
    });
  });

  it('devrait gérer l\'inscription réussie', async () => {
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

    render(<RegisterForm />);

    const emailInput = screen.getByLabelText(/email/i);
    const usernameInput = screen.getByLabelText(/nom d'utilisateur/i);
    const passwordInput = screen.getByLabelText(/mot de passe/i);
    const confirmPasswordInput = screen.getByLabelText(/confirmer le mot de passe/i);
    const submitButton = screen.getByRole('button', { name: /s'inscrire/i });

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(usernameInput, { target: { value: 'testuser' } });
    fireEvent.change(passwordInput, { target: { value: 'Password123!' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'Password123!' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(useAuthStore.getState().isAuthenticated).toBe(true);
      expect(useAuthStore.getState().user).toEqual(mockUser);
    });
  });

  it('devrait afficher les erreurs d\'inscription', async () => {
    const errorMessage = 'Cet email est déjà utilisé';
    vi.spyOn(global, 'fetch').mockRejectedValueOnce({
      response: {
        status: 400,
        data: { message: errorMessage },
      },
    });

    render(<RegisterForm />);

    const emailInput = screen.getByLabelText(/email/i);
    const usernameInput = screen.getByLabelText(/nom d'utilisateur/i);
    const passwordInput = screen.getByLabelText(/mot de passe/i);
    const confirmPasswordInput = screen.getByLabelText(/confirmer le mot de passe/i);
    const submitButton = screen.getByRole('button', { name: /s'inscrire/i });

    fireEvent.change(emailInput, { target: { value: 'existing@example.com' } });
    fireEvent.change(usernameInput, { target: { value: 'testuser' } });
    fireEvent.change(passwordInput, { target: { value: 'Password123!' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'Password123!' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
  });
}); 