'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { ArrowLeft, Plus, Search, Edit, Trash2 } from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import { useAdminGenres, useDeleteGenre } from '@/lib/api/hooks/useAdminGenresTags';

export default function AdminGenresPage() {
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user, isLoading } = useAuthStore();
  const [searchQuery, setSearchQuery] = useState('');
  const [page, setPage] = useState(1);
  
  const hasAccess = isAuthenticated && isAdmin(user);

  useEffect(() => {
    if (!isLoading && !hasAccess) {
      router.replace(`/${locale}`);
    }
  }, [isLoading, hasAccess, router, locale]);

  const { data: genresData, isLoading } = useAdminGenres({
    query: searchQuery,
    lang: locale,
    page,
    limit: 50,
  });

  const deleteGenre = useDeleteGenre();

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Удалить жанр "${name}"?`)) return;
    
    try {
      await deleteGenre.mutateAsync(id);
    } catch (error: any) {
      alert(error.response?.data?.error?.message || 'Ошибка при удалении');
    }
  };

  if (isLoading) return null;
  if (!hasAccess) return null;
  
  return (
    <div className="container-custom py-6">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-4">
          <Link href={`/${locale}/admin`} className="btn-ghost p-2">
            <ArrowLeft className="w-5 h-5" />
          </Link>
          <h1 className="text-2xl font-heading font-bold">Управление жанрами</h1>
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
            placeholder="Поиск жанров..."
          />
        </div>
      </div>
      
      <div className="bg-background-secondary rounded-card p-6">
        {isLoading ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Загрузка...</p></div>
        ) : !genresData || !genresData.genres || genresData.genres.length === 0 ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Жанры не найдены</p></div>
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
                  {genresData.genres.map((genre) => (
                    <tr key={genre.id} className="border-b border-background-tertiary hover:bg-background-hover">
                      <td className="py-3 px-4">{genre.name || genre.slug}</td>
                      <td className="py-3 px-4 font-mono text-sm text-foreground-secondary">{genre.slug}</td>
                      <td className="py-3 px-4">
                        <div className="flex items-center justify-end gap-2">
                          <button
                            onClick={() => handleDelete(genre.id, genre.name || genre.slug)}
                            className="btn-ghost p-2 text-status-error"
                            title="Удалить"
                            disabled={deleteGenre.isPending}
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
            
            {genresData.totalCount > genresData.limit && (
              <div className="flex items-center justify-between mt-6">
                <p className="text-sm text-foreground-secondary">
                  Показано {genresData.genres.length} из {genresData.totalCount}
                </p>
                <div className="flex gap-2">
                  <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1} className="btn-secondary">Назад</button>
                  <button onClick={() => setPage(p => p + 1)} disabled={page * genresData.limit >= genresData.totalCount} className="btn-secondary">Вперед</button>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
