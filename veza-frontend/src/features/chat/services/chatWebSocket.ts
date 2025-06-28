// WebSocket Service pour le chat - Compatible avec Chat-Server Rust
// Adapté pour le nouveau store d'authentification
import { useAuthStore } from '@/features/auth/store/authStore';

// Types pour les messages entrants (du serveur Rust vers le client)
export type WebSocketInboundMessageType =
  | 'NewMessage'       // Nouveau message
  | 'ActionConfirmed'  // Confirmation d'action
  | 'Error'           // Erreur
  | 'Pong';           // Réponse au ping

// Types pour les messages sortants (du client vers le serveur Rust)
export type WebSocketOutboundMessageType = 
  | 'SendMessage'      // Envoyer un message
  | 'JoinConversation' // Rejoindre une conversation
  | 'LeaveConversation'// Quitter une conversation
  | 'MarkAsRead'       // Marquer comme lu
  | 'Ping';            // Ping de connexion

// Structure des messages sortants vers le serveur Rust
export interface WebSocketOutboundMessage {
  type: WebSocketOutboundMessageType;
  conversation_id?: string;    // UUID de la conversation
  content?: string;           // Contenu du message
  message_id?: string;        // UUID du message
  parent_message_id?: string | null; // UUID du message parent
}

// Structure des messages entrants du serveur Rust
export interface WebSocketInboundMessage {
  type: WebSocketInboundMessageType;
  data?: any;
  message?: string;   // Pour les erreurs
  room?: string;      // Salon concerné
  user?: any;         // Utilisateur concerné
}

// Message individuel (adapté de l'ancien frontend)
export interface WebSocketChatMessage {
  id?: number;
  fromUser?: number;
  username: string;
  content: string;
  timestamp: string;
  room?: string;      // Pour les messages de salon
  to?: number;        // Pour les messages directs
  status?: 'sent' | 'delivered' | 'read'; // Statut du message
}

// Utilisateur connecté
export interface ConnectedUser {
  id: number;
  username: string;
  email?: string;
  isOnline: boolean;
}

// Events émis par le WebSocket manager
export type ChatEventType =
  | 'connected'
  | 'disconnected'
  | 'message_received'    // Nouveau message de salon
  | 'dm_received'         // Nouveau message direct reçu
  | 'room_history'        // Historique salon reçu
  | 'dm_history'          // Historique DM reçu
  | 'room_joined'         // Salon rejoint avec succès
  | 'room_left'           // Salon quitté
  | 'user_joined'         // Utilisateur a rejoint salon
  | 'user_left'           // Utilisateur a quitté salon
  | 'users_list'          // Liste des utilisateurs connectés
  | 'typing_start'        // Quelqu'un commence à écrire
  | 'typing_stop'         // Quelqu'un arrête d'écrire
  | 'error';              // Erreur serveur

export class ChatWebSocketManager extends EventTarget {
  private ws: WebSocket | null = null;
  private eventListeners: { [K in ChatEventType]: Function[] } = {
    connected: [],
    disconnected: [],
    message_received: [],
    dm_received: [],
    room_history: [],
    dm_history: [],
    room_joined: [],
    room_left: [],
    user_joined: [],
    user_left: [],
    users_list: [],
    typing_start: [],
    typing_stop: [],
    error: []
  };
  
  private reconnectAttempts = 0;
  private readonly maxReconnectAttempts = 5;
  private readonly reconnectDelay = 2000;
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private pingInterval: NodeJS.Timeout | null = null;
  private isManualDisconnect = false;
  
  // Mapping des rooms vers des UUIDs fixes
  private roomToUuidMap: Map<string, string> = new Map();
  
  // UUIDs fixes pour les rooms communes
  private readonly defaultRoomUuids = {
    'general': '00000000-0000-4000-8000-000000000001',
    'random': '00000000-0000-4000-8000-000000000002',
    'tech': '00000000-0000-4000-8000-000000000003',
    'music': '00000000-0000-4000-8000-000000000004',
  };

  constructor() {
    super();
    // Initialiser le mapping avec les rooms par défaut
    Object.entries(this.defaultRoomUuids).forEach(([room, uuid]) => {
      this.roomToUuidMap.set(room, uuid);
    });
  }
  
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
    const DEBUG = import.meta.env.VITE_DEBUG === 'true';
    if (DEBUG) {
      console.log(`🔵 [Chat WebSocket] Événement: ${event}`, data);
    }
    
    this.eventListeners[event].forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        console.error(`🔴 [Chat WebSocket] Erreur callback ${event}:`, error);
      }
    });
  }
  
  // === CONNEXION/DÉCONNEXION ===
  
  async connect(): Promise<boolean> {
    if (this.ws?.readyState === WebSocket.OPEN) {
      console.log('🟡 [Chat WebSocket] Déjà connecté');
      return true;
    }
    
    try {
      // Pour l'instant, le serveur de chat ne gère pas l'authentification WebSocket
      // On se connecte directement sans token
      
      // URL WebSocket vers le serveur de chat Rust
      const wsUrl = import.meta.env.VITE_WS_CHAT_URL || 'ws://10.5.191.108:3001/ws';
      
      console.log('🔵 [Chat WebSocket] Connexion à:', wsUrl);
      
      this.ws = new WebSocket(wsUrl);
      
      return new Promise((resolve) => {
        if (!this.ws) {
          resolve(false);
          return;
        }
        
        const timeout = setTimeout(() => {
          console.error('🔴 [Chat WebSocket] Timeout de connexion');
          resolve(false);
        }, 10000); // 10 secondes de timeout
        
        this.ws.onopen = () => {
          clearTimeout(timeout);
          console.log('🟢 [Chat WebSocket] Connexion établie');
          this.reconnectAttempts = 0;
          this.startPing();
          this.emit('connected');
          resolve(true);
        };
        
        this.ws.onclose = (event) => {
          clearTimeout(timeout);
          console.log(`🟡 [Chat WebSocket] Connexion fermée: ${event.code} - ${event.reason}`);
          this.stopPing();
          this.emit('disconnected', { code: event.code, reason: event.reason });
          
          // Reconnexion automatique si pas une fermeture normale
          if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect();
          }
          
          if (this.reconnectAttempts === 0) {
            resolve(false);
          }
        };
        
        this.ws.onerror = (error) => {
          clearTimeout(timeout);
          console.error('🔴 [Chat WebSocket] Erreur:', error);
          this.emit('error', { message: 'Erreur de connexion WebSocket' });
          resolve(false);
        };
        
        this.ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            this.handleInboundMessage(data);
          } catch (error) {
            console.error('🔴 [Chat WebSocket] Erreur parsing message:', error, event.data);
          }
        };
      });
      
    } catch (error) {
      console.error('🔴 [Chat WebSocket] Erreur de connexion:', error);
      this.emit('error', { message: 'Impossible de se connecter au chat' });
      return false;
    }
  }
  
  private scheduleReconnect() {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
    }
    
    this.reconnectAttempts++;
    const delay = this.reconnectDelay * this.reconnectAttempts;
    
    console.log(`🟡 [Chat WebSocket] Reconnexion ${this.reconnectAttempts}/${this.maxReconnectAttempts} dans ${delay}ms`);
    
    this.reconnectTimeout = setTimeout(() => {
      this.connect();
    }, delay);
  }
  
  private startPing() {
    this.stopPing();
    this.pingInterval = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'Ping' }));
      }
    }, 30000); // Ping toutes les 30 secondes
  }
  
  private stopPing() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }
  
  disconnect() {
    this.stopPing();
    
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    
    if (this.ws) {
      this.ws.onclose = null; // Empêcher la reconnexion automatique
      this.ws.close(1000, 'Déconnexion demandée');
      this.ws = null;
    }
    
    this.reconnectAttempts = 0;
    console.log('🟡 [Chat WebSocket] Déconnecté');
  }
  
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
  
  // === GESTION DES MESSAGES ENTRANTS ===
  
  private handleInboundMessage(data: WebSocketInboundMessage) {
    const DEBUG = import.meta.env.VITE_DEBUG === 'true';
    if (DEBUG) {
      console.log('📥 [Chat WebSocket] Message reçu:', data);
    }
    
    switch (data.type) {
      case 'NewMessage':
        // Nouveau message de salon
        this.emit('message_received', data.data);
        break;
        
      case 'ActionConfirmed':
        // Confirmation salon rejoint
        this.emit('room_joined', data.data);
        break;
        
      case 'Error':
        // Erreur serveur
        console.error('🔴 [Chat WebSocket] Erreur serveur:', data.message);
        this.emit('error', { message: data.message || 'Erreur inconnue' });
        break;
        
      case 'Pong':
        // Réponse au ping
        this.emit('connected');
        break;
        
      default:
        console.warn('⚠️ [Chat WebSocket] Type de message inconnu:', data.type);
    }
  }
  
  // === ENVOI DE MESSAGES ===
  
  private sendMessage(payload: WebSocketOutboundMessage): boolean {
    if (!this.isConnected()) {
      console.error('🔴 [Chat WebSocket] Pas connecté, impossible d\'envoyer:', payload.type);
      return false;
    }
    
    try {
      this.ws!.send(JSON.stringify(payload));
      
      const DEBUG = import.meta.env.VITE_DEBUG === 'true';
      if (DEBUG) {
        console.log('📤 [Chat WebSocket] Message envoyé:', payload);
      }
      
      return true;
    } catch (error) {
      console.error('🔴 [Chat WebSocket] Erreur envoi message:', error);
      return false;
    }
  }
  
  // === MÉTHODES PUBLIQUES ===
  
  joinRoom(roomName: string): boolean {
    const conversationId = this.getRoomUuid(roomName);
    return this.sendMessage({
      type: 'JoinConversation',
      conversation_id: conversationId
    });
  }
  
  leaveRoom(roomName: string): boolean {
    const conversationId = this.getRoomUuid(roomName);
    return this.sendMessage({
      type: 'LeaveConversation',
      conversation_id: conversationId
    });
  }
  
  // Fonction utilitaire pour générer un UUID simple
  private generateUUID(): string {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
      const r = Math.random() * 16 | 0;
      const v = c === 'x' ? r : (r & 0x3 | 0x8);
      return v.toString(16);
    });
  }

  // Obtenir ou créer l'UUID d'une room
  private getRoomUuid(roomName: string): string {
    if (!this.roomToUuidMap.has(roomName)) {
      this.roomToUuidMap.set(roomName, this.generateUUID());
    }
    return this.roomToUuidMap.get(roomName)!;
  }

  sendRoomMessage(roomName: string, content: string): boolean {
    // Utiliser l'UUID fixe de la room
    const conversationId = this.getRoomUuid(roomName);
    
    return this.sendMessage({
      type: 'SendMessage',
      conversation_id: conversationId,
      content: content.trim(),
      parent_message_id: null
    });
  }
  
  sendDirectMessage(toUserId: number, content: string): boolean {
    return this.sendMessage({
      type: 'SendMessage',
      conversation_id: this.generateUUID(),
      content: content.trim(),
      message_id: this.generateUUID(),
      parent_message_id: null
    });
  }
  
  async getRoomHistory(roomName: string, limit = 50): Promise<boolean> {
    try {
      // Utiliser l'API REST pour récupérer l'historique
      const conversationId = this.getRoomUuid(roomName);
      const baseUrl = import.meta.env.VITE_API_URL || 'http://10.5.191.175:8080/api/v1';
      // Correction: utiliser l'API Go au lieu de l'API Rust
      const response = await fetch(`${baseUrl}/chat/rooms/${encodeURIComponent(conversationId)}/messages?limit=${limit}`);
      
      if (response.ok) {
        const data = await response.json();
        // Émettre l'événement d'historique reçu
        this.emit('room_history', data.data || []);
        return true;
      }
    } catch (error) {
      console.warn('Erreur récupération historique:', error);
    }
    return false;
  }
  
  getDMHistory(withUserId: number, limit = 50): boolean {
    // Pour l'instant, les DM ne sont pas implémentés dans l'API
    return false;
  }
  
  sendTyping(roomName?: string, toUserId?: number): boolean {
    // Le serveur actuel ne supporte pas les messages de typing
    // On peut ignorer silencieusement cette fonctionnalité
    return true;
  }
}

// Instance singleton
export const chatWebSocket = new ChatWebSocketManager(); 