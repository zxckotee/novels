import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../client';
import type { Chapter, ChapterWithContent, ReadingProgress } from '../types';

type BackendChapterListItem = {
  id: string;
  number: number;
  slug?: string | null;
  title?: string | null;
  views: number;
  published_at?: string | null;
  is_read: boolean;
  is_new: boolean;
  comments_count: number;
};

type BackendNovelBrief = {
  id: string;
  slug: string;
  title: string;
  cover_url?: string | null;
};

type BackendChaptersListResponse = {
  chapters: BackendChapterListItem[];
  novel?: BackendNovelBrief | null;
  pagination?: { page: number; limit: number; total: number; totalPages: number };
};

type BackendChapterWithContent = {
  id: string;
  novel_id: string;
  number: number;
  slug?: string | null;
  title?: string | null;
  views: number;
  published_at?: string | null;
  created_at: string;
  updated_at: string;

  content: string;
  word_count: number;
  reading_time_minutes: number;
  source: string;

  prev_chapter?: { id: string; number: number; title?: string | null } | null;
  next_chapter?: { id: string; number: number; title?: string | null } | null;
  novel_slug: string;
  novel_title: string;
};

function mapBackendChapterListItem(
  item: BackendChapterListItem,
  novelId?: string
): Chapter {
  return {
    id: item.id,
    novelId: novelId || '',
    number: item.number,
    slug: item.slug ?? undefined,
    title: item.title ?? `Глава ${item.number}`,
    wordCount: 0,
    publishedAt: item.published_at ?? new Date(0).toISOString(),
    createdAt: new Date(0).toISOString(),
    updatedAt: new Date(0).toISOString(),
  };
}

function mapBackendChapterWithContent(ch: BackendChapterWithContent): ChapterWithContent {
  return {
    id: ch.id,
    novelId: ch.novel_id,
    number: ch.number,
    slug: ch.slug ?? undefined,
    title: ch.title ?? `Глава ${ch.number}`,
    wordCount: ch.word_count ?? 0,
    publishedAt: ch.published_at ?? ch.created_at,
    createdAt: ch.created_at,
    updatedAt: ch.updated_at,
    content: ch.content,
    novel: {
      id: ch.novel_id,
      slug: ch.novel_slug,
      title: ch.novel_title,
    },
    prevChapter: ch.prev_chapter
      ? { id: ch.prev_chapter.id, number: ch.prev_chapter.number, title: ch.prev_chapter.title ?? undefined }
      : undefined,
    nextChapter: ch.next_chapter
      ? { id: ch.next_chapter.id, number: ch.next_chapter.number, title: ch.next_chapter.title ?? undefined }
      : undefined,
  };
}

// Query keys
export const chapterKeys = {
  all: ['chapters'] as const,
  lists: () => [...chapterKeys.all, 'list'] as const,
  list: (novelSlug: string, page: number, limit: number) => [...chapterKeys.lists(), novelSlug, page, limit] as const,
  details: () => [...chapterKeys.all, 'detail'] as const,
  detail: (novelSlug: string, chapterId: string) => [...chapterKeys.details(), novelSlug, chapterId] as const,
};

export const progressKeys = {
  all: ['progress'] as const,
  novel: (novelId: string) => [...progressKeys.all, novelId] as const,
  user: () => [...progressKeys.all, 'user'] as const,
};

// Fetch chapters list for a novel
export function useChapters(novelSlug: string, page = 1, limit = 100) {
  return useQuery({
    queryKey: chapterKeys.list(novelSlug, page, limit),
    queryFn: async () => {
      const response = await api.get<BackendChaptersListResponse>(
        `/novels/${novelSlug}/chapters?page=${page}&limit=${limit}`
      );
      // Backend typically wraps payload as { data: T, meta: { timestamp } }.
      // Some endpoints might be unwrapped; support both shapes defensively.
      const payload: BackendChaptersListResponse =
        // @ts-expect-error - defensive runtime shape support
        (response as any)?.data?.data && typeof (response as any)?.data?.data === 'object'
          ? // @ts-expect-error - defensive runtime shape support
            (response as any).data.data
          : response.data;

      const novelId = payload?.novel?.id;
      const chapters = (payload?.chapters || []).map((c) => mapBackendChapterListItem(c, novelId));
      const p = payload?.pagination;
      const meta = p
        ? { page: p.page, limit: p.limit, total: p.total, totalPages: p.totalPages }
        : undefined;
      return { data: chapters, meta };
    },
    enabled: !!novelSlug,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

// Fetch single chapter with content
export function useChapter(novelSlug: string, chapterId: string, lang?: string) {
  return useQuery({
    queryKey: chapterKeys.detail(novelSlug, chapterId),
    queryFn: async () => {
      const params = lang ? `?lang=${lang}` : '';
      const response = await api.get<BackendChapterWithContent>(`/chapters/${chapterId}${params}`);
      return mapBackendChapterWithContent(response.data);
    },
    enabled: !!novelSlug && !!chapterId,
    staleTime: 10 * 60 * 1000, // 10 minutes
  });
}

// Fetch reading progress for a novel
export function useReadingProgress(novelId: string) {
  return useQuery({
    queryKey: progressKeys.novel(novelId),
    queryFn: async () => {
      const response = await api.get<ReadingProgress>(`/progress/${novelId}`);
      return response.data;
    },
    enabled: !!novelId,
    staleTime: 0, // Always fresh
    retry: false, // No progress is OK
  });
}

// Fetch all user's reading progress
export function useAllReadingProgress() {
  return useQuery({
    queryKey: progressKeys.user(),
    queryFn: async () => {
      const response = await api.get<ReadingProgress[]>('/progress');
      return response.data;
    },
    staleTime: 30 * 1000, // 30 seconds
  });
}

// Save reading progress
export function useSaveProgress() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async ({ chapterId, position }: { chapterId: string; position: number }) => {
      const response = await api.post<{ message: string }>(`/chapters/${chapterId}/progress`, { position });
      return response.data;
    },
    onSuccess: () => {
      // Best-effort: progress endpoints are novel-scoped; invalidate all
      queryClient.invalidateQueries({ queryKey: progressKeys.all });
    },
  });
}

// Mark chapter as read (alias for save progress)
export function useMarkChapterRead() {
  return useSaveProgress();
}
