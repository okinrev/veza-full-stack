# API Client - Configuration et Utilisation

## Vue d'ensemble

Le client API est le point central pour toutes les communications HTTP avec le backend Go. Il utilise Axios avec des intercepteurs pour l'authentification automatique, la gestion d'erreurs et le retry logic.

## Fichiers Concernés

- `src/shared/api/client.ts` - Client API principal
- `src/shared/api/endpoints.ts` - Définition des endpoints
- `src/shared/api/types.ts` - Types TypeScript pour l'API

## Configuration du Client

### Création de l'Instance

```typescript
// src/shared/api/client.ts
class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL, // http://localhost:8080
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      withCredentials: false // Important pour CORS
    });
    
    this.setupInterceptors();
  }
}
```

### Variables d'Environnement

```typescript
const API_BASE_URL = import.meta.env.VITE_API_URL || 
                     import.meta.env.VITE_API_BASE_URL || 
                     'http://localhost:8080';
```

**Variables requises dans `.env`** :
```env
VITE_API_URL=http://localhost:8080
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## Intercepteurs

### Request Interceptor - Authentification

```typescript
private setupInterceptors() {
  // Injection automatique du token JWT
  this.client.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
      const token = localStorage.getItem('access_token');
      if (token && config.headers) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      
      // Debug logging en développement
      if (import.meta.env.DEV) {
        console.log('🚀 API Request:', {
          method: config.method?.toUpperCase(),
          url: config.url,
          data: config.data
        });
      }
      
      return config;
    },
    (error) => Promise.reject(error)
  );
}
```

### Response Interceptor - Gestion d'Erreurs

```typescript
this.client.interceptors.response.use(
  (response) => {
    // Log des réponses en développement
    if (import.meta.env.DEV) {
      console.log('✅ API Response:', {
        status: response.status,
        url: response.config.url,
        data: response.data
      });
    }
    
    // Extraction automatique des données
    return response.data?.data || response.data;
  },
  async (error) => {
    const originalRequest = error.config;
    
    // Gestion de l'expiration du token (401)
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      try {
        await this.refreshToken();
        return this.client(originalRequest);
      } catch (refreshError) {
        // Redirection vers login si refresh échoue
        this.handleAuthError();
        return Promise.reject(refreshError);
      }
    }
    
    // Gestion des autres erreurs
    return this.handleApiError(error);
  }
);
```

## Méthodes Principales

### GET Request

```typescript
async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return this.client.get(url, config);
}

// Utilisation
const users = await apiClient.get<User[]>('/users');
const user = await apiClient.get<User>(`/users/${id}`);
```

### POST Request

```typescript
async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
  return this.client.post(url, data, config);
}

// Utilisation
const newUser = await apiClient.post<User>('/users', {
  username: 'john_doe',
  email: 'john@example.com'
});
```

### PUT/PATCH Request

```typescript
async put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
  return this.client.put(url, data, config);
}

async patch<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
  return this.client.patch(url, data, config);
}

// Utilisation
const updatedUser = await apiClient.put<User>(`/users/${id}`, updateData);
const partialUpdate = await apiClient.patch<User>(`/users/${id}`, { email: 'new@email.com' });
```

### DELETE Request

```typescript
async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
  return this.client.delete(url, config);
}

// Utilisation
await apiClient.delete(`/users/${id}`);
```

### Upload de Fichiers

```typescript
async uploadFile<T>(
  url: string, 
  file: File, 
  additionalData?: Record<string, any>,
  onProgress?: (progress: number) => void
): Promise<T> {
  const formData = new FormData();
  formData.append('file', file);
  
  // Ajout de données supplémentaires
  if (additionalData) {
    Object.entries(additionalData).forEach(([key, value]) => {
      formData.append(key, String(value));
    });
  }
  
  return this.client.post(url, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
    onUploadProgress: (progressEvent) => {
      if (onProgress && progressEvent.total) {
        const progress = Math.round(
          (progressEvent.loaded * 100) / progressEvent.total
        );
        onProgress(progress);
      }
    },
  });
}

// Utilisation
const result = await apiClient.uploadFile(
  '/tracks/upload',
  audioFile,
  { title: 'Ma piste', artist: 'Artiste' },
  (progress) => console.log(`Upload: ${progress}%`)
);
```

## Gestion d'Authentification

### Token Management

```typescript
setAuthToken(token: string) {
  localStorage.setItem('access_token', token);
  this.client.defaults.headers.common['Authorization'] = `Bearer ${token}`;
}

removeAuthToken() {
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  delete this.client.defaults.headers.common['Authorization'];
}
```

### Refresh Token Flow

```typescript
private async refreshToken(): Promise<void> {
  const refreshToken = localStorage.getItem('refresh_token');
  if (!refreshToken) {
    throw new Error('No refresh token available');
  }
  
  try {
    const response = await axios.post(`${API_BASE_URL}/auth/refresh`, {
      refresh_token: refreshToken
    });
    
    const { access_token, refresh_token: newRefreshToken } = response.data;
    
    localStorage.setItem('access_token', access_token);
    localStorage.setItem('refresh_token', newRefreshToken);
    
    this.client.defaults.headers.common['Authorization'] = `Bearer ${access_token}`;
  } catch (error) {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    throw error;
  }
}
```

## Gestion d'Erreurs

### Types d'Erreurs

```typescript
export interface ApiError {
  message: string;
  code?: string;
  details?: any;
}

export interface ApiResponse<T = any> {
  success?: boolean;
  data?: T;
  message?: string;
  error?: string;
  errors?: string[];
}
```

### Traitement des Erreurs

```typescript
private handleApiError(error: AxiosError): Promise<never> {
  let errorMessage = 'Une erreur est survenue';
  let errorCode = 'UNKNOWN_ERROR';
  
  if (error.response) {
    // Erreur avec réponse du serveur
    const { status, data } = error.response;
    
    switch (status) {
      case 400:
        errorMessage = data?.message || 'Requête invalide';
        errorCode = 'BAD_REQUEST';
        break;
      case 401:
        errorMessage = 'Non autorisé';
        errorCode = 'UNAUTHORIZED';
        break;
      case 403:
        errorMessage = 'Accès refusé';
        errorCode = 'FORBIDDEN';
        break;
      case 404:
        errorMessage = 'Ressource non trouvée';
        errorCode = 'NOT_FOUND';
        break;
      case 422:
        errorMessage = data?.errors?.join(', ') || 'Données invalides';
        errorCode = 'VALIDATION_ERROR';
        break;
      case 500:
        errorMessage = 'Erreur serveur interne';
        errorCode = 'INTERNAL_ERROR';
        break;
      default:
        errorMessage = data?.message || `Erreur ${status}`;
    }
  } else if (error.request) {
    // Erreur réseau
    errorMessage = 'Erreur de connexion au serveur';
    errorCode = 'NETWORK_ERROR';
  }
  
  const apiError: ApiError = {
    message: errorMessage,
    code: errorCode,
    details: error.response?.data
  };
  
  // Log de l'erreur en développement
  if (import.meta.env.DEV) {
    console.error('❌ API Error:', apiError);
  }
  
  return Promise.reject(apiError);
}
```

## Utilisation dans les Services

### Pattern Service

```typescript
// Example: UserService
class UserService {
  private apiClient = ApiClient.getInstance();
  
  async getUsers(filters?: UserFilters): Promise<User[]> {
    const params = new URLSearchParams();
    if (filters?.search) params.append('search', filters.search);
    if (filters?.role) params.append('role', filters.role);
    
    return this.apiClient.get<User[]>(`/users?${params}`);
  }
  
  async getUserById(id: number): Promise<User> {
    return this.apiClient.get<User>(`/users/${id}`);
  }
  
  async createUser(userData: CreateUserData): Promise<User> {
    return this.apiClient.post<User>('/users', userData);
  }
  
  async updateUser(id: number, userData: UpdateUserData): Promise<User> {
    return this.apiClient.put<User>(`/users/${id}`, userData);
  }
  
  async deleteUser(id: number): Promise<void> {
    return this.apiClient.delete(`/users/${id}`);
  }
}

export const userService = new UserService();
```

## Configuration Backend Go Requise

### CORS Headers

```go
func setupCORS() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
        c.Header("Access-Control-Allow-Credentials", "true")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    })
}
```

### Format de Réponse Standardisé

```go
type ApiResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Message string      `json:"message,omitempty"`
    Error   string      `json:"error,omitempty"`
    Errors  []string    `json:"errors,omitempty"`
}

// Exemple d'utilisation
func (h *UserHandler) GetUsers(c *gin.Context) {
    users, err := h.userService.GetAll()
    if err != nil {
        c.JSON(500, ApiResponse{
            Success: false,
            Error:   err.Error(),
        })
        return
    }
    
    c.JSON(200, ApiResponse{
        Success: true,
        Data:    users,
    })
}
```

### Authentification JWT

```go
func AuthMiddleware() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(401, ApiResponse{
                Success: false,
                Error:   "Authorization header required",
            })
            c.Abort()
            return
        }
        
        // Supprimer "Bearer " du token
        if strings.HasPrefix(tokenString, "Bearer ") {
            tokenString = tokenString[7:]
        }
        
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(os.Getenv("JWT_SECRET")), nil
        })
        
        if err != nil || !token.Valid {
            c.JSON(401, ApiResponse{
                Success: false,
                Error:   "Invalid token",
            })
            c.Abort()
            return
        }
        
        // Extraire les claims et les stocker dans le contexte
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            c.Set("user_id", claims["user_id"])
            c.Set("username", claims["username"])
        }
        
        c.Next()
    })
}
```

## Testing du Client API

### Mock pour les Tests

```typescript
// src/test/mocks/apiClient.ts
export const mockApiClient = {
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  patch: vi.fn(),
  delete: vi.fn(),
  uploadFile: vi.fn(),
};

// Dans les tests
import { mockApiClient } from '@/test/mocks/apiClient';

beforeEach(() => {
  vi.clearAllMocks();
});

test('should fetch users', async () => {
  const mockUsers = [{ id: 1, username: 'test' }];
  mockApiClient.get.mockResolvedValue(mockUsers);
  
  const users = await userService.getUsers();
  
  expect(mockApiClient.get).toHaveBeenCalledWith('/users?');
  expect(users).toEqual(mockUsers);
});
```

## Debugging et Monitoring

### Logs de Développement

Le client API inclut des logs détaillés en mode développement :

```typescript
if (import.meta.env.DEV) {
  console.log('🚀 API Request:', {
    method: config.method?.toUpperCase(),
    url: config.url,
    data: config.data
  });
}
```

### Métriques de Performance

```typescript
// Ajout possible pour monitoring
const startTime = performance.now();
// ... requête API
const duration = performance.now() - startTime;
console.log(`API call took ${duration}ms`);
```

Cette configuration garantit une communication robuste et sécurisée avec le backend Go, avec une gestion d'erreurs appropriée et une expérience développeur optimisée. 