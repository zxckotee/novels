import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../client';
import type { Novel, NovelListItem, NovelFilters, PaginationMeta, Genre, Tag } from '../types';

type BackendNovelListPagination = {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

type BackendNovelListItem = {
  id: string;
  slug: string;
  title: string;
  cover_url?: string | null;
  translation_status?: 'ongoing' | 'completed' | 'paused' | 'dropped';
  rating?: number;
  rating_count?: number;
  updated_at?: string;
  description?: string | null;
  // Backend uses snake_case; we only map what UI needs for lists
};

type BackendNovelListResponse = {
  novels: BackendNovelListItem[];
  pagination: BackendNovelListPagination;
};

type BackendNovelDetail = {
  id: string;
  slug: string;
  cover_url?: string | null;
  translation_status: 'ongoing' | 'completed' | 'paused' | 'dropped';
  original_chapters_count: number;
  release_year?: number | null;
  author?: string | null;
  rating: number;
  rating_count: number;
  views_total: number;
  bookmarks_count: number;
  created_at: string;
  updated_at: string;
  title: string;
  description?: string | null;
  alt_titles?: string[];
  genres?: Genre[];
  tags?: Tag[];
};

function mapBackendNovelToNovelListItem(novel: BackendNovelListItem): NovelListItem {
  return {
    id: novel.id,
    slug: novel.slug,
    title: novel.title,
    description: novel.description ?? undefined,
    coverUrl: novel.cover_url ?? undefined,
    rating: novel.rating ?? 0,
    // Fallback: UI uses updatedAt only for relative time
    updatedAt: novel.updated_at ?? new Date(0).toISOString(),
    translationStatus: novel.translation_status ?? 'ongoing',
  };
}

function mapBackendNovelDetailToNovel(novel: BackendNovelDetail): Novel {
  return {
    id: novel.id,
    slug: novel.slug,
    coverUrl: novel.cover_url ?? undefined,
    translationStatus: novel.translation_status,
    originalChaptersCount: novel.original_chapters_count ?? 0,
    releaseYear: novel.release_year ?? 0,
    rating: typeof novel.rating === 'number' ? novel.rating : 0,
    ratingsCount: novel.rating_count ?? 0,
    viewsCount: novel.views_total ?? 0,
    bookmarksCount: novel.bookmarks_count ?? 0,
    // Backend doesn't currently provide these in NovelDetail; keep safe defaults
    chaptersCount: 0,
    createdAt: novel.created_at,
    updatedAt: novel.updated_at,
    title: novel.title,
    description: novel.description ?? undefined,
    altTitles: novel.alt_titles ?? undefined,
    genres: novel.genres,
    tags: novel.tags,
    author: novel.author
      ? {
          id: 'legacy',
          name: novel.author,
        }
      : undefined,
  };
}

function mapFrontendSortToBackend(sort?: NovelFilters['sort']): string | undefined {
  switch (sort) {
    case 'popular':
      return 'views_daily';
    case 'views':
      return 'views_total';
    case 'bookmarks':
      return 'bookmarks_count';
    case 'rating':
      return 'rating';
    case 'created':
      return 'created_at';
    case 'updated':
      return 'updated_at';
    default:
      return undefined;
  }
}

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
      
      // Backend search is a separate endpoint: /novels/search?q=...
      const isSearch = Boolean(filters.search && filters.search.trim().length > 0);
      if (filters.genres?.length) params.append('genres', filters.genres.join(','));
      if (filters.tags?.length) params.append('tags', filters.tags.join(','));
      if (filters.status) params.append('status', filters.status);
      const backendSort = mapFrontendSortToBackend(filters.sort);
      if (backendSort) params.append('sort', backendSort);
      if (filters.order) params.append('order', filters.order);
      if (filters.page) params.append('page', String(filters.page));
      if (filters.limit) params.append('limit', String(filters.limit));
      if (filters.lang) params.append('lang', filters.lang);
      
      const url = isSearch
        ? `/novels/search?q=${encodeURIComponent(filters.search!.trim())}&${params.toString()}`
        : `/novels?${params.toString()}`;

      const response = await api.get<BackendNovelListResponse>(url);

      const novels = (response.data?.novels || []).map(mapBackendNovelToNovelListItem);
      const pagination = response.data?.pagination;

      const meta: PaginationMeta | undefined = pagination
        ? {
            page: pagination.page,
            limit: pagination.limit,
            total: pagination.total,
            totalPages: pagination.totalPages,
          }
        : undefined;

      return { data: novels, meta };
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
      const response = await api.get<BackendNovelDetail>(`/novels/${slug}${params}`);
      return mapBackendNovelDetailToNovel(response.data);
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
      const response = await api.get<BackendNovelListItem[]>(`/novels/popular?period=${period}&limit=${limit}`);
      return (response.data || []).map(mapBackendNovelToNovelListItem);
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

// Fetch trending novels
export function useTrendingNovels(limit = 10) {
  return useQuery({
    queryKey: novelKeys.trending(),
    queryFn: async () => {
      const response = await api.get<BackendNovelListItem[]>(`/novels/trending?limit=${limit}`);
      return (response.data || []).map(mapBackendNovelToNovelListItem);
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

// Fetch latest updated novels
export function useLatestNovels(limit = 20) {
  return useQuery({
    queryKey: novelKeys.latest(),
    queryFn: async () => {
      const response = await api.get<BackendNovelListItem[]>(`/novels/latest?limit=${limit}`);
      return (response.data || []).map(mapBackendNovelToNovelListItem);
    },
    staleTime: 30 * 1000, // 30 seconds
  });
}

// Fetch new releases
export function useNewReleases(limit = 10) {
  return useQuery({
    queryKey: [...novelKeys.all, 'new', limit] as const,
    queryFn: async () => {
      const response = await api.get<BackendNovelListItem[]>(`/novels/new?limit=${limit}`);
      return (response.data || []).map(mapBackendNovelToNovelListItem);
    },
    staleTime: 60 * 1000,
  });
}

// Fetch top rated novels
export function useTopRatedNovels(limit = 10) {
  return useQuery({
    queryKey: [...novelKeys.all, 'top-rated', limit] as const,
    queryFn: async () => {
      const response = await api.get<BackendNovelListItem[]>(`/novels/top-rated?limit=${limit}`);
      return (response.data || []).map(mapBackendNovelToNovelListItem);
    },
    staleTime: 5 * 60 * 1000,
  });
}

// Search novels
export function useSearchNovels(query: string, enabled = true) {
  return useQuery({
    queryKey: novelKeys.search(query),
    queryFn: async () => {
      const response = await api.get<BackendNovelListResponse>(`/novels/search?q=${encodeURIComponent(query)}`);
      return (response.data?.novels || []).map(mapBackendNovelToNovelListItem);
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
