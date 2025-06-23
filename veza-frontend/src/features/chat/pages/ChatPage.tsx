import React, { useState, useEffect, useRef } from 'react';
import { useAuthStore } from '@/features/auth/store/authStore';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Hash, Users, Plus, Search, Send, RefreshCw } from 'lucide-react';

interface ChatMessage {
  id: number;
  content: string;
  username?: string;
  fromUser?: number;
  to?: number;
  room?: string;
  timestamp: string;
  status?: 'sending' | 'sent' | 'delivered';
}

interface Room {
  id: number;
  name: string;
  description?: string;
  user_count?: number;
}

interface User {
  id: number;
  username: string;
  email?: string;
  isOnline?: boolean;
}

interface ConnectedUser {
  id: number;
  username: string;
}

export function ChatPage() {
  const { user, isAuthenticated, accessToken } = useAuthStore();
  
  // États principaux
  const [activeTab, setActiveTab] = useState<'rooms' | 'dm'>('rooms');
  const [isConnected, setIsConnected] = useState(false);
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [myUserId, setMyUserId] = useState<number | null>(null);
  
  // Messages
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [messageContent, setMessageContent] = useState('');
  const [loadingMessages, setLoadingMessages] = useState(false);
  const [sending, setSending] = useState(false);
  
  // Salons
  const [rooms, setRooms] = useState<Room[]>([]);
  const [currentRoom, setCurrentRoom] = useState('');
  const [connectedUsers, setConnectedUsers] = useState<ConnectedUser[]>([]);
  const [newRoomName, setNewRoomName] = useState('');
  const [creating, setCreating] = useState(false);
  
  // Messages privés
  const [allUsers, setAllUsers] = useState<User[]>([]);
  const [filteredUsers, setFilteredUsers] = useState<User[]>([]);
  const [userSearch, setUserSearch] = useState('');
  const [loadingUsers, setLoadingUsers] = useState(false);
  const [otherUserId, setOtherUserId] = useState<number | null>(null);
  const [otherUserInfo, setOtherUserInfo] = useState({
    username: '',
    isOnline: false
  });
  
  // Refs
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const socketRef = useRef<WebSocket | null>(null);

  // Initialisation
  useEffect(() => {
    if (!isAuthenticated || !user || !accessToken) return;

    const init = async () => {
      console.log('🔄 Chat: Initialisation...');
      
      await loadUserInfo();
      await loadRooms();
      
      if (activeTab === 'dm') {
        await loadUsers();
      }
      
      await connectWebSocket();
    };

    init();

    return () => {
      if (socketRef.current) {
        socketRef.current.onclose = null;
        socketRef.current.onerror = null;
        socketRef.current.close();
      }
    };
  }, [isAuthenticated, user, accessToken]);

  // Chargement des informations utilisateur
  const loadUserInfo = async () => {
    try {
      if (accessToken) {
        // Décoder le token JWT pour obtenir l'ID utilisateur
        const payload = JSON.parse(atob(accessToken.split('.')[1]));
        const userId = payload.user_id || payload.id || payload.sub;
        setMyUserId(userId);
        console.log(`👤 User ID: ${userId}, Username: ${user?.username}`);
      }
    } catch (error) {
      console.error('Erreur chargement utilisateur:', error);
    }
  };

  // Chargement des salons
  const loadRooms = async () => {
      try {
      console.log('[Chat] Chargement des salons...');
      const response = await fetch('/api/v1/chat/rooms', {
        headers: {
          'Authorization': `Bearer ${accessToken}`
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        setRooms(data.rooms || []);
        console.log(`[Chat] ${data.rooms?.length || 0} salons chargés`);
      }
    } catch (error) {
      console.error('[Chat] Erreur chargement salons:', error);
    }
  };

  // Chargement des utilisateurs
  const loadUsers = async () => {
    try {
      setLoadingUsers(true);
      console.log('[Chat] Chargement des utilisateurs...');
      
      // Pour l'instant, utiliser les conversations existantes
      const response = await fetch('/api/v1/chat/conversations', {
        headers: {
          'Authorization': `Bearer ${accessToken}`
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        const users = data.conversations?.map((conv: any) => ({
          id: conv.user_id,
          username: conv.username,
          email: conv.email || '',
          isOnline: conv.is_online || false
        })) || [];
        
        setAllUsers(users);
        setFilteredUsers(users);
        console.log(`[Chat] ${users.length} utilisateurs chargés`);
        }
      } catch (error) {
      console.error('[Chat] Erreur chargement utilisateurs:', error);
    } finally {
      setLoadingUsers(false);
    }
  };

  // Connexion WebSocket
  const connectWebSocket = async () => {
    try {
      if (socketRef.current) {
        socketRef.current.onclose = null;
        socketRef.current.onerror = null;
        socketRef.current.close();
      }

      console.log('🔌 Connexion WebSocket...');
      
      // Utiliser l'URL HAProxy pour WebSocket
      const wsUrl = `ws://10.5.191.133/ws/chat?token=${accessToken}`;
      const ws = new WebSocket(wsUrl);
      
      ws.onopen = () => {
        console.log('✅ WebSocket connecté');
        setIsConnected(true);
        setSocket(ws);
        socketRef.current = ws;
        
        // Si on est en mode DM avec un utilisateur sélectionné
        if (activeTab === 'dm' && otherUserId) {
          setLoadingMessages(true);
          ws.send(JSON.stringify({
            type: "dm_history",
            with: otherUserId,
            limit: 50
          }));
        }
      };

      ws.onclose = (event) => {
        console.log('❌ WebSocket fermé:', event.code);
        setIsConnected(false);
        setSocket(null);
        socketRef.current = null;
        
        // Reconnexion automatique si ce n'est pas une fermeture normale
        if (event.code !== 1000) {
          setTimeout(() => {
            if ((activeTab === 'rooms') || (activeTab === 'dm' && otherUserId)) {
              connectWebSocket();
            }
          }, 5000);
        }
      };

      ws.onerror = (error) => {
        console.error('❌ Erreur WebSocket:', error);
      };

      ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        handleWebSocketMessage(data);
      };

    } catch (error) {
      console.error('Erreur connexion WebSocket:', error);
    }
  };

  // Gestion des messages WebSocket
  const handleWebSocketMessage = (data: any) => {
    console.log("📥 WS reçu:", data);
    
    if (activeTab === 'rooms') {
      handleRoomMessage(data);
    } else if (activeTab === 'dm') {
      handleDMMessage(data);
    }
  };

  // Gestion des messages de salon
  const handleRoomMessage = (data: any) => {
    if (data.type === "message" && data.data?.room === currentRoom) {
      setMessages(prev => [...prev, data.data]);
      scrollToBottom();
    } else if (data.username && data.content) {
      if (data.room === currentRoom || !data.room) {
        setMessages(prev => [...prev, data]);
        scrollToBottom();
      }
    } else if (Array.isArray(data)) {
      const roomMessages = data
        .filter(m => m.room === currentRoom)
        .sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
      setMessages(roomMessages);
      setLoadingMessages(false);
      scrollToBottom();
    } else if (data.type === "room_history") {
      if (data.messages && Array.isArray(data.messages)) {
        const roomMessages = data.messages
          .filter(m => m.room === currentRoom)
          .sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
        setMessages(roomMessages);
      }
      setLoadingMessages(false);
      scrollToBottom();
    } else if (data.type === "user_joined") {
      if (data.user && !connectedUsers.find(u => u.id === data.user.id)) {
        setConnectedUsers(prev => [...prev, data.user]);
      }
    } else if (data.type === "user_left") {
      setConnectedUsers(prev => prev.filter(u => u.id !== data.user.id));
    } else if (data.type === "room_users") {
      setConnectedUsers(data.users || []);
    }
  };

  // Gestion des messages privés
  const handleDMMessage = (data: any) => {
    if (data.type === "dm" && (data.data?.fromUser === otherUserId || data.data?.to === otherUserId)) {
      setMessages(prev => [...prev, {
        ...data.data,
        id: Date.now(),
        status: 'delivered'
      }]);
      scrollToBottom();
    } else if (data.type === "dm_history") {
      if (data.data && Array.isArray(data.data)) {
        const dmMessages = data.data
          .filter(msg => msg.content)
          .map(msg => ({
            ...msg,
            id: msg.id || Date.now() + Math.random(),
            status: 'delivered'
          }))
          .sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
        setMessages(dmMessages);
      } else {
        setMessages([]);
      }
      setLoadingMessages(false);
      scrollToBottom();
    } else if (data.type === "dm_sent") {
      setMessages(prev => {
        const updated = [...prev];
        const lastMessage = updated[updated.length - 1];
        if (lastMessage && lastMessage.fromUser === myUserId) {
          lastMessage.status = 'sent';
        }
        return updated;
      });
    } else if (data.type === "user_status") {
      if (data.userId === otherUserId) {
        setOtherUserInfo(prev => ({ ...prev, isOnline: data.isOnline }));
      }
    }
  };

  // Rejoindre un salon
  const joinRoom = async (roomName: string) => {
    try {
      console.log(`[Chat] Rejoindre salon: ${roomName}`);
      setCurrentRoom(roomName);
      setMessages([]);
      setLoadingMessages(true);

      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({
          type: "join_room",
          room: roomName
        }));

        // Demander l'historique
        socket.send(JSON.stringify({
          type: "room_history",
          room: roomName,
          limit: 50
        }));
      }
    } catch (error) {
      console.error('Erreur rejoindre salon:', error);
    }
  };

  // Sélectionner un utilisateur pour DM
  const selectUser = async (userId: number) => {
    try {
      console.log(`[Chat] Sélectionner utilisateur: ${userId}`);
      setOtherUserId(userId);
      setMessages([]);
      setLoadingMessages(true);

      const selectedUser = allUsers.find(u => u.id === userId);
      if (selectedUser) {
        setOtherUserInfo({
          username: selectedUser.username,
          isOnline: selectedUser.isOnline || false
        });
      }

      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({
          type: "dm_history",
          with: userId,
          limit: 50
        }));
      }
    } catch (error) {
      console.error('Erreur sélection utilisateur:', error);
    }
  };

  // Créer un salon
  const createRoom = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newRoomName.trim() || creating) return;

    try {
      setCreating(true);
      
      const response = await fetch('/api/v1/chat/rooms', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${accessToken}`
        },
        body: JSON.stringify({
          name: newRoomName.trim(),
          description: `Salon créé par ${user?.username}`,
          is_private: false
        })
      });

      if (response.ok) {
        setNewRoomName('');
        await loadRooms();
        console.log('✅ Salon créé avec succès');
      } else {
        console.error('❌ Erreur création salon:', response.status);
      }
    } catch (error) {
      console.error('❌ Erreur création salon:', error);
    } finally {
      setCreating(false);
    }
  };

  // Envoyer un message
  const sendMessage = async (e: React.FormEvent) => {
    e.preventDefault();
    const content = messageContent.trim();
    if (!content || sending || !isConnected || !socket) return;

    setSending(true);

    try {
      if (activeTab === 'rooms' && currentRoom) {
        socket.send(JSON.stringify({
          type: "room_message",
          room: currentRoom,
          content: content
        }));
      } else if (activeTab === 'dm' && otherUserId) {
        const outgoingMessage = {
          id: Date.now(),
          fromUser: myUserId,
          to: otherUserId,
          content: content,
          timestamp: new Date().toISOString(),
          username: user?.username,
          status: 'sending'
        };

        setMessages(prev => [...prev, outgoingMessage]);
        scrollToBottom();

        socket.send(JSON.stringify({
          type: "direct_message",
          to_user_id: otherUserId,
          content: content
        }));
      }

      setMessageContent('');
    } catch (error) {
      console.error('Erreur envoi message:', error);
    } finally {
      setSending(false);
    }
  };

  // Recherche d'utilisateurs
  const searchUsers = () => {
    const query = userSearch.toLowerCase();
    if (!query) {
      setFilteredUsers(allUsers);
    } else {
      setFilteredUsers(allUsers.filter(user => 
        user.username.toLowerCase().includes(query) ||
        (user.email && user.email.toLowerCase().includes(query))
      ));
    }
  };

  // Changement d'onglet
  const switchTab = (newTab: 'rooms' | 'dm') => {
    setActiveTab(newTab);
    setMessages([]);
    setLoadingMessages(false);
    setMessageContent('');

    if (newTab === 'rooms') {
      setCurrentRoom('');
      setConnectedUsers([]);
    } else if (newTab === 'dm') {
      setOtherUserId(null);
      setOtherUserInfo({ username: '', isOnline: false });
      if (allUsers.length === 0) {
        loadUsers();
      }
    }
  };

  // Scroll vers le bas
  const scrollToBottom = () => {
    setTimeout(() => {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, 100);
  };

  // Format du temps
  const formatTime = (timestamp: string) => {
    try {
      const date = new Date(timestamp);
      return date.toLocaleTimeString('fr-FR', { 
      hour: '2-digit',
      minute: '2-digit'
    });
    } catch {
      return '';
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <p className="text-gray-600">Connexion requise</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto p-6 space-y-6">
      {/* Header */}
      <header className="flex items-center justify-between bg-white rounded-lg shadow p-6">
        <div className="flex items-center gap-4">
          <h1 className="text-4xl font-extrabold tracking-tight bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
            💬 Talas — Chat
          </h1>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-gray-600">👤 {user?.username}</span>
            <div className="flex items-center gap-2">
            <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`}></div>
            <span className="text-xs text-gray-500">{isConnected ? 'Connecté' : 'Déconnecté'}</span>
          </div>
        </div>
      </header>
          
      {/* Navigation */}
      <Tabs value={activeTab} onValueChange={(value) => switchTab(value as 'rooms' | 'dm')}>
            <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="rooms">🏠 Salons publics</TabsTrigger>
          <TabsTrigger value="dm">💬 Messages privés</TabsTrigger>
            </TabsList>

        {/* Interface principale */}
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6 mt-6">
          {/* Sidebar */}
          <div className="lg:col-span-1 space-y-4">
            <TabsContent value="rooms" className="mt-0">
              {/* Créer un salon */}
              <div className="bg-white rounded-lg shadow p-4 space-y-3">
                <h3 className="font-semibold text-gray-800">➕ Créer un salon</h3>
                <form onSubmit={createRoom} className="space-y-2">
                  <Input
                    value={newRoomName}
                    onChange={(e) => setNewRoomName(e.target.value)}
                    placeholder="Nom du salon"
                    required
                  />
                  <Button 
                    type="submit"
                    disabled={creating}
                    className="w-full"
                  >
                    {creating ? 'Création...' : 'Créer'}
                  </Button>
                </form>
              </div>

              {/* Liste des salons */}
              <div className="bg-white rounded-lg shadow p-4">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="font-semibold text-gray-800">🏠 Salons disponibles</h3>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={loadRooms}
                  >
                    <RefreshCw className="w-4 h-4" />
                  </Button>
                </div>
                
                <div className="space-y-1 max-h-96 overflow-y-auto">
                  {rooms.map((room) => (
                    <button
                      key={room.id}
                      onClick={() => joinRoom(room.name)}
                      className={`w-full text-left p-3 rounded-lg transition-colors text-sm ${
                        currentRoom === room.name 
                          ? 'bg-blue-600 text-white' 
                          : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <span className="font-medium">{room.name}</span>
                        <div className="flex items-center gap-2">
                          <span className="text-xs opacity-75">{room.user_count || 0}</span>
                          <Users className="w-3 h-3 opacity-75" />
                        </div>
                      </div>
                      {room.description && (
                        <div className="text-xs opacity-75 mt-1">{room.description}</div>
                      )}
                    </button>
                  ))}
                </div>
              </div>

              {/* Utilisateurs connectés */}
              {currentRoom && (
                <div className="bg-white rounded-lg shadow p-4">
                  <h3 className="font-semibold text-gray-800 mb-3">👥 Utilisateurs connectés</h3>
                  <div className="space-y-1 max-h-48 overflow-y-auto">
                    {connectedUsers.map((connectedUser) => (
                      <div key={connectedUser.id} className="flex items-center gap-2 p-2 rounded bg-gray-50">
                        <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                        <span className="text-sm">{connectedUser.username}</span>
                        {connectedUser.username === user?.username && (
                          <span className="text-xs text-blue-600">(vous)</span>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </TabsContent>

            <TabsContent value="dm" className="mt-0">
              {/* Recherche d'utilisateurs */}
              <div className="bg-white rounded-lg shadow p-4 space-y-3">
                <h3 className="font-semibold text-gray-800">👥 Choisir un correspondant</h3>
                <div className="relative">
                  <Input
                    value={userSearch}
                    onChange={(e) => {
                      setUserSearch(e.target.value);
                      searchUsers();
                    }}
                    placeholder="Rechercher un utilisateur..."
                    className="pl-8"
                  />
                  <Search className="absolute left-2 top-2.5 w-4 h-4 text-gray-400" />
                </div>
              </div>
              
              {/* Liste des utilisateurs */}
              <div className="bg-white rounded-lg shadow p-4">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="font-semibold text-gray-800">💬 Conversations</h3>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={loadUsers}
                  >
                    <RefreshCw className="w-4 h-4" />
                  </Button>
                </div>
                
                <div className="space-y-1 max-h-96 overflow-y-auto">
                  {loadingUsers && (
                    <div className="flex justify-center py-4">
                      <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
                    </div>
                  )}

                  {filteredUsers.length === 0 && !loadingUsers && (
                    <div className="text-center py-4 text-gray-500 text-sm">
                      <p>Aucun utilisateur trouvé</p>
                    </div>
                  )}

                  {filteredUsers.map((userItem) => (
                    <button
                      key={userItem.id}
                      onClick={() => selectUser(userItem.id)}
                      className={`w-full text-left p-3 rounded-lg transition-colors text-sm ${
                        otherUserId === userItem.id 
                          ? 'bg-blue-600 text-white' 
                          : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                      }`}
                    >
                      <div className="flex items-center gap-3">
                        <div className="relative">
                          <Avatar className="w-8 h-8">
                            <AvatarFallback className="bg-gradient-to-r from-blue-500 to-purple-500 text-white text-xs">
                              {userItem.username.charAt(0).toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          {userItem.isOnline && (
                            <div className="absolute -bottom-1 -right-1 w-3 h-3 bg-green-500 border-2 border-white rounded-full"></div>
                          )}
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="font-medium truncate">{userItem.username}</div>
                          <div className="text-xs opacity-75 truncate">{userItem.email}</div>
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            </TabsContent>
      </div>

          {/* Zone de discussion principale */}
          <div className="lg:col-span-3 bg-white rounded-lg shadow flex flex-col h-[600px]">
            {/* En-tête */}
            <div className="p-4 border-b border-gray-200">
              {activeTab === 'rooms' ? (
                  <div>
                  <h2 className="text-xl font-semibold">
                    {currentRoom ? `💬 ${currentRoom}` : 'Sélectionnez un salon'}
                  </h2>
                  {currentRoom && (
                    <p className="text-sm text-gray-500">
                      {connectedUsers.length} utilisateur(s) connecté(s)
                    </p>
                  )}
                  </div>
              ) : (
                <div className="flex items-center gap-3">
                  {otherUserId && (
                    <Avatar className="w-10 h-10">
                      <AvatarFallback className="bg-gradient-to-r from-blue-500 to-purple-500 text-white">
                        {(otherUserInfo.username || 'U').charAt(0).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  )}
                  <div>
                    <h2 className="text-xl font-semibold">
                      {otherUserId ? (otherUserInfo.username || `Utilisateur #${otherUserId}`) : 'Sélectionnez un correspondant'}
                    </h2>
                    {otherUserId && (
                      <div className="flex items-center gap-2">
                        <span className={`text-sm ${otherUserInfo.isOnline ? 'text-green-600' : 'text-gray-500'}`}>
                          {otherUserInfo.isOnline ? 'En ligne' : 'Hors ligne'}
                        </span>
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
            
            {/* Messages */}
            <ScrollArea className="flex-1 p-4 bg-gray-50">
              {/* État vide */}
              {((activeTab === 'rooms' && !currentRoom) || (activeTab === 'dm' && !otherUserId)) && (
                <div className="flex items-center justify-center h-full text-gray-500">
                  <div className="text-center">
                    <div className="w-16 h-16 mx-auto mb-4 text-gray-300">💬</div>
                    <p className="text-lg">
                      {activeTab === 'rooms' 
                        ? 'Sélectionnez un salon pour commencer à discuter' 
                        : 'Sélectionnez un correspondant pour commencer une conversation'}
                    </p>
            </div>
                </div>
              )}

              {/* Loader */}
              {loadingMessages && (
                <div className="flex justify-center py-4">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                </div>
              )}

              {/* Messages vides */}
              {((activeTab === 'rooms' && currentRoom) || (activeTab === 'dm' && otherUserId)) && 
               messages.length === 0 && !loadingMessages && (
                <div className="flex items-center justify-center h-full text-gray-500">
                  <div className="text-center">
                    <div className="w-16 h-16 mx-auto mb-4 text-gray-300">💬</div>
                    <p className="text-lg">Aucun message pour le moment</p>
                    <p className="text-sm text-gray-400">Envoyez le premier message pour commencer</p>
          </div>
        </div>
              )}

              {/* Messages des salons */}
              {activeTab === 'rooms' && messages.map((message, index) => (
                <div key={message.id || index} className="flex gap-3 mb-4">
                  <Avatar className="w-8 h-8 flex-shrink-0">
                    <AvatarFallback className="bg-gradient-to-r from-blue-500 to-purple-500 text-white text-sm">
                      {(message.username || `User #${message.fromUser}`).charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-medium text-sm">
                        {message.username || `User #${message.fromUser}`}
                    </span>
                    <span className="text-xs text-gray-500">
                        {formatTime(message.timestamp)}
                    </span>
                    </div>
                    <div className="bg-white rounded-lg px-3 py-2 shadow-sm">
                      <p className="text-sm text-gray-800">{message.content}</p>
                    </div>
                  </div>
                </div>
              ))}

              {/* Messages privés */}
              {activeTab === 'dm' && messages.map((message, index) => (
                <div key={message.id || index} className={`flex mb-4 ${
                  message.fromUser === myUserId ? 'justify-end' : 'justify-start'
                }`}>
                  <div className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg shadow-sm ${
                    message.fromUser === myUserId 
                      ? 'bg-blue-600 text-white' 
                      : 'bg-white text-gray-800'
                  }`}>
                    <p className="text-sm">{message.content}</p>
                    <div className="flex items-center justify-between mt-1">
                      <span className={`text-xs ${
                        message.fromUser === myUserId ? 'text-blue-200' : 'text-gray-500'
                      }`}>
                        {formatTime(message.timestamp)}
                      </span>
                      {message.fromUser === myUserId && message.status && (
                        <div className="flex items-center gap-1">
                          <div className={`w-2 h-2 rounded-full ${
                            message.status === 'sent' ? 'bg-blue-200' : 
                            message.status === 'delivered' ? 'bg-green-200' : 'bg-gray-200'
                          }`}></div>
                        </div>
                      )}
                    </div>
                </div>
              </div>
            ))}

            <div ref={messagesEndRef} />
        </ScrollArea>

            {/* Zone de saisie */}
            {((activeTab === 'rooms' && currentRoom) || (activeTab === 'dm' && otherUserId)) && (
              <div className="p-4 border-t border-gray-200 bg-white">
                <form onSubmit={sendMessage} className="flex gap-3">
              <Input
                    value={messageContent}
                    onChange={(e) => setMessageContent(e.target.value)}
                    placeholder="Tapez votre message..."
                className="flex-1"
                disabled={!isConnected}
                    maxLength={activeTab === 'rooms' ? 500 : 1000}
              />
                  <Button
                    type="submit"
                    disabled={!messageContent.trim() || sending || !isConnected}
                    className="flex items-center gap-2"
                  >
                    {sending ? 'Envoi...' : 'Envoyer'}
                <Send className="w-4 h-4" />
              </Button>
            </form>
              </div>
            )}
          </div>
      </div>
      </Tabs>
    </div>
  );
}