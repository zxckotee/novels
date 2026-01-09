'use client';

import { useState, useEffect } from 'react';
import { useTranslations } from 'next-intl';
import Link from 'next/link';
import Image from 'next/image';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';

interface Collection {
  id: string;
  slug: string;
  title: string;
  description: string;
  coverUrl: string;
  isPublic: boolean;
  items: CollectionItem[];
}

interface CollectionItem {
  novelId: string;
  position: number;
  note: string;
  novel: {
    id: string;
    slug: string;
    title: string;
    coverUrl: string;
  };
}

interface Novel {
  id: string;
  slug: string;
  title: string;
  coverUrl: string;
}

interface CollectionFormClientProps {
  locale: string;
  slug?: string;
}

export default function CollectionFormClient({ locale, slug }: CollectionFormClientProps) {
  const t = useTranslations('community');
  const router = useRouter();
  const { isAuthenticated } = useAuth();

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [coverUrl, setCoverUrl] = useState('');
  const [isPublic, setIsPublic] = useState(true);
  const [items, setItems] = useState<{ novelId: string; note: string; novel?: Novel }[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<Novel[]>([]);
  const [searching, setSearching] = useState(false);

  const isEditing = !!slug;

  useEffect(() => {
    if (!isAuthenticated) {
      router.push(`/${locale}/auth/login`);
    }
  }, [isAuthenticated, locale, router]);

  useEffect(() => {
    if (isEditing) {
      loadCollection();
    }
  }, [slug]);

  const loadCollection = async () => {
    setLoading(true);
    try {
      const response = await fetch(`/api/v1/collections/${slug}`);
      if (!response.ok) throw new Error('Failed to load');
      
      const data: Collection = await response.json();
      setTitle(data.title);
      setDescription(data.description || '');
      setCoverUrl(data.coverUrl || '');
      setIsPublic(data.isPublic);
      setItems(data.items.map(item => ({
        novelId: item.novelId,
        note: item.note || '',
        novel: item.novel,
      })));
    } catch (err) {
      setError(t('loadError'));
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = async () => {
    if (!searchQuery.trim()) return;

    setSearching(true);
    try {
      const params = new URLSearchParams({
        q: searchQuery,
        limit: '10',
      });
      const response = await fetch(`/api/v1/novels?${params}`);
      if (response.ok) {
        const data = await response.json();
        setSearchResults(data.novels || []);
      }
    } catch (err) {
      console.error('Search error:', err);
    } finally {
      setSearching(false);
    }
  };

  const addNovel = (novel: Novel) => {
    if (items.find(item => item.novelId === novel.id)) {
      alert(t('novelAlreadyAdded'));
      return;
    }

    setItems([...items, { novelId: novel.id, note: '', novel }]);
    setSearchQuery('');
    setSearchResults([]);
  };

  const removeNovel = (novelId: string) => {
    setItems(items.filter(item => item.novelId !== novelId));
  };

  const updateNote = (novelId: string, note: string) => {
    setItems(items.map(item =>
      item.novelId === novelId ? { ...item, note } : item
    ));
  };

  const moveItem = (index: number, direction: 'up' | 'down') => {
    const newIndex = direction === 'up' ? index - 1 : index + 1;
    if (newIndex < 0 || newIndex >= items.length) return;

    const newItems = [...items];
    [newItems[index], newItems[newIndex]] = [newItems[newIndex], newItems[index]];
    setItems(newItems);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) {
      setError(t('titleRequired'));
      return;
    }

    setSaving(true);
    setError(null);

    try {
      const body = {
        title: title.trim(),
        description: description.trim() || null,
        coverUrl: coverUrl.trim() || null,
        isPublic,
        items: items.map((item, index) => ({
          novelId: item.novelId,
          note: item.note || null,
          position: index,
        })),
      };

      const url = isEditing 
        ? `/api/v1/collections/${slug}`
        : '/api/v1/collections';

      const response = await fetch(url, {
        method: isEditing ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to save');
      }

      const data = await response.json();
      router.push(`/${locale}/collections/${data.slug}`);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError(t('saveError'));
      }
    } finally {
      setSaving(false);
    }
  };

  if (!isAuthenticated) {
    return null;
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-[#121212] flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-4 border-purple-500 border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#121212]">
      <div className="max-w-4xl mx-auto px-4 py-8">
        {/* Заголовок */}
        <div className="mb-8">
          <Link
            href={`/${locale}/collections`}
            className="text-purple-400 hover:text-purple-300 mb-4 inline-flex items-center gap-2"
          >
            ← {t('backToCollections')}
          </Link>
          <h1 className="text-3xl font-bold text-white">
            {isEditing ? t('editCollection') : t('createCollection')}
          </h1>
        </div>

        {error && (
          <div className="mb-6 p-4 bg-red-900/50 border border-red-500 rounded-lg text-red-200">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Основная информация */}
          <div className="bg-[#1a1a2e] rounded-xl p-6 space-y-4">
            <h2 className="text-xl font-bold text-white mb-4">{t('basicInfo')}</h2>

            <div>
              <label className="block text-gray-300 mb-2">{t('collectionTitle')} *</label>
              <input
                type="text"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                required
                maxLength={100}
                className="w-full px-4 py-3 bg-[#121212] border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-purple-500"
                placeholder={t('enterTitle')}
              />
            </div>

            <div>
              <label className="block text-gray-300 mb-2">{t('description')}</label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={4}
                maxLength={1000}
                className="w-full px-4 py-3 bg-[#121212] border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-purple-500 resize-none"
                placeholder={t('enterDescription')}
              />
            </div>

            <div>
              <label className="block text-gray-300 mb-2">{t('coverUrl')}</label>
              <input
                type="url"
                value={coverUrl}
                onChange={(e) => setCoverUrl(e.target.value)}
                className="w-full px-4 py-3 bg-[#121212] border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-purple-500"
                placeholder="https://example.com/cover.jpg"
              />
              {coverUrl && (
                <div className="mt-2">
                  <Image
                    src={coverUrl}
                    alt="Preview"
                    width={100}
                    height={140}
                    className="rounded-lg object-cover"
                    onError={() => setCoverUrl('')}
                  />
                </div>
              )}
            </div>

            <div className="flex items-center gap-3">
              <input
                type="checkbox"
                id="isPublic"
                checked={isPublic}
                onChange={(e) => setIsPublic(e.target.checked)}
                className="w-5 h-5 rounded bg-[#121212] border-gray-700 text-purple-600 focus:ring-purple-500"
              />
              <label htmlFor="isPublic" className="text-gray-300">
                {t('publicCollection')}
              </label>
            </div>
          </div>

          {/* Добавление новелл */}
          <div className="bg-[#1a1a2e] rounded-xl p-6">
            <h2 className="text-xl font-bold text-white mb-4">{t('addNovels')}</h2>

            {/* Поиск */}
            <div className="flex gap-2 mb-4">
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), handleSearch())}
                className="flex-1 px-4 py-3 bg-[#121212] border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:border-purple-500"
                placeholder={t('searchNovels')}
              />
              <button
                type="button"
                onClick={handleSearch}
                disabled={searching}
                className="px-6 py-3 bg-purple-600 hover:bg-purple-700 text-white rounded-lg transition-colors disabled:opacity-50"
              >
                {searching ? '...' : t('search')}
              </button>
            </div>

            {/* Результаты поиска */}
            {searchResults.length > 0 && (
              <div className="mb-6 space-y-2 max-h-60 overflow-y-auto">
                {searchResults.map((novel) => (
                  <div
                    key={novel.id}
                    className="flex items-center justify-between p-3 bg-[#121212] rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      {novel.coverUrl ? (
                        <Image
                          src={novel.coverUrl}
                          alt={novel.title}
                          width={40}
                          height={56}
                          className="rounded object-cover"
                        />
                      ) : (
                        <div className="w-10 h-14 bg-gray-700 rounded flex items-center justify-center">
                          📖
                        </div>
                      )}
                      <span className="text-white">{novel.title}</span>
                    </div>
                    <button
                      type="button"
                      onClick={() => addNovel(novel)}
                      className="px-3 py-1 bg-green-600 hover:bg-green-700 text-white rounded transition-colors"
                    >
                      +
                    </button>
                  </div>
                ))}
              </div>
            )}

            {/* Список добавленных новелл */}
            <div className="space-y-3">
              <h3 className="text-lg font-medium text-gray-300">
                {t('addedNovels')} ({items.length})
              </h3>

              {items.length === 0 ? (
                <p className="text-gray-500 text-center py-8">{t('noNovelsAdded')}</p>
              ) : (
                items.map((item, index) => (
                  <div
                    key={item.novelId}
                    className="flex flex-col gap-3 p-4 bg-[#121212] rounded-lg"
                  >
                    <div className="flex items-center gap-3">
                      {/* Порядок */}
                      <div className="flex flex-col gap-1">
                        <button
                          type="button"
                          onClick={() => moveItem(index, 'up')}
                          disabled={index === 0}
                          className="p-1 text-gray-400 hover:text-white disabled:opacity-30"
                        >
                          ▲
                        </button>
                        <button
                          type="button"
                          onClick={() => moveItem(index, 'down')}
                          disabled={index === items.length - 1}
                          className="p-1 text-gray-400 hover:text-white disabled:opacity-30"
                        >
                          ▼
                        </button>
                      </div>

                      {/* Номер */}
                      <span className="w-8 h-8 flex items-center justify-center bg-gray-700 rounded-full text-white font-bold">
                        {index + 1}
                      </span>

                      {/* Обложка */}
                      {item.novel?.coverUrl ? (
                        <Image
                          src={item.novel.coverUrl}
                          alt={item.novel.title}
                          width={40}
                          height={56}
                          className="rounded object-cover"
                        />
                      ) : (
                        <div className="w-10 h-14 bg-gray-700 rounded flex items-center justify-center">
                          📖
                        </div>
                      )}

                      {/* Название */}
                      <span className="flex-1 text-white font-medium">
                        {item.novel?.title || item.novelId}
                      </span>

                      {/* Удалить */}
                      <button
                        type="button"
                        onClick={() => removeNovel(item.novelId)}
                        className="p-2 text-red-400 hover:text-red-300 hover:bg-red-900/30 rounded transition-colors"
                      >
                        ✕
                      </button>
                    </div>

                    {/* Заметка */}
                    <input
                      type="text"
                      value={item.note}
                      onChange={(e) => updateNote(item.novelId, e.target.value)}
                      maxLength={200}
                      className="w-full px-3 py-2 bg-[#1a1a2e] border border-gray-700 rounded text-gray-300 text-sm placeholder-gray-500 focus:outline-none focus:border-purple-500"
                      placeholder={t('addNote')}
                    />
                  </div>
                ))
              )}
            </div>
          </div>

          {/* Кнопки */}
          <div className="flex justify-end gap-4">
            <Link
              href={`/${locale}/collections`}
              className="px-6 py-3 bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-colors"
            >
              {t('cancel')}
            </Link>
            <button
              type="submit"
              disabled={saving || !title.trim()}
              className="px-8 py-3 bg-purple-600 hover:bg-purple-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {saving ? t('saving') : isEditing ? t('saveChanges') : t('createCollection')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
