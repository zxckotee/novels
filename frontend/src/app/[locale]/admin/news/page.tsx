'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { ArrowLeft, Plus, Edit, Trash2, Eye, Pin } from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import { useAdminNews, useDeleteNews, usePublishNews, usePinNews } from '@/lib/api/hooks/useAdminNews';

export default function AdminNewsPage() {
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user, isLoading: authLoading } = useAuthStore();
  const [page, setPage] = useState(1);
  
  const hasAccess = isAuthenticated && isAdmin(user);

  useEffect(() => {
    if (!authLoading && !hasAccess) {
      router.replace(`/${locale}`);
    }
  }, [authLoading, hasAccess, router, locale]);

  const { data: newsData, isLoading: newsLoading } = useAdminNews({ page, limit: 20 });
  const deleteNews = useDeleteNews();
  const publishNews = usePublishNews();
  const pinNews = usePinNews();

  const handleDelete = async (id: string, title: string) => {
    if (!confirm(`Удалить новость "${title}"?`)) return;
    try {
      await deleteNews.mutateAsync(id);
    } catch (error: any) {
      alert(error.response?.data?.error?.message || 'Ошибка при удалении');
    }
  };

  const handlePublish = async (id: string) => {
    try {
      await publishNews.mutateAsync(id);
    } catch (error: any) {
      alert(error.response?.data?.error?.message || 'Ошибка');
    }
  };

  const handlePin = async (id: string, currentlyPinned: boolean) => {
    try {
      await pinNews.mutateAsync({ id, pinned: !currentlyPinned });
    } catch (error: any) {
      alert(error.response?.data?.error?.message || 'Ошибка');
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
          <h1 className="text-2xl font-heading font-bold">Управление новостями</h1>
        </div>
        <Link href={`/${locale}/admin/news/new`} className="btn-primary flex items-center gap-2">
          <Plus className="w-4 h-4" />
          Добавить новость
        </Link>
      </div>
      
      <div className="bg-background-secondary rounded-card p-6">
        {newsLoading ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Загрузка...</p></div>
        ) : !newsData || !newsData.news || newsData.news.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-foreground-secondary mb-4">Новости не найдены</p>
            <Link href={`/${locale}/admin/news/new`} className="btn-primary">
              Создать первую новость
            </Link>
          </div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-background-tertiary">
                    <th className="text-left py-3 px-4">Заголовок</th>
                    <th className="text-left py-3 px-4">Slug</th>
                    <th className="text-left py-3 px-4">Дата публикации</th>
                    <th className="text-right py-3 px-4">Действия</th>
                  </tr>
                </thead>
                <tbody>
                  {newsData.news.map((news) => (
                    <tr key={news.id} className="border-b border-background-tertiary hover:bg-background-hover">
                      <td className="py-3 px-4">
                        <div>
                          {news.title}
                          {news.id.includes('pinned') && (
                            <Pin className="inline w-3 h-3 ml-2 text-accent-primary" />
                          )}
                        </div>
                      </td>
                      <td className="py-3 px-4 font-mono text-sm text-foreground-secondary">
                        {news.slug}
                      </td>
                      <td className="py-3 px-4 text-sm text-foreground-secondary">
                        {news.publishedAt ? new Date(news.publishedAt).toLocaleDateString('ru-RU') : 'Черновик'}
                      </td>
                      <td className="py-3 px-4">
                        <div className="flex items-center justify-end gap-2">
                          <Link
                            href={`/${locale}/news/${news.slug}`}
                            className="btn-ghost p-2"
                            title="Просмотр"
                            target="_blank"
                          >
                            <Eye className="w-4 h-4" />
                          </Link>
                          <Link
                            href={`/${locale}/admin/news/${news.id}/edit`}
                            className="btn-ghost p-2"
                            title="Редактировать"
                          >
                            <Edit className="w-4 h-4" />
                          </Link>
                          {!news.publishedAt && (
                            <button
                              onClick={() => handlePublish(news.id)}
                              className="btn-ghost p-2 text-status-success"
                              title="Опубликовать"
                              disabled={publishNews.isPending}
                            >
                              <Eye className="w-4 h-4" />
                            </button>
                          )}
                          <button
                            onClick={() => handleDelete(news.id, news.title)}
                            className="btn-ghost p-2 text-status-error"
                            title="Удалить"
                            disabled={deleteNews.isPending}
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
            
            {newsData.total > 20 && (
              <div className="flex items-center justify-between mt-6">
                <p className="text-sm text-foreground-secondary">
                  Показано {newsData.news.length} из {newsData.total}
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
                    disabled={page * 20 >= newsData.total}
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
