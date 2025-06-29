import axios, { AxiosError } from 'axios';
import type { AxiosResponse, AxiosRequestConfig } from 'axios';
import { useAuthStore } from '@/features/auth/store/authStore';

// Configuration des URLs
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';
const FILES_BASE_URL = import.meta.env.VITE_FILES_BASE_URL || 'http://localhost:8080';
const DEBUG = import.meta.env.VITE_DEBUG === 'true';

// Interface pour les réponses API standardisées
interface ApiResponse<T = any> {
  success: boolean;
  message?: string;
  data?: T;
  error?: string;
  details?: any;
}

// Interface pour les erreurs API
interface ApiError {
  message: string;
  status?: number;
  code?: string;
  details?: any;
}

// Types pour les méthodes HTTP
type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

// Configuration du client Axios principal
export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000, // 30 secondes
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
  withCredentials: false
});

// Intercepteur de requête pour ajouter le token d'authentification JWT unifié
apiClient.interceptors.request.use(
  (config) => {
    // Ajouter le token d'authentification
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
      
      // Ajouter les headers JWT unifiés pour compatibilité inter-services Talas
      config.headers['X-JWT-Issuer'] = import.meta.env.VITE_JWT_ISSUER || 'veza-platform';
      config.headers['X-JWT-Audience'] = import.meta.env.VITE_JWT_AUDIENCE || 'veza-services';
      config.headers['X-Service-Name'] = 'frontend-react';
    }

    // Logs de debug
    if (DEBUG) {
      console.log(`🔵 [API] ${config.method?.toUpperCase()} ${config.url}`, {
        params: config.params,
        data: config.data,
        jwtHeaders: token ? { issuer: config.headers['X-JWT-Issuer'], audience: config.headers['X-JWT-Audience'] } : null
      });
    }

    return config;
  },
  (error) => {
    if (DEBUG) {
      console.error('🔴 [API] Erreur de requête:', error);
    }
    return Promise.reject(error);
  }
);

// Intercepteur de réponse pour gérer les erreurs et l'authentification
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    if (DEBUG) {
      console.log(`🟢 [API] ${response.config.method?.toUpperCase()} ${response.config.url}`, {
        status: response.status,
        data: response.data
      });
    }
    return response;
  },
  async (error: AxiosError) => {
    if (DEBUG) {
      console.error(`🔴 [API] ${error.config?.method?.toUpperCase()} ${error.config?.url}`, {
        status: error.response?.status,
        data: error.response?.data
      });
    }

    // Gestion de l'authentification expirée
    if (error.response?.status === 401) {
      const authStore = useAuthStore.getState();
      
      // Essayer de rafraîchir le token
      const refreshed = await authStore.refreshAuth();
      if (refreshed && error.config) {
        // Retry la requête avec le nouveau token
        const newToken = localStorage.getItem('access_token');
        if (newToken) {
          error.config.headers.Authorization = `Bearer ${newToken}`;
          return apiClient.request(error.config);
        }
      }
      
      // Si le refresh a échoué, déconnecter l'utilisateur
      authStore.logout();
      window.location.href = '/login';
    }

    return Promise.reject(error);
  }
);

// Client pour les uploads de fichiers
export const uploadClient = axios.create({
  baseURL: FILES_BASE_URL,
  timeout: 300000, // 5 minutes pour les uploads
  headers: {
    'Accept': 'application/json',
  },
  withCredentials: false
});

// Intercepteur pour les uploads
uploadClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Classe principale pour les appels API
class ApiService {
  // Méthode générique pour les requêtes
  private async request<T>(
    method: HttpMethod,
    url: string,
    options: {
      params?: any;
      data?: any;
      config?: AxiosRequestConfig;
    } = {}
  ): Promise<T> {
    try {
      const response = await apiClient.request<ApiResponse<T>>({
        method,
        url,
        params: options.params,
        data: options.data,
        ...options.config
      });

      // Gérer les réponses avec format standard API
      if (response.data && typeof response.data === 'object' && 'success' in response.data) {
        const apiResponse = response.data as ApiResponse<T>;
        if (!apiResponse.success) {
          throw new Error(apiResponse.error || apiResponse.message || 'Erreur API');
        }
        return apiResponse.data as T;
      }

      // Retourner directement les données si pas de format standard
      return response.data as T;
    } catch (error) {
      this.handleError(error as AxiosError);
      throw error;
    }
  }

  // Gestion des erreurs
  private handleError(error: AxiosError) {
    if (error.response) {
      // Erreur HTTP avec réponse du serveur
      const status = error.response.status;
      const data = error.response.data as any;
      
      let message = 'Erreur du serveur';
      if (data?.error) message = data.error;
      else if (data?.message) message = data.message;
      else if (status === 404) message = 'Ressource non trouvée';
      else if (status === 403) message = 'Accès interdit';
      else if (status === 500) message = 'Erreur interne du serveur';

      throw new Error(message);
    } else if (error.request) {
      // Erreur réseau
      throw new Error('Impossible de contacter le serveur');
    } else {
      // Autre erreur
      throw new Error(error.message || 'Erreur inconnue');
    }
  }

  // Méthodes HTTP conveniences
  async get<T>(url: string, params?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.request<T>('GET', url, { params, config });
  }

  async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.request<T>('POST', url, { data, config });
  }

  async put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.request<T>('PUT', url, { data, config });
  }

  async patch<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.request<T>('PATCH', url, { data, config });
  }

  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return this.request<T>('DELETE', url, { config });
  }

  // Upload de fichiers
  async upload<T>(
    url: string, 
    file: File | FormData, 
    onProgress?: (progress: number) => void
  ): Promise<T> {
    const formData = file instanceof FormData ? file : new FormData();
    if (file instanceof File) {
      formData.append('file', file);
    }

    try {
      const response = await uploadClient.post<ApiResponse<T>>(url, formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
        onUploadProgress: (progressEvent) => {
          if (onProgress && progressEvent.total) {
            const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
            onProgress(progress);
          }
        }
      });

      if (response.data && typeof response.data === 'object' && 'success' in response.data) {
        const apiResponse = response.data as ApiResponse<T>;
        if (!apiResponse.success) {
          throw new Error(apiResponse.error || apiResponse.message || 'Erreur upload');
        }
        return apiResponse.data as T;
      }

      return response.data as T;
    } catch (error) {
      this.handleError(error as AxiosError);
      throw error;
    }
  }

  // Download de fichiers
  async download(url: string, filename?: string): Promise<void> {
    try {
      const response = await apiClient.get(url, {
        responseType: 'blob'
      });

      // Créer un lien de téléchargement
      const blob = new Blob([response.data]);
      const downloadUrl = window.URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = downloadUrl;
      link.download = filename || 'download';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(downloadUrl);
    } catch (error) {
      this.handleError(error as AxiosError);
      throw error;
    }
  }
}

// Instance singleton du service API
export const api = new ApiService();

// Fonctions utilitaires pour les URLs
export const urls = {
  // Authentification
  auth: {
    login: '/auth/login',
    signup: '/auth/signup',
    logout: '/auth/logout',
    refresh: '/auth/refresh',
    me: '/auth/me'
  },
  
  // Utilisateurs
  users: {
    list: '/users',
    me: '/users/me',
    byId: (id: number) => `/users/${id}`,
    exceptMe: '/users/except-me'
  },
  
  // Chat
  chat: {
    rooms: '/chat/rooms',
    room: (name: string) => `/chat/rooms/${name}`,
    messages: (roomName: string) => `/chat/rooms/${roomName}/messages`,
    directMessages: (userId: number) => `/chat/dm/${userId}`
  },
  
  // Produits
  products: {
    list: '/products',
    byId: (id: number) => `/products/${id}`,
    create: '/products',
    update: (id: number) => `/products/${id}`,
    delete: (id: number) => `/products/${id}`
  },
  
  // Admin
  admin: {
    products: '/admin/products',
    categories: '/admin/categories',
    users: '/admin/users'
  },
  
  // Fichiers/uploads
  files: {
    upload: '/files/upload',
    download: (id: string) => `/files/${id}`,
    delete: (id: string) => `/files/${id}`
  }
};

// Export par défaut
export default api; 