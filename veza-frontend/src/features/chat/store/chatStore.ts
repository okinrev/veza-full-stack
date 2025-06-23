import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

// Types pour le chat selon la documentation backend
export interface ChatUser {
  id: number;
  username: string;
  first_name: string;
  last_name: string;
  avatar: string;
  is_online: boolean;
  last_seen: string;
}

export interface ChatMessage {
  id: number;
  from_user: number;
  to_user?: number;
  room?: string;
  content: string;
  is_read: boolean;
  timestamp: string;
  username: string;
  avatar?: string;
}

export interface ChatRoom {
  id: number;
  name: string;
  description: string;
  is_private: boolean;
  creator_id: number;
  creator_name: string;
  created_at: string;
  updated_at: string;
}

export interface Conversation {
  user_id: number;
  username: string;
  first_name: string;
  last_name: string;
  avatar: string;
  last_message: string;
  last_activity: string;
  unread_count: number;
  is_online: boolean;
  last_seen: string;
}

// État global selon la documentation
interface ChatState {
  // Authentification
  user: {
    id: number | null;
    username: string;
    avatar: string;
  };
  isAuthenticated: boolean;
  
  // Connexions
  socket: WebSocket | null;
  isConnected: boolean;
  reconnectAttempts: number;
  
  // Messages Directs
  conversations: Conversation[];
  currentDMUser: ChatUser | null;
  dmMessages: ChatMessage[];
  totalUnreadDM: number;
  
  // Rooms
  rooms: ChatRoom[];
  currentRoom: string;
  roomMessages: ChatMessage[];
  
  // UI
  activeTab: 'rooms' | 'dm';
  isLoading: boolean;
  error: string | null;

  // Actions - Authentification
  setUser: (user: { id: number; username: string; avatar: string }) => void;
  setAuthenticated: (authenticated: boolean) => void;
  
  // Actions - Connexion
  setSocket: (socket: WebSocket | null) => void;
  setConnected: (connected: boolean) => void;
  incrementReconnectAttempts: () => void;
  resetReconnectAttempts: () => void;
  
  // Actions - Messages Directs
  setConversations: (conversations: Conversation[]) => void;
  setCurrentDMUser: (user: ChatUser | null) => void;
  setDmMessages: (messages: ChatMessage[]) => void;
  addDmMessage: (message: ChatMessage) => void;
  setTotalUnreadDM: (count: number) => void;
  updateConversationUnread: (userId: number, count: number) => void;
  
  // Actions - Rooms
  setRooms: (rooms: ChatRoom[]) => void;
  setCurrentRoom: (room: string) => void;
  setRoomMessages: (messages: ChatMessage[]) => void;
  addRoomMessage: (message: ChatMessage) => void;
  
  // Actions - UI
  setActiveTab: (tab: 'rooms' | 'dm') => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  
  // Actions - Utils
  reset: () => void;
}

const initialState = {
  user: {
    id: null,
    username: '',
    avatar: ''
  },
  isAuthenticated: false,
  socket: null,
  isConnected: false,
  reconnectAttempts: 0,
  conversations: [],
  currentDMUser: null,
  dmMessages: [],
  totalUnreadDM: 0,
  rooms: [],
  currentRoom: '',
  roomMessages: [],
  activeTab: 'rooms' as const,
  isLoading: false,
  error: null,
};

export const useChatStore = create<ChatState>()(
  devtools(
    immer((set) => ({
      ...initialState,
      
      // Actions - Authentification
      setUser: (user) =>
        set((state) => {
          state.user = user;
        }),
        
      setAuthenticated: (authenticated) =>
        set((state) => {
          state.isAuthenticated = authenticated;
        }),
      
      // Actions - Connexion
      setSocket: (socket) =>
        set((state) => {
          state.socket = socket;
        }),
        
      setConnected: (connected) =>
        set((state) => {
          state.isConnected = connected;
        }),
        
      incrementReconnectAttempts: () =>
        set((state) => {
          state.reconnectAttempts += 1;
        }),
        
      resetReconnectAttempts: () =>
        set((state) => {
          state.reconnectAttempts = 0;
        }),
      
      // Actions - Messages Directs
      setConversations: (conversations) =>
        set((state) => {
          state.conversations = conversations;
        }),
        
      setCurrentDMUser: (user) =>
        set((state) => {
          state.currentDMUser = user;
        }),
        
      setDmMessages: (messages) =>
        set((state) => {
          state.dmMessages = messages;
        }),
        
      addDmMessage: (message) =>
        set((state) => {
          state.dmMessages.push(message);
        }),
        
      setTotalUnreadDM: (count) =>
        set((state) => {
          state.totalUnreadDM = count;
        }),
        
      updateConversationUnread: (userId, count) =>
        set((state) => {
          const conversation = state.conversations.find(c => c.user_id === userId);
          if (conversation) {
            conversation.unread_count = count;
          }
        }),
      
      // Actions - Rooms
      setRooms: (rooms) =>
        set((state) => {
          state.rooms = rooms;
        }),
        
      setCurrentRoom: (room) =>
        set((state) => {
          state.currentRoom = room;
        }),
        
      setRoomMessages: (messages) =>
        set((state) => {
          state.roomMessages = messages;
        }),
        
      addRoomMessage: (message) =>
        set((state) => {
          state.roomMessages.push(message);
        }),
      
      // Actions - UI
      setActiveTab: (tab) =>
        set((state) => {
          state.activeTab = tab;
        }),
        
      setLoading: (loading) =>
        set((state) => {
          state.isLoading = loading;
        }),
        
      setError: (error) =>
        set((state) => {
          state.error = error;
        }),
      
      // Actions - Utils
      reset: () =>
        set((state) => {
          Object.assign(state, initialState);
        }),
    }))
  )
); 