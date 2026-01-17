import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../client';
import type { AppSetting, UpdateSettingRequest, AdminAuditLog, AuditLogsResponse, AdminStats } from '../types';

export function useAdminSettings() {
  return useQuery<AppSetting[]>({
    queryKey: ['admin', 'settings'],
    queryFn: async () => {
      const { data } = await api.get<AppSetting[]>('/admin/settings');
      return data;
    },
  });
}

export function useAdminSetting(key: string) {
  return useQuery<AppSetting>({
    queryKey: ['admin', 'settings', key],
    queryFn: async () => {
      const { data } = await api.get<AppSetting>(`/admin/settings/${key}`);
      return data;
    },
    enabled: !!key,
  });
}

export function useUpdateSetting(key: string) {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (value: any) => {
      await api.put(`/admin/settings/${key}`, { value });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'settings'] });
      queryClient.invalidateQueries({ queryKey: ['admin', 'settings', key] });
    },
  });
}

export function useAdminLogs(params?: {
  actorUserId?: string;
  action?: string;
  entityType?: string;
  entityId?: string;
  startDate?: string;
  endDate?: string;
  page?: number;
  limit?: number;
}) {
  return useQuery<AuditLogsResponse>({
    queryKey: ['admin', 'logs', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams();
      if (params?.actorUserId) searchParams.set('actorUserId', params.actorUserId);
      if (params?.action) searchParams.set('action', params.action);
      if (params?.entityType) searchParams.set('entityType', params.entityType);
      if (params?.entityId) searchParams.set('entityId', params.entityId);
      if (params?.startDate) searchParams.set('startDate', params.startDate);
      if (params?.endDate) searchParams.set('endDate', params.endDate);
      if (params?.page) searchParams.set('page', params.page.toString());
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      
      const { data } = await api.get<AuditLogsResponse>(
        `/admin/logs?${searchParams.toString()}`
      );
      return data;
    },
  });
}

export function useAdminStats() {
  return useQuery<AdminStats>({
    queryKey: ['admin', 'stats'],
    queryFn: async () => {
      const { data } = await api.get<AdminStats>('/admin/stats');
      return data;
    },
  });
}
