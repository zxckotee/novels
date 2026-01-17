import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../client';
import type { Comment, CommentsResponse, CommentReport, ReportsResponse, ResolveReportRequest } from '../types';

export function useAdminComments(params?: {
  targetType?: string;
  targetId?: string;
  userId?: string;
  isDeleted?: boolean;
  page?: number;
  limit?: number;
}) {
  return useQuery<CommentsResponse>({
    queryKey: ['admin', 'comments', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams();
      if (params?.targetType) searchParams.set('targetType', params.targetType);
      if (params?.targetId) searchParams.set('targetId', params.targetId);
      if (params?.userId) searchParams.set('userId', params.userId);
      if (params?.isDeleted !== undefined) searchParams.set('isDeleted', params.isDeleted.toString());
      if (params?.page) searchParams.set('page', params.page.toString());
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      
      const { data } = await api.get<CommentsResponse>(
        `/admin/comments?${searchParams.toString()}`
      );
      return data;
    },
  });
}

export function useSoftDeleteComment() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/admin/comments/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'comments'] });
    },
  });
}

export function useHardDeleteComment() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/admin/comments/${id}/hard`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'comments'] });
    },
  });
}

export function useAdminReports(params?: {
  status?: string;
  commentId?: string;
  reporterId?: string;
  page?: number;
  limit?: number;
}) {
  return useQuery<ReportsResponse>({
    queryKey: ['admin', 'reports', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams();
      if (params?.status) searchParams.set('status', params.status);
      if (params?.commentId) searchParams.set('commentId', params.commentId);
      if (params?.reporterId) searchParams.set('reporterId', params.reporterId);
      if (params?.page) searchParams.set('page', params.page.toString());
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      
      const { data } = await api.get<ReportsResponse>(
        `/admin/reports?${searchParams.toString()}`
      );
      return data;
    },
  });
}

export function useResolveReport() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ id, action, reason }: { id: string; action: string; reason?: string }) => {
      await api.post(`/admin/reports/${id}/resolve`, { action, reason });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'reports'] });
      queryClient.invalidateQueries({ queryKey: ['admin', 'comments'] });
    },
  });
}
