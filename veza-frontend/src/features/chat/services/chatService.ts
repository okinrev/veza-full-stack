// Service de chat centralisé - Implémente toutes les méthodes critiques de la documentation
import { useChatStore } from '../store/chatStore';
import { useAuthStore } from '../../../shared/stores/authStore';
import { chatApi } from './chatApi';
import { chatWebSocket } from './chatWebSocket';
import type { BackendConversation, BackendRoom, BackendMessage, BackendUser } from './chatApi';
import { extractNullString } from '../../../shared/api/types';

export class ChatService {
  private isInitialized = false;
  private initPromise: Promise<boolean> | null = null;

  /**
   * Initialisation du chat selon la documentation
   * 1. Vérifier authentification
   * 2. Charger données initiales
   * 3. Établir connexion WebSocket
   */
  async initialize(): Promise<boolean> {
    if (this.isInitialized) {
      return true;
    }

    if (this.initPromise) {
      return this.initPromise;
    }

    this.initPromise = this._performInit();
    return this.initPromise;
  }

  private async _performInit(): Promise<boolean> {
    const chatStore = useChatStore.getState();
    
    try {
      chatStore.setLoading(true);
      chatStore.setError(null);

      // 1. Vérifier authentification
      const isAuthenticated = await this.checkAuth();
      if (!isAuthenticated) {
        chatStore.setError('Authentification requise');
        return false;
      }

      // 2. Charger données initiales en parallèle
      await Promise.all([
        this.loadRooms(),
        this.loadConversations(),
        this.loadUnreadCount()
      ]);

      // 3. Établir connexion WebSocket
      const wsConnected = await this.connectWebSocket();
      if (!wsConnected) {
        chatStore.setError('Impossible de se connecter au chat temps réel');
        return false;
      }

      this.isInitialized = true;
      chatStore.setLoading(false);
      console.log('[Chat Service] Initialisation réussie');
      return true;

    } catch (error) {
      console.error('[Chat Service] Erreur initialisation:', error);
      chatStore.setError('Erreur lors de l\'initialisation du chat');
      chatStore.setLoading(false);
      return false;
    }
  }

  /**
   * Vérification authentification selon la documentation
   */
  async checkAuth(): Promise<boolean> {
    const authStore = useAuthStore.getState();
    const token = authStore.accessToken;
    
    if (!token) {
      console.log('[Chat Service] Pas de token, redirection vers login');
      window.location.href = '/login';
      return false;
    }

    try {
      const currentUser = authStore.user;

      if (!currentUser) {
        console.log('[Chat Service] Pas d\'utilisateur, redirection vers login');
        window.location.href = '/login';
        return false;
      }

      // Mettre à jour le store chat avec les infos utilisateur
      const chatStore = useChatStore.getState();
      
      chatStore.setUser({
        id: currentUser.id,
        username: extractNullString(currentUser.username) || 'User',
        avatar: extractNullString(currentUser.avatar) || ''
      });
      chatStore.setAuthenticated(true);

      console.log('[Chat Service] Authentification validée pour:', currentUser.username);
      return true;

    } catch (error) {
      console.error('[Chat Service] Erreur vérification auth:', error);
      localStorage.removeItem('access_token');
      window.location.href = '/login';
      return false;
    }
  }

  /**
   * Charger les rooms publiques
   */
  async loadRooms(): Promise<void> {
    try {
      console.log('[Chat Service] Chargement des rooms...');
      const rooms = await chatApi.getRooms();
      
      const chatStore = useChatStore.getState();
      chatStore.setRooms(rooms);
      
      console.log('[Chat Service] Rooms chargées:', rooms.length);
    } catch (error) {
      console.error('[Chat Service] Erreur chargement rooms:', error);
    }
  }

  /**
   * Charger les conversations DM
   */
  async loadConversations(): Promise<void> {
    try {
      console.log('[Chat Service] Chargement des conversations...');
      const conversations = await chatApi.getConversations();
      
      const chatStore = useChatStore.getState();
      chatStore.setConversations(conversations);
      
      console.log('[Chat Service] Conversations chargées:', conversations.length);
    } catch (error) {
      console.error('[Chat Service] Erreur chargement conversations:', error);
    }
  }

  /**
   * Charger le compteur de messages non lus
   */
  async loadUnreadCount(): Promise<void> {
    try {
      console.log('[Chat Service] Chargement compteur non lus...');
      const count = await chatApi.getUnreadCount();
      
      const chatStore = useChatStore.getState();
      chatStore.setTotalUnreadDM(count);
      
      console.log('[Chat Service] Messages non lus:', count);
    } catch (error) {
      console.error('[Chat Service] Erreur chargement compteur non lus:', error);
    }
  }

  /**
   * Établir connexion WebSocket
   */
  async connectWebSocket(): Promise<boolean> {
    try {
      console.log('[Chat Service] Connexion WebSocket...');
      
      // Configurer les écouteurs d'événements
      this.setupWebSocketListeners();
      
      // Se connecter
      const connected = await chatWebSocket.connect();
      
      const chatStore = useChatStore.getState();
      chatStore.setConnected(connected);
      chatStore.setSocket(connected ? chatWebSocket as any : null);
      
      if (connected) {
        console.log('[Chat Service] WebSocket connecté avec succès');
      } else {
        console.error('[Chat Service] Échec connexion WebSocket');
      }
      
      return connected;

    } catch (error) {
      console.error('[Chat Service] Erreur connexion WebSocket:', error);
      return false;
    }
  }

  /**
   * Configurer les écouteurs WebSocket
   */
  private setupWebSocketListeners(): void {
    const chatStore = useChatStore.getState();

    // Connexion établie
    chatWebSocket.on('connected', () => {
      console.log('[Chat Service] WebSocket connecté');
      chatStore.setConnected(true);
      chatStore.resetReconnectAttempts();
    });

    // Connexion fermée
    chatWebSocket.on('disconnected', () => {
      console.log('[Chat Service] WebSocket déconnecté');
      chatStore.setConnected(false);
    });

    // Nouveau message de salon
    chatWebSocket.on('message_received', (data: any) => {
      console.log('[Chat Service] Nouveau message salon:', data);
      if (data.room === chatStore.currentRoom) {
        chatStore.addRoomMessage({
          id: Date.now(), // ID temporaire
          from_user: data.fromUser,
          content: data.content,
          timestamp: data.timestamp,
          username: data.username,
          room: data.room,
          is_read: true
        });
      }
    });

    // Nouveau message privé
    chatWebSocket.on('dm_received', (data: any) => {
      console.log('[Chat Service] Nouveau message privé:', data);
      if (data.fromUser === chatStore.currentDMUser?.id) {
        chatStore.addDmMessage({
          id: data.id,
          from_user: data.fromUser,
          content: data.content,
          timestamp: data.timestamp,
          username: data.username,
          is_read: false
        });
      }
      // Mettre à jour le compteur de non lus
      this.loadUnreadCount();
    });

    // Historique salon reçu
    chatWebSocket.on('room_history', (messages: any[]) => {
      console.log('[Chat Service] Historique salon reçu:', messages.length);
      const formattedMessages = messages.map(msg => ({
        id: msg.id || Date.now(),
        from_user: msg.fromUser || 0,
        content: msg.content,
        timestamp: msg.timestamp,
        username: msg.username,
        room: chatStore.currentRoom,
        is_read: true
      }));
      chatStore.setRoomMessages(formattedMessages);
    });

    // Historique DM reçu
    chatWebSocket.on('dm_history', (messages: any[]) => {
      console.log('[Chat Service] Historique DM reçu:', messages.length);
      const formattedMessages = messages.map(msg => ({
        id: msg.id || Date.now(),
        from_user: msg.fromUser || 0,
        content: msg.content,
        timestamp: msg.timestamp,
        username: msg.username,
        is_read: false
      }));
      chatStore.setDmMessages(formattedMessages);
    });

    // Erreur WebSocket
    chatWebSocket.on('error', (error: any) => {
      console.error('[Chat Service] Erreur WebSocket:', error);
      chatStore.setError(error.message || 'Erreur WebSocket');
    });
  }

  /**
   * Rejoindre un salon
   */
  async joinRoom(roomName: string): Promise<boolean> {
    try {
      console.log('[Chat Service] Rejoint salon:', roomName);
      
      const chatStore = useChatStore.getState();
      chatStore.setCurrentRoom(roomName);
      chatStore.setRoomMessages([]); // Clear messages
      
      // Rejoindre via WebSocket
      const joined = chatWebSocket.joinRoom(roomName);
      if (joined) {
        // Demander l'historique
        chatWebSocket.getRoomHistory(roomName, 50);
      }
      
      return joined;

    } catch (error) {
      console.error('[Chat Service] Erreur join salon:', error);
      return false;
    }
  }

  /**
   * Envoyer message dans salon
   */
  async sendRoomMessage(roomName: string, content: string): Promise<boolean> {
    try {
      console.log('[Chat Service] Envoi message salon:', roomName, content);
      return chatWebSocket.sendRoomMessage(roomName, content);
    } catch (error) {
      console.error('[Chat Service] Erreur envoi message salon:', error);
      return false;
    }
  }

  /**
   * Ouvrir conversation privée
   */
  async openDirectMessage(user: BackendUser): Promise<boolean> {
    try {
      console.log('[Chat Service] Ouverture DM avec:', user.username);
      
      const chatStore = useChatStore.getState();
      chatStore.setCurrentDMUser({
        id: user.id,
        username: user.username,
        first_name: user.first_name,
        last_name: user.last_name,
        avatar: user.avatar,
        is_online: user.is_online,
        last_seen: user.last_seen
      });
      chatStore.setDmMessages([]); // Clear messages
      
      // Demander l'historique via WebSocket
      chatWebSocket.getDMHistory(user.id, 50);
      
      return true;

    } catch (error) {
      console.error('[Chat Service] Erreur ouverture DM:', error);
      return false;
    }
  }

  /**
   * Envoyer message privé
   */
  async sendDirectMessage(userId: number, content: string): Promise<boolean> {
    try {
      console.log('[Chat Service] Envoi DM à:', userId, content);
      return chatWebSocket.sendDirectMessage(userId, content);
    } catch (error) {
      console.error('[Chat Service] Erreur envoi DM:', error);
      return false;
    }
  }

  /**
   * Marquer messages comme lus
   */
  async markMessagesAsRead(userId: number): Promise<boolean> {
    try {
      console.log('[Chat Service] Marquage messages lus:', userId);
      const success = await chatApi.markMessagesAsRead(userId);
      
      if (success) {
        // Mettre à jour le store
        const chatStore = useChatStore.getState();
        chatStore.updateConversationUnread(userId, 0);
        this.loadUnreadCount(); // Recharger le compteur global
      }
      
      return success;
    } catch (error) {
      console.error('[Chat Service] Erreur marquage messages lus:', error);
      return false;
    }
  }

  /**
   * Créer une nouvelle room
   */
  async createRoom(name: string, description?: string, isPrivate: boolean = false): Promise<boolean> {
    try {
      console.log('[Chat Service] Création room:', name);
      
      const room = await chatApi.createRoom({
        name,
        description,
        is_private: isPrivate
      });
      
      if (room) {
        // Recharger la liste des rooms
        await this.loadRooms();
        return true;
      }
      
      return false;
    } catch (error) {
      console.error('[Chat Service] Erreur création room:', error);
      return false;
    }
  }

  /**
   * Récupérer utilisateurs pour DM
   */
  async getUsersForDM(): Promise<BackendUser[]> {
    try {
      console.log('[Chat Service] Récupération utilisateurs pour DM...');
      const users = await chatApi.getUsersForDM();
      console.log('[Chat Service] Utilisateurs DM récupérés:', users.length);
      return users;
    } catch (error) {
      console.error('[Chat Service] Erreur récupération utilisateurs DM:', error);
      return [];
    }
  }

  /**
   * Réinitialiser le chat
   */
  reset(): void {
    console.log('[Chat Service] Réinitialisation...');
    
    // Déconnecter WebSocket
    chatWebSocket.disconnect();
    
    // Réinitialiser le store
    const chatStore = useChatStore.getState();
    chatStore.reset();
    
    // Réinitialiser les flags
    this.isInitialized = false;
    this.initPromise = null;
  }

  /**
   * Nettoyage à la fermeture de l'application
   */
  cleanup(): void {
    try {
      console.log('[Chat Service] Nettoyage en cours...');
      
      // Déconnecter le WebSocket
      chatWebSocket.disconnect();
      
      // Réinitialiser le service
      this.reset();
      
      console.log('[Chat Service] Nettoyage terminé');
    } catch (error) {
      console.error('[Chat Service] Erreur lors du nettoyage:', error);
    }
  }

  /**
   * Obtenir l'état de connexion
   */
  isConnected(): boolean {
    return chatWebSocket.isConnected();
  }
}

// Instance singleton
export const chatService = new ChatService(); 