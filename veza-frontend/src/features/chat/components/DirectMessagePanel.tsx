import React, { useState, useEffect, useRef } from 'react';
import { Send, Phone, Video, MoreVertical, Search, Paperclip, Smile } from 'lucide-react';
import type { DirectMessage, ChatUser } from '../services/chatApi';

interface DirectMessagePanelProps {
  user: ChatUser;
  messages: DirectMessage[];
  currentUserId: number;
  onSendMessage: (content: string) => void;
  onClose: () => void;
}

export function DirectMessagePanel({ 
  user, 
  messages, 
  currentUserId, 
  onSendMessage, 
  onClose 
}: DirectMessagePanelProps) {
  const [newMessage, setNewMessage] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll vers le bas quand de nouveaux messages arrivent
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    if (newMessage.trim()) {
      onSendMessage(newMessage.trim());
      setNewMessage('');
    }
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffInHours = (now.getTime() - date.getTime()) / (1000 * 60 * 60);
    
    if (diffInHours < 24) {
      return date.toLocaleTimeString('fr-FR', { 
        hour: '2-digit', 
        minute: '2-digit' 
      });
    } else {
      return date.toLocaleDateString('fr-FR', { 
        day: '2-digit', 
        month: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      });
    }
  };

  const getInitials = (firstName?: string, lastName?: string, username?: string) => {
    if (firstName && lastName) {
      return `${firstName[0]}${lastName[0]}`.toUpperCase();
    }
    if (username) {
      return username.slice(0, 2).toUpperCase();
    }
    return 'U';
  };

  return (
    <div className="flex flex-col h-full bg-white">
      {/* Header de la conversation */}
      <div className="flex items-center justify-between p-4 border-b border-gray-200 bg-gray-50">
        <div className="flex items-center space-x-3">
          {/* Avatar utilisateur */}
          <div className="relative">
            {user.avatar_url ? (
              <img 
                src={user.avatar_url} 
                alt={user.username}
                className="w-10 h-10 rounded-full object-cover"
              />
            ) : (
              <div className="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-semibold">
                {getInitials(user.first_name, user.last_name, user.username)}
              </div>
            )}
            {/* Indicateur en ligne */}
            {user.isOnline && (
              <div className="absolute -bottom-1 -right-1 w-4 h-4 bg-green-500 border-2 border-white rounded-full"></div>
            )}
          </div>
          
          {/* Infos utilisateur */}
          <div>
            <h3 className="font-semibold text-gray-900">
              {user.first_name && user.last_name 
                ? `${user.first_name} ${user.last_name}` 
                : user.username
              }
            </h3>
            <p className="text-sm text-gray-500">
              {user.isOnline ? (
                <span className="text-green-600">En ligne</span>
              ) : (
                `Vu ${formatTime(user.lastSeen)}`
              )}
            </p>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center space-x-2">
          <button className="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-full transition-colors">
            <Phone size={18} />
          </button>
          <button className="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-full transition-colors">
            <Video size={18} />
          </button>
          <button className="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-full transition-colors">
            <Search size={18} />
          </button>
          <button className="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-full transition-colors">
            <MoreVertical size={18} />
          </button>
          <button 
            onClick={onClose}
            className="p-2 text-gray-500 hover:text-red-600 hover:bg-red-50 rounded-full transition-colors"
          >
            ✕
          </button>
        </div>
      </div>

      {/* Zone des messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-gray-500">
            <div className="w-16 h-16 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-bold text-xl mb-4">
              {getInitials(user.first_name, user.last_name, user.username)}
            </div>
            <h3 className="text-lg font-semibold text-gray-700 mb-2">
              Commencez une conversation avec {user.first_name || user.username}
            </h3>
            <p className="text-center text-gray-500">
              C'est le début de votre conversation privée. Dites bonjour ! 👋
            </p>
          </div>
        ) : (
          messages.map((message, index) => {
            const isFromCurrentUser = message.from_user_id === currentUserId;
            const showAvatar = index === 0 || messages[index - 1].from_user_id !== message.from_user_id;
            
            return (
              <div
                key={message.id}
                className={`flex ${isFromCurrentUser ? 'justify-end' : 'justify-start'} ${
                  showAvatar ? 'mt-4' : 'mt-1'
                }`}
              >
                {/* Avatar pour les messages des autres */}
                {!isFromCurrentUser && showAvatar && (
                  <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center text-white font-semibold text-sm mr-2 flex-shrink-0">
                    {getInitials(user.first_name, user.last_name, user.username)}
                  </div>
                )}
                
                {/* Espace pour aligner les messages suivants */}
                {!isFromCurrentUser && !showAvatar && (
                  <div className="w-8 mr-2 flex-shrink-0"></div>
                )}

                <div className={`max-w-xs lg:max-w-md ${isFromCurrentUser ? 'order-1' : 'order-2'}`}>
                  {/* Nom de l'utilisateur (seulement pour le premier message d'une série) */}
                  {showAvatar && !isFromCurrentUser && (
                    <p className="text-xs text-gray-500 mb-1 ml-1">
                      {message.from_username}
                    </p>
                  )}
                  
                  {/* Bulle de message */}
                  <div
                    className={`px-4 py-2 rounded-2xl ${
                      isFromCurrentUser
                        ? 'bg-blue-500 text-white rounded-br-md'
                        : 'bg-gray-100 text-gray-900 rounded-bl-md'
                    }`}
                  >
                    <p className="text-sm">{message.content}</p>
                  </div>
                  
                  {/* Timestamp */}
                  <p className={`text-xs text-gray-400 mt-1 ${
                    isFromCurrentUser ? 'text-right' : 'text-left'
                  }`}>
                    {formatTime(message.timestamp)}
                    {isFromCurrentUser && (
                      <span className="ml-1">
                        {message.is_read ? '✓✓' : '✓'}
                      </span>
                    )}
                  </p>
                </div>
              </div>
            );
          })
        )}
        
        {/* Indicateur de frappe */}
        {isTyping && (
          <div className="flex items-center space-x-2 text-gray-500">
            <div className="w-8 h-8 rounded-full bg-gray-200 flex items-center justify-center">
              <div className="flex space-x-1">
                <div className="w-1 h-1 bg-gray-400 rounded-full animate-bounce"></div>
                <div className="w-1 h-1 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
                <div className="w-1 h-1 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
              </div>
            </div>
            <span className="text-sm">{user.username} est en train d'écrire...</span>
          </div>
        )}
        
        <div ref={messagesEndRef} />
      </div>

      {/* Zone de saisie */}
      <div className="border-t border-gray-200 p-4">
        <form onSubmit={handleSendMessage} className="flex items-end space-x-2">
          {/* Boutons d'actions */}
          <div className="flex space-x-1">
            <button
              type="button"
              className="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-full transition-colors"
            >
              <Paperclip size={18} />
            </button>
            <button
              type="button"
              className="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-full transition-colors"
            >
              <Smile size={18} />
            </button>
          </div>

          {/* Zone de texte */}
          <div className="flex-1 relative">
            <textarea
              value={newMessage}
              onChange={(e) => setNewMessage(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault();
                  handleSendMessage(e);
                }
              }}
              placeholder={`Écrivez à ${user.first_name || user.username}...`}
              className="w-full px-4 py-2 border border-gray-300 rounded-2xl resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent max-h-32"
              rows={1}
              style={{
                minHeight: '40px',
                height: 'auto',
              }}
            />
          </div>

          {/* Bouton d'envoi */}
          <button
            type="submit"
            disabled={!newMessage.trim()}
            className={`p-2 rounded-full transition-colors ${
              newMessage.trim()
                ? 'bg-blue-500 text-white hover:bg-blue-600'
                : 'bg-gray-200 text-gray-400 cursor-not-allowed'
            }`}
          >
            <Send size={18} />
          </button>
        </form>
      </div>
    </div>
  );
} 