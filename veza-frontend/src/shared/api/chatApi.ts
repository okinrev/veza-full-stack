import { apiClient } from './client';

// Types compatibles avec le backend
export interface ChatRoom {
  id: number;
  name: string;
  description?: string;
  is_private: boolean;
  creator_id?: number;
  member_count?: number;
  created_at: string;
  updated_at?: string;
}

export interface ChatMessage {
  id: number;
  content: string;
  sender_id: number;
  room_id?: number;
  receiver_id?: number;
  username?: string;
  sender_username?: string;
  type?: 'text' | 'image' | 'file';
  created_at: string;
  timestamp?: string;
  is_read?: boolean;
}

export interface DirectMessage {
  id: number;
  content: string;
  sender_id: number;
  receiver_id: number;
  sender_username?: string;
  receiver_username?: string;
  is_read: boolean;
  created_at: string;
}

export interface ChatUser {
  id: number;
  username: string;
  email: string;
  first_name?: {
    String: string;
    Valid: boolean;
  } | string;
  last_name?: {
    String: string;
    Valid: boolean;
  } | string;
  avatar?: {
    String: string;
    Valid: boolean;
  } | string;
  role: string;
  is_active: boolean;
  created_at: string;
}

export interface CreateRoomData {
  name: string;
  description?: string;
  is_private?: boolean;
}

export interface SendMessageData {
  content: string;
  room_id?: string;
  receiver_id?: string;
}

export interface MessageHistory {
  messages: ChatMessage[];
  total: number;
  page: number;
  limit: number;
}

export class ChatApiService {
  // Gestion des salons
  async getRooms(): Promise<ChatRoom[]> {
    return apiClient.get<ChatRoom[]>('/api/v1/rooms');
  }

  async createRoom(data: CreateRoomData): Promise<ChatRoom> {
    return apiClient.post<ChatRoom>('/api/v1/rooms', data);
  }

  async joinRoom(roomId: string): Promise<void> {
    return apiClient.post(`/api/v1/rooms/${roomId}/join`);
  }

  async leaveRoom(roomId: string): Promise<void> {
    return apiClient.post(`/api/v1/rooms/${roomId}/leave`);
  }

  async getRoomMembers(roomId: string): Promise<ChatUser[]> {
    return apiClient.get<ChatUser[]>(`/api/v1/rooms/${roomId}/members`);
  }

  // Gestion des messages
  async getMessageHistory(roomId: string, page = 1, limit = 50): Promise<MessageHistory> {
    return apiClient.get<MessageHistory>(`/api/v1/rooms/${roomId}/messages`, {
      params: { page, limit }
    });
  }

  async sendMessage(data: SendMessageData): Promise<ChatMessage> {
    return apiClient.post<ChatMessage>('/api/v1/messages', data);
  }

  async getDirectMessages(userId: string, page = 1, limit = 50): Promise<MessageHistory> {
    return apiClient.get<MessageHistory>(`/api/v1/messages/conversations/${userId}`, {
      params: { page, limit }
    });
  }

  // Gestion des utilisateurs
  async getUsers(): Promise<ChatUser[]> {
    return apiClient.get<ChatUser[]>('/api/v1/users');
  }

  async getUsersExceptMe(): Promise<ChatUser[]> {
    return apiClient.get<ChatUser[]>('/api/v1/users/except-me');
  }

  async getOnlineUsers(): Promise<ChatUser[]> {
    return apiClient.get<ChatUser[]>('/api/v1/chat/online-users');
  }

  // Statistiques
  async getChatStats(): Promise<{
    totalRooms: number;
    activeUsers: number;
    todayMessages: number;
  }> {
    return apiClient.get('/api/v1/chat/stats');
  }

  // Utilitaires
  async getUser(userId: number): Promise<ChatUser | null> {
    try {
      return await apiClient.get<ChatUser>(`/api/v1/users/${userId}`);
    } catch (error) {
      console.error('Erreur lors de la récupération de l\'utilisateur:', error);
      return null;
    }
  }

  async createRoomWithData(name: string, description?: string): Promise<ChatRoom | null> {
    try {
      return await this.createRoom({ name, description, is_private: false });
    } catch (error) {
      console.error('Erreur lors de la création du salon:', error);
      return null;
    }
  }

  // Messages privés
  async sendDirectMessage(toUserId: number, content: string): Promise<DirectMessage> {
    return apiClient.post<DirectMessage>('/api/v1/messages/direct', {
      to_user_id: toUserId,
      content
    });
  }

  async getDirectMessageHistory(withUserId: number, limit = 50): Promise<DirectMessage[]> {
    return apiClient.get<DirectMessage[]>(`/api/v1/messages/direct/${withUserId}`, {
      params: { limit }
    });
  }

  async markMessagesAsRead(userId: number): Promise<void> {
    return apiClient.put(`/api/v1/messages/direct/${userId}/read`);
  }

  // Recherche
  async searchUsers(query: string): Promise<ChatUser[]> {
    return apiClient.get<ChatUser[]>('/api/v1/search/users', {
      params: { q: query }
    });
  }

  async searchMessages(query: string, roomId?: string): Promise<ChatMessage[]> {
    const params: any = { q: query };
    if (roomId) {
      params.room_id = roomId;
    }
    return apiClient.get<ChatMessage[]>('/api/v1/search/messages', { params });
  }
}

// Instance singleton
export const chatApi = new ChatApiService();
export default chatApi; 