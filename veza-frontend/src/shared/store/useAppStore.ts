import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';
import { immer } from 'zustand/middleware/immer';

interface Notification {
  id: string;
  title: string;
  description?: string;
  type: 'info' | 'success' | 'warning' | 'error';
  timestamp: Date;
}

interface AppState {
  // UI State
  sidebarOpen: boolean;
  theme: 'light' | 'dark' | 'system';
  notifications: Notification[];
  
  // Actions
  toggleSidebar: () => void;
  setTheme: (theme: AppState['theme']) => void;
  addNotification: (notification: Omit<Notification, 'id'>) => void;
  removeNotification: (id: string) => void;
}

export const useAppStore = create<AppState>()(
  devtools(
    persist(
      immer((set) => ({
        // Initial state
        sidebarOpen: true,
        theme: 'system',
        notifications: [],
        
        // Actions
        toggleSidebar: () =>
          set((state) => {
            state.sidebarOpen = !state.sidebarOpen;
          }),
          
        setTheme: (theme) =>
          set((state) => {
            state.theme = theme;
          }),
          
        addNotification: (notification) =>
          set((state) => {
            state.notifications.push({
              ...notification,
              id: Date.now().toString(),
              timestamp: new Date(),
            });
          }),
          
        removeNotification: (id) =>
          set((state) => {
            state.notifications = state.notifications.filter(
              (n) => n.id !== id
            );
          }),
      })),
      {
        name: 'app-store',
        partialize: (state) => ({ theme: state.theme }),
      }
    )
  )
); 