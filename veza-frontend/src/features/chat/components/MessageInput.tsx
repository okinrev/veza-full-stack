import { useState, useCallback, useRef } from 'react';
import { useChatStore } from '../store/chatStore';
import { wsManager } from '@/shared/api/websocket';
import { Button } from '@/components/ui/button';
import { Send } from 'lucide-react';

interface MessageInputProps {
  roomId: string;
  className?: string;
}

export const MessageInput = ({ roomId, className }: MessageInputProps) => {
  const [message, setMessage] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const typingTimeoutRef = useRef<NodeJS.Timeout>();
  const currentUserId = useChatStore((state) => state.currentUserId);

  const handleTyping = useCallback(() => {
    if (!isTyping) {
      setIsTyping(true);
      wsManager.send({
        type: 'user_typing',
        payload: { roomId, userId: currentUserId, isTyping: true },
      });
    }

    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current);
    }

    typingTimeoutRef.current = setTimeout(() => {
      setIsTyping(false);
      wsManager.send({
        type: 'user_typing',
        payload: { roomId, userId: currentUserId, isTyping: false },
      });
    }, 2000);
  }, [roomId, currentUserId, isTyping]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!message.trim()) return;

    wsManager.send({
      type: 'message',
      payload: {
        roomId,
        content: message.trim(),
      },
    });

    setMessage('');
  };

  return (
    <form
      onSubmit={handleSubmit}
      className={`flex items-center gap-2 p-4 border-t ${className}`}
    >
      <input
        type="text"
        value={message}
        onChange={(e) => setMessage(e.target.value)}
        onKeyPress={handleTyping}
        placeholder="Écrivez votre message..."
        className="flex-1 px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
      <Button
        type="submit"
        disabled={!message.trim()}
        className="p-2 rounded-lg bg-blue-500 text-white hover:bg-blue-600 disabled:opacity-50"
      >
        <Send size={20} />
      </Button>
    </form>
  );
}; 