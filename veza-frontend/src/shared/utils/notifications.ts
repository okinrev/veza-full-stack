import { useAppStore } from '../store/useAppStore';

export const useNotifications = () => {
  const { addNotification } = useAppStore();

  return {
    success: (title: string, description?: string) => {
      addNotification({
        title,
        description,
        type: 'success',
        timestamp: new Date(),
      });
    },
    error: (title: string, description?: string) => {
      addNotification({
        title,
        description,
        type: 'error',
        timestamp: new Date(),
      });
    },
    warning: (title: string, description?: string) => {
      addNotification({
        title,
        description,
        type: 'warning',
        timestamp: new Date(),
      });
    },
    info: (title: string, description?: string) => {
      addNotification({
        title,
        description,
        type: 'info',
        timestamp: new Date(),
      });
    },
  };
}; 