import { useNotifications } from '../hooks/useNotifications';
import { X } from 'lucide-react';
import { cn } from '../utils/helpers';

export const Notifications = () => {
  const { notifications, removeNotification } = useNotifications();

  return (
    <div className="fixed top-4 right-4 z-50 flex flex-col gap-2">
      {notifications.map((notification) => (
        <div
          key={notification.id}
          className={cn(
            'p-4 rounded-lg shadow-lg flex items-center justify-between min-w-[300px]',
            {
              'bg-green-500 text-white': notification.type === 'success',
              'bg-red-500 text-white': notification.type === 'error',
              'bg-blue-500 text-white': notification.type === 'info',
              'bg-yellow-500 text-white': notification.type === 'warning',
            }
          )}
        >
          <span>{notification.message}</span>
          <button
            onClick={() => removeNotification(notification.id)}
            className="ml-4 text-white hover:text-gray-200"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      ))}
    </div>
  );
}; 