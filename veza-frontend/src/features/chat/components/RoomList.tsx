import { useChatStore } from '../store/chatStore';
import { cn } from '@/shared/utils/helpers';
import { MessageSquare, Users } from 'lucide-react';

interface RoomListProps {
  className?: string;
}

export const RoomList = ({ className }: RoomListProps) => {
  const rooms = useChatStore((state) => state.rooms);
  const currentRoom = useChatStore((state) => state.currentRoom);
  const joinRoom = useChatStore((state) => state.joinRoom);

  const formatLastMessage = (message?: { content: string; timestamp: Date }) => {
    if (!message) return '';
    const date = new Date(message.timestamp);
    return `${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })} - ${message.content}`;
  };

  return (
    <div className={cn('w-64 border-r bg-gray-50', className)}>
      <div className="p-4 border-b">
        <h2 className="text-lg font-semibold">Salles de chat</h2>
      </div>
      <div className="overflow-y-auto h-[calc(100vh-4rem)]">
        {rooms.map((room) => (
          <button
            key={room.id}
            onClick={() => joinRoom(room.id)}
            className={cn(
              'w-full p-4 text-left hover:bg-gray-100 transition-colors',
              currentRoom === room.id && 'bg-blue-50 hover:bg-blue-100'
            )}
          >
            <div className="flex items-center gap-2 mb-1">
              <MessageSquare size={16} className="text-gray-500" />
              <span className="font-medium">{room.name}</span>
            </div>
            {room.description && (
              <p className="text-sm text-gray-500 mb-2">{room.description}</p>
            )}
            <div className="flex items-center justify-between text-sm">
              <span className="text-gray-500">
                {room.lastMessage && formatLastMessage(room.lastMessage)}
              </span>
              <div className="flex items-center gap-1 text-gray-500">
                <Users size={14} />
                <span>{room.userCount}</span>
              </div>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}; 