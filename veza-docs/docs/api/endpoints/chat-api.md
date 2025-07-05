# API Chat - Documentation Complète

## Vue d'ensemble

L'API Chat de Veza fournit un système de messagerie instantanée complet avec support WebSocket pour le temps réel, messages directs, salons publics/privés, et fonctionnalités avancées de chat.

### Architecture Chat

```mermaid
graph TB
    subgraph "Chat Architecture"
        A[Client Apps] --> B[WebSocket Gateway]
        A --> C[REST API]
        
        B --> D[Chat Manager]
        C --> E[Chat Service]
        
        D --> F[Connection Pool]
        D --> G[Message Broadcasting]
        D --> H[Presence System]
        
        E --> I[PostgreSQL]
        E --> J[Redis Cache]
        
        F --> K[User Sessions]
        G --> L[Room Management]
        H --> M[Online Status]
        
        I --> N[Messages Table]
        I --> O[Rooms Table]
        I --> P[Members Table]
        
        J --> Q[Recent Messages]
        J --> R[Active Users]
        J --> S[Typing Indicators]
    end
```

## Endpoints REST

### 1. Messages directs

#### Récupérer conversation

**GET** `/api/v1/chat/dm/{user_id}`

Récupère l'historique des messages directs avec un utilisateur.

**Authentification :** Requis (JWT)

**Paramètres :**
- `user_id` (path) : ID de l'utilisateur cible
- `page` (query) : Numéro de page (défaut: 1)
- `limit` (query) : Messages par page (défaut: 50, max: 100)
- `before` (query) : ID message pour pagination chronologique
- `after` (query) : ID message pour nouveaux messages

**Exemple :**
```bash
curl -X GET "http://localhost:8080/api/v1/chat/dm/123?limit=20&page=1" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Réponse (200) :**
```json
{
  "success": true,
  "message": "Messages retrieved successfully",
  "data": {
    "conversation": {
      "user_id": 123,
      "username": "alice_doe",
      "display_name": "Alice Doe",
      "avatar_url": "https://cdn.veza.app/avatars/123.jpg",
      "is_online": true,
      "last_seen": "2024-01-20T16:45:00Z"
    },
    "messages": [
      {
        "id": 1001,
        "from_user": 456,
        "to_user": 123,
        "content": "Hey! How's your new track coming along?",
        "is_read": true,
        "timestamp": "2024-01-20T15:30:00Z",
        "created_at": "2024-01-20T15:30:00Z",
        "edited_at": null,
        "from_username": "bob_producer",
        "from_avatar": "https://cdn.veza.app/avatars/456.jpg"
      },
      {
        "id": 1002,
        "from_user": 123,
        "to_user": 456,
        "content": "Going great! Almost finished with the mixing. Want to hear a preview?",
        "is_read": true,
        "timestamp": "2024-01-20T15:32:00Z",
        "created_at": "2024-01-20T15:32:00Z",
        "edited_at": null,
        "from_username": "alice_doe",
        "from_avatar": "https://cdn.veza.app/avatars/123.jpg"
      }
    ],
    "meta": {
      "page": 1,
      "per_page": 20,
      "total": 45,
      "total_pages": 3,
      "has_next": true,
      "has_prev": false,
      "unread_count": 2
    }
  }
}
```

#### Envoyer message direct

**POST** `/api/v1/chat/dm/{user_id}`

Envoie un message direct à un utilisateur.

**Authentification :** Requis (JWT)

**Corps de requête :**
```json
{
  "content": "Message content here",
  "reply_to": 1001,
  "attachments": [
    {
      "type": "image",
      "url": "https://cdn.veza.app/attachments/abc123.jpg",
      "filename": "preview.jpg"
    }
  ]
}
```

**Exemple :**
```bash
curl -X POST "http://localhost:8080/api/v1/chat/dm/123" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Check out this new track preview!",
    "attachments": [
      {
        "type": "audio",
        "url": "https://cdn.veza.app/previews/track123.mp3",
        "filename": "preview.mp3"
      }
    ]
  }'
```

**Réponse (201) :**
```json
{
  "success": true,
  "message": "Message sent successfully",
  "data": {
    "id": 1003,
    "from_user": 456,
    "to_user": 123,
    "content": "Check out this new track preview!",
    "is_read": false,
    "timestamp": "2024-01-20T16:45:00Z",
    "created_at": "2024-01-20T16:45:00Z",
    "attachments": [
      {
        "type": "audio",
        "url": "https://cdn.veza.app/previews/track123.mp3",
        "filename": "preview.mp3",
        "duration": 30
      }
    ]
  }
}
```

### 2. Conversations

#### Liste des conversations

**GET** `/api/v1/chat/conversations`

Récupère la liste des conversations actives de l'utilisateur.

**Authentification :** Requis (JWT)

**Paramètres :**
- `limit` (query) : Nombre de conversations (défaut: 20, max: 50)
- `search` (query) : Recherche par nom d'utilisateur
- `unread_only` (query) : Afficher uniquement non lues

**Exemple :**
```bash
curl -X GET "http://localhost:8080/api/v1/chat/conversations?limit=10&unread_only=false" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Réponse (200) :**
```json
{
  "success": true,
  "message": "Conversations retrieved successfully",
  "data": [
    {
      "user_id": 123,
      "username": "alice_doe",
      "display_name": "Alice Doe",
      "avatar_url": "https://cdn.veza.app/avatars/123.jpg",
      "last_message": "Going great! Almost finished with the mixing.",
      "last_activity": "2024-01-20T15:32:00Z",
      "unread_count": 2,
      "is_online": true,
      "last_seen": "2024-01-20T16:45:00Z"
    },
    {
      "user_id": 789,
      "username": "charlie_beats",
      "display_name": "Charlie Beats",
      "avatar_url": "https://cdn.veza.app/avatars/789.jpg",
      "last_message": "Thanks for the feedback on the track!",
      "last_activity": "2024-01-20T14:20:00Z",
      "unread_count": 0,
      "is_online": false,
      "last_seen": "2024-01-20T15:00:00Z"
    }
  ]
}
```

### 3. Salons (Rooms)

#### Liste des salons publics

**GET** `/api/v1/chat/rooms`

Récupère la liste des salons publics disponibles.

**Authentification :** Requis (JWT)

**Paramètres :**
- `page` (query) : Numéro de page
- `limit` (query) : Salons par page
- `category` (query) : Catégorie de salon
- `search` (query) : Recherche par nom

**Exemple :**
```bash
curl -X GET "http://localhost:8080/api/v1/chat/rooms?category=music&limit=10" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Réponse (200) :**
```json
{
  "success": true,
  "message": "Rooms retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "general",
      "description": "General discussion for all topics",
      "is_private": false,
      "creator_id": 1,
      "creator_name": "admin",
      "member_count": 156,
      "online_count": 23,
      "last_activity": "2024-01-20T16:40:00Z",
      "last_message": "Welcome everyone to the general chat!",
      "created_at": "2024-01-01T00:00:00Z",
      "is_member": true,
      "user_role": "member"
    },
    {
      "id": 2,
      "name": "electronic-music",
      "description": "Discussion about electronic music production",
      "is_private": false,
      "creator_id": 5,
      "creator_name": "dj_master",
      "member_count": 89,
      "online_count": 12,
      "last_activity": "2024-01-20T16:35:00Z",
      "last_message": "Anyone tried the new synthesizer plugin?",
      "created_at": "2024-01-05T10:00:00Z",
      "is_member": false,
      "user_role": null
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 25,
    "total_pages": 3
  }
}
```

#### Créer un salon

**POST** `/api/v1/chat/rooms`

Crée un nouveau salon de chat.

**Authentification :** Requis (JWT)

**Corps de requête :**
```json
{
  "name": "my-music-room",
  "description": "Room for discussing my music projects",
  "is_private": false,
  "password": "optional_password",
  "category": "music",
  "max_members": 100
}
```

**Validation :**
- `name` : 3-50 caractères, alphanumériques et tirets
- `description` : Maximum 500 caractères
- `password` : Si salon privé, minimum 8 caractères
- `max_members` : 2-1000 membres

**Exemple :**
```bash
curl -X POST "http://localhost:8080/api/v1/chat/rooms" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ambient-producers",
    "description": "Room for ambient music producers to share ideas",
    "is_private": false,
    "category": "music",
    "max_members": 50
  }'
```

**Réponse (201) :**
```json
{
  "success": true,
  "message": "Room created successfully",
  "data": {
    "id": 26,
    "name": "ambient-producers",
    "description": "Room for ambient music producers to share ideas",
    "is_private": false,
    "creator_id": 456,
    "member_count": 1,
    "online_count": 1,
    "created_at": "2024-01-20T16:50:00Z",
    "user_role": "owner",
    "invite_link": "https://veza.app/rooms/26/join?token=abc123"
  }
}
```

#### Messages d'un salon

**GET** `/api/v1/chat/rooms/{room_id}/messages`

Récupère les messages d'un salon spécifique.

**Authentification :** Requis (JWT)

**Paramètres :**
- `room_id` (path) : ID du salon
- `page` (query) : Numéro de page
- `limit` (query) : Messages par page
- `since` (query) : Timestamp pour nouveaux messages

**Exemple :**
```bash
curl -X GET "http://localhost:8080/api/v1/chat/rooms/1/messages?limit=20" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Réponse (200) :**
```json
{
  "success": true,
  "message": "Room messages retrieved successfully",
  "data": {
    "room": {
      "id": 1,
      "name": "general",
      "description": "General discussion",
      "member_count": 156
    },
    "messages": [
      {
        "id": 2001,
        "room_id": 1,
        "user_id": 123,
        "content": "Hey everyone! Just released a new ambient track",
        "message_type": "message",
        "created_at": "2024-01-20T16:30:00Z",
        "edited_at": null,
        "username": "alice_doe",
        "avatar": "https://cdn.veza.app/avatars/123.jpg"
      },
      {
        "id": 2002,
        "room_id": 1,
        "user_id": null,
        "content": "bob_producer joined the room",
        "message_type": "join",
        "created_at": "2024-01-20T16:35:00Z",
        "username": null,
        "avatar": null
      }
    ],
    "meta": {
      "page": 1,
      "per_page": 20,
      "total": 1247,
      "has_next": true
    }
  }
}
```

#### Envoyer message dans salon

**POST** `/api/v1/chat/rooms/{room_id}/messages`

Envoie un message dans un salon.

**Authentification :** Requis (JWT)

**Corps de requête :**
```json
{
  "content": "Check out this new track I'm working on!",
  "message_type": "message",
  "reply_to": 2001,
  "attachments": []
}
```

### 4. Messages non lus

**GET** `/api/v1/chat/unread`

Récupère le nombre de messages non lus par conversation.

**Authentification :** Requis (JWT)

**Réponse (200) :**
```json
{
  "success": true,
  "message": "Unread messages retrieved successfully",
  "data": {
    "total_unread": 5,
    "direct_messages": {
      "123": 2,
      "789": 1
    },
    "rooms": {
      "1": 1,
      "2": 1
    }
  }
}
```

## WebSocket API

### Connection

**URL :** `ws://localhost:8080/ws/chat`

**Authentification :** Token JWT en query parameter

**Exemple de connexion :**
```javascript
const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...";
const ws = new WebSocket(`ws://localhost:8080/ws/chat?token=${token}`);

ws.onopen = function() {
    console.log("Connected to chat WebSocket");
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    handleMessage(message);
};
```

### Messages WebSocket

#### 1. Message direct

**Envoi :**
```json
{
  "type": "direct_message",
  "data": {
    "to_user": 123,
    "content": "Hello from WebSocket!",
    "attachments": []
  }
}
```

**Réception :**
```json
{
  "type": "direct_message",
  "data": {
    "id": 1004,
    "from_user": 456,
    "to_user": 123,
    "content": "Hello from WebSocket!",
    "timestamp": "2024-01-20T17:00:00Z",
    "from_username": "bob_producer",
    "from_avatar": "https://cdn.veza.app/avatars/456.jpg"
  }
}
```

#### 2. Message salon

**Envoi :**
```json
{
  "type": "room_message",
  "data": {
    "room_id": 1,
    "content": "Hello room!",
    "message_type": "message"
  }
}
```

**Réception :**
```json
{
  "type": "room_message",
  "data": {
    "id": 2003,
    "room_id": 1,
    "user_id": 456,
    "content": "Hello room!",
    "message_type": "message",
    "created_at": "2024-01-20T17:05:00Z",
    "username": "bob_producer",
    "avatar": "https://cdn.veza.app/avatars/456.jpg"
  }
}
```

#### 3. Indicateur de frappe

**Envoi :**
```json
{
  "type": "typing_start",
  "data": {
    "target_type": "user",
    "target_id": 123
  }
}
```

**Réception :**
```json
{
  "type": "typing_indicator",
  "data": {
    "user_id": 456,
    "username": "bob_producer",
    "target_type": "user",
    "target_id": 123,
    "is_typing": true
  }
}
```

#### 4. Statut de présence

**Réception :**
```json
{
  "type": "presence_update",
  "data": {
    "user_id": 123,
    "username": "alice_doe",
    "status": "online",
    "last_seen": "2024-01-20T17:10:00Z"
  }
}
```

#### 5. Jointure/départ salon

**Réception :**
```json
{
  "type": "room_join",
  "data": {
    "room_id": 1,
    "user_id": 789,
    "username": "charlie_beats",
    "avatar": "https://cdn.veza.app/avatars/789.jpg",
    "timestamp": "2024-01-20T17:15:00Z"
  }
}
```

#### 6. Accusé de réception

**Envoi :**
```json
{
  "type": "message_read",
  "data": {
    "message_id": 1004,
    "conversation_type": "direct"
  }
}
```

**Réception :**
```json
{
  "type": "message_read_receipt",
  "data": {
    "message_id": 1004,
    "read_by": 123,
    "read_at": "2024-01-20T17:20:00Z"
  }
}
```

## Modèles de données

### Message
```typescript
interface Message {
  id: number;
  from_user: number;
  to_user?: number;
  room?: string;
  content: string;
  is_read: boolean;
  edited_at?: string;
  timestamp: string;
  created_at: string;
  updated_at: string;
}
```

### Room
```typescript
interface Room {
  id: number;
  name: string;
  description?: string;
  is_private: boolean;
  creator_id?: number;
  created_at: string;
  updated_at: string;
}
```

### WebSocketMessage
```typescript
interface WebSocketMessage {
  type: 'direct_message' | 'room_message' | 'typing_start' | 'typing_stop' | 'presence_update' | 'room_join' | 'room_leave' | 'message_read';
  data: any;
  timestamp?: string;
}
```

## Rate Limiting

| Action | Limite | Fenêtre |
|--------|--------|---------|
| Envoyer message | 30 | 1 minute |
| Créer salon | 5 | 1 heure |
| Joindre salon | 20 | 1 minute |
| WebSocket messages | 60 | 1 minute |

## Codes d'erreur

| Code | Message | Description |
|------|---------|-------------|
| `CHAT_USER_NOT_FOUND` | 404 | Utilisateur cible introuvable |
| `CHAT_ROOM_NOT_FOUND` | 404 | Salon introuvable |
| `CHAT_ACCESS_DENIED` | 403 | Accès au salon interdit |
| `CHAT_MESSAGE_TOO_LONG` | 400 | Message trop long (>2000 chars) |
| `CHAT_RATE_LIMITED` | 429 | Trop de messages envoyés |
| `CHAT_ROOM_FULL` | 423 | Salon plein |
| `CHAT_BLOCKED_USER` | 403 | Utilisateur bloqué |
| `WEBSOCKET_AUTH_FAILED` | 401 | Authentification WebSocket échouée |

## Sécurité

### Protection anti-spam
- Rate limiting par utilisateur
- Détection de contenu répétitif
- Filtrage automatique de contenu
- Système de signalement

### Modération
- Mots-clés interdits configurables
- Système de permissions par salon
- Modérateurs et administrateurs
- Logs d'audit complets

### Privacy
- Messages directs chiffrés en transit
- Option suppression automatique
- Contrôle de visibilité
- Blocage d'utilisateurs

## Performance

### Optimisations WebSocket
- Connection pooling efficace
- Message batching pour broadcast
- Compression des gros messages
- Reconnexion automatique côté client

### Cache strategy
- Messages récents en Redis (1h)
- Statuts de présence en mémoire
- Métadonnées salons cachées (5min)
- Indicateurs typing éphémères

### Métriques
- Connexions WebSocket actives
- Messages par seconde
- Latence moyenne des messages
- Taux d'erreur WebSocket 