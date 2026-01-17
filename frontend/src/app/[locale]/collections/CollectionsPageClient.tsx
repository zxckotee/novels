'use client';

import { useState, useEffect } from 'react';
import { useTranslations } from 'next-intl';
import Link from 'next/link';
import Image from 'next/image';
import { useAuth } from '@/contexts/AuthContext';
import { useAuthStore } from '@/store/auth';

function nextImageSrcFromApi(value: string): string | null {
  const src = value.trim();
  if (!src) return null;
  if (src.startsWith('/') || src.startsWith('http://') || src.startsWith('https://')) return src;
  return `/${src}`;
}

interface Collection {
  id: string;
  slug: string;
  title: string;
  description?: string;
  coverUrl?: string;
  votesCount: number;
  itemsCount: number;
  hasVoted?: boolean;
  user?: {
    id: string;
    displayName: string;
    avatarUrl: string;
  };
  previewCovers?: string[];
  createdAt?: string;
  updatedAt?: string;
}

interface CollectionListResponse {
  collections: Collection[];
  total: number;
  page: number;
  limit: number;
  totalPages?: number;
}

type ApiEnvelope<T> = { data: T };

interface CollectionsPageClientProps {
  locale: string;
  searchParams: { [key: string]: string | string[] | undefined };
}

export default function CollectionsPageClient({ locale, searchParams }: CollectionsPageClientProps) {
  const t = useTranslations('community');
  const { user, isAuthenticated } = useAuth();
  const accessToken = useAuthStore((s) => s.accessToken);
  const [collections, setCollections] = useState<Collection[]>([]);
  const [featuredCollections, setFeaturedCollections] = useState<Collection[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [sort, setSort] = useState<string>((searchParams.sort as string) || 'popular');

  useEffect(() => {
    loadFeaturedCollections();
  }, []);

  useEffect(() => {
    loadCollections();
  }, [page, sort]);

  const loadFeaturedCollections = async () => {
    try {
      const response = await fetch('/api/v1/collections/featured?limit=6');
      if (response.ok) {
        const result: ApiEnvelope<{ collections: Collection[] }> = await response.json();
        setFeaturedCollections(result.data.collections || []);
      }
    } catch (err) {
      console.error('Failed to load featured collections:', err);
    }
  };

  const loadCollections = async () => {
    setLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams({
        page: String(page),
        limit: '20',
        sort,
      });

      const response = await fetch(`/api/v1/collections?${params}`);
      if (!response.ok) throw new Error('Failed to load collections');

      const result: ApiEnvelope<CollectionListResponse> = await response.json();
      setCollections(result.data.collections || []);
      setTotal(result.data.total);
    } catch (err) {
      setError(t('loadError'));
    } finally {
      setLoading(false);
    }
  };

  const handleVote = async (collectionId: string) => {
    if (!isAuthenticated) {
      alert(t('loginRequired'));
      return;
    }

    try {
      const headers: Record<string, string> = {};
      if (accessToken) headers.Authorization = `Bearer ${accessToken}`;
      const response = await fetch(`/api/v1/collections/${collectionId}/vote`, {
        method: 'POST',
        headers,
        credentials: 'include',
      });

      if (response.ok) {
        // Обновляем локальное состояние
        setCollections(collections.map(c => {
          if (c.id === collectionId) {
            const currentlyVoted = !!c.hasVoted;
            return {
              ...c,
              hasVoted: !currentlyVoted,
              votesCount: currentlyVoted ? Math.max(0, c.votesCount - 1) : c.votesCount + 1,
            };
          }
          return c;
        }));
      }
    } catch (err) {
      console.error('Vote error:', err);
    }
  };

  const totalPages = Math.ceil(total / 20);

  return (
    <div className="min-h-screen bg-[#121212]">
      <div className="max-w-7xl mx-auto px-4 py-8">
        {/* Шапка */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-white mb-2">{t('collections')}</h1>
            <p className="text-gray-400">{t('collectionsDescription')}</p>
          </div>
          {isAuthenticated && (
            <Link
              href={`/${locale}/collections/new`}
              className="px-6 py-3 bg-purple-600 hover:bg-purple-700 text-white font-medium rounded-lg transition-colors"
            >
              {t('createCollection')}
            </Link>
          )}
        </div>

        {/* Рекомендуемые коллекции */}
        {featuredCollections.length > 0 && (
          <section className="mb-12">
            <h2 className="text-xl font-bold text-white mb-4 flex items-center gap-2">
              <span className="text-yellow-500">⭐</span>
              {t('featuredCollections')}
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {featuredCollections.map((collection) => (
                <CollectionCard
                  key={collection.id}
                  collection={collection}
                  locale={locale}
                  onVote={handleVote}
                  featured
                />
              ))}
            </div>
          </section>
        )}

        {/* Фильтры и сортировка */}
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-xl font-bold text-white">{t('allCollections')}</h2>
          <div className="flex items-center gap-4">
            <select
              value={sort}
              onChange={(e) => {
                setSort(e.target.value);
                setPage(1);
              }}
              className="px-4 py-2 bg-[#1a1a2e] border border-gray-700 rounded-lg text-white focus:outline-none focus:border-purple-500"
            >
              <option value="popular">{t('sortByPopular')}</option>
              <option value="recent">{t('sortByRecent')}</option>
              <option value="votes">{t('sortByVotes')}</option>
              <option value="items">{t('sortByItems')}</option>
            </select>
          </div>
        </div>

        {/* Список коллекций */}
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-12 w-12 border-4 border-purple-500 border-t-transparent" />
          </div>
        ) : error ? (
          <div className="text-center py-12">
            <p className="text-red-400">{error}</p>
            <button
              onClick={loadCollections}
              className="mt-4 px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg"
            >
              {t('retry')}
            </button>
          </div>
        ) : collections.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-400">{t('noCollections')}</p>
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
              {collections.map((collection) => (
                <CollectionCard
                  key={collection.id}
                  collection={collection}
                  locale={locale}
                  onVote={handleVote}
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

// Компонент карточки коллекции
function CollectionCard({
  collection,
  locale,
  onVote,
  featured = false,
}: {
  collection: Collection;
  locale: string;
  onVote: (id: string) => void;
  featured?: boolean;
}) {
  const t = useTranslations('community');
  const coverSrc = collection.coverUrl ? nextImageSrcFromApi(collection.coverUrl) : null;
  const previewCovers = (collection.previewCovers || []).map((c) => nextImageSrcFromApi(c)).filter(Boolean) as string[];
  const authorName = collection.user?.displayName || '—';
  const avatarSrc = collection.user?.avatarUrl ? nextImageSrcFromApi(collection.user.avatarUrl) : null;
  const hasVoted = !!collection.hasVoted;
  
  return (
    <Link href={`/${locale}/collections/${collection.id}`}>
      <div className={`bg-[#1a1a2e] rounded-xl overflow-hidden group hover:ring-2 hover:ring-purple-500 transition-all ${featured ? 'ring-1 ring-yellow-500/50' : ''}`}>
        {/* Обложка коллекции */}
        <div className="relative h-40 bg-gradient-to-br from-purple-900/50 to-blue-900/50">
          {coverSrc ? (
            <Image
              src={coverSrc}
              alt={collection.title}
              fill
              className="object-cover"
            />
          ) : previewCovers.length > 0 ? (
            <div className="absolute inset-0 flex">
              {previewCovers.slice(0, 3).map((src, idx) => (
                <div
                  key={src + idx}
                  className="relative flex-1"
                  style={{ zIndex: 3 - idx }}
                >
                  <Image
                    src={src}
                    alt=""
                    fill
                    className="object-cover opacity-60"
                  />
                </div>
              ))}
            </div>
          ) : null}
          <div className="absolute inset-0 bg-gradient-to-t from-[#1a1a2e] via-transparent to-transparent" />
          
          {/* Бейджи */}
          <div className="absolute top-2 right-2 flex gap-1">
            {featured && (
              <span className="px-2 py-1 bg-yellow-500/90 text-black text-xs font-bold rounded">
                ⭐ {t('featured')}
              </span>
            )}
          </div>

          {/* Счетчик */}
          <div className="absolute bottom-2 left-2">
            <span className="px-2 py-1 bg-black/60 text-white text-sm rounded">
              {collection.itemsCount} {t('novels')}
            </span>
          </div>
        </div>

        {/* Контент */}
        <div className="p-4">
          <h3 className="text-lg font-bold text-white mb-1 line-clamp-1 group-hover:text-purple-400 transition-colors">
            {collection.title}
          </h3>
          {collection.description && (
            <p className="text-gray-400 text-sm mb-3 line-clamp-2">
              {collection.description}
            </p>
          )}

          {/* Автор и голоса */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              {avatarSrc ? (
                <Image
                  src={avatarSrc}
                  alt={authorName}
                  width={24}
                  height={24}
                  className="rounded-full"
                />
              ) : (
                <div className="w-6 h-6 rounded-full bg-purple-600 flex items-center justify-center text-white text-xs font-bold">
                  {authorName[0]}
                </div>
              )}
              <span className="text-sm text-gray-400">
                {authorName}
              </span>
            </div>

            <button
              onClick={(e) => {
                e.preventDefault();
                onVote(collection.id);
              }}
              className={`flex items-center gap-1 px-3 py-1 rounded-full transition-colors ${
                hasVoted
                  ? 'bg-purple-600 text-white'
                  : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
              }`}
            >
              <span>❤️</span>
              <span className="text-sm font-medium">{collection.votesCount}</span>
            </button>
          </div>
        </div>
      </div>
    </Link>
  );
}
