// Types pour le chat compatibles avec le serveur Rust

export interface Message {
  id: string;                    // ID du message (converti en string pour React)
  content: string;               // Contenu du message
  senderId?: string;             // Pour compatibilité React
  fromUser?: number;             // ID utilisateur depuis Rust
  sender: {
    id: string;                  // Converti en string pour React
    username?: string;
    first_name?: any;
    last_name?: any;
  };
  roomId?: string;               // Pour compatibilité React
  room?: string;                 // Nom du salon depuis Rust
  receiverId?: string;           // Pour compatibilité React
  to?: number;                   // ID destinataire depuis Rust
  type: 'text' | 'image' | 'file' | 'system';
  createdAt: string;             // Pour compatibilité React
  timestamp?: string;            // Timestamp depuis Rust
  username?: string;             // Nom d'utilisateur depuis Rust
}

export interface Room {
  id: string;                    // Pour compatibilité React
  name: string;                  // Nom du salon
  description?: string;
  isPrivate: boolean;
  members: any[];
  memberCount?: number;
  user_count?: number;           // Depuis Rust
  lastMessage?: Message;
  createdAt: string;
  updatedAt: string;
}

export interface Conversation {
  userId: string;                // Converti en string pour React
  username: string;
  displayName?: string;          // Nom d'affichage
  first_name?: any;
  last_name?: any;
  avatar?: any;
  avatarUrl?: string;            // URL de l'avatar
  lastMessage?: Message;
  lastMessageTime?: string | null;
  unreadCount: number;
  isOnline?: boolean;
  lastSeen?: string;
}

// Types spécifiques au serveur Rust
export interface RustUser {
  id: number;
  username: string;
  email?: string;
  isOnline?: boolean;
}

export interface RustChatMessage {
  id: number;
  fromUser?: number;
  to?: number;                   // Pour les DM
  username: string;
  content: string;
  timestamp: string;
  room?: string;                 // Pour les messages de salon
}

// Stats du chat
export interface ChatStats {
  totalRooms: number;
  activeUsers: number;
  todayMessages: number;
}

export interface DMStats {
  totalMessages: number;
  todayMessages: number;
}

// Utilitaires de conversion entre formats React et Rust
export const convertRustMessageToReactMessage = (rustMsg: RustChatMessage, currentUserId?: number): Message => {
  return {
    id: rustMsg.id.toString(),
    content: rustMsg.content,
    senderId: rustMsg.fromUser?.toString() || '',
    fromUser: rustMsg.fromUser,
    sender: {
      id: rustMsg.fromUser?.toString() || '',
      username: rustMsg.username,
    },
    roomId: rustMsg.room,
    room: rustMsg.room,
    receiverId: rustMsg.to?.toString(),
    to: rustMsg.to,
    type: 'text',
    createdAt: rustMsg.timestamp,
    timestamp: rustMsg.timestamp,
    username: rustMsg.username,
  };
};

export const convertRustUserToConversation = (rustUser: RustUser): Conversation => {
  return {
    userId: rustUser.id.toString(),
    username: rustUser.username,
    first_name: null,
    last_name: null,
    avatar: null,
    unreadCount: 0,
    isOnline: rustUser.isOnline || false,
  };
}; 