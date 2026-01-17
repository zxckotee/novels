import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../client';
import type {
  GenreAdmin,
  TagAdmin,
  GenresResponse,
  TagsResponse,
  CreateGenreRequest,
  CreateTagRequest,
} from '../types';

// Genres
export function useAdminGenres(params?: {
  query?: string;
  lang?: string;
  page?: number;
  limit?: number;
}) {
  return useQuery<GenresResponse>({
    queryKey: ['admin', 'genres', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams();
      if (params?.query) searchParams.set('query', params.query);
      if (params?.lang) searchParams.set('lang', params.lang);
      if (params?.page) searchParams.set('page', params.page.toString());
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      
      const { data } = await api.get<GenresResponse>(
        `/admin/genres?${searchParams.toString()}`
      );
      return data;
    },
  });
}

export function useAdminGenre(id: string) {
  return useQuery<GenreAdmin>({
    queryKey: ['admin', 'genres', id],
    queryFn: async () => {
      const { data } = await api.get<GenreAdmin>(`/admin/genres/${id}`);
      return data;
    },
    enabled: !!id,
  });
}

export function useCreateGenre() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (req: CreateGenreRequest) => {
      const { data } = await api.post<GenreAdmin>('/admin/genres', req);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'genres'] });
    },
  });
}

export function useUpdateGenre(id: string) {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (req: Partial<CreateGenreRequest>) => {
      const { data } = await api.put<GenreAdmin>(`/admin/genres/${id}`, req);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'genres'] });
      queryClient.invalidateQueries({ queryKey: ['admin', 'genres', id] });
    },
  });
}

export function useDeleteGenre() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/admin/genres/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'genres'] });
    },
  });
}

// Tags
export function useAdminTags(params?: {
  query?: string;
  lang?: string;
  page?: number;
  limit?: number;
}) {
  return useQuery<TagsResponse>({
    queryKey: ['admin', 'tags', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams();
      if (params?.query) searchParams.set('query', params.query);
      if (params?.lang) searchParams.set('lang', params.lang);
      if (params?.page) searchParams.set('page', params.page.toString());
      if (params?.limit) searchParams.set('limit', params.limit.toString());
      
      const { data } = await api.get<TagsResponse>(
        `/admin/tags?${searchParams.toString()}`
      );
      return data;
    },
  });
}

export function useAdminTag(id: string) {
  return useQuery<TagAdmin>({
    queryKey: ['admin', 'tags', id],
    queryFn: async () => {
      const { data } = await api.get<TagAdmin>(`/admin/tags/${id}`);
      return data;
    },
    enabled: !!id,
  });
}

export function useCreateTag() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (req: CreateTagRequest) => {
      const { data } = await api.post<TagAdmin>('/admin/tags', req);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'tags'] });
    },
  });
}

export function useUpdateTag(id: string) {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (req: Partial<CreateTagRequest>) => {
      const { data } = await api.put<TagAdmin>(`/admin/tags/${id}`, req);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'tags'] });
      queryClient.invalidateQueries({ queryKey: ['admin', 'tags', id] });
    },
  });
}

export function useDeleteTag() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/admin/tags/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'tags'] });
    },
  });
}
