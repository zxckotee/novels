import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../client';
import type { Novel, NovelListItem, NovelFilters, PaginationMeta, Genre, Tag } from '../types';

// Query keys
export const novelKeys = {
  all: ['novels'] as const,
  lists: () => [...novelKeys.all, 'list'] as const,
  list: (filters: NovelFilters) => [...novelKeys.lists(), filters] as const,
  details: () => [...novelKeys.all, 'detail'] as const,
  detail: (slug: string) => [...novelKeys.details(), slug] as const,
  popular: (period: string) => [...novelKeys.all, 'popular', period] as const,
  trending: () => [...novelKeys.all, 'trending'] as const,
  latest: () => [...novelKeys.all, 'latest'] as const,
  search: (query: string) => [...novelKeys.all, 'search', query] as const,
};

export const metaKeys = {
  genres: ['genres'] as const,
  tags: ['tags'] as const,
};

// Fetch novels list with filters
export function useNovels(filters: NovelFilters = {}) {
  return useQuery({
    queryKey: novelKeys.list(filters),
    queryFn: async () => {
      const params = new URLSearchParams();
      
      if (filters.search) params.append('search', filters.search);
      if (filters.genres?.length) params.append('genres', filters.genres.join(','));
      if (filters.tags?.length) params.append('tags', filters.tags.join(','));
      if (filters.status) params.append('status', filters.status);
      if (filters.sort) params.append('sort', filters.sort);
      if (filters.order) params.append('order', filters.order);
      if (filters.page) params.append('page', String(filters.page));
      if (filters.limit) params.append('limit', String(filters.limit));
      if (filters.lang) params.append('lang', filters.lang);
      
      const response = await api.get<NovelListItem[]>(`/novels?${params.toString()}`);
      return response;
    },
    staleTime: 30 * 1000, // 30 seconds
  });
}

// Fetch single novel by slug
export function useNovel(slug: string, lang?: string) {
  return useQuery({
    queryKey: novelKeys.detail(slug),
    queryFn: async () => {
      const params = lang ? `?lang=${lang}` : '';
      const response = await api.get<Novel>(`/novels/${slug}${params}`);
      return response.data;
    },
    enabled: !!slug,
    staleTime: 60 * 1000, // 1 minute
  });
}

// Fetch popular novels
export function usePopularNovels(period: 'day' | 'week' | 'month' = 'week', limit = 10) {
  return useQuery({
    queryKey: novelKeys.popular(period),
    queryFn: async () => {
      const response = await api.get<NovelListItem[]>(`/novels/popular?period=${period}&limit=${limit}`);
      return response.data;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

// Fetch trending novels
export function useTrendingNovels(limit = 10) {
  return useQuery({
    queryKey: novelKeys.trending(),
    queryFn: async () => {
      const response = await api.get<NovelListItem[]>(`/novels/trending?limit=${limit}`);
      return response.data;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

// Fetch latest updated novels
export function useLatestNovels(limit = 20) {
  return useQuery({
    queryKey: novelKeys.latest(),
    queryFn: async () => {
      const response = await api.get<NovelListItem[]>(`/novels/latest?limit=${limit}`);
      return response.data;
    },
    staleTime: 30 * 1000, // 30 seconds
  });
}

// Search novels
export function useSearchNovels(query: string, enabled = true) {
  return useQuery({
    queryKey: novelKeys.search(query),
    queryFn: async () => {
      const response = await api.get<NovelListItem[]>(`/novels/search?q=${encodeURIComponent(query)}`);
      return response.data;
    },
    enabled: enabled && query.length >= 2,
    staleTime: 60 * 1000, // 1 minute
  });
}

// Fetch all genres
export function useGenres() {
  return useQuery({
    queryKey: metaKeys.genres,
    queryFn: async () => {
      const response = await api.get<Genre[]>('/genres');
      return response.data;
    },
    staleTime: 24 * 60 * 60 * 1000, // 24 hours
  });
}

// Fetch all tags
export function useTags() {
  return useQuery({
    queryKey: metaKeys.tags,
    queryFn: async () => {
      const response = await api.get<Tag[]>('/tags');
      return response.data;
    },
    staleTime: 24 * 60 * 60 * 1000, // 24 hours
  });
}

// Rate a novel
export function useRateNovel() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ novelId, rating }: { novelId: string; rating: number }) => {
      const response = await api.post<{ rating: number }>(`/novels/${novelId}/rate`, { rating });
      return response.data;
    },
    onSuccess: (_, { novelId }) => {
      // Invalidate the novel detail cache
      queryClient.invalidateQueries({ queryKey: novelKeys.details() });
    },
  });
}
