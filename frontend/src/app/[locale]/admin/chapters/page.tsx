'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { 
  ArrowLeft, 
  Plus, 
  Search,
  FileText,
  Trash2
} from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import { useRouter } from 'next/navigation';
import { useAdminChapters, useDeleteChapter } from '@/lib/api/hooks/useAdminChapters';

export default function AdminChaptersPage() {
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

  const { data: chaptersData, isLoading: chaptersLoading } = useAdminChapters({
    page,
    limit: 50,
  });

  const deleteChapter = useDeleteChapter();

  const handleDelete = async (id: string, title: string) => {
    if (!confirm(`Удалить главу "${title}"?`)) return;
    
    try {
      await deleteChapter.mutateAsync(id);
    } catch (error: any) {
      alert(error.response?.data?.error?.message || 'Ошибка при удалении');
    }
  };

  if (authLoading) return null;
  if (!hasAccess) return null;

  // Filter chapters by search query (slug or title)
  const filteredChapters = chaptersData?.chapters.filter(ch => 
    !searchQuery || 
    ch.title?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    ch.slug?.toLowerCase().includes(searchQuery.toLowerCase())
  ) || [];
  
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
          <h1 className="text-2xl font-heading font-bold">Управление главами</h1>
        </div>
        
        <Link
          href={`/${locale}/admin/chapters/new`}
          className="btn-primary flex items-center gap-2"
        >
          <Plus className="w-4 h-4" />
          Добавить главу
        </Link>
      </div>
      
      {/* Search */}
      <div className="mb-6">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground-muted" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="input w-full pl-10"
            placeholder="Поиск глав по названию..."
          />
        </div>
      </div>
      
      {/* Content */}
      <div className="bg-background-secondary rounded-card p-6">
        {chaptersLoading ? (
          <div className="text-center py-12">
            <FileText className="w-16 h-16 mx-auto mb-4 text-foreground-muted" />
            <p className="text-foreground-secondary">Загрузка глав...</p>
          </div>
        ) : filteredChapters.length === 0 ? (
          <div className="text-center py-12">
            <FileText className="w-16 h-16 mx-auto mb-4 text-foreground-muted" />
            <p className="text-foreground-secondary mb-4">Главы не найдены</p>
            {searchQuery && (
              <button onClick={() => setSearchQuery('')} className="btn-secondary">
                Сбросить поиск
              </button>
            )}
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-background-tertiary">
                    <th className="text-left py-3 px-4">Номер</th>
                    <th className="text-left py-3 px-4">Название</th>
                    <th className="text-left py-3 px-4">ID Новеллы</th>
                    <th className="text-left py-3 px-4">Просмотры</th>
                    <th className="text-left py-3 px-4">Дата публикации</th>
                    <th className="text-right py-3 px-4">Действия</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredChapters.map((chapter) => (
                    <tr 
                      key={chapter.id}
                      className="border-b border-background-tertiary hover:bg-background-hover"
                    >
                      <td className="py-3 px-4 font-semibold">{chapter.number}</td>
                      <td className="py-3 px-4">{chapter.title || '—'}</td>
                      <td className="py-3 px-4 font-mono text-xs text-foreground-secondary">
                        {chapter.novelId.substring(0, 8)}...
                      </td>
                      <td className="py-3 px-4 text-foreground-secondary">
                        {chapter.wordCount || 0}
                      </td>
                      <td className="py-3 px-4 text-sm text-foreground-secondary">
                        {chapter.publishedAt ? new Date(chapter.publishedAt).toLocaleDateString('ru-RU') : 'Черновик'}
                      </td>
                      <td className="py-3 px-4">
                        <div className="flex items-center justify-end gap-2">
                          <button
                            onClick={() => handleDelete(chapter.id, chapter.title || `Глава ${chapter.number}`)}
                            className="btn-ghost p-2 text-status-error"
                            title="Удалить"
                            disabled={deleteChapter.isPending}
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
            
            {chaptersData && chaptersData.total > 50 && (
              <div className="flex items-center justify-between mt-6">
                <p className="text-sm text-foreground-secondary">
                  Показано {filteredChapters.length} из {chaptersData.total}
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
                    disabled={page * 50 >= chaptersData.total}
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
