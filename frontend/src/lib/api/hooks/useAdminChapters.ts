import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../client';
import type { Chapter } from '../types';

export function useAdminChapters(params?: {
  novelId?: string;
  page?: number;
  limit?: number;
}) {
  return useQuery({
    queryKey: ['admin', 'chapters', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams();
      if (params?.page) searchParams.set('page', params.page.toString());
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      searchParams.set('sort', 'number');
      searchParams.set('order', 'desc');
      
      // Use admin endpoint for full list
      const endpoint = `/admin/chapters?${searchParams.toString()}`;
      
      const { data } = await api.get<Chapter[]>(endpoint);
      return { chapters: Array.isArray(data) ? data : [], total: Array.isArray(data) ? data.length : 0 };
    },
    enabled: true,
  });
}

export function useDeleteChapter() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/admin/chapters/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'chapters'] });
    },
  });
}
