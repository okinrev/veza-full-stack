// WebSocket Service pour le chat - Conforme à la documentation backend
// Se connecte au Chat-Server Rust sur le port 9001
import { useAuthStore } from '../../../shared/stores/authStore';

// Types pour les messages entrants (du serveur Rust vers le client)
export type WebSocketInboundMessageType =
  | 'message'      // Nouveau message room
  | 'dm'           // Nouveau message privé
  | 'room_history' // Historique room
  | 'dm_history'   // Historique DM
  | 'error';       // Erreur

// Types pour les messages sortants (du client vers le serveur Rust)
export type WebSocketOutboundMessageType = 
  | 'join'         // Rejoindre un salon
  | 'message'      // Envoyer message salon
  | 'dm'           // Envoyer message direct
  | 'room_history' // Demander historique salon
  | 'dm_history';  // Demander historique DM

// Structure des messages sortants vers le serveur Rust
export interface WebSocketOutboundMessage {
  type: WebSocketOutboundMessageType;
  room?: string;      // Pour join, message, room_history
  content?: string;   // Pour message, dm
  to?: number;        // Pour dm (user_id destinataire)
  with?: number;      // Pour dm_history (user_id correspondant)
  limit?: number;     // Pour *_history (nombre max de messages)
}

// Structure des messages entrants du serveur Rust
export interface WebSocketInboundMessage {
  type: WebSocketInboundMessageType;
  data?: any;
  message?: string;   // Pour les erreurs
}

// Message individuel selon la documentation
export interface WebSocketChatMessage {
  id?: number;
  fromUser?: number;
  username: string;
  content: string;
  timestamp: string;
  room?: string;      // Pour les messages de salon
  to?: number;        // Pour les messages directs
}

// Events émis par le WebSocket manager
export type ChatEventType =
  | 'connected'
  | 'disconnected'
  | 'message_received'    // Nouveau message de salon
  | 'dm_received'         // Nouveau message direct reçu
  | 'room_history'        // Historique salon reçu
  | 'dm_history'          // Historique DM reçu
  | 'error';              // Erreur serveur

export class ChatWebSocketManager {
  private ws: WebSocket | null = null;
  private eventListeners: { [K in ChatEventType]: Function[] } = {
    connected: [],
    disconnected: [],
    message_received: [],
    dm_received: [],
    room_history: [],
    dm_history: [],
    error: []
  };
  
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 3000; // ms
  
  // === MÉTHODES D'ÉCOUTE ===
  
  on<T extends ChatEventType>(event: T, callback: Function) {
    this.eventListeners[event].push(callback);
  }
  
  off<T extends ChatEventType>(event: T, callback: Function) {
    const listeners = this.eventListeners[event];
    const index = listeners.indexOf(callback);
    if (index > -1) {
      listeners.splice(index, 1);
    }
  }
  
  private emit<T extends ChatEventType>(event: T, data?: any) {
    console.log(`[Chat WebSocket] Événement émis: ${event}`, data);
    this.eventListeners[event].forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        console.error(`[Chat WebSocket] Erreur dans le callback ${event}:`, error);
      }
    });
  }
  
  // === CONNEXION/DÉCONNEXION ===
  
  async connect(): Promise<boolean> {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('[Chat WebSocket] Déjà connecté');
      return true;
    }
    
    try {
      const authStore = useAuthStore.getState();
      const token = authStore.accessToken;
    
    if (!token) {
        console.error('[Chat WebSocket] Pas de token d\'authentification');
      return false;
    }
    
      // URL WebSocket vers le serveur de chat dédié sur le port 8081
      const wsUrl = import.meta.env.VITE_WS_CHAT_URL ? 
        `${import.meta.env.VITE_WS_CHAT_URL}?token=${token}` :
        `ws://localhost:8081/ws?token=${token}`;
      console.log('[Chat WebSocket] Connexion à:', wsUrl);
      
      this.ws = new WebSocket(wsUrl);
      
      return new Promise((resolve) => {
        if (!this.ws) {
          resolve(false);
          return;
        }
        
        this.ws.onopen = () => {
          console.log('[Chat WebSocket] Connexion établie');
          this.reconnectAttempts = 0;
          this.emit('connected');
          resolve(true);
        };
        
        this.ws.onclose = (event) => {
          console.log('[Chat WebSocket] Connexion fermée:', event.code, event.reason);
          this.emit('disconnected');
          
          // Reconnexion automatique si pas une fermeture normale
          if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.attemptReconnect();
          }
          
          if (this.reconnectAttempts === 0) {
            resolve(false);
          }
        };
        
        this.ws.onerror = (error) => {
          console.error('[Chat WebSocket] Erreur:', error);
          this.emit('error', { message: 'Erreur de connexion WebSocket' });
          resolve(false);
        };
        
        this.ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            this.handleInboundMessage(data);
          } catch (error) {
            console.error('[Chat WebSocket] Erreur parsing message:', error, event.data);
          }
        };
      });
      
    } catch (error) {
      console.error('[Chat WebSocket] Erreur de connexion:', error);
      this.emit('error', { message: 'Impossible de se connecter' });
      return false;
    }
  }
  
  private attemptReconnect() {
    this.reconnectAttempts++;
    const delay = this.reconnectDelay * this.reconnectAttempts;
    
    console.log(`[Chat WebSocket] Tentative de reconnexion ${this.reconnectAttempts}/${this.maxReconnectAttempts} dans ${delay}ms`);
    
    setTimeout(() => {
      this.connect();
    }, delay);
  }
  
  disconnect() {
    if (this.ws) {
      this.ws.onclose = null; // Empêcher la reconnexion
      this.ws.close(1000, 'Déconnexion demandée');
      this.ws = null;
    }
    this.reconnectAttempts = 0;
  }
  
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
  
  // === GESTION DES MESSAGES ENTRANTS ===
  
  private handleInboundMessage(message: WebSocketInboundMessage) {
    console.log('[Chat WebSocket] Message reçu:', message);
    
    switch (message.type) {
      case 'message':
        // Nouveau message de salon selon la documentation
        if (message.data) {
          this.emit('message_received', {
            username: message.data.username,
            fromUser: message.data.fromUser,
            content: message.data.content,
            timestamp: message.data.timestamp,
            room: message.data.room
          });
        }
        break;
        
      case 'dm':
        // Nouveau message privé selon la documentation
        if (message.data) {
          this.emit('dm_received', {
            id: message.data.id,
            fromUser: message.data.fromUser,
            username: message.data.username,
            content: message.data.content,
            timestamp: message.data.timestamp
          });
        }
        break;
        
      case 'room_history':
        // Historique salon selon la documentation
        if (Array.isArray(message.data)) {
          this.emit('room_history', message.data);
        }
        break;
        
      case 'dm_history':
        // Historique DM selon la documentation
        if (Array.isArray(message.data)) {
          this.emit('dm_history', message.data);
        }
        break;
        
      case 'error':
        // Erreur serveur
        console.error('[Chat WebSocket] Erreur serveur:', message.data);
        this.emit('error', { message: message.data?.message || 'Erreur inconnue' });
        break;
        
      default:
        console.warn('[Chat WebSocket] Type de message non géré:', message.type);
    }
  }

  // === MÉTHODES POUR ENVOYER DES MESSAGES AU SERVEUR RUST ===
  
  private sendMessage(payload: WebSocketOutboundMessage) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[Chat WebSocket] WebSocket non connecté');
      this.emit('error', { message: 'WebSocket non connecté' });
      return false;
    }
    
    const message = JSON.stringify(payload);
    console.log('[Chat WebSocket] Envoi vers serveur:', message);
    this.ws.send(message);
    return true;
  }
  
  /**
   * Rejoindre un salon
   */
  joinRoom(roomName: string): boolean {
    return this.sendMessage({
      type: 'join',
      room: roomName
    });
  }

  /**
   * Envoyer un message dans un salon
   */
  sendRoomMessage(roomName: string, content: string): boolean {
    return this.sendMessage({
      type: 'message',
      room: roomName,
      content
    });
  }

  /**
   * Envoyer un message privé
   */
  sendDirectMessage(toUserId: number, content: string): boolean {
    return this.sendMessage({
      type: 'dm',
      to: toUserId,
      content
    });
  }

  /**
   * Demander l'historique d'un salon
   */
  getRoomHistory(roomName: string, limit = 50): boolean {
    return this.sendMessage({
      type: 'room_history',
      room: roomName,
      limit
    });
  }

  /**
   * Demander l'historique des messages privés avec un utilisateur
   */
  getDMHistory(withUserId: number, limit = 50): boolean {
    return this.sendMessage({
      type: 'dm_history',
      with: withUserId,
      limit
    });
  }
}

// Instance singleton
export const chatWebSocket = new ChatWebSocketManager(); 