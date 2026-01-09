import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../client';
import type { Chapter, ChapterWithContent, ReadingProgress } from '../types';

// Query keys
export const chapterKeys = {
  all: ['chapters'] as const,
  lists: () => [...chapterKeys.all, 'list'] as const,
  list: (novelSlug: string) => [...chapterKeys.lists(), novelSlug] as const,
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
    queryKey: chapterKeys.list(novelSlug),
    queryFn: async () => {
      const response = await api.get<Chapter[]>(
        `/novels/${novelSlug}/chapters?page=${page}&limit=${limit}`
      );
      return response;
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
      const response = await api.get<ChapterWithContent>(
        `/novels/${novelSlug}/chapters/${chapterId}${params}`
      );
      return response.data;
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
    mutationFn: async ({
      novelId,
      chapterId,
    }: {
      novelId: string;
      chapterId: string;
    }) => {
      const response = await api.post<ReadingProgress>('/progress', {
        novelId,
        chapterId,
      });
      return response.data;
    },
    onSuccess: (data, { novelId }) => {
      // Update progress cache
      queryClient.setQueryData(progressKeys.novel(novelId), data);
      // Invalidate user progress list
      queryClient.invalidateQueries({ queryKey: progressKeys.user() });
    },
  });
}

// Mark chapter as read (alias for save progress)
export function useMarkChapterRead() {
  return useSaveProgress();
}
