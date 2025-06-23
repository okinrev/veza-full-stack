import { apiClient } from '../../../shared/api/client';
import { ENDPOINTS } from '../../../shared/api/endpoints';
import type { ApiResponse } from '../../../shared/api/types';

// Types conformes à la documentation backend
export interface BackendUser {
  id: number;
  username: string;
  first_name: string;
  last_name: string;
  avatar: string;
  is_online: boolean;
  last_seen: string;
}

export interface BackendConversation {
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

export interface BackendMessage {
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

export interface BackendRoom {
  id: number;
  name: string;
  description: string;
  is_private: boolean;
  creator_id: number;
  creator_name: string;
  created_at: string;
  updated_at: string;
}

export interface BackendUnreadCount {
  count: number;
}

// API Chat selon la documentation backend
export const chatApi = {
  /**
   * Récupérer conversations DM
   * GET /api/v1/chat/conversations
   */
  async getConversations(): Promise<BackendConversation[]> {
    try {
      console.log('[Chat API] Récupération des conversations...');
      const response = await apiClient.get(ENDPOINTS.CHAT_CONVERSATIONS) as any;
      console.log('[Chat API] Conversations reçues:', response);
      
      // Le backend retourne { conversations: [...] }
      if (response && Array.isArray(response.conversations)) {
        return response.conversations;
      }
      
      return [];
    } catch (error) {
      console.error('[Chat API] Erreur récupération conversations:', error);
      return [];
    }
  },

  /**
   * Récupérer rooms publiques
   * GET /api/v1/chat/rooms
   */
  async getRooms(): Promise<BackendRoom[]> {
    try {
      console.log('[Chat API] Récupération des rooms...');
      const response = await apiClient.get(ENDPOINTS.CHAT_ROOMS) as any;
      console.log('[Chat API] Rooms reçues:', response);
      
      // Le backend retourne { rooms: [...] }
      if (response && Array.isArray(response.rooms)) {
        return response.rooms;
      }
      
      return [];
    } catch (error) {
      console.error('[Chat API] Erreur récupération rooms:', error);
      return [];
    }
  },

  /**
   * Créer une room
   * POST /api/v1/chat/rooms
   */
  async createRoom(data: {
    name: string;
    description?: string;
    is_private?: boolean;
  }): Promise<BackendRoom | null> {
    try {
      console.log('[Chat API] Création room:', data);
      const response = await apiClient.post(ENDPOINTS.CHAT_CREATE_ROOM, data) as ApiResponse<BackendRoom>;
      console.log('[Chat API] Room créée:', response.data);
      
      if (response.success && response.data) {
        return response.data;
      }
      
      return null;
    } catch (error) {
      console.error('[Chat API] Erreur création room:', error);
      throw error;
    }
  },

  /**
   * Récupérer messages d'une room
   * GET /api/v1/chat/rooms/{room_name}/messages?limit=50
   */
  async getRoomMessages(roomName: string, limit: number = 50): Promise<BackendMessage[]> {
    try {
      console.log('[Chat API] Récupération messages room:', roomName);
      const response = await apiClient.get(ENDPOINTS.CHAT_ROOM_MESSAGES(roomName), {
        params: { limit }
      }) as ApiResponse<BackendMessage[]>;
      console.log('[Chat API] Messages room reçus:', response.data);
      
      if (response.success && Array.isArray(response.data)) {
        return response.data;
      }
      
      return [];
    } catch (error) {
      console.error('[Chat API] Erreur récupération messages room:', error);
      return [];
    }
  },

  /**
   * Récupérer messages directs avec un utilisateur
   * GET /api/v1/chat/dm/{user_id}?page=1&limit=50
   */
  async getDirectMessages(userId: number, page: number = 1, limit: number = 50): Promise<BackendMessage[]> {
    try {
      console.log('[Chat API] Récupération DM avec utilisateur:', userId);
      const response = await apiClient.get(ENDPOINTS.CHAT_DM(userId), {
        params: { page, limit }
      }) as ApiResponse<BackendMessage[]>;
      console.log('[Chat API] Messages DM reçus:', response.data);
      
      if (response.success && Array.isArray(response.data)) {
        return response.data;
      }
      
      return [];
    } catch (error) {
      console.error('[Chat API] Erreur récupération DM:', error);
      return [];
    }
  },

  /**
   * Envoyer message direct
   * POST /api/v1/chat/dm/{user_id}
   */
  async sendDirectMessage(userId: number, content: string): Promise<BackendMessage | null> {
    try {
      console.log('[Chat API] Envoi DM à utilisateur:', userId, content);
      const response = await apiClient.post(ENDPOINTS.CHAT_DM(userId), {
        content
      }) as ApiResponse<BackendMessage>;
      console.log('[Chat API] DM envoyé:', response.data);
      
      if (response.success && response.data) {
        return response.data;
      }
      
      return null;
    } catch (error) {
      console.error('[Chat API] Erreur envoi DM:', error);
      throw error;
    }
  },

  /**
   * Marquer messages comme lus
   * PUT /api/v1/chat/messages/{user_id}/read
   */
  async markMessagesAsRead(userId: number): Promise<boolean> {
    try {
      console.log('[Chat API] Marquage messages comme lus:', userId);
      const response = await apiClient.put(ENDPOINTS.CHAT_DM_READ(userId)) as ApiResponse<{ message: string }>;
      console.log('[Chat API] Messages marqués comme lus:', response);
      
      return response.success === true;
    } catch (error) {
      console.error('[Chat API] Erreur marquage messages lus:', error);  
      return false;
    }
  },

  /**
   * Compter messages non lus
   * GET /api/v1/chat/unread
   */
  async getUnreadCount(): Promise<number> {
    try {
      console.log('[Chat API] Récupération compteur non lus...');
      const response = await apiClient.get(ENDPOINTS.CHAT_UNREAD) as any;
      console.log('[Chat API] Compteur non lus reçu:', response);
      
      // Le backend retourne { unread_count: 0, user_id: 16 }
      if (response && typeof response.unread_count === 'number') {
        return response.unread_count;
      }
      
      return 0;
    } catch (error) {
      console.error('[Chat API] Erreur compteur non lus:', error);
      return 0;
    }
  },

  /**
   * Récupérer utilisateurs pour DM
   * GET /api/v1/users/except-me
   */
  async getUsersForDM(): Promise<BackendUser[]> {
    try {
      console.log('[Chat API] Récupération utilisateurs pour DM...');
      const response = await apiClient.get(ENDPOINTS.CHAT_USERS) as ApiResponse<BackendUser[]>;
      console.log('[Chat API] Utilisateurs DM reçus:', response.data);
      
      if (response.success && Array.isArray(response.data)) {
        return response.data;
      }
      
      return [];
    } catch (error) {
      console.error('[Chat API] Erreur récupération utilisateurs:', error);
      return [];
    }
  },
}; 