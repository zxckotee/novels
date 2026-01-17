'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { ArrowLeft, TrendingUp, Star, Flame, Plus } from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import { useQuery } from '@tanstack/react-query';
import api from '@/lib/api/client';

export default function AdminPopularPage() {
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user, isLoading: authLoading } = useAuthStore();
  const [activeTab, setActiveTab] = useState<'trending' | 'top-rated' | 'popular'>('trending');
  
  const hasAccess = isAuthenticated && isAdmin(user);

  useEffect(() => {
    if (!authLoading && !hasAccess) router.replace(`/${locale}`);
  }, [authLoading, hasAccess, router, locale]);

  const { data: novelsData, isLoading } = useQuery({
    queryKey: ['novels', activeTab, locale],
    queryFn: async () => {
      const sort = activeTab === 'trending' ? 'views' : activeTab === 'top-rated' ? 'rating' : 'bookmarks';
      const { data } = await api.get(`/novels?sort=${sort}&limit=20&lang=${locale}`);
      return Array.isArray(data) ? data : [];
    },
  });

  if (authLoading) return null;
  if (!hasAccess) return null;
  
  return (
    <div className="container-custom py-6">
      <div className="flex items-center gap-4 mb-6">
        <Link href={`/${locale}/admin`} className="btn-ghost p-2">
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <h1 className="text-2xl font-heading font-bold">Популярный контент</h1>
      </div>

      <div className="flex gap-2 mb-6">
        <button
          onClick={() => setActiveTab('trending')}
          className={`px-4 py-2 rounded flex items-center gap-2 ${
            activeTab === 'trending' ? 'bg-accent-primary text-white' : 'bg-background-secondary'
          }`}
        >
          <Flame className="w-4 h-4" />
          Trending (по просмотрам)
        </button>
        <button
          onClick={() => setActiveTab('top-rated')}
          className={`px-4 py-2 rounded flex items-center gap-2 ${
            activeTab === 'top-rated' ? 'bg-accent-primary text-white' : 'bg-background-secondary'
          }`}
        >
          <Star className="w-4 h-4" />
          Топ рейтинг
        </button>
        <button
          onClick={() => setActiveTab('popular')}
          className={`px-4 py-2 rounded flex items-center gap-2 ${
            activeTab === 'popular' ? 'bg-accent-primary text-white' : 'bg-background-secondary'
          }`}
        >
          <TrendingUp className="w-4 h-4" />
          По закладкам
        </button>
      </div>
      
      <div className="bg-background-secondary rounded-card p-6">
        {isLoading ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Загрузка...</p></div>
        ) : !novelsData || novelsData.length === 0 ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Новеллы не найдены</p></div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-background-tertiary">
                  <th className="text-left py-3 px-4">#</th>
                  <th className="text-left py-3 px-4">Название</th>
                  <th className="text-left py-3 px-4">Рейтинг</th>
                  <th className="text-left py-3 px-4">Просмотры</th>
                  <th className="text-left py-3 px-4">Закладки</th>
                  <th className="text-right py-3 px-4">Действия</th>
                </tr>
              </thead>
              <tbody>
                {novelsData.map((novel: any, idx: number) => (
                  <tr key={novel.id} className="border-b border-background-tertiary hover:bg-background-hover">
                    <td className="py-3 px-4 font-bold text-accent-primary">{idx + 1}</td>
                    <td className="py-3 px-4">
                      <Link href={`/${locale}/novel/${novel.slug}`} className="hover:text-accent-primary" target="_blank">
                        {novel.title}
                      </Link>
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-1">
                        <Star className="w-4 h-4 text-accent-primary fill-accent-primary" />
                        <span>{novel.rating?.toFixed(1) || '—'}</span>
                        <span className="text-xs text-foreground-muted">({novel.ratingsCount || 0})</span>
                      </div>
                    </td>
                    <td className="py-3 px-4 text-foreground-secondary">{novel.viewsCount?.toLocaleString('ru-RU') || 0}</td>
                    <td className="py-3 px-4 text-foreground-secondary">{novel.bookmarksCount?.toLocaleString('ru-RU') || 0}</td>
                    <td className="py-3 px-4">
                      <Link href={`/${locale}/admin/novels/${novel.slug}`} className="btn-ghost p-2" title="Управление">
                        <Plus className="w-4 h-4" />
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        
        <div className="mt-6 space-y-4">
          <div className="p-4 bg-background-tertiary rounded">
            <h3 className="font-semibold mb-2">Управление Featured-новеллами</h3>
            <p className="text-sm text-foreground-secondary mb-2">
              Для установки новелл в раздел "Рекомендуемое" используйте коллекции:
            </p>
            <ul className="list-disc list-inside text-sm text-foreground-secondary ml-4 space-y-1">
              <li>Создайте коллекцию "Featured" в разделе Коллекции</li>
              <li>Добавьте туда нужные новеллы</li>
              <li>Установите её как featured: POST /api/v1/admin/collections/:id/featured</li>
            </ul>
            <Link href={`/${locale}/collections/new`} className="btn-primary mt-4 inline-flex items-center gap-2">
              <Plus className="w-4 h-4" />
              Создать Featured коллекцию
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
