'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { ArrowLeft, Check, X, Trash2 } from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import { useAdminReports, useResolveReport } from '@/lib/api/hooks/useAdminComments';

export default function AdminReportsPage() {
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user, isLoading: authLoading } = useAuthStore();
  const [page, setPage] = useState(1);
  const [statusFilter, setStatusFilter] = useState('pending');
  const hasAccess = isAuthenticated && isAdmin(user);

  useEffect(() => {
    if (!authLoading && !hasAccess) router.replace(`/${locale}`);
  }, [authLoading, hasAccess, router, locale]);

  const { data: reportsData, isLoading } = useAdminReports({
    status: statusFilter,
    page,
    limit: 20,
  });

  const resolveReport = useResolveReport();

  const handleResolve = async (id: string, action: string) => {
    const confirmTexts = {
      resolve: 'Отметить жалобу как решенную?',
      dismiss: 'Отклонить жалобу?',
      delete_comment: 'Удалить комментарий и закрыть жалобу?',
    };
    
    if (!confirm(confirmTexts[action as keyof typeof confirmTexts])) return;
    
    try {
      await resolveReport.mutateAsync({ id, action });
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
        <h1 className="text-2xl font-heading font-bold">Жалобы на комментарии</h1>
      </div>

      <div className="mb-6">
        <select value={statusFilter} onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }} className="input">
          <option value="pending">Ожидают</option>
          <option value="resolved">Решенные</option>
          <option value="dismissed">Отклоненные</option>
        </select>
      </div>

      <div className="bg-background-secondary rounded-card p-6">
        {isLoading ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Загрузка...</p></div>
        ) : !reportsData || !reportsData.reports || reportsData.reports.length === 0 ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Жалобы не найдены</p></div>
        ) : (
          <>
            <div className="space-y-4">
              {reportsData.reports?.map((report) => (
                <div key={report.id} className="p-4 border border-background-tertiary rounded">
                  <div className="flex justify-between mb-2">
                    <span className="text-sm text-foreground-secondary">
                      Жалоба #{report.id.substring(0, 8)} • {new Date(report.createdAt).toLocaleString('ru-RU')}
                    </span>
                    <span className={`text-xs px-2 py-1 rounded ${
                      report.status === 'pending' ? 'bg-status-warning text-white' :
                      report.status === 'resolved' ? 'bg-status-success text-white' :
                      'bg-background-tertiary'
                    }`}>
                      {report.status}
                    </span>
                  </div>
                  <p className="mb-3">{report.reason}</p>
                  {report.status === 'pending' && (
                    <div className="flex gap-2">
                      <button onClick={() => handleResolve(report.id, 'resolve')} className="btn-secondary text-sm flex items-center gap-1" disabled={resolveReport.isPending}>
                        <Check className="w-4 h-4" /> Решить
                      </button>
                      <button onClick={() => handleResolve(report.id, 'dismiss')} className="btn-secondary text-sm flex items-center gap-1" disabled={resolveReport.isPending}>
                        <X className="w-4 h-4" /> Отклонить
                      </button>
                      <button onClick={() => handleResolve(report.id, 'delete_comment')} className="btn-secondary text-sm flex items-center gap-1 text-status-error" disabled={resolveReport.isPending}>
                        <Trash2 className="w-4 h-4" /> Удалить комментарий
                      </button>
                    </div>
                  )}
                </div>
              ))}
            </div>
            
            {reportsData.totalCount > reportsData.limit && (
              <div className="flex justify-between mt-6">
                <p className="text-sm text-foreground-secondary">Показано {reportsData.reports.length} из {reportsData.totalCount}</p>
                <div className="flex gap-2">
                  <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1} className="btn-secondary">Назад</button>
                  <button onClick={() => setPage(p => p + 1)} disabled={page * reportsData.limit >= reportsData.totalCount} className="btn-secondary">Вперед</button>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
