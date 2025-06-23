import { create } from 'zustand';
import { createContext, useContext, ReactNode, useEffect, useCallback } from 'react';

interface WebSocketState {
  isConnected: boolean;
  connect: () => void;
  disconnect: () => void;
  sendMessage: (message: any) => void;
}

const useWebSocketStore = create<WebSocketState>((set, get) => ({
  isConnected: false,
  connect: () => {
    const ws = new WebSocket(import.meta.env.VITE_WS_URL || 'ws://localhost:8080');
    
    ws.onopen = () => {
      set({ isConnected: true });
    };

    ws.onclose = () => {
      set({ isConnected: false });
      // Tentative de reconnexion après 5 secondes
      setTimeout(() => get().connect(), 5000);
    };

    ws.onerror = (error) => {
      console.error('Erreur WebSocket:', error);
    };

    // Stocker l'instance WebSocket
    (window as any).__ws = ws;
  },
  disconnect: () => {
    const ws = (window as any).__ws;
    if (ws) {
      ws.close();
      (window as any).__ws = null;
    }
    set({ isConnected: false });
  },
  sendMessage: (message) => {
    const ws = (window as any).__ws;
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(message));
    } else {
      console.error('WebSocket non connecté');
    }
  },
}));

const WebSocketContext = createContext<WebSocketState | null>(null);

export function WebSocketProvider({ children }: { children: ReactNode }) {
  const ws = useWebSocketStore();

  useEffect(() => {
    ws.connect();
    return () => ws.disconnect();
  }, []);

  return <WebSocketContext.Provider value={ws}>{children}</WebSocketContext.Provider>;
}

export function useWebSocket() {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket doit être utilisé à l\'intérieur d\'un WebSocketProvider');
  }
  return context;
} 