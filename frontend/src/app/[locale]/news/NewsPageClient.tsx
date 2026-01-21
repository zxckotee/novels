'use client';

import { useState, useEffect } from 'react';
import { useTranslations } from 'next-intl';
import Link from 'next/link';
import Image from 'next/image';

interface NewsPost {
  id: string;
  slug: string;
  title: string;
  summary: string;
  content: string;
  category: string;
  coverUrl: string;
  isPinned: boolean;
  author: {
    id: string;
    displayName: string;
    avatarUrl: string;
  };
  publishedAt: string;
  createdAt: string;
}

interface NewsListResponse {
  news: NewsPost[];
  total: number;
  page: number;
  limit: number;
}

const CATEGORIES = [
  { value: '', label: 'all' },
  { value: 'announcement', label: 'announcement' },
  { value: 'update', label: 'update' },
  { value: 'event', label: 'event' },
  { value: 'contest', label: 'contest' },
  { value: 'guide', label: 'guide' },
  { value: 'other', label: 'other' },
];

interface NewsPageClientProps {
  locale: string;
  searchParams: { [key: string]: string | string[] | undefined };
}

export default function NewsPageClient({ locale, searchParams }: NewsPageClientProps) {
  const t = useTranslations('community');
  const [news, setNews] = useState<NewsPost[]>([]);
  const [pinnedNews, setPinnedNews] = useState<NewsPost[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [category, setCategory] = useState<string>((searchParams.category as string) || '');

  useEffect(() => {
    loadPinnedNews();
  }, []);

  useEffect(() => {
    loadNews();
  }, [page, category, locale]);

  const loadPinnedNews = async () => {
    try {
      const response = await fetch('/api/v1/news/pinned');
      if (response.ok) {
        const data = await response.json();
        setPinnedNews(data.news || []);
      }
    } catch (err) {
      console.error('Failed to load pinned news:', err);
    }
  };

  const loadNews = async () => {
    setLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: '20',
        lang: locale,
      });
      if (category) {
        params.append('category', category);
      }

      const response = await fetch(`/api/v1/news?${params}`);
      if (!response.ok) throw new Error('Failed to load news');

      const data: NewsListResponse = await response.json();
      setNews(data.news || []);
      setTotal(data.total);
    } catch (err) {
      setError(t('loadError'));
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString(locale === 'ru' ? 'ru-RU' : 'en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const getCategoryColor = (cat: string) => {
    switch (cat) {
      case 'announcement':
        return 'bg-red-600';
      case 'update':
        return 'bg-blue-600';
      case 'event':
        return 'bg-purple-600';
      case 'contest':
        return 'bg-yellow-600';
      case 'guide':
        return 'bg-green-600';
      default:
        return 'bg-gray-600';
    }
  };

  const totalPages = Math.ceil(total / 20);

  return (
    <div className="min-h-screen bg-[#121212]">
      <div className="max-w-7xl mx-auto px-4 py-8">
        {/* Заголовок */}
        <h1 className="text-3xl font-bold text-white mb-2">{t('news')}</h1>
        <p className="text-gray-400 mb-8">{t('newsDescription')}</p>

        {/* Закрепленные новости */}
        {pinnedNews.length > 0 && (
          <section className="mb-12">
            <h2 className="text-xl font-bold text-white mb-4 flex items-center gap-2">
              <span className="text-red-500">📌</span>
              {t('pinnedNews')}
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {pinnedNews.map((post) => (
                <NewsCard
                  key={post.id}
                  post={post}
                  locale={locale}
                  formatDate={formatDate}
                  getCategoryColor={getCategoryColor}
                  pinned
                />
              ))}
            </div>
          </section>
        )}

        {/* Фильтр по категориям */}
        <div className="flex flex-wrap gap-2 mb-6">
          {CATEGORIES.map((cat) => (
            <button
              key={cat.value}
              onClick={() => {
                setCategory(cat.value);
                setPage(1);
              }}
              className={`px-4 py-2 rounded-lg font-medium transition-colors ${
                category === cat.value
                  ? 'bg-purple-600 text-white'
                  : 'bg-[#1a1a2e] text-gray-300 hover:bg-[#252540]'
              }`}
            >
              {t(`category.${cat.label}`)}
            </button>
          ))}
        </div>

        {/* Список новостей */}
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-4 border-purple-500 border-t-transparent" />
          </div>
        ) : error ? (
          <div className="text-center py-12">
            <p className="text-red-400">{error}</p>
            <button
              onClick={loadNews}
              className="mt-4 px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg"
            >
              {t('retry')}
            </button>
          </div>
        ) : news.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-400">{t('noNews')}</p>
          </div>
        ) : (
          <>
            <div className="space-y-6">
              {news.map((post) => (
                <NewsCard
                  key={post.id}
                  post={post}
                  locale={locale}
                  formatDate={formatDate}
                  getCategoryColor={getCategoryColor}
                />
              ))}
            </div>

            {/* Пагинация */}
            {totalPages > 1 && (
              <div className="flex justify-center mt-8 gap-2">
                <button
                  onClick={() => setPage(p => Math.max(1, p - 1))}
                  disabled={page === 1}
                  className="px-4 py-2 bg-[#1a1a2e] text-white rounded-lg disabled:opacity-50"
                >
                  ←
                </button>
                {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                  let pageNum: number;
                  if (totalPages <= 5) {
                    pageNum = i + 1;
                  } else if (page <= 3) {
                    pageNum = i + 1;
                  } else if (page >= totalPages - 2) {
                    pageNum = totalPages - 4 + i;
                  } else {
                    pageNum = page - 2 + i;
                  }
                  return (
                    <button
                      key={pageNum}
                      onClick={() => setPage(pageNum)}
                      className={`px-4 py-2 rounded-lg ${
                        page === pageNum
                          ? 'bg-purple-600 text-white'
                          : 'bg-[#1a1a2e] text-white hover:bg-[#252540]'
                      }`}
                    >
                      {pageNum}
                    </button>
                  );
                })}
                <button
                  onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                  disabled={page === totalPages}
                  className="px-4 py-2 bg-[#1a1a2e] text-white rounded-lg disabled:opacity-50"
                >
                  →
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}

// Компонент карточки новости
function NewsCard({
  post,
  locale,
  formatDate,
  getCategoryColor,
  pinned = false,
}: {
  post: NewsPost;
  locale: string;
  formatDate: (date: string) => string;
  getCategoryColor: (cat: string) => string;
  pinned?: boolean;
}) {
  const t = useTranslations('community');

  return (
    <Link href={`/${locale}/news/${post.slug}`}>
      <article className={`bg-[#1a1a2e] rounded-xl overflow-hidden hover:ring-2 hover:ring-purple-500 transition-all ${pinned ? 'ring-1 ring-red-500/50' : ''}`}>
        <div className="flex flex-col md:flex-row">
          {/* Изображение */}
          {post.coverUrl && (
            <div className="relative w-full md:w-48 h-40 md:h-auto flex-shrink-0">
              <Image
                src={post.coverUrl}
                alt={post.title}
                fill
                sizes="(max-width: 768px) 100vw, 192px"
                className="object-cover"
              />
            </div>
          )}

          {/* Контент */}
          <div className="flex-1 p-5">
            <div className="flex items-center gap-2 mb-2">
              <span className={`px-2 py-0.5 text-xs font-medium text-white rounded ${getCategoryColor(post.category)}`}>
                {t(`category.${post.category}`)}
              </span>
              {pinned && (
                <span className="px-2 py-0.5 text-xs font-medium text-red-400 bg-red-900/50 rounded">
                  📌 {t('pinned')}
                </span>
              )}
            </div>

            <h2 className="text-xl font-bold text-white mb-2 line-clamp-2 group-hover:text-purple-400 transition-colors">
              {post.title}
            </h2>

            {post.summary && (
              <p className="text-gray-400 mb-4 line-clamp-2">
                {post.summary}
              </p>
            )}

            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                {post.author.avatarUrl ? (
                  <Image
                    src={post.author.avatarUrl}
                    alt={post.author.displayName}
                    width={24}
                    height={24}
                    className="rounded-full"
                  />
                ) : (
                  <div className="w-6 h-6 rounded-full bg-purple-600 flex items-center justify-center text-white text-xs font-bold">
                    {post.author.displayName[0]}
                  </div>
                )}
                <span className="text-sm text-gray-400">
                  {post.author.displayName}
                </span>
              </div>

              <span className="text-sm text-gray-500">
                {formatDate(post.publishedAt || post.createdAt)}
              </span>
            </div>
          </div>
        </div>
      </article>
    </Link>
  );
}
