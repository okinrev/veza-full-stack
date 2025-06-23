# Documentation - Intégration Frontend React

## Vue d'ensemble

Cette documentation détaille l'intégration complète entre le backend Go API et un frontend React. Elle couvre l'authentification, la communication API, les WebSockets, et les bonnes pratiques pour une application moderne.

## Configuration Frontend

### Variables d'Environnement

Créer un fichier `.env` dans le projet React :

```env
# API Configuration
REACT_APP_API_URL=http://localhost:8080/api/v1
REACT_APP_WS_URL=ws://localhost:8080/ws
REACT_APP_UPLOAD_URL=http://localhost:8080/uploads

# Environment
REACT_APP_ENV=development
REACT_APP_DEBUG=true

# Features
REACT_APP_ENABLE_CHAT=true
REACT_APP_ENABLE_AUDIO=true
REACT_APP_MAX_FILE_SIZE=10485760
```

### Configuration des URLs

```javascript
// config/api.js
const config = {
    API_BASE_URL: process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1',
    WS_BASE_URL: process.env.REACT_APP_WS_URL || 'ws://localhost:8080/ws',
    UPLOAD_URL: process.env.REACT_APP_UPLOAD_URL || 'http://localhost:8080/uploads',
    
    // Timeouts
    REQUEST_TIMEOUT: 10000,
    UPLOAD_TIMEOUT: 30000,
    
    // Pagination
    DEFAULT_PAGE_SIZE: 20,
    MAX_PAGE_SIZE: 100,
};

export default config;
```

## Client API

### Service API de Base

```javascript
// services/apiClient.js
import config from '../config/api';

class APIClient {
    constructor() {
        this.baseURL = config.API_BASE_URL;
        this.timeout = config.REQUEST_TIMEOUT;
    }

    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const token = this.getAuthToken();

        const defaultOptions = {
            headers: {
                'Content-Type': 'application/json',
                ...(token && { 'Authorization': `Bearer ${token}` }),
            },
            timeout: this.timeout,
        };

        const finalOptions = {
            ...defaultOptions,
            ...options,
            headers: {
                ...defaultOptions.headers,
                ...options.headers,
            },
        };

        try {
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), this.timeout);

            const response = await fetch(url, {
                ...finalOptions,
                signal: controller.signal,
            });

            clearTimeout(timeoutId);

            if (!response.ok) {
                await this.handleErrorResponse(response);
            }

            return await response.json();
        } catch (error) {
            if (error.name === 'AbortError') {
                throw new Error('Request timeout');
            }
            throw error;
        }
    }

    async handleErrorResponse(response) {
        const data = await response.json().catch(() => ({}));
        
        switch (response.status) {
            case 401:
                this.handleUnauthorized();
                throw new Error('Unauthorized');
            case 403:
                throw new Error('Forbidden');
            case 404:
                throw new Error('Not found');
            case 422:
                throw new ValidationError(data.errors || {});
            case 429:
                throw new Error('Too many requests');
            default:
                throw new Error(data.error || 'Server error');
        }
    }

    handleUnauthorized() {
        // Supprimer le token et rediriger vers login
        localStorage.removeItem('authToken');
        localStorage.removeItem('refreshToken');
        window.location.href = '/login';
    }

    getAuthToken() {
        return localStorage.getItem('authToken');
    }

    // Méthodes HTTP de base
    async get(endpoint, params = {}) {
        const queryString = new URLSearchParams(params).toString();
        const url = queryString ? `${endpoint}?${queryString}` : endpoint;
        return this.request(url, { method: 'GET' });
    }

    async post(endpoint, data) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async put(endpoint, data) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data),
        });
    }

    async delete(endpoint) {
        return this.request(endpoint, { method: 'DELETE' });
    }

    // Upload de fichiers
    async upload(endpoint, file, onProgress) {
        const formData = new FormData();
        formData.append('file', file);

        const token = this.getAuthToken();

        return new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest();

            xhr.upload.onprogress = (event) => {
                if (event.lengthComputable && onProgress) {
                    const progress = (event.loaded / event.total) * 100;
                    onProgress(progress);
                }
            };

            xhr.onload = () => {
                if (xhr.status >= 200 && xhr.status < 300) {
                    resolve(JSON.parse(xhr.responseText));
                } else {
                    reject(new Error(`Upload failed: ${xhr.statusText}`));
                }
            };

            xhr.onerror = () => reject(new Error('Upload failed'));

            xhr.open('POST', `${this.baseURL}${endpoint}`);
            if (token) {
                xhr.setRequestHeader('Authorization', `Bearer ${token}`);
            }

            xhr.send(formData);
        });
    }
}

// Erreur de validation personnalisée
class ValidationError extends Error {
    constructor(errors) {
        super('Validation failed');
        this.name = 'ValidationError';
        this.errors = errors;
    }
}

export default new APIClient();
export { ValidationError };
```

## Services Spécialisés

### Service d'Authentification

```javascript
// services/authService.js
import apiClient from './apiClient';

class AuthService {
    async register(userData) {
        const response = await apiClient.post('/auth/register', userData);
        
        if (response.success) {
            this.storeTokens(response.data.token, response.data.refresh_token);
            return response.data.user;
        }
        throw new Error(response.error);
    }

    async login(email, password) {
        const response = await apiClient.post('/auth/login', { email, password });
        
        if (response.success) {
            this.storeTokens(response.data.token, response.data.refresh_token);
            return response.data.user;
        }
        throw new Error(response.error);
    }

    async refreshToken() {
        const refreshToken = localStorage.getItem('refreshToken');
        if (!refreshToken) {
            throw new Error('No refresh token available');
        }

        const response = await apiClient.post('/auth/refresh', {
            refresh_token: refreshToken
        });

        if (response.success) {
            this.storeTokens(response.data.token, response.data.refresh_token);
            return response.data.token;
        }
        throw new Error('Token refresh failed');
    }

    async logout() {
        const refreshToken = localStorage.getItem('refreshToken');
        
        try {
            await apiClient.post('/auth/logout', { refresh_token: refreshToken });
        } catch (error) {
            console.warn('Logout request failed:', error);
        } finally {
            this.clearTokens();
        }
    }

    storeTokens(token, refreshToken) {
        localStorage.setItem('authToken', token);
        localStorage.setItem('refreshToken', refreshToken);
    }

    clearTokens() {
        localStorage.removeItem('authToken');
        localStorage.removeItem('refreshToken');
    }

    getToken() {
        return localStorage.getItem('authToken');
    }

    isAuthenticated() {
        return !!this.getToken();
    }
}

export default new AuthService();
```

### Service Utilisateur

```javascript
// services/userService.js
import apiClient from './apiClient';

class UserService {
    async getProfile() {
        const response = await apiClient.get('/users/profile');
        return response.data;
    }

    async updateProfile(userData) {
        const response = await apiClient.put('/users/profile', userData);
        return response.data;
    }

    async uploadAvatar(file, onProgress) {
        const response = await apiClient.upload('/users/avatar', file, onProgress);
        return response.data;
    }

    async getUserById(id) {
        const response = await apiClient.get(`/users/${id}`);
        return response.data;
    }

    async searchUsers(query, page = 1, limit = 20) {
        const response = await apiClient.get('/users/search', {
            q: query,
            page,
            limit
        });
        return response.data;
    }
}

export default new UserService();
```

## Hooks React

### Hook d'Authentification

```javascript
// hooks/useAuth.js
import { useState, useEffect, useContext, createContext } from 'react';
import authService from '../services/authService';
import userService from '../services/userService';

const AuthContext = createContext();

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        initializeAuth();
    }, []);

    const initializeAuth = async () => {
        try {
            if (authService.isAuthenticated()) {
                const userData = await userService.getProfile();
                setUser(userData);
            }
        } catch (error) {
            console.error('Auth initialization failed:', error);
            authService.clearTokens();
        } finally {
            setLoading(false);
        }
    };

    const login = async (email, password) => {
        try {
            setError(null);
            const userData = await authService.login(email, password);
            setUser(userData);
            return userData;
        } catch (error) {
            setError(error.message);
            throw error;
        }
    };

    const register = async (userData) => {
        try {
            setError(null);
            const user = await authService.register(userData);
            setUser(user);
            return user;
        } catch (error) {
            setError(error.message);
            throw error;
        }
    };

    const logout = async () => {
        try {
            await authService.logout();
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            setUser(null);
        }
    };

    const updateProfile = async (userData) => {
        try {
            const updatedUser = await userService.updateProfile(userData);
            setUser(updatedUser);
            return updatedUser;
        } catch (error) {
            setError(error.message);
            throw error;
        }
    };

    const value = {
        user,
        loading,
        error,
        login,
        register,
        logout,
        updateProfile,
        isAuthenticated: !!user,
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
};
```

### Hook WebSocket

```javascript
// hooks/useWebSocket.js
import { useEffect, useRef, useState, useCallback } from 'react';
import { useAuth } from './useAuth';
import config from '../config/api';

export const useWebSocket = (url = '/chat') => {
    const { user } = useAuth();
    const [connected, setConnected] = useState(false);
    const [messages, setMessages] = useState([]);
    const [error, setError] = useState(null);
    const ws = useRef(null);
    const reconnectAttempts = useRef(0);
    const maxReconnectAttempts = 5;

    const connect = useCallback(() => {
        if (!user) return;

        const token = localStorage.getItem('authToken');
        const wsUrl = `${config.WS_BASE_URL}${url}?token=${token}`;

        ws.current = new WebSocket(wsUrl);

        ws.current.onopen = () => {
            console.log('WebSocket connected');
            setConnected(true);
            setError(null);
            reconnectAttempts.current = 0;
        };

        ws.current.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                handleMessage(message);
            } catch (error) {
                console.error('Failed to parse WebSocket message:', error);
            }
        };

        ws.current.onclose = (event) => {
            console.log('WebSocket closed:', event.code, event.reason);
            setConnected(false);
            
            if (event.code !== 1000 && reconnectAttempts.current < maxReconnectAttempts) {
                setTimeout(() => {
                    reconnectAttempts.current++;
                    console.log(`Reconnection attempt ${reconnectAttempts.current}`);
                    connect();
                }, 1000 * Math.pow(2, reconnectAttempts.current));
            }
        };

        ws.current.onerror = (error) => {
            console.error('WebSocket error:', error);
            setError('WebSocket connection failed');
        };
    }, [user, url]);

    const handleMessage = (message) => {
        switch (message.type) {
            case 'room_message':
            case 'direct_message':
                setMessages(prev => [...prev, message.data]);
                break;
            case 'message_history':
                setMessages(prev => [...message.data.messages, ...prev]);
                break;
            case 'error':
                setError(message.data.message);
                break;
            default:
                console.log('Unknown message type:', message.type);
        }
    };

    const sendMessage = useCallback((type, data) => {
        if (ws.current && ws.current.readyState === WebSocket.OPEN) {
            const message = {
                type,
                data,
                timestamp: new Date().toISOString()
            };
            ws.current.send(JSON.stringify(message));
        } else {
            console.error('WebSocket not connected');
        }
    }, []);

    const sendRoomMessage = useCallback((room, content) => {
        sendMessage('room_message', {
            room,
            content,
            message_type: 'text'
        });
    }, [sendMessage]);

    const joinRoom = useCallback((room) => {
        sendMessage('join_room', { room });
    }, [sendMessage]);

    useEffect(() => {
        if (user) {
            connect();
        }

        return () => {
            if (ws.current) {
                ws.current.close(1000, 'Component unmounting');
            }
        };
    }, [connect, user]);

    return {
        connected,
        messages,
        error,
        sendMessage,
        sendRoomMessage,
        joinRoom,
    };
};
```

### Hook pour les API

```javascript
// hooks/useApi.js
import { useState, useEffect, useCallback } from 'react';
import apiClient from '../services/apiClient';

export const useApi = (endpoint, options = {}) => {
    const [data, setData] = useState(null);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const {
        immediate = true,
        params = {},
        onSuccess,
        onError,
    } = options;

    const execute = useCallback(async (customParams = {}) => {
        try {
            setLoading(true);
            setError(null);
            
            const response = await apiClient.get(endpoint, { ...params, ...customParams });
            setData(response.data);
            
            if (onSuccess) {
                onSuccess(response.data);
            }
            
            return response.data;
        } catch (err) {
            setError(err.message);
            if (onError) {
                onError(err);
            }
            throw err;
        } finally {
            setLoading(false);
        }
    }, [endpoint, params, onSuccess, onError]);

    useEffect(() => {
        if (immediate) {
            execute();
        }
    }, [execute, immediate]);

    return {
        data,
        loading,
        error,
        refetch: execute,
    };
};

// Hook pour les mutations (POST, PUT, DELETE)
export const useMutation = (mutationFn, options = {}) => {
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);

    const { onSuccess, onError } = options;

    const mutate = useCallback(async (variables) => {
        try {
            setLoading(true);
            setError(null);
            
            const result = await mutationFn(variables);
            
            if (onSuccess) {
                onSuccess(result);
            }
            
            return result;
        } catch (err) {
            setError(err.message);
            if (onError) {
                onError(err);
            }
            throw err;
        } finally {
            setLoading(false);
        }
    }, [mutationFn, onSuccess, onError]);

    return {
        mutate,
        loading,
        error,
    };
};
```

## Composants React

### Composant d'Authentification

```javascript
// components/LoginForm.jsx
import React, { useState } from 'react';
import { useAuth } from '../hooks/useAuth';
import { ValidationError } from '../services/apiClient';

const LoginForm = ({ onSuccess }) => {
    const { login, loading, error } = useAuth();
    const [formData, setFormData] = useState({
        email: '',
        password: '',
    });
    const [validationErrors, setValidationErrors] = useState({});

    const handleSubmit = async (e) => {
        e.preventDefault();
        setValidationErrors({});

        try {
            await login(formData.email, formData.password);
            if (onSuccess) {
                onSuccess();
            }
        } catch (error) {
            if (error instanceof ValidationError) {
                setValidationErrors(error.errors);
            }
        }
    };

    const handleChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: value
        }));
        
        // Nettoyer l'erreur de validation pour ce champ
        if (validationErrors[name]) {
            setValidationErrors(prev => ({
                ...prev,
                [name]: null
            }));
        }
    };

    return (
        <form onSubmit={handleSubmit} className="login-form">
            <div className="form-group">
                <label htmlFor="email">Email</label>
                <input
                    type="email"
                    id="email"
                    name="email"
                    value={formData.email}
                    onChange={handleChange}
                    required
                    disabled={loading}
                    className={validationErrors.email ? 'error' : ''}
                />
                {validationErrors.email && (
                    <span className="error-message">{validationErrors.email}</span>
                )}
            </div>

            <div className="form-group">
                <label htmlFor="password">Mot de passe</label>
                <input
                    type="password"
                    id="password"
                    name="password"
                    value={formData.password}
                    onChange={handleChange}
                    required
                    disabled={loading}
                    className={validationErrors.password ? 'error' : ''}
                />
                {validationErrors.password && (
                    <span className="error-message">{validationErrors.password}</span>
                )}
            </div>

            {error && (
                <div className="error-message global-error">
                    {error}
                </div>
            )}

            <button type="submit" disabled={loading} className="submit-btn">
                {loading ? 'Connexion...' : 'Se connecter'}
            </button>
        </form>
    );
};

export default LoginForm;
```

### Composant de Chat

```javascript
// components/ChatRoom.jsx
import React, { useState, useEffect, useRef } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';
import { useAuth } from '../hooks/useAuth';

const ChatRoom = ({ room = 'general' }) => {
    const { user } = useAuth();
    const { messages, connected, sendRoomMessage, joinRoom } = useWebSocket();
    const [messageText, setMessageText] = useState('');
    const messagesEndRef = useRef(null);

    useEffect(() => {
        if (connected) {
            joinRoom(room);
        }
    }, [connected, room, joinRoom]);

    useEffect(() => {
        scrollToBottom();
    }, [messages]);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    };

    const handleSendMessage = (e) => {
        e.preventDefault();
        if (messageText.trim() && connected) {
            sendRoomMessage(room, messageText.trim());
            setMessageText('');
        }
    };

    const formatTime = (timestamp) => {
        return new Date(timestamp).toLocaleTimeString('fr-FR', {
            hour: '2-digit',
            minute: '2-digit'
        });
    };

    const roomMessages = messages.filter(msg => msg.room === room);

    return (
        <div className="chat-room">
            <div className="chat-header">
                <h3>#{room}</h3>
                <span className={`connection-status ${connected ? 'connected' : 'disconnected'}`}>
                    {connected ? 'Connecté' : 'Déconnecté'}
                </span>
            </div>

            <div className="messages-container">
                {roomMessages.map((message) => (
                    <div
                        key={message.id}
                        className={`message ${message.user_id === user?.id ? 'own-message' : ''}`}
                    >
                        <div className="message-header">
                            <span className="username">{message.username}</span>
                            <span className="timestamp">
                                {formatTime(message.created_at)}
                            </span>
                        </div>
                        <div className="message-content">
                            {message.content}
                        </div>
                    </div>
                ))}
                <div ref={messagesEndRef} />
            </div>

            <form onSubmit={handleSendMessage} className="message-form">
                <input
                    type="text"
                    value={messageText}
                    onChange={(e) => setMessageText(e.target.value)}
                    placeholder={connected ? "Tapez votre message..." : "Connexion..."}
                    disabled={!connected}
                    maxLength={500}
                />
                <button type="submit" disabled={!connected || !messageText.trim()}>
                    Envoyer
                </button>
            </form>
        </div>
    );
};

export default ChatRoom;
```

## Gestion des Fichiers

### Composant d'Upload

```javascript
// components/FileUpload.jsx
import React, { useState, useRef } from 'react';
import apiClient from '../services/apiClient';

const FileUpload = ({ endpoint, accept, maxSize, onSuccess, onError }) => {
    const [uploading, setUploading] = useState(false);
    const [progress, setProgress] = useState(0);
    const fileInputRef = useRef(null);

    const handleFileSelect = async (e) => {
        const file = e.target.files[0];
        if (!file) return;

        // Validation de la taille
        if (maxSize && file.size > maxSize) {
            onError?.(`Le fichier est trop volumineux. Taille maximum: ${formatFileSize(maxSize)}`);
            return;
        }

        try {
            setUploading(true);
            setProgress(0);

            const result = await apiClient.upload(endpoint, file, (progress) => {
                setProgress(progress);
            });

            onSuccess?.(result);
        } catch (error) {
            onError?.(error.message);
        } finally {
            setUploading(false);
            setProgress(0);
            if (fileInputRef.current) {
                fileInputRef.current.value = '';
            }
        }
    };

    const formatFileSize = (bytes) => {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    return (
        <div className="file-upload">
            <input
                ref={fileInputRef}
                type="file"
                accept={accept}
                onChange={handleFileSelect}
                disabled={uploading}
                style={{ display: 'none' }}
            />
            
            <button
                onClick={() => fileInputRef.current?.click()}
                disabled={uploading}
                className="upload-btn"
            >
                {uploading ? 'Upload en cours...' : 'Choisir un fichier'}
            </button>

            {uploading && (
                <div className="upload-progress">
                    <div className="progress-bar">
                        <div
                            className="progress-fill"
                            style={{ width: `${progress}%` }}
                        />
                    </div>
                    <span className="progress-text">{Math.round(progress)}%</span>
                </div>
            )}
        </div>
    );
};

export default FileUpload;
```

## Routing et Navigation

### Routes Protégées

```javascript
// components/ProtectedRoute.jsx
import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

const ProtectedRoute = ({ children, roles = [] }) => {
    const { user, loading } = useAuth();
    const location = useLocation();

    if (loading) {
        return <div className="loading-spinner">Chargement...</div>;
    }

    if (!user) {
        return <Navigate to="/login" state={{ from: location }} replace />;
    }

    if (roles.length > 0 && !roles.includes(user.role)) {
        return <Navigate to="/unauthorized" replace />;
    }

    return children;
};

export default ProtectedRoute;
```

### Configuration des Routes

```javascript
// App.jsx
import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './hooks/useAuth';
import ProtectedRoute from './components/ProtectedRoute';
import LoginForm from './components/LoginForm';
import ChatRoom from './components/ChatRoom';
import Dashboard from './components/Dashboard';
import AdminPanel from './components/AdminPanel';

function App() {
    return (
        <BrowserRouter>
            <AuthProvider>
                <div className="app">
                    <Routes>
                        <Route path="/login" element={<LoginForm />} />
                        
                        <Route path="/dashboard" element={
                            <ProtectedRoute>
                                <Dashboard />
                            </ProtectedRoute>
                        } />
                        
                        <Route path="/chat" element={
                            <ProtectedRoute>
                                <ChatRoom />
                            </ProtectedRoute>
                        } />
                        
                        <Route path="/admin" element={
                            <ProtectedRoute roles={['admin', 'super_admin']}>
                                <AdminPanel />
                            </ProtectedRoute>
                        } />
                        
                        <Route path="/" element={<Navigate to="/dashboard" />} />
                        <Route path="*" element={<div>Page non trouvée</div>} />
                    </Routes>
                </div>
            </AuthProvider>
        </BrowserRouter>
    );
}

export default App;
```

## Gestion des Erreurs

### Boundary d'Erreur

```javascript
// components/ErrorBoundary.jsx
import React from 'react';

class ErrorBoundary extends React.Component {
    constructor(props) {
        super(props);
        this.state = { hasError: false, error: null };
    }

    static getDerivedStateFromError(error) {
        return { hasError: true, error };
    }

    componentDidCatch(error, errorInfo) {
        console.error('Error caught by boundary:', error, errorInfo);
        
        // Envoyer l'erreur à un service de monitoring
        if (process.env.NODE_ENV === 'production') {
            // reportError(error, errorInfo);
        }
    }

    render() {
        if (this.state.hasError) {
            return (
                <div className="error-boundary">
                    <h2>Oops! Une erreur s'est produite</h2>
                    <p>Nous nous excusons pour ce problème technique.</p>
                    <button onClick={() => window.location.reload()}>
                        Recharger la page
                    </button>
                </div>
            );
        }

        return this.props.children;
    }
}

export default ErrorBoundary;
```

## Build et Déploiement

### Configuration Build

```json
// package.json
{
    "name": "veza-frontend",
    "version": "1.0.0",
    "scripts": {
        "start": "react-scripts start",
        "build": "react-scripts build",
        "build:prod": "REACT_APP_ENV=production npm run build",
        "test": "react-scripts test",
        "eject": "react-scripts eject"
    },
    "dependencies": {
        "react": "^18.2.0",
        "react-dom": "^18.2.0",
        "react-router-dom": "^6.8.0",
        "react-scripts": "5.0.1"
    }
}
```

### Dockerfile

```dockerfile
# Frontend Dockerfile
FROM node:18-alpine as builder

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Configuration Nginx

```nginx
# nginx.conf
server {
    listen 80;
    server_name localhost;
    
    location / {
        root /usr/share/nginx/html;
        index index.html index.htm;
        try_files $uri $uri/ /index.html;
    }
    
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
    
    location /ws/ {
        proxy_pass http://backend:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
    }
}
```

Cette documentation fournit tout ce qui est nécessaire pour intégrer parfaitement un frontend React avec votre backend Go API, incluant l'authentification, les WebSockets, et les bonnes pratiques de développement moderne. 