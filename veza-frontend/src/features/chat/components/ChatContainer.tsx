import { useEffect } from 'react';
import { useChatStore } from '../store/chatStore';
import { MessageList } from './MessageList';
import { MessageInput } from './MessageInput';
import { RoomList } from './RoomList';
import { wsManager } from '@/shared/api/websocket';
import { LoadingSpinner } from '@/shared/components/common/LoadingSpinner';
import { EmptyState } from '@/shared/components/common/EmptyState';
import { WifiOff } from 'lucide-react';

export const ChatContainer = () => {
  const { currentRoom, isConnected, setCurrentUserId } = useChatStore();

  useEffect(() => {
    // TODO: Récupérer l'ID de l'utilisateur depuis le store d'authentification
    setCurrentUserId(1); // Temporaire pour les tests
    wsManager.connect();

    return () => {
      wsManager.disconnect();
    };
  }, [setCurrentUserId]);

  if (!isConnected) {
    return (
      <EmptyState
        icon={WifiOff}
        title="Connexion en cours..."
        description="Tentative de connexion au serveur de chat"
      />
    );
  }

  return (
    <div className="flex h-screen">
      <RoomList />
      <div className="flex-1 flex flex-col">
        {currentRoom ? (
          <>
            <MessageList roomId={currentRoom} />
            <MessageInput roomId={currentRoom} />
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center">
            <p className="text-gray-500">
              Sélectionnez une salle de chat pour commencer
            </p>
          </div>
        )}
      </div>
    </div>
  );
}; 