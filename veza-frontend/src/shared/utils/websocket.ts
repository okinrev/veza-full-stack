interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: number;
}

interface WebSocketOptions {
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onMessage?: (message: WebSocketMessage) => void;
  onError?: (error: Event) => void;
}

class WebSocketManager {
  private socket: WebSocket | null = null;
  private url: string;
  private options: WebSocketOptions;
  private reconnectAttempts = 0;
  private messageQueue: WebSocketMessage[] = [];
  private isConnected = false;
  private reconnectTimeout: NodeJS.Timeout | null = null;
  
  constructor(url: string, options: WebSocketOptions = {}) {
    this.url = url;
    this.options = {
      reconnectInterval: 3000,
      maxReconnectAttempts: 5,
      ...options,
    };
  }
  
  connect(token?: string) {
    try {
      const wsUrl = token ? `${this.url}?token=${token}` : this.url;
      this.socket = new WebSocket(wsUrl);
      
      this.socket.onopen = () => {
        console.log('WebSocket connecté');
        this.isConnected = true;
        this.reconnectAttempts = 0;
        this.options.onConnect?.();
        this.flushMessageQueue();
      };
      
      this.socket.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          this.options.onMessage?.(message);
        } catch (error) {
          console.error('Erreur parsing message WebSocket:', error);
        }
      };
      
      this.socket.onclose = () => {
        console.log('WebSocket déconnecté');
        this.isConnected = false;
        this.options.onDisconnect?.();
        this.attemptReconnect();
      };
      
      this.socket.onerror = (error) => {
        console.error('Erreur WebSocket:', error);
        this.options.onError?.(error);
      };
      
    } catch (error) {
      console.error('Erreur connexion WebSocket:', error);
    }
  }
  
  disconnect() {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    
    if (this.socket) {
      this.socket.close();
      this.socket = null;
      this.isConnected = false;
    }
  }
  
  sendMessage(type: string, data: any) {
    const message: WebSocketMessage = {
      type,
      data,
      timestamp: Date.now(),
    };
    
    if (this.isConnected && this.socket) {
      this.socket.send(JSON.stringify(message));
    } else {
      // Ajouter à la queue si pas connecté
      this.messageQueue.push(message);
    }
  }
  
  private flushMessageQueue() {
    while (this.messageQueue.length > 0 && this.isConnected) {
      const message = this.messageQueue.shift();
      if (message && this.socket) {
        this.socket.send(JSON.stringify(message));
      }
    }
  }
  
  private attemptReconnect() {
    if (this.reconnectAttempts < (this.options.maxReconnectAttempts || 5)) {
      this.reconnectAttempts++;
      console.log(`Tentative de reconnexion ${this.reconnectAttempts}/${this.options.maxReconnectAttempts}`);
      
      this.reconnectTimeout = setTimeout(() => {
        this.connect();
      }, this.options.reconnectInterval);
    } else {
      console.error('Nombre maximum de tentatives de reconnexion atteint');
    }
  }
  
  isSocketConnected(): boolean {
    return this.isConnected;
  }
}

export default WebSocketManager; 