import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface User {
  id: string;
  email: string;
  displayName: string;
  avatarUrl?: string;
  roles: string[];
}

interface AuthState {
  user: User | null;
  accessToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  // Actions
  setUser: (user: User | null) => void;
  setAccessToken: (token: string | null) => void;
  login: (user: User, accessToken: string) => void;
  logout: () => void;
  setLoading: (loading: boolean) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      accessToken: null,
      isAuthenticated: false,
      isLoading: true,

      setUser: (user) =>
        set({
          user,
          isAuthenticated: !!user,
        }),

      setAccessToken: (accessToken) =>
        set({
          accessToken,
        }),

      login: (user, accessToken) =>
        set({
          user,
          accessToken,
          isAuthenticated: true,
          isLoading: false,
        }),

      logout: () =>
        set({
          user: null,
          accessToken: null,
          isAuthenticated: false,
        }),

      setLoading: (isLoading) =>
        set({
          isLoading,
        }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);

// Helper to check if user has a specific role
export const hasRole = (user: User | null, role: string): boolean => {
  if (!user) return false;
  return user.roles.includes(role);
};

// Helper to check if user is admin
export const isAdmin = (user: User | null): boolean => {
  return hasRole(user, 'admin');
};

// Helper to check if user is moderator
export const isModerator = (user: User | null): boolean => {
  return hasRole(user, 'moderator') || hasRole(user, 'admin');
};

// Helper to check if user has premium
export const isPremium = (user: User | null): boolean => {
  return hasRole(user, 'premium') || hasRole(user, 'admin');
};
