# Chat Service - Documentation

## Vue d'ensemble

Le ChatService (`src/features/chat/services/chatService.ts`) est le service principal qui gère toutes les opérations de chat, incluant les WebSockets, les API REST et la synchronisation des données.

## Architecture

```
ChatService
├── WebSocket Manager (Rust)  ← Communication temps réel
├── API Client (Go)           ← Persistance et historique
└── Chat Store (Zustand)      ← État local
```

## Initialisation

```typescript
export class ChatService {
  private isInitialized = false;
  private wsManager: ChatWebSocketManager | null = null;
  private apiClient = ApiClient.getInstance();
  
  async initialize(): Promise<boolean> {
    if (this.isInitialized) return true;
    
    try {
      // 1. Vérifier l'authentification
      await this.checkAuth();
      
      // 2. Charger les données initiales
      await Promise.all([
        this.loadRooms(),
        this.loadConversations(),
        this.loadUnreadCount()
      ]);
      
      // 3. Connecter WebSocket
      await this.connectWebSocket();
      
      this.isInitialized = true;
      return true;
    } catch (error) {
      console.error('Chat initialization failed:', error);
      return false;
    }
  }
}
```

## Gestion WebSocket

### Connexion

```typescript
async connectWebSocket(): Promise<boolean> {
  if (this.wsManager?.isConnected()) return true;
  
  this.wsManager = new ChatWebSocketManager();
  
  // Configuration des event listeners
  this.setupWebSocketListeners();
  
  try {
    await this.wsManager.connect();
    return true;
  } catch (error) {
    console.error('WebSocket connection failed:', error);
    return false;
  }
}
```

### Event Listeners

```typescript
private setupWebSocketListeners(): void {
  if (!this.wsManager) return;
  
  // Nouveau message de salon
  this.wsManager.on('message_received', (message: WebSocketChatMessage) => {
    const chatStore = useChatStore.getState();
    chatStore.addRoomMessage(this.convertMessage(message));
  });
  
  // Message direct reçu
  this.wsManager.on('dm_received', (message: WebSocketChatMessage) => {
    const chatStore = useChatStore.getState();
    chatStore.addDmMessage(this.convertMessage(message));
    
    // Mettre à jour les conversations
    this.updateConversationFromMessage(message);
  });
  
  // Historique reçu
  this.wsManager.on('room_history', (data: { room: string, messages: WebSocketChatMessage[] }) => {
    const chatStore = useChatStore.getState();
    const convertedMessages = data.messages.map(this.convertMessage);
    chatStore.setRoomMessages(convertedMessages);
  });
  
  // Gestion des erreurs
  this.wsManager.on('error', (error: any) => {
    console.error('WebSocket error:', error);
    // Optionnel: afficher une notification
  });
}
```

## Opérations Principales

### Rejoindre un Salon

```typescript
async joinRoom(roomName: string): Promise<boolean> {
  try {
    // 1. Joindre via WebSocket pour temps réel
    if (this.wsManager?.isConnected()) {
      this.wsManager.joinRoom(roomName);
    }
    
    // 2. Charger l'historique via API
    await this.loadRoomHistory(roomName);
    
    // 3. Mettre à jour le store
    const chatStore = useChatStore.getState();
    chatStore.setCurrentRoom(roomName);
    
    return true;
  } catch (error) {
    console.error('Failed to join room:', error);
    return false;
  }
}
```

### Envoyer Message de Salon

```typescript
async sendRoomMessage(roomName: string, content: string): Promise<boolean> {
  try {
    // Envoyer via WebSocket pour diffusion immédiate
    if (this.wsManager?.isConnected()) {
      this.wsManager.sendRoomMessage(roomName, content);
      return true;
    }
    
    // Fallback API si WebSocket indisponible
    await this.apiClient.post(`/api/v1/chat/rooms/${roomName}/messages`, {
      content
    });
    
    return true;
  } catch (error) {
    console.error('Failed to send room message:', error);
    return false;
  }
}
```

### Messages Directs

```typescript
async openDirectMessage(user: BackendUser): Promise<boolean> {
  try {
    // 1. Charger l'historique des DM
    const messages = await this.apiClient.get<BackendMessage[]>(
      `/api/v1/chat/dm/${user.id}?limit=50`
    );
    
    // 2. Convertir et stocker
    const chatStore = useChatStore.getState();
    const convertedMessages = messages.map(this.convertBackendMessage);
    chatStore.setDmMessages(convertedMessages);
    chatStore.setCurrentDMUser(this.convertBackendUser(user));
    
    // 3. Marquer comme lu
    await this.markMessagesAsRead(user.id);
    
    return true;
  } catch (error) {
    console.error('Failed to open DM:', error);
    return false;
  }
}

async sendDirectMessage(userId: number, content: string): Promise<boolean> {
  try {
    // WebSocket en priorité
    if (this.wsManager?.isConnected()) {
      this.wsManager.sendDirectMessage(userId, content);
      return true;
    }
    
    // Fallback API
    await this.apiClient.post(`/api/v1/chat/dm/${userId}`, { content });
    return true;
  } catch (error) {
    console.error('Failed to send DM:', error);
    return false;
  }
}
```

## Chargement des Données

### Salles de Chat

```typescript
async loadRooms(): Promise<void> {
  try {
    const rooms = await this.apiClient.get<BackendRoom[]>('/api/v1/chat/rooms');
    
    const chatStore = useChatStore.getState();
    const convertedRooms = rooms.map(this.convertBackendRoom);
    chatStore.setRooms(convertedRooms);
  } catch (error) {
    console.error('Failed to load rooms:', error);
    throw error;
  }
}
```

### Conversations

```typescript
async loadConversations(): Promise<void> {
  try {
    const conversations = await this.apiClient.get<BackendConversation[]>(
      '/api/v1/chat/conversations'
    );
    
    const chatStore = useChatStore.getState();
    chatStore.setConversations(conversations);
  } catch (error) {
    console.error('Failed to load conversations:', error);
    throw error;
  }
}
```

## Convertisseurs de Données

### Message Backend → Frontend

```typescript
private convertBackendMessage(backendMessage: BackendMessage): ChatMessage {
  return {
    id: backendMessage.id,
    from_user: backendMessage.from_user,
    to_user: backendMessage.to_user,
    room: backendMessage.room,
    content: backendMessage.content,
    is_read: backendMessage.is_read,
    timestamp: backendMessage.timestamp,
    username: backendMessage.username,
    avatar: backendMessage.avatar
  };
}
```

### User Backend → Frontend

```typescript
private convertBackendUser(backendUser: BackendUser): ChatUser {
  return {
    id: backendUser.id,
    username: backendUser.username,
    first_name: backendUser.first_name,
    last_name: backendUser.last_name,
    avatar: backendUser.avatar,
    is_online: backendUser.is_online,
    last_seen: backendUser.last_seen
  };
}
```

## Gestion des Erreurs

```typescript
private handleError(operation: string, error: any): void {
  console.error(`Chat ${operation} failed:`, error);
  
  // Optionnel: notification utilisateur
  const { toast } = useToast();
  toast({
    title: 'Erreur Chat',
    description: `Impossible de ${operation}`,
    variant: 'destructive'
  });
  
  // Analytics/monitoring
  if (import.meta.env.PROD) {
    // Envoyer erreur à service de monitoring
  }
}
```

## Nettoyage et Cleanup

```typescript
cleanup(): void {
  // Déconnecter WebSocket
  if (this.wsManager) {
    this.wsManager.disconnect();
    this.wsManager = null;
  }
  
  // Réinitialiser l'état
  this.isInitialized = false;
  
  // Nettoyer le store
  const chatStore = useChatStore.getState();
  chatStore.reset();
}

reset(): void {
  this.cleanup();
  
  // Vider les caches locaux
  // Réinitialiser les timers
}
```

## Utilisation dans les Composants

```typescript
// Hook personnalisé pour simplifier l'usage
export const useChat = () => {
  const chatService = useMemo(() => ChatService.getInstance(), []);
  const [isInitialized, setIsInitialized] = useState(false);
  
  useEffect(() => {
    const initChat = async () => {
      const success = await chatService.initialize();
      setIsInitialized(success);
    };
    
    initChat();
    
    return () => {
      chatService.cleanup();
    };
  }, [chatService]);
  
  return {
    chatService,
    isInitialized,
    joinRoom: chatService.joinRoom.bind(chatService),
    sendMessage: chatService.sendRoomMessage.bind(chatService),
    sendDM: chatService.sendDirectMessage.bind(chatService),
    openDM: chatService.openDirectMessage.bind(chatService)
  };
};
```

## Configuration Backend Requise

### API Endpoints Go

```go
// Groupe chat
chatGroup := v1.Group("/chat")
chatGroup.Use(authMiddleware)
{
    chatGroup.GET("/conversations", chatHandler.GetConversations)
    chatGroup.GET("/rooms", chatHandler.GetRooms)
    chatGroup.POST("/rooms", chatHandler.CreateRoom)
    chatGroup.GET("/rooms/:roomName/messages", chatHandler.GetRoomMessages)
    chatGroup.GET("/dm/:userId", chatHandler.GetDirectMessages)
    chatGroup.POST("/dm/:userId", chatHandler.SendDirectMessage)
    chatGroup.PUT("/messages/:userId/read", chatHandler.MarkAsRead)
    chatGroup.GET("/unread", chatHandler.GetUnreadCount)
}
```

### WebSocket Rust

```rust
// Configuration serveur WebSocket Rust
#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let addr = "127.0.0.1:8081";
    let listener = TcpListener::bind(&addr).await?;
    println!("Chat WebSocket server running on: {}", addr);
    
    while let Ok((stream, _)) = listener.accept().await {
        tokio::spawn(handle_connection(stream));
    }
    
    Ok(())
}

async fn handle_connection(raw_stream: TcpStream) {
    // Gérer l'upgrade WebSocket et l'authentification
    // Traiter les messages selon le protocole défini
}
```

Cette architecture garantit une communication chat robuste avec fallback automatique et synchronisation entre temps réel et persistance.
