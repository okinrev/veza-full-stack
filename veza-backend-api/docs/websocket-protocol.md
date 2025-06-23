# Documentation - Protocole WebSocket

## Vue d'ensemble

Le protocole WebSocket permet la communication en temps réel entre le frontend et le backend pour les fonctionnalités de chat, notifications, et mises à jour en direct. Il utilise l'authentification JWT et supporte plusieurs types de messages.

## Connexion WebSocket

### Endpoint de Connexion

```
ws://localhost:8080/ws/chat?token=<jwt_token>
```

### Authentification

La connexion WebSocket nécessite un token JWT valide passé en paramètre de query :

```javascript
const token = localStorage.getItem('authToken');
const ws = new WebSocket(`ws://localhost:8080/ws/chat?token=${token}`);
```

### Headers d'Authentification (Alternative)

```javascript
// Le token peut aussi être passé dans les headers
const ws = new WebSocket('ws://localhost:8080/ws/chat', [], {
    headers: {
        'Authorization': `Bearer ${token}`
    }
});
```

## Types de Messages

### Structure Générale

Tous les messages WebSocket suivent cette structure JSON :

```json
{
    "type": "message_type",
    "data": {
        // Données spécifiques au type de message
    },
    "timestamp": "2023-11-20T10:30:00Z",
    "user_id": 123
}
```

### Messages Entrants (Client → Serveur)

#### 1. Rejoindre un Salon

**Type** : `join_room`

```json
{
    "type": "join_room",
    "data": {
        "room": "general"
    }
}
```

**Réponse du serveur** :
```json
{
    "type": "join_room_ack",
    "data": {
        "room": "general",
        "success": true,
        "message": "Joined room successfully"
    },
    "timestamp": "2023-11-20T10:30:00Z"
}
```

#### 2. Message de Salon

**Type** : `room_message`

```json
{
    "type": "room_message",
    "data": {
        "room": "general",
        "content": "Bonjour tout le monde !",
        "message_type": "text"
    }
}
```

#### 3. Message Privé

**Type** : `direct_message`

```json
{
    "type": "direct_message",
    "data": {
        "to_user_id": 456,
        "content": "Salut, comment ça va ?",
        "message_type": "text"
    }
}
```

#### 4. Demande d'Historique de Salon

**Type** : `room_history`

```json
{
    "type": "room_history",
    "data": {
        "room": "general",
        "limit": 50,
        "offset": 0
    }
}
```

#### 5. Demande d'Historique de Messages Privés

**Type** : `dm_history`

```json
{
    "type": "dm_history",
    "data": {
        "with_user": 456,
        "limit": 50,
        "offset": 0
    }
}
```

#### 6. Ping

**Type** : `ping`

```json
{
    "type": "ping",
    "data": {}
}
```

### Messages Sortants (Serveur → Client)

#### 1. Nouveau Message de Salon

**Type** : `room_message`

```json
{
    "type": "room_message",
    "data": {
        "id": 123,
        "room": "general",
        "user_id": 789,
        "username": "alice",
        "content": "Bonjour tout le monde !",
        "message_type": "text",
        "created_at": "2023-11-20T10:30:00Z"
    },
    "timestamp": "2023-11-20T10:30:00Z"
}
```

#### 2. Nouveau Message Privé

**Type** : `direct_message`

```json
{
    "type": "direct_message",
    "data": {
        "id": 124,
        "from_user_id": 789,
        "from_username": "alice",
        "to_user_id": 456,
        "content": "Salut, comment ça va ?",
        "message_type": "text",
        "created_at": "2023-11-20T10:30:00Z"
    },
    "timestamp": "2023-11-20T10:30:00Z"
}
```

#### 3. Historique de Messages

**Type** : `message_history`

```json
{
    "type": "message_history",
    "data": {
        "room": "general",
        "messages": [
            {
                "id": 120,
                "user_id": 789,
                "username": "alice",
                "content": "Message plus ancien",
                "message_type": "text",
                "created_at": "2023-11-20T09:00:00Z"
            },
            {
                "id": 121,
                "user_id": 456,
                "username": "bob",
                "content": "Réponse au message",
                "message_type": "text",
                "created_at": "2023-11-20T09:15:00Z"
            }
        ],
        "total": 2,
        "has_more": false
    },
    "timestamp": "2023-11-20T10:30:00Z"
}
```

#### 4. Utilisateur Connecté

**Type** : `user_connected`

```json
{
    "type": "user_connected",
    "data": {
        "user_id": 789,
        "username": "alice",
        "room": "general"
    },
    "timestamp": "2023-11-20T10:30:00Z"
}
```

#### 5. Utilisateur Déconnecté

**Type** : `user_disconnected`

```json
{
    "type": "user_disconnected",
    "data": {
        "user_id": 789,
        "username": "alice",
        "room": "general"
    },
    "timestamp": "2023-11-20T10:30:00Z"
}
```

#### 6. Pong

**Type** : `pong`

```json
{
    "type": "pong",
    "data": {
        "timestamp": "2023-11-20T10:30:00Z"
    },
    "timestamp": "2023-11-20T10:30:00Z"
}
```

#### 7. Erreur

**Type** : `error`

```json
{
    "type": "error",
    "data": {
        "message": "Room not found",
        "code": "ROOM_NOT_FOUND",
        "details": {
            "room": "nonexistent_room"
        }
    },
    "timestamp": "2023-11-20T10:30:00Z"
}
```

## Types de Messages Supportés

### Messages Texte

```json
{
    "message_type": "text",
    "content": "Bonjour tout le monde !"
}
```

### Messages avec Média

```json
{
    "message_type": "image",
    "content": "Regardez cette photo !",
    "media_url": "/uploads/images/photo123.jpg",
    "media_type": "image/jpeg"
}
```

### Messages Audio

```json
{
    "message_type": "audio",
    "content": "Message vocal",
    "media_url": "/uploads/audio/voice123.mp3",
    "duration": 15.5
}
```

### Messages de Réaction

```json
{
    "message_type": "reaction",
    "content": "👍",
    "reply_to": 123
}
```

## Gestion des Erreurs

### Codes d'Erreur

| Code | Description | Action |
|------|-------------|--------|
| `AUTH_FAILED` | Authentification échouée | Reconnecter avec nouveau token |
| `ROOM_NOT_FOUND` | Salon inexistant | Vérifier le nom du salon |
| `USER_NOT_FOUND` | Utilisateur inexistant | Vérifier l'ID utilisateur |
| `PERMISSION_DENIED` | Permissions insuffisantes | Vérifier les droits d'accès |
| `RATE_LIMITED` | Trop de messages | Attendre avant de renvoyer |
| `MESSAGE_TOO_LONG` | Message trop long | Réduire la taille du message |
| `INVALID_FORMAT` | Format de message invalide | Corriger le format JSON |

### Exemple de Gestion d'Erreur

```javascript
ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    
    if (message.type === 'error') {
        switch (message.data.code) {
            case 'AUTH_FAILED':
                // Reconnecter avec nouveau token
                refreshTokenAndReconnect();
                break;
            case 'ROOM_NOT_FOUND':
                console.error('Salon non trouvé:', message.data.details.room);
                break;
            case 'RATE_LIMITED':
                showRateLimitWarning();
                break;
            default:
                console.error('Erreur WebSocket:', message.data.message);
        }
    }
};
```

## Implémentation Client

### Client JavaScript/React

```javascript
class WebSocketClient {
    constructor(token) {
        this.token = token;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
        this.messageHandlers = new Map();
    }

    connect() {
        this.ws = new WebSocket(`ws://localhost:8080/ws/chat?token=${this.token}`);
        
        this.ws.onopen = () => {
            console.log('WebSocket connecté');
            this.reconnectAttempts = 0;
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };

        this.ws.onclose = (event) => {
            console.log('WebSocket fermé:', event.code, event.reason);
            this.handleReconnect();
        };

        this.ws.onerror = (error) => {
            console.error('Erreur WebSocket:', error);
        };
    }

    handleMessage(message) {
        const handler = this.messageHandlers.get(message.type);
        if (handler) {
            handler(message.data);
        }
    }

    onMessage(type, handler) {
        this.messageHandlers.set(type, handler);
    }

    send(type, data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            const message = {
                type,
                data,
                timestamp: new Date().toISOString()
            };
            this.ws.send(JSON.stringify(message));
        }
    }

    joinRoom(room) {
        this.send('join_room', { room });
    }

    sendRoomMessage(room, content) {
        this.send('room_message', {
            room,
            content,
            message_type: 'text'
        });
    }

    sendDirectMessage(toUserId, content) {
        this.send('direct_message', {
            to_user_id: toUserId,
            content,
            message_type: 'text'
        });
    }

    getRoomHistory(room, limit = 50, offset = 0) {
        this.send('room_history', { room, limit, offset });
    }

    handleReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            setTimeout(() => {
                console.log(`Tentative de reconnexion ${this.reconnectAttempts}`);
                this.connect();
            }, this.reconnectDelay * this.reconnectAttempts);
        }
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
        }
    }
}
```

### Hook React pour WebSocket

```javascript
import { useEffect, useRef, useState } from 'react';

export const useWebSocket = (token) => {
    const [connected, setConnected] = useState(false);
    const [messages, setMessages] = useState([]);
    const ws = useRef(null);

    useEffect(() => {
        if (!token) return;

        const client = new WebSocketClient(token);
        ws.current = client;

        // Gestionnaires de messages
        client.onMessage('room_message', (data) => {
            setMessages(prev => [...prev, data]);
        });

        client.onMessage('direct_message', (data) => {
            setMessages(prev => [...prev, data]);
        });

        client.onMessage('message_history', (data) => {
            setMessages(prev => [...data.messages, ...prev]);
        });

        client.connect();

        return () => {
            client.disconnect();
        };
    }, [token]);

    const sendMessage = (room, content) => {
        if (ws.current) {
            ws.current.sendRoomMessage(room, content);
        }
    };

    const joinRoom = (room) => {
        if (ws.current) {
            ws.current.joinRoom(room);
        }
    };

    return {
        connected,
        messages,
        sendMessage,
        joinRoom,
        client: ws.current
    };
};
```

## Intégration avec les Modules Rust

### Chat Server Rust

Le chat server Rust peut se connecter à la même base de données et utiliser le même protocole :

```rust
// Structure de message compatible
#[derive(Serialize, Deserialize)]
struct WebSocketMessage {
    #[serde(rename = "type")]
    message_type: String,
    data: serde_json::Value,
    timestamp: String,
    user_id: Option<i32>,
}

// Envoi de message via WebSocket
async fn send_websocket_message(
    websocket: &mut WebSocket,
    message_type: &str,
    data: serde_json::Value,
) -> Result<(), Box<dyn std::error::Error>> {
    let message = WebSocketMessage {
        message_type: message_type.to_string(),
        data,
        timestamp: chrono::Utc::now().to_rfc3339(),
        user_id: None,
    };

    let json = serde_json::to_string(&message)?;
    websocket.send(Message::Text(json)).await?;
    Ok(())
}
```

## Performance et Optimisation

### Batching des Messages

Pour améliorer les performances, les messages peuvent être regroupés :

```json
{
    "type": "message_batch",
    "data": {
        "messages": [
            {
                "type": "room_message",
                "data": { /* message 1 */ }
            },
            {
                "type": "room_message", 
                "data": { /* message 2 */ }
            }
        ]
    }
}
```

### Compression

Utiliser la compression WebSocket pour réduire la bande passante :

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/chat', ['permessage-deflate']);
```

### Heartbeat

Maintenir la connexion active avec des pings réguliers :

```javascript
setInterval(() => {
    if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: 'ping', data: {} }));
    }
}, 30000); // Ping toutes les 30 secondes
```

## Sécurité

### Validation des Messages

Tous les messages entrants sont validés côté serveur :
- Taille maximum des messages
- Format JSON valide
- Types de messages autorisés
- Permissions utilisateur

### Rate Limiting

Limitation du nombre de messages par utilisateur :
- 10 messages par minute par utilisateur
- 100 messages par heure par utilisateur

### Authentification Continue

Le token JWT est validé à chaque message critique pour maintenir la sécurité.

## Monitoring

### Métriques à Surveiller

- Nombre de connexions actives
- Messages envoyés/reçus par seconde
- Latence moyenne des messages
- Taux de reconnexions
- Erreurs de connexion

### Logs

```go
log.Printf("WS_CONNECT: UserID=%d, IP=%s", userID, clientIP)
log.Printf("WS_MESSAGE: Type=%s, UserID=%d, Room=%s", msgType, userID, room)
log.Printf("WS_DISCONNECT: UserID=%d, Duration=%v", userID, duration)
```

Cette documentation fournit toutes les informations nécessaires pour implémenter et utiliser le protocole WebSocket dans votre application. 