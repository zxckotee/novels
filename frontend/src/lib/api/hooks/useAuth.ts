import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../client';
import { useAuthStore } from '@/store/auth';
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
  accessToken: string;
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
  const { isAuthenticated } = useAuthStore();
  
  return useQuery({
    queryKey: authKeys.profile,
    queryFn: async () => {
      const response = await api.get<UserProfile>('/profile');
      return response.data;
    },
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

// Login mutation
export function useLogin() {
  const queryClient = useQueryClient();
  const { setAuth } = useAuthStore();
  
  return useMutation({
    mutationFn: async (credentials: LoginRequest) => {
      const response = await api.post<AuthResponse>('/auth/login', credentials);
      return response.data;
    },
    onSuccess: (data) => {
      setAuth(data.user, data.accessToken);
      queryClient.setQueryData(authKeys.user, data.user);
    },
  });
}

// Register mutation
export function useRegister() {
  const queryClient = useQueryClient();
  const { setAuth } = useAuthStore();
  
  return useMutation({
    mutationFn: async (data: RegisterRequest) => {
      const response = await api.post<AuthResponse>('/auth/register', data);
      return response.data;
    },
    onSuccess: (data) => {
      setAuth(data.user, data.accessToken);
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
      const response = await api.post<{ accessToken: string }>('/auth/refresh');
      return response.data;
    },
    onSuccess: (data) => {
      setAccessToken(data.accessToken);
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
