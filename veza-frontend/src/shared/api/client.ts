import axios from 'axios';
import type { AxiosInstance, AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios';

// Configuration par défaut - Utiliser HAProxy comme point d'entrée unique
const API_BASE_URL = import.meta.env.VITE_API_URL || import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

// Interfaces pour les réponses API
export interface ApiResponse<T = any> {
  success?: boolean;
  data?: T;
  message?: string;
  error?: string;
  errors?: string[];
}

export interface ApiError {
  message: string;
  code?: string;
  details?: any;
}

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors() {
    // Intercepteur de requête pour ajouter le token d'authentification
    this.client.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        // Récupérer le token depuis le localStorage ou le store
        const authStore = JSON.parse(localStorage.getItem('auth-store') || '{}');
        const accessToken = authStore.state?.accessToken;

        if (accessToken && config.headers) {
          config.headers.Authorization = `Bearer ${accessToken}`;
        }

        // Log pour debugging
        console.log(`[API] ${config.method?.toUpperCase()} ${config.url}`, {
          headers: config.headers,
          data: config.data
        });

        return config;
      },
      (error) => {
        console.error('[API] Erreur de requête:', error);
        return Promise.reject(error);
      }
    );

    // Intercepteur de réponse pour gérer les erreurs et le refresh de token
    this.client.interceptors.response.use(
      (response) => {
        console.log(`[API] Réponse ${response.status}:`, response.data);
        
        // Extraire les données selon le format de réponse du backend
        if (response.data && typeof response.data === 'object') {
          // Si la réponse a une structure avec success/data, on retourne directement data
          if ('data' in response.data) {
            return response.data.data;
          }
          // Sinon on retourne la réponse telle quelle
          return response.data;
        }
        
        return response.data;
      },
      async (error) => {
        const originalRequest = error.config;

        console.error('[API] Erreur de réponse:', {
          status: error.response?.status,
          data: error.response?.data,
          url: originalRequest?.url
        });

        // Gestion de l'erreur 401 (token expiré)
        if (error.response?.status === 401 && !originalRequest._retry) {
          originalRequest._retry = true;

          try {
            const authStore = JSON.parse(localStorage.getItem('auth-store') || '{}');
            const refreshToken = authStore.state?.refreshToken;

            if (refreshToken) {
              console.log('[API] Tentative de refresh du token...');
              
              const refreshResponse = await axios.post(`${API_BASE_URL}/api/v1/auth/refresh`, {
                refresh_token: refreshToken
              });

              const newAccessToken = refreshResponse.data.access_token;

              // Mettre à jour le token dans le store
              const updatedStore = {
                ...authStore,
                state: {
                  ...authStore.state,
                  accessToken: newAccessToken
                }
              };
              localStorage.setItem('auth-store', JSON.stringify(updatedStore));

              // Réessayer la requête originale avec le nouveau token
              originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
              return this.client(originalRequest);
            }
          } catch (refreshError) {
            console.error('[API] Erreur lors du refresh du token:', refreshError);
            
            // Nettoyer le store et rediriger vers login
            localStorage.removeItem('auth-store');
            window.location.href = '/login';
            
            return Promise.reject(refreshError);
          }
        }

        // Transformer l'erreur pour une gestion plus facile
        const apiError: ApiError = {
          message: error.response?.data?.error || 
                   error.response?.data?.message || 
                   error.message || 
                   'Une erreur est survenue',
          code: error.response?.data?.code,
          details: error.response?.data
        };

        return Promise.reject(apiError);
      }
    );
  }

  // Méthodes HTTP génériques
  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return this.client.get(url, config);
  }

  async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.client.post(url, data, config);
  }

  async put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.client.put(url, data, config);
  }

  async patch<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    return this.client.patch(url, data, config);
  }

  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return this.client.delete(url, config);
  }

  // Méthode spécialisée pour l'upload de fichiers
  async uploadFile<T>(
    url: string, 
    file: File, 
    additionalData?: Record<string, any>,
    onProgress?: (progress: number) => void
  ): Promise<T> {
    const formData = new FormData();
    formData.append('file', file);
    
    if (additionalData) {
      Object.entries(additionalData).forEach(([key, value]) => {
        formData.append(key, value);
      });
    }

    return this.client.post(url, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(progress);
        }
      },
    });
  }

  // Méthodes utilitaires
  setAuthToken(token: string) {
    this.client.defaults.headers.common['Authorization'] = `Bearer ${token}`;
  }

  removeAuthToken() {
    delete this.client.defaults.headers.common['Authorization'];
  }

  // Instance singleton
  static getInstance(): ApiClient {
    if (!ApiClient.instance) {
      ApiClient.instance = new ApiClient();
    }
    return ApiClient.instance;
  }

  private static instance: ApiClient;
}

// Export de l'instance par défaut
export const apiClient = ApiClient.getInstance();

// Export des utilitaires
export const createFormData = (file: File, additionalData?: Record<string, any>) => {
  const formData = new FormData();
  formData.append('file', file);
  
  if (additionalData) {
    Object.entries(additionalData).forEach(([key, value]) => {
      formData.append(key, value);
    });
  }
  
  return formData;
};

export default apiClient; 