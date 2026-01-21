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
  updatedAt: string;
}

interface NewsDetailClientProps {
  locale: string;
  slug: string;
}

export default function NewsDetailClient({ locale, slug }: NewsDetailClientProps) {
  const t = useTranslations('community');
  const [news, setNews] = useState<NewsPost | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadNews();
  }, [slug, locale]);

  const loadNews = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetch(`/api/v1/news/${slug}?lang=${locale}`);
      if (!response.ok) {
        if (response.status === 404) {
          setError(t('newsNotFound'));
        } else {
          throw new Error('Failed to load');
        }
        return;
      }
      const data = await response.json();
      setNews(data);
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
      hour: '2-digit',
      minute: '2-digit',
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

  if (loading) {
    return (
      <div className="min-h-screen bg-[#121212] flex items-center justify-center">
        <div className="animate-spin rounded-full h-12 w-12 border-4 border-purple-500 border-t-transparent" />
      </div>
    );
  }

  if (error || !news) {
    return (
      <div className="min-h-screen bg-[#121212] flex flex-col items-center justify-center">
        <p className="text-red-400 text-xl mb-4">{error || t('newsNotFound')}</p>
        <Link
          href={`/${locale}/news`}
          className="px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg"
        >
          {t('backToNews')}
        </Link>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#121212]">
      {/* Обложка */}
      {news.coverUrl && (
        <div className="relative h-80 md:h-96 overflow-hidden">
          <Image
            src={news.coverUrl}
            alt={news.title}
            fill
            sizes="100vw"
            className="object-cover"
          />
          <div className="absolute inset-0 bg-gradient-to-b from-transparent via-[#121212]/50 to-[#121212]" />
        </div>
      )}

      <div className="max-w-4xl mx-auto px-4 py-8">
        {/* Хлебные крошки */}
        <nav className="mb-6">
          <Link
            href={`/${locale}/news`}
            className="text-purple-400 hover:text-purple-300"
          >
            ← {t('backToNews')}
          </Link>
        </nav>

        {/* Заголовок и метаданные */}
        <header className="mb-8">
          <div className="flex items-center gap-2 mb-4">
            <span className={`px-3 py-1 text-sm font-medium text-white rounded ${getCategoryColor(news.category)}`}>
              {t(`category.${news.category}`)}
            </span>
            {news.isPinned && (
              <span className="px-3 py-1 text-sm font-medium text-red-400 bg-red-900/50 rounded">
                📌 {t('pinned')}
              </span>
            )}
          </div>

          <h1 className="text-4xl md:text-5xl font-bold text-white mb-6">
            {news.title}
          </h1>

          <div className="flex items-center justify-between flex-wrap gap-4">
            <Link
              href={`/${locale}/users/${news.author.id}`}
              className="flex items-center gap-3 hover:opacity-80"
            >
              {news.author.avatarUrl ? (
                <Image
                  src={news.author.avatarUrl}
                  alt={news.author.displayName}
                  width={40}
                  height={40}
                  className="rounded-full"
                />
              ) : (
                <div className="w-10 h-10 rounded-full bg-purple-600 flex items-center justify-center text-white font-bold">
                  {news.author.displayName[0]}
                </div>
              )}
              <span className="text-gray-300 font-medium">
                {news.author.displayName}
              </span>
            </Link>

            <div className="text-gray-500 text-sm">
              {formatDate(news.publishedAt || news.createdAt)}
            </div>
          </div>
        </header>

        {/* Краткое содержание */}
        {news.summary && (
          <div className="mb-8 p-4 bg-[#1a1a2e] rounded-xl border-l-4 border-purple-500">
            <p className="text-lg text-gray-300 italic">{news.summary}</p>
          </div>
        )}

        {/* Контент */}
        <article className="prose prose-invert prose-lg max-w-none">
          <div 
            className="text-gray-300 leading-relaxed space-y-4"
            dangerouslySetInnerHTML={{ __html: formatContent(news.content) }}
          />
        </article>

        {/* Дата обновления */}
        {news.updatedAt && news.updatedAt !== news.createdAt && (
          <div className="mt-8 pt-4 border-t border-gray-800 text-sm text-gray-500">
            {t('lastUpdated')}: {formatDate(news.updatedAt)}
          </div>
        )}

        {/* Навигация */}
        <div className="mt-12 pt-8 border-t border-gray-800">
          <Link
            href={`/${locale}/news`}
            className="inline-flex items-center gap-2 px-6 py-3 bg-[#1a1a2e] hover:bg-[#252540] text-white rounded-lg transition-colors"
          >
            ← {t('moreNews')}
          </Link>
        </div>
      </div>
    </div>
  );
}

// Форматирование контента (простая разметка)
function formatContent(content: string): string {
  if (!content) return '';

  // Преобразуем переносы строк в параграфы
  const paragraphs = content.split(/\n\n+/);
  
  return paragraphs
    .map(p => {
      // Обрабатываем заголовки
      if (p.startsWith('# ')) {
        return `<h2 class="text-2xl font-bold text-white mt-8 mb-4">${p.slice(2)}</h2>`;
      }
      if (p.startsWith('## ')) {
        return `<h3 class="text-xl font-bold text-white mt-6 mb-3">${p.slice(3)}</h3>`;
      }
      if (p.startsWith('### ')) {
        return `<h4 class="text-lg font-bold text-white mt-4 mb-2">${p.slice(4)}</h4>`;
      }

      // Обрабатываем списки
      if (p.startsWith('- ') || p.startsWith('* ')) {
        const items = p.split('\n').map(item => 
          item.replace(/^[-*]\s+/, '')
        );
        return `<ul class="list-disc list-inside space-y-1">${items.map(i => `<li>${i}</li>`).join('')}</ul>`;
      }

      // Обрабатываем цитаты
      if (p.startsWith('> ')) {
        return `<blockquote class="border-l-4 border-purple-500 pl-4 italic text-gray-400">${p.slice(2)}</blockquote>`;
      }

      // Обычный параграф
      return `<p class="mb-4">${p.replace(/\n/g, '<br />')}</p>`;
    })
    .join('');
}
