import React, { useEffect, useState, useRef, useCallback } from 'react';
import { useAuthStore } from '@/features/auth/store/authStore';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import { MessageCircle, Users, Send, Plus, Bell, BellOff, LogOut, RefreshCw, Dot, Mail, AlertCircle, CheckCircle } from 'lucide-react';
import { Navigate } from 'react-router-dom';
import { chatWebSocket, type WebSocketChatMessage, type ConnectedUser } from '../services/chatWebSocket';

export function ChatPage() {
  const { user, logout } = useAuthStore();
  if (!user) return <Navigate to="/login" replace />;

  // UI state
  const [activeTab, setActiveTab] = useState<'rooms' | 'dm'>('rooms');
  const [currentRoom, setCurrentRoom] = useState<string>('general'); // Salon par défaut
  const [messages, setMessages] = useState<WebSocketChatMessage[]>([]);
  const [connectedUsers, setConnectedUsers] = useState<ConnectedUser[]>([]);
  const [dmUser, setDmUser] = useState<ConnectedUser | null>(null);
  const [dmList, setDmList] = useState<any[]>([]);
  const [messageInput, setMessageInput] = useState('');
  const [newRoomName, setNewRoomName] = useState('');
  const [creating, setCreating] = useState(false);
  const [loading, setLoading] = useState(true);
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState<string>('');
  const [typingUsers, setTypingUsers] = useState<string[]>([]);
  const [notificationsEnabled, setNotificationsEnabled] = useState(true);
  const messagesContainerRef = useRef<HTMLDivElement>(null);
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Liste des salons disponibles (peut être étendue par l'API)
  const [availableRooms] = useState<string[]>(['general', 'tech', 'random', 'help']);

  // --- WebSocket events ---
  useEffect(() => {
    setLoading(true);
    setConnectionError('');
    
    // Fonction de connexion
    const connectToChat = async () => {
      try {
        const connected = await chatWebSocket.connect();
        if (connected) {
          setIsConnected(true);
          setConnectionError('');
          // Rejoindre le salon par défaut
          if (currentRoom) {
            chatWebSocket.joinRoom(currentRoom);
            chatWebSocket.getRoomHistory(currentRoom);
          }
        } else {
          setConnectionError('Impossible de se connecter au chat');
          setIsConnected(false);
        }
      } catch (error) {
        console.error('Erreur connexion chat:', error);
        setConnectionError('Erreur de connexion');
        setIsConnected(false);
      } finally {
        setLoading(false);
      }
    };

    // Gestionnaires d'événements WebSocket
    const handleConnected = () => {
      console.log('✅ Chat connecté');
      setIsConnected(true);
      setConnectionError('');
    };

    const handleDisconnected = (data?: { code: number; reason: string }) => {
      console.log('❌ Chat déconnecté:', data);
      setIsConnected(false);
      if (data?.code !== 1000) {
        setConnectionError('Connexion perdue, tentative de reconnexion...');
      }
    };

    const handleMessageReceived = (msg: WebSocketChatMessage) => {
      console.log('📨 Nouveau message:', msg);
      setMessages(prev => [...prev, msg]);
      scrollToBottom();
      
      // Notification si permissions accordées
      if (notificationsEnabled && 'Notification' in window && Notification.permission === 'granted') {
        new Notification(`💬 ${msg.username}`, {
          body: msg.content,
          icon: '/favicon.ico'
        });
      }
    };

    const handleRoomHistory = (msgs: WebSocketChatMessage[]) => {
      console.log('📚 Historique reçu:', msgs.length, 'messages');
      setMessages(msgs || []);
      scrollToBottom();
    };

    const handleDMReceived = (msg: WebSocketChatMessage) => {
      console.log('📩 Message privé reçu:', msg);
      // TODO: Gérer les messages privés
    };

    const handleRoomJoined = (data: any) => {
      console.log('✅ Salon rejoint:', data);
    };

    const handleTypingStart = (data: { username: string; room?: string }) => {
      if (data.room === currentRoom) {
        setTypingUsers(prev => [...prev.filter(u => u !== data.username), data.username]);
      }
    };

    const handleTypingStop = (data: { username: string; room?: string }) => {
      if (data.room === currentRoom) {
        setTypingUsers(prev => prev.filter(u => u !== data.username));
      }
    };

    const handleError = (error: { message: string }) => {
      console.error('❌ Erreur chat:', error.message);
      setConnectionError(error.message);
    };

    // Enregistrer les gestionnaires
    chatWebSocket.on('connected', handleConnected);
    chatWebSocket.on('disconnected', handleDisconnected);
    chatWebSocket.on('message_received', handleMessageReceived);
    chatWebSocket.on('room_history', handleRoomHistory);
    chatWebSocket.on('dm_received', handleDMReceived);
    chatWebSocket.on('room_joined', handleRoomJoined);
    chatWebSocket.on('typing_start', handleTypingStart);
    chatWebSocket.on('typing_stop', handleTypingStop);
    chatWebSocket.on('error', handleError);

    // Démarrer la connexion
    connectToChat();

    // Nettoyage
    return () => {
      chatWebSocket.off('connected', handleConnected);
      chatWebSocket.off('disconnected', handleDisconnected);
      chatWebSocket.off('message_received', handleMessageReceived);
      chatWebSocket.off('room_history', handleRoomHistory);
      chatWebSocket.off('dm_received', handleDMReceived);
      chatWebSocket.off('room_joined', handleRoomJoined);
      chatWebSocket.off('typing_start', handleTypingStart);
      chatWebSocket.off('typing_stop', handleTypingStop);
      chatWebSocket.off('error', handleError);
      chatWebSocket.disconnect();
    };
  }, [currentRoom, notificationsEnabled]);

  // --- Demander les permissions de notification ---
  useEffect(() => {
    if ('Notification' in window && Notification.permission === 'default') {
      Notification.requestPermission().then(permission => {
        setNotificationsEnabled(permission === 'granted');
      });
    }
  }, []);

  // --- Room management ---
  const joinRoom = useCallback((roomName: string) => {
    if (currentRoom) {
      chatWebSocket.leaveRoom(currentRoom);
    }
    
    setCurrentRoom(roomName);
    setActiveTab('rooms');
    setMessages([]);
    setTypingUsers([]);
    
    if (chatWebSocket.isConnected()) {
      chatWebSocket.joinRoom(roomName);
      chatWebSocket.getRoomHistory(roomName);
    }
  }, [currentRoom]);

  // --- Message sending ---
  const sendMessage = useCallback(() => {
    if (!messageInput.trim() || !currentRoom || !chatWebSocket.isConnected()) return;
    
    const success = chatWebSocket.sendRoomMessage(currentRoom, messageInput.trim());
    if (success) {
      setMessageInput('');
      // Arrêter l'indicateur de frappe
      if (typingTimeoutRef.current) {
        clearTimeout(typingTimeoutRef.current);
        typingTimeoutRef.current = null;
      }
    }
  }, [messageInput, currentRoom]);

  // --- Typing indicator ---
  const handleTyping = useCallback(() => {
    if (!currentRoom || !chatWebSocket.isConnected()) return;
    
    // Envoyer indicateur de frappe
    chatWebSocket.sendTyping(currentRoom);
    
    // Programmer l'arrêt après 3 secondes
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }
    
    typingTimeoutRef.current = setTimeout(() => {
      // L'indicateur s'arrêtera automatiquement côté serveur
    }, 3000);
  }, [currentRoom]);

  // --- Scroll to bottom ---
  const scrollToBottom = () => {
    setTimeout(() => {
      if (messagesContainerRef.current) {
        messagesContainerRef.current.scrollTop = messagesContainerRef.current.scrollHeight;
      }
    }, 100);
  };

  // --- Retry connection ---
  const retryConnection = useCallback(async () => {
    setLoading(true);
    setConnectionError('');
    
    const connected = await chatWebSocket.connect();
    if (connected && currentRoom) {
      chatWebSocket.joinRoom(currentRoom);
      chatWebSocket.getRoomHistory(currentRoom);
    }
    
    setLoading(false);
  }, [currentRoom]);

  // --- UI rendering ---
  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <LoadingSpinner size="lg" className="text-blue-600 mx-auto mb-4" />
          <p className="text-gray-600">Connexion au chat...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto space-y-6">
        {/* Header */}
        <header className="flex items-center justify-between bg-white rounded-lg shadow p-6">
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
            💬 Chat Veza
          </h1>
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-600">👤 {user?.username}</span>
            
            {/* Statut de connexion */}
            <div className="flex items-center gap-2">
              <Dot className={`w-4 h-4 ${isConnected ? 'text-green-500' : 'text-red-500'}`} />
              <span className="text-xs text-gray-500">
                {isConnected ? 'Connecté' : 'Déconnecté'}
              </span>
              {!isConnected && (
                <Button 
                  onClick={retryConnection} 
                  variant="ghost" 
                  size="sm" 
                  className="text-blue-600 hover:text-blue-700 p-1"
                  disabled={loading}
                >
                  <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
                </Button>
              )}
            </div>

            {/* Notifications */}
            <Button
              onClick={() => setNotificationsEnabled(!notificationsEnabled)}
              variant="ghost"
              size="sm"
              className="text-gray-600 hover:text-gray-700"
            >
              {notificationsEnabled ? <Bell className="w-4 h-4" /> : <BellOff className="w-4 h-4" />}
            </Button>
          </div>
        </header>

        {/* Erreur de connexion */}
        {connectionError && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-red-500 flex-shrink-0" />
            <span className="text-red-700 flex-1">{connectionError}</span>
            <Button onClick={retryConnection} variant="outline" size="sm" className="text-red-600 border-red-300">
              Réessayer
            </Button>
          </div>
        )}

        {/* Tabs navigation */}
        <nav className="bg-white rounded-lg shadow p-2 flex gap-2">
          <Button 
            onClick={() => setActiveTab('rooms')} 
            variant={activeTab === 'rooms' ? 'default' : 'ghost'} 
            size="sm"
          >
            <MessageCircle className="w-4 h-4 mr-2" /> 
            Salons publics
          </Button>
          <Button 
            onClick={() => setActiveTab('dm')} 
            variant={activeTab === 'dm' ? 'default' : 'ghost'} 
            size="sm"
          >
            <Mail className="w-4 h-4 mr-2" /> 
            Messages privés
          </Button>
        </nav>

        {/* Main content */}
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Sidebar - Liste des salons */}
          {activeTab === 'rooms' && (
            <div className="lg:col-span-1 space-y-4">
              {/* Liste des salons */}
              <div className="bg-white rounded-lg shadow p-4">
                <h3 className="font-semibold text-gray-800 flex items-center gap-2 mb-3">
                  <MessageCircle className="w-4 h-4" /> Salons disponibles
                </h3>
                <div className="space-y-1">
                  {availableRooms.map((roomName) => (
                    <Button
                      key={roomName}
                      onClick={() => joinRoom(roomName)}
                      variant={currentRoom === roomName ? 'default' : 'ghost'}
                      size="sm"
                      className={`w-full justify-start ${
                        currentRoom === roomName 
                          ? 'bg-blue-100 text-blue-700 border-blue-300' 
                          : 'hover:bg-gray-50'
                      }`}
                    >
                      <span className="mr-2">#</span>
                      {roomName}
                      {currentRoom === roomName && <CheckCircle className="w-4 h-4 ml-auto" />}
                    </Button>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* Messages privés */}
          {activeTab === 'dm' && (
            <div className="lg:col-span-1 space-y-4">
              <div className="bg-white rounded-lg shadow p-4">
                <h3 className="font-semibold text-gray-800 flex items-center gap-2 mb-3">
                  <Users className="w-4 h-4" /> Utilisateurs connectés
                </h3>
                <div className="text-center text-gray-500 py-8">
                  <Mail className="w-12 h-12 mx-auto mb-2 text-gray-300" />
                  <p className="text-sm">Messages privés</p>
                  <p className="text-xs text-gray-400">À implémenter</p>
                </div>
              </div>
            </div>
          )}

          {/* Zone de discussion principale */}
          <div className="lg:col-span-3 bg-white rounded-lg shadow flex flex-col h-[600px]">
            {/* En-tête du salon */}
            <div className="p-4 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <h2 className="text-xl font-semibold">
                    {currentRoom ? `# ${currentRoom}` : 'Sélectionnez un salon'}
                  </h2>
                  {isConnected && currentRoom && (
                    <span className="text-xs bg-green-100 text-green-700 px-2 py-1 rounded-full">
                      Connecté
                    </span>
                  )}
                </div>
                
                {/* Indicateur de frappe */}
                {typingUsers.length > 0 && (
                  <div className="text-sm text-gray-500 italic">
                    {typingUsers.join(', ')} {typingUsers.length === 1 ? 'écrit...' : 'écrivent...'}
                  </div>
                )}
              </div>
            </div>

            {/* Messages */}
            <div ref={messagesContainerRef} className="flex-1 overflow-y-auto p-4 space-y-3 bg-gray-50">
              {!currentRoom ? (
                <div className="flex items-center justify-center h-full text-gray-500">
                  <div className="text-center">
                    <MessageCircle className="w-16 h-16 mx-auto mb-4 text-gray-300" />
                    <p className="text-lg">Sélectionnez un salon pour commencer</p>
                    <p className="text-sm text-gray-400">Choisissez un salon dans la liste à gauche</p>
                  </div>
                </div>
              ) : !isConnected ? (
                <div className="flex items-center justify-center h-full text-gray-500">
                  <div className="text-center">
                    <AlertCircle className="w-16 h-16 mx-auto mb-4 text-red-300" />
                    <p className="text-lg">Connexion au chat...</p>
                    <Button onClick={retryConnection} variant="outline" className="mt-4">
                      <RefreshCw className="w-4 h-4 mr-2" />
                      Reconnecter
                    </Button>
                  </div>
                </div>
              ) : messages.length === 0 ? (
                <div className="flex items-center justify-center h-full text-gray-500">
                  <div className="text-center">
                    <MessageCircle className="w-16 h-16 mx-auto mb-4 text-gray-300" />
                    <p className="text-lg">Aucun message</p>
                    <p className="text-sm text-gray-400">Soyez le premier à écrire dans #{currentRoom} !</p>
                  </div>
                </div>
              ) : (
                <>
                  {messages.map((message, index) => (
                    <div key={message.id || index} className="flex gap-3">
                      <div className="flex-shrink-0">
                        <div className="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full flex items-center justify-center text-white text-sm font-semibold">
                          <span>{(message.username || `U#${message.fromUser}`).charAt(0).toUpperCase()}</span>
                        </div>
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 mb-1">
                          <span className="font-medium text-sm text-gray-700">
                            {message.username || `User #${message.fromUser}`}
                          </span>
                          <span className="text-xs text-gray-500">
                            {new Date(message.timestamp).toLocaleTimeString('fr-FR', {
                              hour: '2-digit',
                              minute: '2-digit'
                            })}
                          </span>
                        </div>
                        <div className="bg-white rounded-lg px-3 py-2 shadow-sm border">
                          <p className="text-sm text-gray-800 whitespace-pre-wrap">{message.content}</p>
                        </div>
                      </div>
                    </div>
                  ))}
                </>
              )}
            </div>

            {/* Zone de saisie */}
            {currentRoom && isConnected && (
              <div className="p-4 border-t border-gray-200 bg-white">
                <div className="flex gap-2">
                  <Input 
                    value={messageInput} 
                    onChange={(e) => {
                      setMessageInput(e.target.value);
                      handleTyping();
                    }}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' && !e.shiftKey) {
                        e.preventDefault();
                        sendMessage();
                      }
                    }}
                    placeholder={`Écrivez dans #${currentRoom}...`}
                    className="flex-1"
                    disabled={!isConnected}
                  />
                  <Button 
                    onClick={sendMessage} 
                    disabled={!messageInput.trim() || !isConnected}
                    className="px-4"
                  >
                    <Send className="w-4 h-4" />
                  </Button>
                </div>
                <p className="text-xs text-gray-500 mt-2">
                  Appuyez sur Entrée pour envoyer • Maj+Entrée pour un retour à la ligne
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}