'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { ArrowLeft, FileText } from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import { useAdminLogs } from '@/lib/api/hooks/useAdminSystem';

export default function AdminLogsPage() {
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user, isLoading: authLoading } = useAuthStore();
  const [page, setPage] = useState(1);
  const hasAccess = isAuthenticated && isAdmin(user);

  useEffect(() => {
    if (!authLoading && !hasAccess) router.replace(`/${locale}`);
  }, [authLoading, hasAccess, router, locale]);

  const { data: logsData, isLoading } = useAdminLogs({ page, limit: 50 });

  if (authLoading) return null;
  if (!hasAccess) return null;
  
  return (
    <div className="container-custom py-6">
      <div className="flex items-center gap-4 mb-6">
        <Link href={`/${locale}/admin`} className="btn-ghost p-2"><ArrowLeft className="w-5 h-5" /></Link>
        <h1 className="text-2xl font-heading font-bold">Логи действий</h1>
      </div>
      
      <div className="bg-background-secondary rounded-card p-6">
        {isLoading ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Загрузка...</p></div>
        ) : !logsData || !logsData.logs || logsData.logs.length === 0 ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Логи не найдены</p></div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-background-tertiary">
                    <th className="text-left py-2 px-3">Дата</th>
                    <th className="text-left py-2 px-3">Действие</th>
                    <th className="text-left py-2 px-3">Сущность</th>
                    <th className="text-left py-2 px-3">ID</th>
                  </tr>
                </thead>
                <tbody>
                  {logsData.logs.map((log) => (
                    <tr key={log.id} className="border-b border-background-tertiary hover:bg-background-hover">
                      <td className="py-2 px-3 text-foreground-secondary whitespace-nowrap">
                        {new Date(log.createdAt).toLocaleString('ru-RU')}
                      </td>
                      <td className="py-2 px-3 font-mono">{log.action}</td>
                      <td className="py-2 px-3">{log.entityType}</td>
                      <td className="py-2 px-3 font-mono text-foreground-secondary">
                        {log.entityId?.substring(0, 8) || '-'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            
            {logsData.totalCount > logsData.limit && (
              <div className="flex justify-between mt-6">
                <p className="text-sm text-foreground-secondary">Показано {logsData.logs.length} из {logsData.totalCount}</p>
                <div className="flex gap-2">
                  <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1} className="btn-secondary">Назад</button>
                  <button onClick={() => setPage(p => p + 1)} disabled={page * logsData.limit >= logsData.totalCount} className="btn-secondary">Вперед</button>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
