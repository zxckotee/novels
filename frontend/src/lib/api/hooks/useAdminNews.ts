import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../client';
import type { NewsPost } from '../types';

export function useAdminNews(params?: {
  page?: number;
  limit?: number;
}) {
  return useQuery({
    queryKey: ['admin', 'news', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams();
      if (params?.page) searchParams.set('page', params.page.toString());
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      
      const { data } = await api.get<{ news: NewsPost[]; total: number }>(
        `/admin/news?${searchParams.toString()}`
      );
      return data;
    },
  });
}

export function useDeleteNews() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/admin/news/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'news'] });
    },
  });
}

export function usePublishNews() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (id: string) => {
      await api.post(`/admin/news/${id}/publish`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'news'] });
    },
  });
}

export function usePinNews() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ id, pinned }: { id: string; pinned: boolean }) => {
      await api.post(`/admin/news/${id}/pin`, { pinned });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'news'] });
    },
  });
}
