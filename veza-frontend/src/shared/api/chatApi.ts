import { api } from './client';

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
    return api.get<ChatRoom[]>('/rooms');
  }

  async createRoom(data: CreateRoomData): Promise<ChatRoom> {
    return api.post<ChatRoom>('/rooms', data);
  }

  async joinRoom(roomId: string): Promise<void> {
    return api.post(`/rooms/${roomId}/join`);
  }

  async leaveRoom(roomId: string): Promise<void> {
    return api.post(`/rooms/${roomId}/leave`);
  }

  async getRoomMembers(roomId: string): Promise<ChatUser[]> {
    return api.get<ChatUser[]>(`/rooms/${roomId}/members`);
  }

  // Gestion des messages
  async getMessageHistory(roomId: string, page = 1, limit = 50): Promise<MessageHistory> {
    return api.get<MessageHistory>(`/rooms/${roomId}/messages`, {
      page, limit
    });
  }

  async sendMessage(data: SendMessageData): Promise<ChatMessage> {
    // Si c'est un message de room, utiliser l'endpoint room
    if (data.room_id) {
      return api.post<ChatMessage>(`/rooms/${data.room_id}/messages`, {
        content: data.content
      });
    }
    // Si c'est un message direct, utiliser l'endpoint direct
    if (data.receiver_id) {
      return api.post<ChatMessage>('/messages/direct', {
        to_user_id: data.receiver_id,
        content: data.content
      });
    }
    throw new Error('room_id ou receiver_id requis pour envoyer un message');
  }

  async getDirectMessages(userId: string, page = 1, limit = 50): Promise<MessageHistory> {
    return api.get<MessageHistory>(`/messages/conversations/${userId}`, {
      page, limit
    });
  }

  // Gestion des utilisateurs
  async getUsers(): Promise<ChatUser[]> {
    return api.get<ChatUser[]>('/users');
  }

  async getUsersExceptMe(): Promise<ChatUser[]> {
    return api.get<ChatUser[]>('/users/except-me');
  }

  async getOnlineUsers(): Promise<ChatUser[]> {
    return api.get<ChatUser[]>('/chat/online-users');
  }

  // Statistiques
  async getChatStats(): Promise<{
    totalRooms: number;
    activeUsers: number;
    todayMessages: number;
  }> {
    return api.get('/chat/stats');
  }

  // Utilitaires
  async getUser(userId: number): Promise<ChatUser | null> {
    try {
      return await api.get<ChatUser>(`/users/${userId}`);
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
    return api.post<DirectMessage>('/messages/direct', {
      to_user_id: toUserId,
      content
    });
  }

  async getDirectMessageHistory(withUserId: number, limit = 50): Promise<DirectMessage[]> {
    return api.get<DirectMessage[]>(`/messages/direct/${withUserId}`, {
      limit
    });
  }

  async markMessagesAsRead(userId: number): Promise<void> {
    return api.put(`/messages/direct/${userId}/read`);
  }

  // Recherche
  async searchUsers(query: string): Promise<ChatUser[]> {
    return api.get<ChatUser[]>('/search/users', {
      q: query
    });
  }

  async searchMessages(query: string, roomId?: string): Promise<ChatMessage[]> {
    const params: any = { q: query };
    if (roomId) {
      params.room_id = roomId;
    }
    return api.get<ChatMessage[]>('/search/messages', params);
  }
}

// Instance singleton
export const chatApi = new ChatApiService();
export default chatApi; 