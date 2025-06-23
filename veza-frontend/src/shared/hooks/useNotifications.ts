import { create } from 'zustand';

type NotificationType = 'success' | 'error' | 'info' | 'warning';

interface Notification {
  id: string;
  message: string;
  type: NotificationType;
}

interface NotificationStore {
  notifications: Notification[];
  showNotification: (message: string, type: NotificationType) => void;
  removeNotification: (id: string) => void;
}

export const useNotifications = create<NotificationStore>((set) => ({
  notifications: [],
  showNotification: (message, type) => {
    const id = Math.random().toString(36).substr(2, 9);
    set((state) => ({
      notifications: [...state.notifications, { id, message, type }],
    }));
    setTimeout(() => {
      set((state) => ({
        notifications: state.notifications.filter((n) => n.id !== id),
      }));
    }, 3000);
  },
  removeNotification: (id) => {
    set((state) => ({
      notifications: state.notifications.filter((n) => n.id !== id),
    }));
  },
})); 