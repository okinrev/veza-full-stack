/**
 * Service WebSocket Unifié Talas
 * Gère les connexions WebSocket avec authentification JWT unifiée
 * Compatible avec Chat Server Rust et Stream Server Rust
 */

interface WebSocketConfig {
  url: string;
  protocols?: string | string[];
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  heartbeatInterval?: number;
}

interface AuthenticatedMessage {
  type: 'auth';
  token: string;
  service: string;
  timestamp: number;
}

interface HeartbeatMessage {
  type: 'heartbeat';
  timestamp: number;
}

type WebSocketMessage = AuthenticatedMessage | HeartbeatMessage | any;

interface WebSocketEventListeners {
  onOpen?: (event: Event) => void;
  onMessage?: (data: any) => void;
  onClose?: (event: CloseEvent) => void;
  onError?: (event: Event) => void;
  onReconnect?: (attempt: number) => void;
  onAuthSuccess?: () => void;
  onAuthFailure?: (error: string) => void;
}

export class TalasWebSocketService {
  private ws: WebSocket | null = null;
  private config: WebSocketConfig;
  private listeners: WebSocketEventListeners;
  private isAuthenticated = false;
  private reconnectAttempts = 0;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private heartbeatTimer: NodeJS.Timeout | null = null;
  private serviceName: string;

  constructor(
    config: WebSocketConfig, 
    listeners: WebSocketEventListeners,
    serviceName: 'chat' | 'stream'
  ) {
    this.config = {
      reconnectInterval: 5000,
      maxReconnectAttempts: 10,
      heartbeatInterval: 30000,
      ...config
    };
    this.listeners = listeners;
    this.serviceName = serviceName;
  }

  /**
   * Établit la connexion WebSocket avec authentification JWT
   */
  public connect(): void {
    try {
      console.log(`🔌 [Talas WebSocket] Connexion au service ${this.serviceName}: ${this.config.url}`);
      
      this.ws = new WebSocket(this.config.url, this.config.protocols);
      this.setupEventListeners();
    } catch (error) {
      console.error(`❌ [Talas WebSocket] Erreur de connexion ${this.serviceName}:`, error);
      this.handleReconnect();
    }
  }

  /**
   * Ferme la connexion WebSocket
   */
  public disconnect(): void {
    console.log(`🔌 [Talas WebSocket] Déconnexion du service ${this.serviceName}`);
    
    this.clearTimers();
    this.isAuthenticated = false;
    this.reconnectAttempts = 0;

    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
  }

  /**
   * Envoie un message via WebSocket (avec vérification d'authentification)
   */
  public send(message: any): boolean {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn(`⚠️ [Talas WebSocket] ${this.serviceName} non connecté`);
      return false;
    }

    if (!this.isAuthenticated) {
      console.warn(`⚠️ [Talas WebSocket] ${this.serviceName} non authentifié`);
      return false;
    }

    try {
      this.ws.send(JSON.stringify(message));
      return true;
    } catch (error) {
      console.error(`❌ [Talas WebSocket] Erreur envoi message ${this.serviceName}:`, error);
      return false;
    }
  }

  /**
   * Vérifie si la connexion est active et authentifiée
   */
  public isConnected(): boolean {
    return this.ws !== null && 
           this.ws.readyState === WebSocket.OPEN && 
           this.isAuthenticated;
  }

  /**
   * Configure les événements WebSocket
   */
  private setupEventListeners(): void {
    if (!this.ws) return;

    this.ws.onopen = (event) => {
      console.log(`✅ [Talas WebSocket] Connexion ${this.serviceName} établie`);
      this.reconnectAttempts = 0;
      this.authenticate();
      this.listeners.onOpen?.(event);
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        this.handleMessage(data);
      } catch (error) {
        console.error(`❌ [Talas WebSocket] Erreur parsing message ${this.serviceName}:`, error);
      }
    };

    this.ws.onclose = (event) => {
      console.log(`🔌 [Talas WebSocket] Connexion ${this.serviceName} fermée:`, event.code, event.reason);
      this.isAuthenticated = false;
      this.clearTimers();
      this.listeners.onClose?.(event);
      
      // Reconnexion automatique si pas une fermeture volontaire
      if (event.code !== 1000) {
        this.handleReconnect();
      }
    };

    this.ws.onerror = (event) => {
      console.error(`❌ [Talas WebSocket] Erreur ${this.serviceName}:`, event);
      this.listeners.onError?.(event);
    };
  }

  /**
   * Authentification JWT unifiée
   */
  private authenticate(): void {
    const token = localStorage.getItem('access_token');
    
    if (!token) {
      console.error(`❌ [Talas WebSocket] Token d'authentification manquant pour ${this.serviceName}`);
      this.listeners.onAuthFailure?.('Token manquant');
      return;
    }

    const authMessage: AuthenticatedMessage = {
      type: 'auth',
      token,
      service: this.serviceName,
      timestamp: Date.now()
    };

    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(authMessage));
      console.log(`🔐 [Talas WebSocket] Authentification envoyée pour ${this.serviceName}`);
    }
  }

  /**
   * Traite les messages reçus
   */
  private handleMessage(data: any): void {
    // Messages d'authentification
    if (data.type === 'auth_success') {
      this.isAuthenticated = true;
      this.startHeartbeat();
      console.log(`✅ [Talas WebSocket] Authentification ${this.serviceName} réussie`);
      this.listeners.onAuthSuccess?.();
      return;
    }

    if (data.type === 'auth_failure' || data.type === 'auth_error') {
      this.isAuthenticated = false;
      console.error(`❌ [Talas WebSocket] Authentification ${this.serviceName} échouée:`, data.message);
      this.listeners.onAuthFailure?.(data.message || 'Authentification échouée');
      return;
    }

    // Messages heartbeat
    if (data.type === 'pong') {
      // Réponse au heartbeat, rien à faire
      return;
    }

    // Transmettre les autres messages au listener
    this.listeners.onMessage?.(data);
  }

  /**
   * Démarre le système de heartbeat
   */
  private startHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
    }

    this.heartbeatTimer = setInterval(() => {
      if (this.isConnected()) {
        const heartbeat: HeartbeatMessage = {
          type: 'heartbeat',
          timestamp: Date.now()
        };
        this.send(heartbeat);
      }
    }, this.config.heartbeatInterval);
  }

  /**
   * Gère la reconnexion automatique
   */
  private handleReconnect(): void {
    if (this.reconnectAttempts >= (this.config.maxReconnectAttempts || 10)) {
      console.error(`❌ [Talas WebSocket] Nombre maximum de tentatives de reconnexion atteint pour ${this.serviceName}`);
      return;
    }

    this.reconnectAttempts++;
    console.log(`🔄 [Talas WebSocket] Tentative de reconnexion ${this.serviceName} (${this.reconnectAttempts}/${this.config.maxReconnectAttempts})`);

    this.reconnectTimer = setTimeout(() => {
      this.listeners.onReconnect?.(this.reconnectAttempts);
      this.connect();
    }, this.config.reconnectInterval);
  }

  /**
   * Nettoie les timers
   */
  private clearTimers(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }
}

/**
 * Factory pour créer les services WebSocket Talas
 */
export class TalasWebSocketFactory {
  /**
   * Crée un service WebSocket pour le Chat
   */
  static createChatService(listeners: WebSocketEventListeners): TalasWebSocketService {
    const chatUrl = import.meta.env.VITE_WS_CHAT_URL || 'ws://localhost:3001/ws';
    
    return new TalasWebSocketService(
      {
        url: chatUrl,
        reconnectInterval: 3000,
        maxReconnectAttempts: 15,
        heartbeatInterval: 30000
      },
      listeners,
      'chat'
    );
  }

  /**
   * Crée un service WebSocket pour le Stream
   */
  static createStreamService(listeners: WebSocketEventListeners): TalasWebSocketService {
    const streamUrl = import.meta.env.VITE_WS_STREAM_URL || 'ws://localhost:3002/ws';
    
    return new TalasWebSocketService(
      {
        url: streamUrl,
        reconnectInterval: 3000,
        maxReconnectAttempts: 15,
        heartbeatInterval: 30000
      },
      listeners,
      'stream'
    );
  }
}

/**
 * Service central pour gérer toutes les connexions WebSocket Talas
 */
export class TalasCentralWebSocketManager {
  private chatService: TalasWebSocketService | null = null;
  private streamService: TalasWebSocketService | null = null;
  private isInitialized = false;

  /**
   * Initialise tous les services WebSocket avec authentification unifiée
   */
  public initialize(
    chatListeners?: WebSocketEventListeners,
    streamListeners?: WebSocketEventListeners
  ): void {
    console.log('🚀 [Talas Central WebSocket] Initialisation des services WebSocket...');

    // Service Chat
    if (chatListeners) {
      this.chatService = TalasWebSocketFactory.createChatService({
        ...chatListeners,
        onAuthFailure: (error) => {
          console.error('❌ [Talas Chat] Authentification échouée:', error);
          chatListeners.onAuthFailure?.(error);
          // Possibilité de redirection vers login
          this.handleAuthFailure();
        }
      });
    }

    // Service Stream
    if (streamListeners) {
      this.streamService = TalasWebSocketFactory.createStreamService({
        ...streamListeners,
        onAuthFailure: (error) => {
          console.error('❌ [Talas Stream] Authentification échouée:', error);
          streamListeners.onAuthFailure?.(error);
          this.handleAuthFailure();
        }
      });
    }

    this.isInitialized = true;
  }

  /**
   * Connecte tous les services initialisés
   */
  public connectAll(): void {
    if (!this.isInitialized) {
      console.warn('⚠️ [Talas Central WebSocket] Services non initialisés');
      return;
    }

    console.log('🔌 [Talas Central WebSocket] Connexion de tous les services...');

    this.chatService?.connect();
    this.streamService?.connect();
  }

  /**
   * Déconnecte tous les services
   */
  public disconnectAll(): void {
    console.log('🔌 [Talas Central WebSocket] Déconnexion de tous les services...');

    this.chatService?.disconnect();
    this.streamService?.disconnect();
  }

  /**
   * Gère les échecs d'authentification
   */
  private handleAuthFailure(): void {
    // Déconnecter tous les services en cas d'échec d'auth
    this.disconnectAll();
    
    // Redirection possible vers la page de login
    console.log('🔒 [Talas Central WebSocket] Authentification échouée, redirection vers login...');
    window.location.href = '/login';
  }

  /**
   * Getters pour accéder aux services
   */
  public getChatService(): TalasWebSocketService | null {
    return this.chatService;
  }

  public getStreamService(): TalasWebSocketService | null {
    return this.streamService;
  }

  /**
   * Vérifie si tous les services sont connectés et authentifiés
   */
  public areAllServicesConnected(): boolean {
    const chatConnected = !this.chatService || this.chatService.isConnected();
    const streamConnected = !this.streamService || this.streamService.isConnected();
    
    return chatConnected && streamConnected;
  }
}

// Instance globale du gestionnaire central
export const talasWebSocketManager = new TalasCentralWebSocketManager(); 