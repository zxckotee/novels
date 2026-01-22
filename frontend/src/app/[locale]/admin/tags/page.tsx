'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { ArrowLeft, Plus, Search, Edit, Trash2 } from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import { useAdminTags, useDeleteTag } from '@/lib/api/hooks/useAdminGenresTags';

export default function AdminTagsPage() {
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

  const { data: tagsData, isLoading: tagsLoading } = useAdminTags({
    query: searchQuery,
    lang: locale,
    page,
    limit: 50,
  });

  const deleteTag = useDeleteTag();

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Удалить тег "${name}"?`)) return;
    
    try {
      await deleteTag.mutateAsync(id);
    } catch (error: any) {
      alert(error.response?.data?.error?.message || 'Ошибка при удалении');
    }
  };

  if (authLoading) return null;
  if (!hasAccess) return null;
  
  return (
    <div className="container-custom py-6">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <Link href={`/${locale}/admin`} className="btn-ghost p-2">
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <h1 className="text-2xl font-heading font-bold">Управление тегами</h1>
        </div>
      </div>
      
      <div className="mb-6">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground-muted" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => { setSearchQuery(e.target.value); setPage(1); }}
            className="input w-full pl-10"
            placeholder="Поиск тегов..."
          />
        </div>
      </div>
      
      <div className="bg-background-secondary rounded-card p-6">
        {tagsLoading ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Загрузка...</p></div>
        ) : !tagsData || !tagsData.tags || tagsData.tags.length === 0 ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Теги не найдены</p></div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-background-tertiary">
                    <th className="text-left py-3 px-4">Название</th>
                    <th className="text-left py-3 px-4">Slug</th>
                    <th className="text-right py-3 px-4">Действия</th>
                  </tr>
                </thead>
                <tbody>
                  {tagsData.tags.map((tag) => (
                    <tr key={tag.id} className="border-b border-background-tertiary hover:bg-background-hover">
                      <td className="py-3 px-4">{tag.name || tag.slug}</td>
                      <td className="py-3 px-4 font-mono text-sm text-foreground-secondary">{tag.slug}</td>
                      <td className="py-3 px-4">
                        <div className="flex items-center justify-end gap-2">
                          <button
                            onClick={() => handleDelete(tag.id, tag.name || tag.slug)}
                            className="btn-ghost p-2 text-status-error"
                            title="Удалить"
                            disabled={deleteTag.isPending}
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
            
            {tagsData.totalCount > tagsData.limit && (
              <div className="flex items-center justify-between mt-6">
                <p className="text-sm text-foreground-secondary">
                  Показано {tagsData.tags.length} из {tagsData.totalCount}
                </p>
                <div className="flex gap-2">
                  <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1} className="btn-secondary">Назад</button>
                  <button onClick={() => setPage(p => p + 1)} disabled={page * tagsData.limit >= tagsData.totalCount} className="btn-secondary">Вперед</button>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
