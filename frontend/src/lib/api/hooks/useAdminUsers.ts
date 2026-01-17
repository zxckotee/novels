import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../client';
import type { UserAdmin, UsersResponse, BanUserRequest, UpdateUserRolesRequest } from '../types';

export function useAdminUsers(params?: {
  query?: string;
  role?: string;
  banned?: boolean;
  page?: number;
  limit?: number;
}) {
  return useQuery<UsersResponse>({
    queryKey: ['admin', 'users', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams();
      if (params?.query) searchParams.set('query', params.query);
      if (params?.role) searchParams.set('role', params.role);
      if (params?.banned !== undefined) searchParams.set('banned', params.banned.toString());
      if (params?.page) searchParams.set('page', params.page.toString());
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      
      const { data } = await api.get<UsersResponse>(
        `/admin/users?${searchParams.toString()}`
      );
      return data;
    },
  });
}

export function useAdminUser(id: string) {
  return useQuery<UserAdmin>({
    queryKey: ['admin', 'users', id],
    queryFn: async () => {
      const { data } = await api.get<UserAdmin>(`/admin/users/${id}`);
      return data;
    },
    enabled: !!id,
  });
}

export function useBanUser() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ id, reason }: { id: string; reason: string }) => {
      await api.post(`/admin/users/${id}/ban`, { reason });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
}

export function useUnbanUser() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (id: string) => {
      await api.post(`/admin/users/${id}/unban`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
}

export function useUpdateUserRoles() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ id, roles }: { id: string; roles: string[] }) => {
      await api.put(`/admin/users/${id}/roles`, { roles });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
  });
}
