'use client';

import { useState, useEffect } from 'react';
import { useTranslations } from 'next-intl';
import Link from 'next/link';
import Image from 'next/image';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import { useAuthStore } from '@/store/auth';

function nextImageSrcFromApi(value: string): string | null {
  const src = value.trim();
  if (!src) return null;
  if (src.startsWith('/') || src.startsWith('http://') || src.startsWith('https://')) return src;
  return `/${src}`;
}

type ApiEnvelope<T> = { data: T };

interface Collection {
  id: string;
  slug: string;
  title: string;
  description: string;
  coverUrl: string;
  isPublic: boolean;
  isFeatured: boolean;
  votesCount: number;
  itemsCount: number;
  hasVoted: boolean;
  user: {
    id: string;
    displayName: string;
    avatarUrl: string;
  };
  items: CollectionItem[];
  createdAt: string;
  updatedAt: string;
}

interface ApiCollection {
  id: string;
  slug: string;
  title: string;
  description?: string;
  coverUrl?: string;
  isPublic: boolean;
  isFeatured: boolean;
  votesCount: number;
  itemsCount: number;
  userVote?: number;
  user?: {
    id: string;
    displayName: string;
    avatarUrl?: string | null;
  };
  items?: CollectionItem[];
  createdAt: string;
  updatedAt: string;
}

interface CollectionItem {
  novelId: string;
  position: number;
  note: string;
  novelSlug?: string;
  novelTitle?: string;
  novelCoverUrl?: string | null;
  novelRating?: number;
  novelDescription?: string | null;
  novelTranslationStatus?: string;
  novel?: {
    id: string;
    slug: string;
    title: string;
    coverUrl: string;
    rating: number;
    description?: string;
    translationStatus?: string;
  };
  addedAt: string;
}

interface CollectionDetailClientProps {
  locale: string;
  slug: string;
}

export default function CollectionDetailClient({ locale, slug }: CollectionDetailClientProps) {
  const t = useTranslations('community');
  const router = useRouter();
  const { user, isAuthenticated } = useAuth();
  const accessToken = useAuthStore((s) => s.accessToken);
  const [collection, setCollection] = useState<Collection | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isOwner, setIsOwner] = useState(false);

  useEffect(() => {
    loadCollection();
  }, [slug]);

  useEffect(() => {
    if (collection && user) {
      setIsOwner(collection.user.id === user.id);
    }
  }, [collection, user]);

  const loadCollection = async () => {
    setLoading(true);
    setError(null);
    try {
      const headers: Record<string, string> = {};
      if (accessToken) headers.Authorization = `Bearer ${accessToken}`;

      const response = await fetch(`/api/v1/collections/${slug}`, {
        headers,
        credentials: 'include',
      });
      if (!response.ok) {
        if (response.status === 404) {
          setError(t('collectionNotFound'));
        } else {
          throw new Error('Failed to load collection');
        }
        return;
      }
      const result: ApiEnvelope<ApiCollection> = await response.json();
      const apiCol = result.data;

      setCollection({
        id: apiCol.id,
        slug: apiCol.slug,
        title: apiCol.title,
        description: apiCol.description || '',
        coverUrl: apiCol.coverUrl || '',
        isPublic: apiCol.isPublic,
        isFeatured: apiCol.isFeatured,
        votesCount: apiCol.votesCount,
        itemsCount: apiCol.itemsCount,
        hasVoted: apiCol.userVote === 1,
        user: {
          id: apiCol.user?.id || '',
          displayName: apiCol.user?.displayName || '',
          avatarUrl: apiCol.user?.avatarUrl || '',
        },
        items: apiCol.items || [],
        createdAt: apiCol.createdAt,
        updatedAt: apiCol.updatedAt,
      });
    } catch (err) {
      setError(t('loadError'));
    } finally {
      setLoading(false);
    }
  };

  const handleVote = async () => {
    if (!isAuthenticated || !collection) {
      alert(t('loginRequired'));
      return;
    }
    if (isOwner) {
      alert('Нельзя голосовать за свою коллекцию');
      return;
    }

    try {
      const headers: Record<string, string> = {};
      if (accessToken) headers.Authorization = `Bearer ${accessToken}`;
      const response = await fetch(`/api/v1/collections/${collection.id}/vote`, {
        method: 'POST',
        headers,
        credentials: 'include',
      });

      if (!response.ok) {
        const errData: any = await response.json().catch(() => ({}));
        const msg = errData?.error?.message || errData?.message || 'Failed to vote';
        throw new Error(msg);
      }

      setCollection({
        ...collection,
        hasVoted: !collection.hasVoted,
        votesCount: collection.hasVoted ? collection.votesCount - 1 : collection.votesCount + 1,
      });
    } catch (err) {
      console.error('Vote error:', err);
      alert(err instanceof Error ? err.message : 'Не удалось проголосовать');
    }
  };

  const handleDelete = async () => {
    if (!collection || !confirm(t('confirmDeleteCollection'))) return;

    try {
      const headers: Record<string, string> = {};
      if (accessToken) headers.Authorization = `Bearer ${accessToken}`;
      const response = await fetch(`/api/v1/collections/${collection.id}`, {
        method: 'DELETE',
        headers,
        credentials: 'include',
      });

      if (response.ok) {
        router.push(`/${locale}/collections`);
      }
    } catch (err) {
      console.error('Delete error:', err);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[#121212] flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-4 border-purple-500 border-t-transparent" />
      </div>
    );
  }

  if (error || !collection) {
    return (
      <div className="min-h-screen bg-[#121212] flex flex-col items-center justify-center">
        <p className="text-red-400 text-xl mb-4">{error || t('collectionNotFound')}</p>
        <Link
          href={`/${locale}/collections`}
          className="px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg"
        >
          {t('backToCollections')}
        </Link>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#121212]">
      {/* Шапка коллекции */}
      <div className="relative">
        {/* Фон */}
        <div className="absolute inset-0 h-80 overflow-hidden">
          {collection.coverUrl && nextImageSrcFromApi(collection.coverUrl) ? (
            <Image
              src={nextImageSrcFromApi(collection.coverUrl)!}
              alt=""
              fill
              className="object-cover blur-xl opacity-30"
            />
          ) : (
            <div className="w-full h-full bg-gradient-to-br from-purple-900/50 to-blue-900/50" />
          )}
          <div className="absolute inset-0 bg-gradient-to-b from-transparent via-[#121212]/70 to-[#121212]" />
        </div>

        <div className="relative max-w-7xl mx-auto px-4 pt-12 pb-8">
          <div className="flex flex-col md:flex-row gap-8">
            {/* Обложка */}
            <div className="w-48 h-64 flex-shrink-0 rounded-xl overflow-hidden shadow-2xl">
              {collection.coverUrl && nextImageSrcFromApi(collection.coverUrl) ? (
                <Image
                  src={nextImageSrcFromApi(collection.coverUrl)!}
                  alt={collection.title}
                  width={192}
                  height={256}
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="w-full h-full bg-gradient-to-br from-purple-700 to-blue-600 flex items-center justify-center">
                  <span className="text-6xl">📚</span>
                </div>
              )}
            </div>

            {/* Информация */}
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-2">
                {collection.isFeatured && (
                  <span className="px-2 py-1 bg-yellow-500 text-black text-xs font-bold rounded">
                    ⭐ {t('featured')}
                  </span>
                )}
                {!collection.isPublic && (
                  <span className="px-2 py-1 bg-gray-700 text-gray-300 text-xs rounded">
                    🔒 {t('private')}
                  </span>
                )}
              </div>

              <h1 className="text-4xl font-bold text-white mb-3">
                {collection.title}
              </h1>

              {collection.description && (
                <p className="text-gray-300 text-lg mb-4 max-w-2xl">
                  {collection.description}
                </p>
              )}

              {/* Автор */}
              <Link
                href={`/${locale}/users/${collection.user.id}`}
                className="inline-flex items-center gap-2 mb-6 hover:opacity-80"
              >
                {collection.user.avatarUrl && nextImageSrcFromApi(collection.user.avatarUrl) ? (
                  <Image
                    src={nextImageSrcFromApi(collection.user.avatarUrl)!}
                    alt={collection.user.displayName}
                    width={32}
                    height={32}
                    className="rounded-full"
                  />
                ) : (
                  <div className="w-8 h-8 rounded-full bg-purple-600 flex items-center justify-center text-white font-bold">
                    {(collection.user.displayName || '?')[0]}
                  </div>
                )}
                <span className="text-gray-300">{collection.user.displayName}</span>
              </Link>

              {/* Статистика и действия */}
              <div className="flex flex-wrap items-center gap-4">
                <div className="text-gray-400">
                  <span className="font-bold text-white text-lg">{collection.itemsCount}</span>{' '}
                  {t('novelsCount')}
                </div>
                <div className="text-gray-400">
                  <span className="font-bold text-white text-lg">{collection.votesCount}</span>{' '}
                  {t('votesCount')}
                </div>

                <button
                  onClick={handleVote}
                  className={`flex items-center gap-2 px-6 py-2 rounded-full font-medium transition-colors ${
                    collection.hasVoted
                      ? 'bg-purple-600 text-white'
                      : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                  }`}
                >
                  <span>❤️</span>
                  {collection.hasVoted ? t('voted') : t('vote')}
                </button>

                {isOwner && (
                  <>
                    <button
                      onClick={handleDelete}
                      className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg"
                    >
                      {t('delete')}
                    </button>
                  </>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Список новелл */}
      <div className="max-w-7xl mx-auto px-4 py-8">
        <h2 className="text-2xl font-bold text-white mb-6">
          {t('novelsInCollection')} ({collection.itemsCount})
        </h2>

        {collection.items.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-400">{t('emptyCollection')}</p>
          </div>
        ) : (
          <div className="space-y-4">
            {collection.items.map((item, index) => {
              const novelSlug = item.novel?.slug || item.novelSlug || '';
              const novelHref = novelSlug ? `/${locale}/novel/${novelSlug}` : null;
              const novelTitle = item.novel?.title || item.novelTitle || item.novelId;
              const novelDescription = item.novel?.description ?? item.novelDescription ?? '';
              const novelCoverUrl = item.novel?.coverUrl || item.novelCoverUrl || '';
              const novelCoverSrc = novelCoverUrl ? nextImageSrcFromApi(novelCoverUrl) : null;
              const novelRating = item.novel?.rating ?? item.novelRating ?? 0;
              const novelTranslationStatus = item.novel?.translationStatus || item.novelTranslationStatus || '';

              const Cover = (
                <div className="w-20 h-28 rounded-lg overflow-hidden">
                  {novelCoverSrc ? (
                    <Image
                      src={novelCoverSrc}
                      alt={novelTitle}
                      width={80}
                      height={112}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full bg-gray-700 flex items-center justify-center">
                      <span className="text-2xl">📖</span>
                    </div>
                  )}
                </div>
              );

              return (
                <div
                  key={item.novelId}
                  className="flex gap-4 bg-[#1a1a2e] rounded-xl p-4 hover:bg-[#252540] transition-colors"
                >
                  {/* Номер позиции */}
                  <div className="w-8 h-8 flex-shrink-0 flex items-center justify-center bg-gray-700 rounded-full text-gray-300 font-bold">
                    {index + 1}
                  </div>

                  {/* Обложка новеллы */}
                  {novelHref ? (
                    <Link href={novelHref} className="flex-shrink-0">
                      {Cover}
                    </Link>
                  ) : (
                    <div className="flex-shrink-0">{Cover}</div>
                  )}

                  {/* Информация о новелле */}
                  <div className="flex-1 min-w-0">
                    {novelHref ? (
                      <Link
                        href={novelHref}
                        className="text-lg font-bold text-white hover:text-purple-400 transition-colors line-clamp-1"
                      >
                        {novelTitle}
                      </Link>
                    ) : (
                      <div className="text-lg font-bold text-white line-clamp-1">{novelTitle}</div>
                    )}

                    {novelDescription && (
                      <p className="text-gray-400 text-sm mt-1 line-clamp-2">{novelDescription}</p>
                    )}

                    {item.note && (
                      <div className="mt-2 p-2 bg-[#121212] rounded-lg">
                        <p className="text-sm text-gray-300 italic">"{item.note}"</p>
                      </div>
                    )}

                    <div className="flex items-center gap-4 mt-2">
                      {novelRating > 0 && (
                        <span className="text-yellow-500 text-sm">⭐ {Number(novelRating).toFixed(1)}</span>
                      )}
                      {novelTranslationStatus && (
                        <span
                          className={`text-xs px-2 py-0.5 rounded ${
                            novelTranslationStatus === 'completed'
                              ? 'bg-green-900/50 text-green-400'
                              : novelTranslationStatus === 'ongoing'
                                ? 'bg-blue-900/50 text-blue-400'
                                : 'bg-gray-700 text-gray-400'
                          }`}
                        >
                          {t(`status.${novelTranslationStatus}`)}
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Кнопка читать */}
                  {novelHref ? (
                    <Link
                      href={novelHref}
                      className="flex-shrink-0 self-center px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg transition-colors"
                    >
                      {t('read')}
                    </Link>
                  ) : null}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
