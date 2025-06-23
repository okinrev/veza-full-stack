# API Endpoints - Documentation Complète

Cette documentation détaille tous les endpoints pour l'intégration avec le backend Go.

## 🔐 Authentification

### POST /auth/login
**Connexion utilisateur avec JWT**

```typescript
interface LoginRequest {
  email: string;
  password: string;
}

interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}
```

**Backend Go Handler requis** :
```go
func (h *AuthHandler) Login(c *gin.Context) {
  var req LoginRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, ApiResponse{Error: "Invalid request"})
    return
  }
  
  user, err := h.authService.Authenticate(req.Email, req.Password)
  if err != nil {
    c.JSON(401, ApiResponse{Error: "Invalid credentials"})
    return
  }
  
  tokens, err := h.authService.GenerateTokens(user.ID)
  c.JSON(200, ApiResponse{Success: true, Data: tokens})
}
```

### POST /auth/register
**Inscription utilisateur**

```typescript
interface RegisterRequest {
  username: string;
  email: string;
  password: string;
  first_name?: string;
  last_name?: string;
}
```

## 💬 Chat - Endpoints REST

### GET /api/v1/chat/conversations
**Liste des conversations privées**

```typescript
interface Conversation {
  user_id: number;
  username: string;
  first_name: string;
  last_name: string;
  avatar: string;
  last_message: string;
  last_activity: string;
  unread_count: number;
  is_online: boolean;
  last_seen: string;
}
```

### GET /api/v1/chat/rooms/{roomName}/messages
**Messages d'une salle**

```typescript
interface ChatMessage {
  id: number;
  from_user: number;
  content: string;
  timestamp: string;
  username: string;
  avatar?: string;
}
```

## 📦 Produits

### GET /products
**Liste avec filtres**

```typescript
interface ProductFilters {
  page?: number;
  limit?: number;
  search?: string;
  category?: string;
  price_min?: number;
  price_max?: number;
}
```

### POST /products
**Création de produit**

```typescript
interface CreateProductRequest {
  name: string;
  description?: string;
  price: number;
  category?: string;
}
```

## 🎵 Pistes Audio

### POST /tracks/upload
**Upload audio avec métadonnées**

```typescript
interface UploadTrackData {
  title: string;
  artist: string;
  genre?: string;
  is_public: boolean;
  tags?: string[];
}
```

### GET /tracks/{id}/stream
**URL de streaming pour le module Rust**

```typescript
interface StreamResponse {
  url: string;
  websocket_url: string; // Pour le module stream Rust
}
```

Cette structure permet l'intégration complète avec les trois composants backend. 