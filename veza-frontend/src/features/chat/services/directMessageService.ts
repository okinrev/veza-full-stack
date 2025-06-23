import { rustChatWebSocket } from './chatWebSocket';
import type { DirectMessage, ChatUser } from './chatApi';

export interface DirectMessageEvent {
  type: 'message_received' | 'message_sent' | 'user_typing' | 'user_online' | 'user_offline';
  data: any;
}

export class DirectMessageService {
  private listeners: Map<string, Function[]> = new Map();
  private activeConversations: Set<number> = new Set();
  private typingTimeouts: Map<number, NodeJS.Timeout> = new Map();

  constructor() {
    this.setupWebSocketListeners();
  }

  private setupWebSocketListeners() {
    // Écouter les messages privés entrants
    rustChatWebSocket.on('dm_received', (message: any) => {
      this.emit('message_received', {
        id: message.id || Date.now(),
        from_user_id: message.fromUser || message.from_user,
        to_user_id: message.to || message.to_user,
        from_username: message.username || message.from_username || 'Utilisateur',
        to_username: 'Vous',
        content: message.content,
        timestamp: message.timestamp || new Date().toISOString(),
        is_read: false,
      } as DirectMessage);
    });

    // Écouter les confirmations d'envoi
    rustChatWebSocket.on('dm_sent', (data: any) => {
      this.emit('message_sent', data);
    });

    // Note: Les événements user_typing et user_status_changed ne sont pas 
    // disponibles dans l'API actuelle du serveur Rust, mais on peut les 
    // simuler ou les ajouter plus tard
  }

  // Envoyer un message privé
  async sendDirectMessage(toUserId: number, content: string): Promise<void> {
    try {
      // Envoyer via WebSocket si connecté
      if (rustChatWebSocket.isConnected) {
        rustChatWebSocket.sendDirectMessage(toUserId, content);
      } else {
        // Fallback : simuler l'envoi
        setTimeout(() => {
          this.emit('message_sent', {
            to_user_id: toUserId,
            content,
            timestamp: new Date().toISOString()
          });
        }, 100);
      }
    } catch (error) {
      console.error('Erreur lors de l\'envoi du message privé:', error);
      throw error;
    }
  }

  // Charger l'historique des messages privés
  async loadDirectMessageHistory(withUserId: number, limit: number = 50): Promise<void> {
    try {
      if (rustChatWebSocket.isConnected) {
        rustChatWebSocket.getDMHistory(withUserId, limit);
      }
    } catch (error) {
      console.error('Erreur lors du chargement de l\'historique DM:', error);
    }
  }

  // Marquer une conversation comme active
  setActiveConversation(userId: number) {
    this.activeConversations.add(userId);
  }

  // Marquer une conversation comme inactive
  setInactiveConversation(userId: number) {
    this.activeConversations.delete(userId);
    
    // Nettoyer le timeout de frappe
    const timeout = this.typingTimeouts.get(userId);
    if (timeout) {
      clearTimeout(timeout);
      this.typingTimeouts.delete(userId);
    }
  }

  // Envoyer un indicateur de frappe (simulé pour l'instant)
  sendTypingIndicator(toUserId: number, isTyping: boolean) {
    // Note: Le serveur Rust actuel ne supporte pas les indicateurs de frappe
    // On peut simuler localement ou ajouter cette fonctionnalité plus tard
    
    if (isTyping) {
      this.emit('user_typing', {
        userId: toUserId,
        username: 'Utilisateur',
        isTyping: true
      });

      // Auto-stop typing après 3 secondes
      const existingTimeout = this.typingTimeouts.get(toUserId);
      if (existingTimeout) {
        clearTimeout(existingTimeout);
      }

      const timeout = setTimeout(() => {
        this.emit('user_typing', {
          userId: toUserId,
          username: 'Utilisateur',
          isTyping: false
        });
        this.typingTimeouts.delete(toUserId);
      }, 3000);

      this.typingTimeouts.set(toUserId, timeout);
    }
  }

  // Marquer les messages comme lus (simulé pour l'instant)
  async markAsRead(userId: number): Promise<void> {
    try {
      // Note: Le serveur Rust actuel ne supporte pas le marquage comme lu
      // On peut simuler localement ou ajouter cette fonctionnalité plus tard
      console.log(`Messages marqués comme lus pour l'utilisateur ${userId}`);
    } catch (error) {
      console.error('Erreur lors du marquage comme lu:', error);
    }
  }

  // Système d'événements
  on(event: string, callback: Function) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event)!.push(callback);
  }

  off(event: string, callback: Function) {
    const eventListeners = this.listeners.get(event);
    if (eventListeners) {
      const index = eventListeners.indexOf(callback);
      if (index > -1) {
        eventListeners.splice(index, 1);
      }
    }
  }

  private emit(event: string, data: any) {
    const eventListeners = this.listeners.get(event);
    if (eventListeners) {
      eventListeners.forEach(callback => callback(data));
    }
  }

  // Nettoyer les ressources
  cleanup() {
    this.activeConversations.clear();
    this.typingTimeouts.forEach(timeout => clearTimeout(timeout));
    this.typingTimeouts.clear();
    this.listeners.clear();
  }
}

// Instance singleton
export const directMessageService = new DirectMessageService(); 