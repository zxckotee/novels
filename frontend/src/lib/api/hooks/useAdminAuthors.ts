import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../client';
import type { AuthorAdmin, AuthorsResponse, CreateAuthorRequest, UpdateAuthorRequest } from '../types';

export function useAdminAuthors(params?: {
  query?: string;
  lang?: string;
  page?: number;
  limit?: number;
}) {
  return useQuery<AuthorsResponse>({
    queryKey: ['admin', 'authors', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams();
      if (params?.query) searchParams.set('query', params.query);
      if (params?.lang) searchParams.set('lang', params.lang);
      if (params?.page) searchParams.set('page', params.page.toString());
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      
      const { data } = await api.get<AuthorsResponse>(
        `/admin/authors?${searchParams.toString()}`
      );
      return data;
    },
  });
}

export function useAdminAuthor(id: string) {
  return useQuery<AuthorAdmin>({
    queryKey: ['admin', 'authors', id],
    queryFn: async () => {
      const { data } = await api.get<AuthorAdmin>(`/admin/authors/${id}`);
      return data;
    },
    enabled: !!id,
  });
}

export function useCreateAuthor() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (req: CreateAuthorRequest) => {
      const { data } = await api.post<AuthorAdmin>('/admin/authors', req);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'authors'] });
    },
  });
}

export function useUpdateAuthor(id: string) {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (req: UpdateAuthorRequest) => {
      const { data } = await api.put<AuthorAdmin>(`/admin/authors/${id}`, req);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'authors'] });
      queryClient.invalidateQueries({ queryKey: ['admin', 'authors', id] });
    },
  });
}

export function useDeleteAuthor() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/admin/authors/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'authors'] });
    },
  });
}
