'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { 
  ArrowLeft, 
  Plus, 
  Search,
  Edit,
  Trash2
} from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import { useRouter } from 'next/navigation';
import { useAdminNovels, useDeleteNovel } from '@/lib/api/hooks/useAdminNovels';

export default function AdminNovelsPage() {
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user, isLoading: authLoading } = useAuthStore();
  const [searchQuery, setSearchQuery] = useState('');
  const [page, setPage] = useState(1);
  
  const hasAccess = isAuthenticated && isAdmin(user);

  useEffect(() => {
    if (!authLoading && !hasAccess) {
      router.replace(`/${locale}`);
    }
  }, [authLoading, hasAccess, router, locale]);

  const { data: novelsData, isLoading: novelsLoading } = useAdminNovels({
    search: searchQuery,
    page,
    limit: 20,
  });

  const deleteNovel = useDeleteNovel();

  const handleDelete = async (id: string, title: string) => {
    if (!confirm(`Удалить новеллу "${title}"? Это также удалит все главы!`)) return;
    
    try {
      await deleteNovel.mutateAsync(id);
    } catch (error: any) {
      alert(error.response?.data?.error?.message || 'Ошибка при удалении');
    }
  };

  if (authLoading) return null;
  if (!hasAccess) return null;
  
  return (
    <div className="container-custom py-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <Link
            href={`/${locale}/admin`}
            className="btn-ghost p-2"
          >
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <h1 className="text-2xl font-heading font-bold">Управление новеллами</h1>
        </div>
        
        <Link
          href={`/${locale}/admin/novels/new`}
          className="btn-primary flex items-center gap-2"
        >
          <Plus className="w-4 h-4" />
          Добавить новеллу
        </Link>
      </div>
      
      {/* Search */}
      <div className="mb-6">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground-muted" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => {
              setSearchQuery(e.target.value);
              setPage(1);
            }}
            className="input w-full pl-10"
            placeholder="Поиск новелл..."
          />
        </div>
      </div>
      
      {/* Content */}
      <div className="bg-background-secondary rounded-card p-6">
        {novelsLoading ? (
          <div className="text-center py-12">
            <p className="text-foreground-secondary">Загрузка...</p>
          </div>
        ) : !novelsData || !novelsData.novels || novelsData.novels.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-foreground-secondary mb-4">Новеллы не найдены</p>
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className="btn-secondary"
              >
                Сбросить поиск
              </button>
            )}
          </div>
        ) : (
          <>
            {/* Table */}
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-background-tertiary">
                    <th className="text-left py-3 px-4">Название</th>
                    <th className="text-left py-3 px-4">Slug</th>
                    <th className="text-left py-3 px-4">Статус</th>
                    <th className="text-left py-3 px-4">Рейтинг</th>
                    <th className="text-right py-3 px-4">Действия</th>
                  </tr>
                </thead>
                <tbody>
                  {novelsData.novels.map((novel) => (
                    <tr 
                      key={novel.id}
                      className="border-b border-background-tertiary hover:bg-background-hover"
                    >
                      <td className="py-3 px-4">{novel.title}</td>
                      <td className="py-3 px-4 font-mono text-sm text-foreground-secondary">
                        {novel.slug}
                      </td>
                      <td className="py-3 px-4">
                        <span className={`text-xs px-2 py-1 rounded ${
                          novel.translationStatus === 'ongoing' ? 'bg-status-info text-white' :
                          novel.translationStatus === 'completed' ? 'bg-status-success text-white' :
                          novel.translationStatus === 'paused' ? 'bg-status-warning text-white' :
                          'bg-background-tertiary'
                        }`}>
                          {novel.translationStatus}
                        </span>
                      </td>
                      <td className="py-3 px-4">
                        <div className="flex items-center gap-1">
                          <span className="text-accent-primary">★</span>
                          <span>{novel.rating?.toFixed(1) || '—'}</span>
                          <span className="text-xs text-foreground-muted">
                            ({novel.ratingsCount || 0})
                          </span>
                        </div>
                      </td>
                      <td className="py-3 px-4">
                        <div className="flex items-center justify-end gap-2">
                          <Link
                            href={`/${locale}/novel/${novel.slug}`}
                            className="btn-ghost p-2 text-accent-primary"
                            title="Просмотр"
                            target="_blank"
                          >
                            <Edit className="w-4 h-4" />
                          </Link>
                          <button
                            onClick={() => handleDelete(novel.id, novel.title)}
                            className="btn-ghost p-2 text-status-error"
                            title="Удалить"
                            disabled={deleteNovel.isPending}
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            
            {/* Pagination */}
            {novelsData.total > 20 && (
              <div className="flex items-center justify-between mt-6">
                <p className="text-sm text-foreground-secondary">
                  Показано {novelsData.novels.length} из {novelsData.total}
                </p>
                <div className="flex gap-2">
                  <button
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    disabled={page === 1}
                    className="btn-secondary"
                  >
                    Назад
                  </button>
                  <button
                    onClick={() => setPage(p => p + 1)}
                    disabled={page * 20 >= novelsData.total}
                    className="btn-secondary"
                  >
                    Вперед
                  </button>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
