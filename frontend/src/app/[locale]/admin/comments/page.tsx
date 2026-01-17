'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { ArrowLeft, Search, Trash2 } from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import { useAdminComments, useSoftDeleteComment, useHardDeleteComment } from '@/lib/api/hooks/useAdminComments';

export default function AdminCommentsPage() {
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user, isLoading: authLoading } = useAuthStore();
  const [page, setPage] = useState(1);
  const [showDeleted, setShowDeleted] = useState(false);
  const hasAccess = isAuthenticated && isAdmin(user);

  useEffect(() => {
    if (!authLoading && !hasAccess) router.replace(`/${locale}`);
  }, [authLoading, hasAccess, router, locale]);

  const { data: commentsData, isLoading } = useAdminComments({
    isDeleted: showDeleted ? true : undefined,
    page,
    limit: 20,
  });

  const softDelete = useSoftDeleteComment();
  const hardDelete = useHardDeleteComment();

  const handleSoftDelete = async (id: string) => {
    if (!confirm('Пометить комментарий как удаленный?')) return;
    try {
      await softDelete.mutateAsync(id);
    } catch (error: any) {
      alert(error.response?.data?.error?.message || 'Ошибка');
    }
  };

  const handleHardDelete = async (id: string) => {
    if (!confirm('НЕОБРАТИМО удалить комментарий из БД?')) return;
    try {
      await hardDelete.mutateAsync(id);
    } catch (error: any) {
      alert(error.response?.data?.error?.message || 'Ошибка');
    }
  };

  if (authLoading) return null;
  if (!hasAccess) return null;
  
  return (
    <div className="container-custom py-6">
      <div className="flex items-center gap-4 mb-6">
        <Link href={`/${locale}/admin`} className="btn-ghost p-2"><ArrowLeft className="w-5 h-5" /></Link>
        <h1 className="text-2xl font-heading font-bold">Управление комментариями</h1>
      </div>

      <div className="mb-6 flex gap-4">
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={showDeleted}
            onChange={(e) => { setShowDeleted(e.target.checked); setPage(1); }}
            className="checkbox"
          />
          <span className="text-sm">Показать удаленные</span>
        </label>
      </div>

      <div className="bg-background-secondary rounded-card p-6">
        {isLoading ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Загрузка...</p></div>
        ) : !commentsData || !commentsData.comments || commentsData.comments.length === 0 ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Комментарии не найдены</p></div>
        ) : (
          <>
            <div className="space-y-4">
              {commentsData.comments?.map((comment) => (
                <div key={comment.id} className={`p-4 rounded border ${comment.isDeleted ? 'bg-background-tertiary border-status-error' : 'border-background-tertiary'}`}>
                  <div className="flex justify-between items-start mb-2">
                    <div className="text-sm text-foreground-secondary">
                      Пользователь #{comment.userId.substring(0, 8)}
                    </div>
                    <div className="flex gap-2">
                      {!comment.isDeleted && (
                        <button onClick={() => handleSoftDelete(comment.id)} className="btn-ghost p-1 text-status-warning" title="Soft delete">
                          <Trash2 className="w-4 h-4" />
                        </button>
                      )}
                      <button onClick={() => handleHardDelete(comment.id)} className="btn-ghost p-1 text-status-error" title="Hard delete">
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                  <p className={comment.isDeleted ? 'text-foreground-muted italic' : ''}>{comment.body}</p>
                  <div className="text-xs text-foreground-muted mt-2">
                    {new Date(comment.createdAt).toLocaleString('ru-RU')} • 👍 {comment.likesCount} 👎 {comment.dislikesCount}
                  </div>
                </div>
              ))}
            </div>
            
            {commentsData.totalCount > commentsData.limit && (
              <div className="flex justify-between mt-6">
                <p className="text-sm text-foreground-secondary">Показано {commentsData.comments.length} из {commentsData.totalCount}</p>
                <div className="flex gap-2">
                  <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1} className="btn-secondary">Назад</button>
                  <button onClick={() => setPage(p => p + 1)} disabled={page * commentsData.limit >= commentsData.totalCount} className="btn-secondary">Вперед</button>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
