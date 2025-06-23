import axios from 'axios';
import { useAuthStore } from '@/features/auth/store/authStore';

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Intercepteur pour ajouter le token d'authentification
api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Intercepteur pour gérer les erreurs
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // Si l'erreur est 401 et qu'on n'a pas déjà tenté de rafraîchir le token
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        // Tentative de rafraîchissement du token
        const response = await api.post('/auth/refresh');
        const { access_token } = response.data.data;

        // Mise à jour du store
        useAuthStore.setState({ accessToken: access_token });

        // Réessai de la requête originale
        originalRequest.headers.Authorization = `Bearer ${access_token}`;
        return api(originalRequest);
      } catch (refreshError) {
        // Si le rafraîchissement échoue, déconnexion
        useAuthStore.getState().logout();
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
); 