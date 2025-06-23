import { useEffect } from 'react';
import { useAuthStore } from '../store/authStore';
import { useNotifications } from '@/shared/utils/notifications';

const CHECK_INTERVAL = 60000; // Vérifier toutes les minutes
const WARNING_THRESHOLD = 5 * 60 * 1000; // Avertir 5 minutes avant la déconnexion

export const useSessionTimeout = () => {
  const { checkSessionTimeout, updateLastActivity, lastActivity } = useAuthStore();
  const notifications = useNotifications();

  useEffect(() => {
    // Mettre à jour l'activité sur les événements utilisateur
    const updateActivity = () => {
      updateLastActivity();
    };

    // Ajouter les écouteurs d'événements
    window.addEventListener('mousemove', updateActivity);
    window.addEventListener('keydown', updateActivity);
    window.addEventListener('click', updateActivity);
    window.addEventListener('scroll', updateActivity);

    // Vérifier régulièrement le timeout de session
    const interval = setInterval(() => {
      if (lastActivity) {
        const timeUntilTimeout = WARNING_THRESHOLD - (Date.now() - lastActivity);
        
        // Avertir l'utilisateur si le timeout approche
        if (timeUntilTimeout > 0 && timeUntilTimeout <= WARNING_THRESHOLD) {
          const minutes = Math.ceil(timeUntilTimeout / 60000);
          notifications.warning(
            'Session expirant',
            `Votre session expirera dans ${minutes} minute${minutes > 1 ? 's' : ''}. Veuillez sauvegarder votre travail.`
          );
        }

        // Vérifier si la session a expiré
        if (checkSessionTimeout()) {
          notifications.error('Session expirée', 'Vous avez été déconnecté pour inactivité.');
        }
      }
    }, CHECK_INTERVAL);

    // Nettoyer les écouteurs d'événements
    return () => {
      window.removeEventListener('mousemove', updateActivity);
      window.removeEventListener('keydown', updateActivity);
      window.removeEventListener('click', updateActivity);
      window.removeEventListener('scroll', updateActivity);
      clearInterval(interval);
    };
  }, [checkSessionTimeout, updateLastActivity, lastActivity, notifications]);
}; 