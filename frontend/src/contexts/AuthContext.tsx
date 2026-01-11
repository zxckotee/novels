'use client';

import { useAuthStore } from '@/store/auth';

/**
 * AuthContext hook that wraps useAuthStore
 * Provides user and isAuthenticated for components
 */
export function useAuth() {
  const { user, isAuthenticated } = useAuthStore();
  
  return {
    user,
    isAuthenticated,
  };
}