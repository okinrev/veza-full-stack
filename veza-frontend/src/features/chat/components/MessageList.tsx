import { useEffect, useRef } from 'react';
import { useChatStore } from '../store/chatStore';
import { cn } from '@/shared/utils/helpers';
import { Message } from '../types';

interface MessageListProps {
  roomId: string;
  className?: string;
}

export const MessageList = ({ roomId, className }: MessageListProps) => {
  const messages = useChatStore((state) => state.messages[roomId] || []);
  const currentUserId = useChatStore((state) => state.currentUserId);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const formatTime = (date: Date) => {
    return new Date(date).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const renderMessage = (message: Message) => {
    const isCurrentUser = message.userId === currentUserId;

    return (
      <div
        key={message.id}
        className={cn(
          'flex flex-col mb-4',
          isCurrentUser ? 'items-end' : 'items-start'
        )}
      >
        {message.type === 'system' ? (
          <div className="text-sm text-gray-500 text-center w-full">
            {message.content}
          </div>
        ) : (
          <>
            <div className="flex items-center gap-2 mb-1">
              <span className="text-sm font-medium">{message.username}</span>
              <span className="text-xs text-gray-500">
                {formatTime(message.timestamp)}
              </span>
            </div>
            <div
              className={cn(
                'max-w-[70%] rounded-lg px-4 py-2',
                isCurrentUser
                  ? 'bg-blue-500 text-white'
                  : 'bg-gray-100 text-gray-900'
              )}
            >
              {message.content}
            </div>
          </>
        )}
      </div>
    );
  };

  return (
    <div className={cn('flex-1 overflow-y-auto p-4', className)}>
      {messages.length === 0 ? (
        <div className="h-full flex items-center justify-center text-gray-500">
          Aucun message dans cette conversation
        </div>
      ) : (
        messages.map(renderMessage)
      )}
      <div ref={messagesEndRef} />
    </div>
  );
}; 