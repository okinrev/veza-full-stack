import { useAuthStore } from '../store/authStore';

export const useAuth = () => {
  const {
    user,
    isAuthenticated,
    isLoading,
    error,
    login,
    register,
    logout,
    refreshAuth,
    clearError,
    updateLastActivity,
    checkSessionTimeout,
  } = useAuthStore();

  return {
    user,
    isAuthenticated,
    isLoading,
    error,
    login,
    register,
    logout,
    refreshAuth,
    clearError,
    updateLastActivity,
    checkSessionTimeout,
  };
}; 