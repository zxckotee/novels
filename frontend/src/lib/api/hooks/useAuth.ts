import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../client';
import { useAuthStore, type User as StoreUser } from '@/store/auth';
import type { User, UserProfile } from '../types';

// Query keys
export const authKeys = {
  user: ['auth', 'user'] as const,
  profile: ['auth', 'profile'] as const,
};

// Types
interface LoginRequest {
  email: string;
  password: string;
}

interface RegisterRequest {
  email: string;
  password: string;
  displayName: string;
}

interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token?: string;
  token_type?: string;
  expires_in?: number;
}

// Convert API User to Store User
function convertApiUserToStoreUser(apiUser: User): StoreUser {
  return {
    id: apiUser.id,
    email: apiUser.email,
    displayName: apiUser.displayName,
    avatarUrl: apiUser.avatarUrl,
    roles: [apiUser.role], // Convert single role to array
  };
}

// Get current user
export function useCurrentUser() {
  const { accessToken, isAuthenticated } = useAuthStore();
  
  return useQuery({
    queryKey: authKeys.user,
    queryFn: async () => {
      const response = await api.get<User>('/auth/me');
      return response.data;
    },
    enabled: !!accessToken && isAuthenticated,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: false,
  });
}

// Get user profile
export function useUserProfile() {
  const { isAuthenticated, user: authUser } = useAuthStore();
  
  return useQuery({
    queryKey: authKeys.profile,
    queryFn: async () => {
      // TODO: Backend doesn't have /profile endpoint yet
      // Using /auth/me as fallback, but it needs to be fully implemented
      try {
        const response = await api.get<User>('/auth/me');
        const user = response.data;
        // Return a minimal profile structure until backend implements full profile endpoint
        // Add defaults for missing fields from authUser store as fallback
        return {
          id: user.id || authUser?.id || '',
          email: user.email || authUser?.email || '',
          displayName: user.displayName || authUser?.displayName || '',
          avatarUrl: user.avatarUrl || authUser?.avatarUrl,
          role: user.role || (authUser?.roles?.[0] as any) || 'user',
          level: user.level ?? 1,
          xp: user.xp ?? 0,
          createdAt: user.createdAt || new Date().toISOString(),
          readChaptersCount: 0,
          readingTime: 0,
          commentsCount: 0,
          bookmarksCount: 0,
        } as UserProfile;
      } catch (error) {
        // If endpoint doesn't exist, return null to show error state
        throw error;
      }
    },
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: false, // Don't retry if endpoint doesn't exist
  });
}

// Login mutation
export function useLogin() {
  const queryClient = useQueryClient();
  const { login } = useAuthStore();
  
  return useMutation({
    mutationFn: async (credentials: LoginRequest) => {
      const response = await api.post<AuthResponse>('/auth/login', credentials);
      return response.data;
    },
    onSuccess: (data) => {
      const storeUser = convertApiUserToStoreUser(data.user);
      login(storeUser, data.access_token);
      queryClient.setQueryData(authKeys.user, data.user);
    },
  });
}

// Register mutation
export function useRegister() {
  const queryClient = useQueryClient();
  const { login } = useAuthStore();
  
  return useMutation({
    mutationFn: async (data: RegisterRequest) => {
      const response = await api.post<AuthResponse>('/auth/register', data);
      return response.data;
    },
    onSuccess: (data) => {
      const storeUser = convertApiUserToStoreUser(data.user);
      login(storeUser, data.access_token);
      queryClient.setQueryData(authKeys.user, data.user);
    },
  });
}

// Logout mutation
export function useLogout() {
  const queryClient = useQueryClient();
  const { logout } = useAuthStore();
  
  return useMutation({
    mutationFn: async () => {
      await api.post('/auth/logout');
    },
    onSuccess: () => {
      logout();
      queryClient.clear();
    },
    onError: () => {
      // Even if logout fails, clear local state
      logout();
      queryClient.clear();
    },
  });
}

// Refresh token mutation
export function useRefreshToken() {
  const { setAccessToken } = useAuthStore();
  
  return useMutation({
    mutationFn: async () => {
      const response = await api.post<{ access_token: string }>('/auth/refresh');
      return response.data;
    },
    onSuccess: (data) => {
      setAccessToken(data.access_token);
    },
  });
}

// Update profile mutation
export function useUpdateProfile() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (data: Partial<UserProfile>) => {
      const response = await api.put<UserProfile>('/profile', data);
      return response.data;
    },
    onSuccess: (data) => {
      queryClient.setQueryData(authKeys.profile, data);
      // Also update the user in auth cache if needed
      queryClient.invalidateQueries({ queryKey: authKeys.user });
    },
  });
}

// Change password mutation
export function useChangePassword() {
  return useMutation({
    mutationFn: async ({
      currentPassword,
      newPassword,
    }: {
      currentPassword: string;
      newPassword: string;
    }) => {
      await api.post('/auth/change-password', {
        currentPassword,
        newPassword,
      });
    },
  });
}
