'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import Image from 'next/image';
import { 
  BookOpen, 
  Clock, 
  Trash2, 
  BookMarked,
  Heart,
  Pause,
  Check,
  ChevronRight,
  Loader2,
  Star
} from 'lucide-react';
import { apiClient } from '@/lib/api/client';
import { useAuthStore } from '@/store/auth';

interface BookmarkNovel {
  id: string;
  slug: string;
  coverImageKey?: string;
  translationStatus: string;
  title: string;
  chaptersCount: number;
  rating: number;
}

interface ReadingProgress {
  chapterId: string;
  chapterNum: number;
  totalChapters: number;
}

interface Bookmark {
  id: string;
  novelId: string;
  listId: string;
  createdAt: string;
  novel?: BookmarkNovel;
  progress?: ReadingProgress;
  hasNewChapter: boolean;
}

interface BookmarkList {
  id: string;
  code: string;
  title: string;
  count: number;
}

interface BookmarksResponse {
  bookmarks: Bookmark[];
  lists: BookmarkList[];
  totalCount: number;
  page: number;
  limit: number;
}

const LIST_ICONS: Record<string, React.ReactNode> = {
  reading: <BookOpen className="w-5 h-5" />,
  planned: <Clock className="w-5 h-5" />,
  dropped: <Trash2 className="w-5 h-5" />,
  completed: <Check className="w-5 h-5" />,
  favorites: <Heart className="w-5 h-5" />,
};

const LIST_COLORS: Record<string, string> = {
  reading: 'text-green-500',
  planned: 'text-blue-500',
  dropped: 'text-gray-500',
  completed: 'text-purple-500',
  favorites: 'text-red-500',
};

interface BookmarksPageClientProps {
  locale: string;
}

export default function BookmarksPageClient({ locale }: BookmarksPageClientProps) {
  const t = useTranslations('bookmarks');
  const { isAuthenticated, isLoading: authLoading } = useAuthStore();
  const queryClient = useQueryClient();
  
  const [activeList, setActiveList] = useState<string | null>(null);
  const [sort, setSort] = useState<'latest_update' | 'date_added' | 'title'>('latest_update');
  const [page, setPage] = useState(1);

  // Fetch bookmarks
  const { data, isLoading, error } = useQuery<BookmarksResponse>({
    queryKey: ['bookmarks', activeList, sort, page],
    queryFn: async () => {
      const params: Record<string, string | number> = {
        sort,
        page,
        limit: 20,
      };
      if (activeList) {
        params.list = activeList;
      }
      const response = await apiClient.get('/bookmarks', { params });
      return response.data;
    },
    enabled: isAuthenticated,
  });

  // Remove bookmark mutation
  const removeMutation = useMutation({
    mutationFn: async (novelId: string) => {
      return apiClient.delete(`/bookmarks/${novelId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bookmarks'] });
    },
  });

  // Move bookmark mutation
  const moveMutation = useMutation({
    mutationFn: async ({ novelId, listCode }: { novelId: string; listCode: string }) => {
      return apiClient.put(`/bookmarks/${novelId}`, { listCode });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bookmarks'] });
    },
  });

  if (authLoading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <Loader2 className="w-8 h-8 animate-spin text-accent-primary" />
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="container mx-auto px-4 py-12">
        <div className="max-w-md mx-auto text-center">
          <BookMarked className="w-16 h-16 mx-auto mb-4 text-foreground-muted" />
          <h1 className="text-2xl font-bold text-foreground-primary mb-2">
            {t('loginRequired')}
          </h1>
          <p className="text-foreground-secondary mb-6">
            {t('loginRequiredDesc')}
          </p>
          <Link
            href={`/${locale}/login`}
            className="inline-block px-6 py-3 bg-accent-primary text-white rounded-lg hover:bg-accent-primary/90"
          >
            {t('login')}
          </Link>
        </div>
      </div>
    );
  }

  const getProgressPercent = (progress?: ReadingProgress) => {
    if (!progress || !progress.totalChapters) return 0;
    return Math.round((progress.chapterNum / progress.totalChapters) * 100);
  };

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground-primary mb-2">
          {t('title')}
        </h1>
        <p className="text-foreground-secondary">
          {t('subtitle', { count: data?.totalCount || 0 })}
        </p>
      </div>

      <div className="flex flex-col lg:flex-row gap-8">
        {/* Sidebar - Lists */}
        <div className="lg:w-64 flex-shrink-0">
          <div className="bg-background-secondary rounded-xl p-4 sticky top-24">
            <h3 className="font-semibold text-foreground-primary mb-4">{t('lists')}</h3>
            <nav className="space-y-1">
              <button
                onClick={() => {
                  setActiveList(null);
                  setPage(1);
                }}
                className={`w-full flex items-center justify-between px-3 py-2 rounded-lg transition-colors ${
                  !activeList
                    ? 'bg-accent-primary text-white'
                    : 'text-foreground-secondary hover:bg-background-tertiary'
                }`}
              >
                <span className="flex items-center gap-2">
                  <BookMarked className="w-5 h-5" />
                  {t('allBooks')}
                </span>
                <span className="text-sm opacity-70">
                  {data?.lists?.reduce((sum, l) => sum + l.count, 0) || 0}
                </span>
              </button>

              {data?.lists?.map((list) => (
                <button
                  key={list.id}
                  onClick={() => {
                    setActiveList(list.code);
                    setPage(1);
                  }}
                  className={`w-full flex items-center justify-between px-3 py-2 rounded-lg transition-colors ${
                    activeList === list.code
                      ? 'bg-accent-primary text-white'
                      : 'text-foreground-secondary hover:bg-background-tertiary'
                  }`}
                >
                  <span className={`flex items-center gap-2 ${activeList !== list.code ? LIST_COLORS[list.code] : ''}`}>
                    {LIST_ICONS[list.code]}
                    {list.title}
                  </span>
                  <span className="text-sm opacity-70">{list.count}</span>
                </button>
              ))}
            </nav>
          </div>
        </div>

        {/* Main content */}
        <div className="flex-1">
          {/* Sort controls */}
          <div className="flex items-center justify-between mb-6">
            <span className="text-foreground-secondary">
              {t('showing', { count: data?.bookmarks?.length || 0, total: data?.totalCount || 0 })}
            </span>
            <select
              value={sort}
              onChange={(e) => setSort(e.target.value as typeof sort)}
              className="bg-background-secondary border border-border-primary rounded-lg px-3 py-2 text-sm text-foreground-primary focus:outline-none focus:border-accent-primary"
            >
              <option value="latest_update">{t('sortLatestUpdate')}</option>
              <option value="date_added">{t('sortDateAdded')}</option>
              <option value="title">{t('sortTitle')}</option>
            </select>
          </div>

          {/* Bookmarks list */}
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="w-8 h-8 animate-spin text-accent-primary" />
            </div>
          ) : error ? (
            <div className="text-center py-12 text-red-500">
              {t('loadError')}
            </div>
          ) : (data?.bookmarks?.length || 0) === 0 ? (
            <div className="text-center py-12">
              <BookMarked className="w-16 h-16 mx-auto mb-4 text-foreground-muted" />
              <h3 className="text-xl font-semibold text-foreground-primary mb-2">
                {t('emptyList')}
              </h3>
              <p className="text-foreground-secondary mb-4">
                {t('emptyListDesc')}
              </p>
              <Link
                href={`/${locale}/catalog`}
                className="inline-block px-4 py-2 bg-accent-primary text-white rounded-lg hover:bg-accent-primary/90"
              >
                {t('browseCatalog')}
              </Link>
            </div>
          ) : (
            <div className="space-y-4">
              {data?.bookmarks?.map((bookmark) => (
                <div
                  key={bookmark.id}
                  className="flex gap-4 p-4 bg-background-secondary rounded-xl hover:bg-background-tertiary transition-colors group"
                >
                  {/* Cover */}
                  <Link
                    href={`/${locale}/novel/${bookmark.novel?.slug}`}
                    className="relative flex-shrink-0"
                  >
                    <div className="w-20 h-28 md:w-24 md:h-32 relative rounded-lg overflow-hidden">
                      {bookmark.novel?.coverImageKey ? (
                        <Image
                          src={`/uploads/${bookmark.novel.coverImageKey}`}
                          alt={bookmark.novel.title}
                          fill
                          className="object-cover"
                        />
                      ) : (
                        <div className="w-full h-full bg-background-tertiary flex items-center justify-center">
                          <BookOpen className="w-8 h-8 text-foreground-muted" />
                        </div>
                      )}
                      {bookmark.hasNewChapter && (
                        <div className="absolute top-1 right-1 px-1.5 py-0.5 bg-accent-primary text-white text-xs rounded">
                          NEW
                        </div>
                      )}
                    </div>
                  </Link>

                  {/* Info */}
                  <div className="flex-1 min-w-0">
                    <Link
                      href={`/${locale}/novel/${bookmark.novel?.slug}`}
                      className="block"
                    >
                      <h3 className="font-semibold text-foreground-primary truncate hover:text-accent-primary">
                        {bookmark.novel?.title}
                      </h3>
                    </Link>
                    
                    <div className="flex items-center gap-3 mt-1 text-sm text-foreground-secondary">
                      <span className="flex items-center gap-1">
                        <Star className="w-4 h-4 text-yellow-500" />
                        {bookmark.novel?.rating?.toFixed(1) || '—'}
                      </span>
                      <span>
                        {bookmark.novel?.chaptersCount || 0} {t('chapters')}
                      </span>
                    </div>

                    {/* Progress bar */}
                    {bookmark.progress && (
                      <div className="mt-3">
                        <div className="flex items-center justify-between text-sm mb-1">
                          <span className="text-foreground-secondary">
                            {t('progress')}: {bookmark.progress.chapterNum} / {bookmark.progress.totalChapters}
                          </span>
                          <span className="text-foreground-muted">
                            {getProgressPercent(bookmark.progress)}%
                          </span>
                        </div>
                        <div className="h-1.5 bg-background-tertiary rounded-full overflow-hidden">
                          <div
                            className="h-full bg-accent-primary rounded-full transition-all"
                            style={{ width: `${getProgressPercent(bookmark.progress)}%` }}
                          />
                        </div>
                      </div>
                    )}

                    {/* Move to list dropdown */}
                    <div className="mt-3 flex items-center gap-2">
                      <select
                        defaultValue=""
                        onChange={(e) => {
                          if (e.target.value) {
                            moveMutation.mutate({
                              novelId: bookmark.novelId,
                              listCode: e.target.value,
                            });
                            e.target.value = '';
                          }
                        }}
                        className="text-sm bg-background-primary border border-border-primary rounded px-2 py-1 text-foreground-secondary focus:outline-none focus:border-accent-primary"
                      >
                        <option value="">{t('moveTo')}</option>
                        {data?.lists
                          ?.filter((l) => l.id !== bookmark.listId)
                          ?.map((list) => (
                            <option key={list.id} value={list.code}>
                              {list.title}
                            </option>
                          ))}
                      </select>

                      <button
                        onClick={() => {
                          if (confirm(t('confirmRemove'))) {
                            removeMutation.mutate(bookmark.novelId);
                          }
                        }}
                        className="p-1 text-foreground-muted hover:text-red-500 transition-colors"
                        title={t('remove')}
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>

                  {/* Continue reading button */}
                  <div className="flex-shrink-0 flex items-center">
                    <Link
                      href={bookmark.progress 
                        ? `/${locale}/novel/${bookmark.novel?.slug}/chapter/${bookmark.progress.chapterId}`
                        : `/${locale}/novel/${bookmark.novel?.slug}`
                      }
                      className="flex items-center gap-2 px-4 py-2 bg-accent-primary text-white rounded-lg hover:bg-accent-primary/90 opacity-0 group-hover:opacity-100 transition-opacity"
                    >
                      {t('continue')}
                      <ChevronRight className="w-4 h-4" />
                    </Link>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Pagination */}
          {data && data.totalCount > 20 && (
            <div className="flex justify-center gap-2 mt-8">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="px-4 py-2 bg-background-secondary rounded-lg text-foreground-primary hover:bg-background-tertiary disabled:opacity-50"
              >
                {t('prevPage')}
              </button>
              <span className="px-4 py-2 text-foreground-secondary">
                {page} / {Math.ceil((data.totalCount || 1) / 20)}
              </span>
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={page >= Math.ceil((data.totalCount || 1) / 20)}
                className="px-4 py-2 bg-background-secondary rounded-lg text-foreground-primary hover:bg-background-tertiary disabled:opacity-50"
              >
                {t('nextPage')}
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
