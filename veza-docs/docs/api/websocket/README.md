# 🔌 WebSocket API - Veza Platform

> **API WebSocket pour la communication en temps réel sur la plateforme Veza**

## 📋 Table des Matières

- [Vue d'ensemble](#vue-densemble)
- [Connexion WebSocket](#connexion-websocket)
- [Authentification](#authentification)
- [Événements](#vnements)
- [Salons de Chat](#salons-de-chat)
- [Messages Privés](#messages-privs)
- [Modération](#modration)
- [Gestion des Erreurs](#gestion-des-erreurs)
- [Exemples d'Implémentation](#exemples-dimplmentation)

## 🎯 Vue d'ensemble

L'API WebSocket de Veza permet une communication bidirectionnelle en temps réel entre les clients et le serveur de chat. Elle supporte les fonctionnalités de chat en temps réel, la modération automatique, et la gestion des salons.

### 🌟 Fonctionnalités Principales

- **Chat en Temps Réel** : Messages instantanés
- **Salons Privés** : Création et gestion de salons
- **Messages Privés** : Communication directe entre utilisateurs
- **Modération Automatique** : Filtrage de contenu et détection de spam
- **Typing Indicators** : Indicateurs de frappe
- **Read Receipts** : Accusés de réception
- **File Sharing** : Partage de fichiers
- **Voice Messages** : Messages vocaux

## 🔌 Connexion WebSocket

### URL de Connexion

```javascript
// Développement
const wsUrl = 'ws://localhost:3001/ws';

// Production
const wsUrl = 'wss://chat.veza.com/ws';
```

### Établissement de Connexion

```javascript
class VezaWebSocket {
  constructor(token) {
    this.token = token;
    this.ws = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectDelay = 1000;
  }

  connect() {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(`wss://chat.veza.com/ws?token=${this.token}`);
      
      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        resolve();
      };
      
      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        reject(error);
      };
      
      this.ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        this.handleReconnect();
      };
      
      this.ws.onmessage = (event) => {
        this.handleMessage(JSON.parse(event.data));
      };
    });
  }

  handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
      
      setTimeout(() => {
        console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
        this.connect();
      }, delay);
    } else {
      console.error('Max reconnection attempts reached');
    }
  }
}
```

### Structure des Messages

Tous les messages WebSocket suivent une structure JSON standard :

```json
{
  "type": "message_type",
  "id": "unique_message_id",
  "timestamp": 1640995200000,
  "data": {
    // Données spécifiques au type de message
  },
  "metadata": {
    "user_id": "user_id",
    "session_id": "session_id",
    "version": "1.0.0"
  }
}
```

## 🔐 Authentification

### Authentification par Token

```javascript
// Connexion avec token JWT
const ws = new WebSocket('wss://chat.veza.com/ws?token=your-jwt-token');

// Ou via header d'authentification
const ws = new WebSocket('wss://chat.veza.com/ws');
ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'auth',
    data: {
      token: 'your-jwt-token'
    }
  }));
};
```

### Événements d'Authentification

#### Authentification Réussie

```json
{
  "type": "auth_success",
  "id": "auth_123",
  "timestamp": 1640995200000,
  "data": {
    "user_id": "user_123",
    "username": "john_doe",
    "roles": ["user"],
    "permissions": ["read:chat", "write:chat"],
    "session_id": "session_456"
  }
}
```

#### Authentification Échouée

```json
{
  "type": "auth_error",
  "id": "auth_124",
  "timestamp": 1640995200000,
  "data": {
    "error": "invalid_token",
    "message": "Token invalide ou expiré",
    "code": 401
  }
}
```

## 📨 Événements

### Types d'Événements Supportés

| Type | Description | Direction |
|------|-------------|-----------|
| `auth` | Authentification | Client → Serveur |
| `auth_success` | Authentification réussie | Serveur → Client |
| `auth_error` | Erreur d'authentification | Serveur → Client |
| `join_room` | Rejoindre un salon | Client → Serveur |
| `leave_room` | Quitter un salon | Client → Serveur |
| `message` | Envoyer un message | Client → Serveur |
| `message_received` | Message reçu | Serveur → Client |
| `typing_start` | Commencer à taper | Client → Serveur |
| `typing_stop` | Arrêter de taper | Client → Serveur |
| `typing_indicator` | Indicateur de frappe | Serveur → Client |
| `read_receipt` | Accusé de réception | Client → Serveur |
| `file_upload` | Upload de fichier | Client → Serveur |
| `file_download` | Téléchargement de fichier | Serveur → Client |
| `moderation_action` | Action de modération | Serveur → Client |
| `user_online` | Utilisateur en ligne | Serveur → Client |
| `user_offline` | Utilisateur hors ligne | Serveur → Client |
| `error` | Erreur générale | Serveur → Client |
| `ping` | Ping de santé | Bidirectionnel |
| `pong` | Pong de santé | Bidirectionnel |

### Envoi de Messages

```javascript
class VezaWebSocket {
  // ... autres méthodes

  sendMessage(roomId, content, type = 'text') {
    const message = {
      type: 'message',
      id: this.generateMessageId(),
      timestamp: Date.now(),
      data: {
        room_id: roomId,
        content: content,
        message_type: type, // text, image, file, voice
        reply_to: null, // ID du message auquel répondre
        mentions: [], // Liste des utilisateurs mentionnés
        attachments: [] // Pièces jointes
      }
    };

    this.ws.send(JSON.stringify(message));
  }

  generateMessageId() {
    return `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }
}
```

### Réception de Messages

```javascript
class VezaWebSocket {
  // ... autres méthodes

  handleMessage(message) {
    switch (message.type) {
      case 'message_received':
        this.handleMessageReceived(message);
        break;
      case 'typing_indicator':
        this.handleTypingIndicator(message);
        break;
      case 'user_online':
        this.handleUserOnline(message);
        break;
      case 'user_offline':
        this.handleUserOffline(message);
        break;
      case 'moderation_action':
        this.handleModerationAction(message);
        break;
      case 'error':
        this.handleError(message);
        break;
      default:
        console.log('Unknown message type:', message.type);
    }
  }

  handleMessageReceived(message) {
    const { room_id, content, user_id, username, timestamp } = message.data;
    
    // Émettre un événement pour le composant React
    this.emit('message', {
      roomId: room_id,
      content,
      userId: user_id,
      username,
      timestamp
    });
  }

  handleTypingIndicator(message) {
    const { room_id, user_id, username, is_typing } = message.data;
    
    this.emit('typing', {
      roomId: room_id,
      userId: user_id,
      username,
      isTyping: is_typing
    });
  }
}
```

## 💬 Salons de Chat

### Rejoindre un Salon

```javascript
joinRoom(roomId) {
  const message = {
    type: 'join_room',
    id: this.generateMessageId(),
    timestamp: Date.now(),
    data: {
      room_id: roomId
    }
  };

  this.ws.send(JSON.stringify(message));
}
```

### Quitter un Salon

```javascript
leaveRoom(roomId) {
  const message = {
    type: 'leave_room',
    id: this.generateMessageId(),
    timestamp: Date.now(),
    data: {
      room_id: roomId
    }
  };

  this.ws.send(JSON.stringify(message));
}
```

### Événements de Salon

#### Salon Rejoint

```json
{
  "type": "room_joined",
  "id": "room_123",
  "timestamp": 1640995200000,
  "data": {
    "room_id": "room_123",
    "room_name": "Général",
    "room_type": "public", // public, private, direct
    "users_online": 15,
    "last_message": {
      "id": "msg_456",
      "content": "Bonjour tout le monde !",
      "user_id": "user_789",
      "username": "alice",
      "timestamp": 1640995000000
    }
  }
}
```

#### Nouveau Message dans le Salon

```json
{
  "type": "message_received",
  "id": "msg_789",
  "timestamp": 1640995200000,
  "data": {
    "room_id": "room_123",
    "message_id": "msg_789",
    "content": "Salut ! Comment ça va ?",
    "user_id": "user_456",
    "username": "bob",
    "message_type": "text",
    "timestamp": 1640995200000,
    "mentions": [],
    "attachments": [],
    "is_edited": false,
    "is_deleted": false
  }
}
```

## 💌 Messages Privés

### Envoyer un Message Privé

```javascript
sendPrivateMessage(recipientId, content) {
  const message = {
    type: 'message',
    id: this.generateMessageId(),
    timestamp: Date.now(),
    data: {
      room_id: `private_${this.userId}_${recipientId}`,
      recipient_id: recipientId,
      content: content,
      message_type: 'text',
      is_private: true
    }
  };

  this.ws.send(JSON.stringify(message));
}
```

### Événements de Messages Privés

#### Message Privé Reçu

```json
{
  "type": "private_message",
  "id": "msg_123",
  "timestamp": 1640995200000,
  "data": {
    "sender_id": "user_456",
    "sender_username": "alice",
    "content": "Salut ! Tu veux qu'on parle ?",
    "message_type": "text",
    "timestamp": 1640995200000,
    "is_read": false
  }
}
```

## 🛡️ Modération

### Actions de Modération

```javascript
// Signaler un message
reportMessage(messageId, reason) {
  const message = {
    type: 'report_message',
    id: this.generateMessageId(),
    timestamp: Date.now(),
    data: {
      message_id: messageId,
      reason: reason, // spam, inappropriate, harassment, etc.
      details: "Description détaillée du problème"
    }
  };

  this.ws.send(JSON.stringify(message));
}
```

### Événements de Modération

#### Message Modéré

```json
{
  "type": "moderation_action",
  "id": "mod_123",
  "timestamp": 1640995200000,
  "data": {
    "action": "message_deleted",
    "message_id": "msg_456",
    "reason": "inappropriate_content",
    "moderator_id": "mod_789",
    "room_id": "room_123",
    "user_id": "user_456",
    "username": "bob"
  }
}
```

#### Utilisateur Banni

```json
{
  "type": "moderation_action",
  "id": "mod_124",
  "timestamp": 1640995200000,
  "data": {
    "action": "user_banned",
    "user_id": "user_456",
    "username": "bob",
    "reason": "repeated_violations",
    "duration": "24h", // permanent, 1h, 24h, 7d
    "moderator_id": "mod_789",
    "room_id": "room_123"
  }
}
```

## ⚠️ Gestion des Erreurs

### Types d'Erreurs

| Code | Type | Description |
|------|------|-------------|
| 1000 | `connection_error` | Erreur de connexion |
| 1001 | `authentication_error` | Erreur d'authentification |
| 1002 | `authorization_error` | Erreur d'autorisation |
| 1003 | `rate_limit_error` | Limite de taux dépassée |
| 1004 | `room_not_found` | Salon non trouvé |
| 1005 | `message_too_long` | Message trop long |
| 1006 | `file_too_large` | Fichier trop volumineux |
| 1007 | `invalid_message_format` | Format de message invalide |
| 1008 | `user_banned` | Utilisateur banni |
| 1009 | `room_full` | Salon plein |
| 1010 | `server_error` | Erreur serveur |

### Structure d'Erreur

```json
{
  "type": "error",
  "id": "error_123",
  "timestamp": 1640995200000,
  "data": {
    "code": 1001,
    "error_type": "authentication_error",
    "message": "Token invalide ou expiré",
    "details": "Le token JWT fourni n'est pas valide",
    "retry_after": 60, // secondes avant nouvelle tentative
    "suggestion": "Veuillez vous reconnecter"
  }
}
```

### Gestion des Erreurs

```javascript
class VezaWebSocket {
  // ... autres méthodes

  handleError(error) {
    const { code, error_type, message, retry_after } = error.data;
    
    switch (code) {
      case 1000:
        console.error('Connection error:', message);
        this.handleReconnect();
        break;
      case 1001:
        console.error('Authentication error:', message);
        this.emit('auth_error', { message, retry_after });
        break;
      case 1002:
        console.error('Authorization error:', message);
        this.emit('auth_error', { message });
        break;
      case 1003:
        console.error('Rate limit exceeded:', message);
        this.emit('rate_limit', { retry_after });
        break;
      default:
        console.error('Unknown error:', error);
        this.emit('error', error);
    }
  }
}
```

## 💻 Exemples d'Implémentation

### Client JavaScript/TypeScript

```typescript
interface VezaMessage {
  type: string;
  id: string;
  timestamp: number;
  data: any;
  metadata?: any;
}

interface ChatMessage {
  roomId: string;
  content: string;
  userId: string;
  username: string;
  timestamp: number;
  messageType: string;
  mentions: string[];
  attachments: any[];
}

class VezaChatClient {
  private ws: WebSocket | null = null;
  private token: string;
  private eventListeners: Map<string, Function[]> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;

  constructor(token: string) {
    this.token = token;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(`wss://chat.veza.com/ws?token=${this.token}`);
      
      this.ws.onopen = () => {
        console.log('Connected to Veza chat');
        this.reconnectAttempts = 0;
        resolve();
      };
      
      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        reject(error);
      };
      
      this.ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        this.handleReconnect();
      };
      
      this.ws.onmessage = (event) => {
        const message: VezaMessage = JSON.parse(event.data);
        this.handleMessage(message);
      };
    });
  }

  private handleMessage(message: VezaMessage) {
    switch (message.type) {
      case 'message_received':
        this.emit('message', this.parseChatMessage(message));
        break;
      case 'typing_indicator':
        this.emit('typing', message.data);
        break;
      case 'user_online':
        this.emit('user_online', message.data);
        break;
      case 'user_offline':
        this.emit('user_offline', message.data);
        break;
      case 'moderation_action':
        this.emit('moderation', message.data);
        break;
      case 'error':
        this.emit('error', message.data);
        break;
      default:
        console.log('Unknown message type:', message.type);
    }
  }

  private parseChatMessage(message: VezaMessage): ChatMessage {
    const { room_id, content, user_id, username, timestamp, message_type, mentions, attachments } = message.data;
    
    return {
      roomId: room_id,
      content,
      userId: user_id,
      username,
      timestamp,
      messageType: message_type,
      mentions: mentions || [],
      attachments: attachments || []
    };
  }

  sendMessage(roomId: string, content: string, messageType: string = 'text'): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket not connected');
    }

    const message: VezaMessage = {
      type: 'message',
      id: this.generateMessageId(),
      timestamp: Date.now(),
      data: {
        room_id: roomId,
        content,
        message_type: messageType,
        mentions: this.extractMentions(content),
        attachments: []
      }
    };

    this.ws.send(JSON.stringify(message));
  }

  joinRoom(roomId: string): void {
    const message: VezaMessage = {
      type: 'join_room',
      id: this.generateMessageId(),
      timestamp: Date.now(),
      data: {
        room_id: roomId
      }
    };

    this.ws?.send(JSON.stringify(message));
  }

  leaveRoom(roomId: string): void {
    const message: VezaMessage = {
      type: 'leave_room',
      id: this.generateMessageId(),
      timestamp: Date.now(),
      data: {
        room_id: roomId
      }
    };

    this.ws?.send(JSON.stringify(message));
  }

  startTyping(roomId: string): void {
    const message: VezaMessage = {
      type: 'typing_start',
      id: this.generateMessageId(),
      timestamp: Date.now(),
      data: {
        room_id: roomId
      }
    };

    this.ws?.send(JSON.stringify(message));
  }

  stopTyping(roomId: string): void {
    const message: VezaMessage = {
      type: 'typing_stop',
      id: this.generateMessageId(),
      timestamp: Date.now(),
      data: {
        room_id: roomId
      }
    };

    this.ws?.send(JSON.stringify(message));
  }

  private generateMessageId(): string {
    return `msg_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  private extractMentions(content: string): string[] {
    const mentions = content.match(/@(\w+)/g);
    return mentions ? mentions.map(mention => mention.substring(1)) : [];
  }

  private handleReconnect(): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = 1000 * Math.pow(2, this.reconnectAttempts - 1);
      
      setTimeout(() => {
        console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
        this.connect();
      }, delay);
    } else {
      this.emit('max_reconnect_attempts_reached');
    }
  }

  // Event handling
  on(event: string, callback: Function): void {
    if (!this.eventListeners.has(event)) {
      this.eventListeners.set(event, []);
    }
    this.eventListeners.get(event)!.push(callback);
  }

  private emit(event: string, data: any): void {
    const listeners = this.eventListeners.get(event);
    if (listeners) {
      listeners.forEach(callback => callback(data));
    }
  }

  disconnect(): void {
    this.ws?.close();
  }
}
```

### Utilisation avec React

```typescript
import React, { useEffect, useState } from 'react';
import { VezaChatClient } from './VezaChatClient';

interface ChatProps {
  token: string;
  roomId: string;
}

const Chat: React.FC<ChatProps> = ({ token, roomId }) => {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [typingUsers, setTypingUsers] = useState<string[]>([]);
  const [client, setClient] = useState<VezaChatClient | null>(null);

  useEffect(() => {
    const chatClient = new VezaChatClient(token);
    
    chatClient.on('message', (message: ChatMessage) => {
      setMessages(prev => [...prev, message]);
    });

    chatClient.on('typing', (data: any) => {
      if (data.isTyping) {
        setTypingUsers(prev => [...prev, data.username]);
      } else {
        setTypingUsers(prev => prev.filter(user => user !== data.username));
      }
    });

    chatClient.on('user_online', (data: any) => {
      console.log(`${data.username} is now online`);
    });

    chatClient.on('user_offline', (data: any) => {
      console.log(`${data.username} is now offline`);
    });

    chatClient.on('error', (error: any) => {
      console.error('Chat error:', error);
    });

    chatClient.connect().then(() => {
      setIsConnected(true);
      chatClient.joinRoom(roomId);
      setClient(chatClient);
    });

    return () => {
      chatClient.disconnect();
    };
  }, [token, roomId]);

  const sendMessage = (content: string) => {
    if (client && isConnected) {
      client.sendMessage(roomId, content);
    }
  };

  const handleTyping = (isTyping: boolean) => {
    if (client && isConnected) {
      if (isTyping) {
        client.startTyping(roomId);
      } else {
        client.stopTyping(roomId);
      }
    }
  };

  return (
    <div className="chat-container">
      <div className="messages">
        {messages.map((message, index) => (
          <div key={index} className="message">
            <span className="username">{message.username}</span>
            <span className="content">{message.content}</span>
            <span className="timestamp">
              {new Date(message.timestamp).toLocaleTimeString()}
            </span>
          </div>
        ))}
      </div>
      
      {typingUsers.length > 0 && (
        <div className="typing-indicator">
          {typingUsers.join(', ')} {typingUsers.length === 1 ? 'is' : 'are'} typing...
        </div>
      )}
      
      <MessageInput 
        onSend={sendMessage}
        onTyping={handleTyping}
        disabled={!isConnected}
      />
    </div>
  );
};
```

---

<div className="alert alert--info">
  <strong>💡 Conseil</strong> : Implémentez toujours la gestion des reconnexions et des erreurs pour une expérience utilisateur robuste.
</div>

<div className="alert alert--warning">
  <strong>⚠️ Important</strong> : Respectez les limites de taux et gérez correctement les événements de modération.
</div>

---

**Dernière mise à jour** : $(date)
**Version** : 1.0.0
**Maintenu par** : Équipe Veza 