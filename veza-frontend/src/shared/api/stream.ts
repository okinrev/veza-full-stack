import { useAuthStore } from "@/shared/stores/authStore";

export class StreamWebSocketClient {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectTimeout = 1000;
  private messageHandlers: Map<string, ((data: any) => void)[]> = new Map();

  constructor(private url: string) {}

  connect() {
    const { user } = useAuthStore.getState();
    if (!user?.token) return;

    this.ws = new WebSocket(`${this.url}?token=${user.token}`);

    this.ws.onopen = () => {
      console.log('Stream WebSocket connected');
      this.reconnectAttempts = 0;
    };

    this.ws.onclose = () => {
      console.log('Stream WebSocket disconnected');
      this.handleReconnect();
    };

    this.ws.onerror = (error) => {
      console.error('Stream WebSocket error:', error);
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        this.handleMessage(data);
      } catch (error) {
        console.error('Error parsing Stream WebSocket message:', error);
      }
    };
  }

  private handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      setTimeout(() => {
        console.log(`Attempting to reconnect stream (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
        this.connect();
      }, this.reconnectTimeout * this.reconnectAttempts);
    }
  }

  private handleMessage(data: any) {
    const handlers = this.messageHandlers.get(data.type) || [];
    handlers.forEach(handler => handler(data));
  }

  subscribe(type: string, handler: (data: any) => void) {
    const handlers = this.messageHandlers.get(type) || [];
    handlers.push(handler);
    this.messageHandlers.set(type, handlers);
  }

  unsubscribe(type: string, handler: (data: any) => void) {
    const handlers = this.messageHandlers.get(type) || [];
    const index = handlers.indexOf(handler);
    if (index !== -1) {
      handlers.splice(index, 1);
      this.messageHandlers.set(type, handlers);
    }
  }

  send(data: any) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    } else {
      console.error('Stream WebSocket is not connected');
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

export const streamWsClient = new StreamWebSocketClient('ws://localhost:8082/stream'); 