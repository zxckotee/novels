'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { ArrowLeft, Search, Ban, Shield } from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import { useAdminUsers, useBanUser, useUnbanUser } from '@/lib/api/hooks/useAdminUsers';

export default function AdminUsersPage() {
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user, isLoading: authLoading } = useAuthStore();
  const [searchQuery, setSearchQuery] = useState('');
  const [page, setPage] = useState(1);
  const hasAccess = isAuthenticated && isAdmin(user);

  useEffect(() => {
    if (!authLoading && !hasAccess) router.replace(`/${locale}`);
  }, [authLoading, hasAccess, router, locale]);

  const { data: usersData, isLoading } = useAdminUsers({
    query: searchQuery,
    page,
    limit: 20,
  });

  const banUser = useBanUser();
  const unbanUser = useUnbanUser();

  const handleBan = async (id: string, email: string) => {
    const reason = prompt(`Причина бана пользователя ${email}:`);
    if (!reason) return;
    try {
      await banUser.mutateAsync({ id, reason });
    } catch (error: any) {
      alert(error.response?.data?.error?.message || 'Ошибка');
    }
  };

  const handleUnban = async (id: string) => {
    try {
      await unbanUser.mutateAsync(id);
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
        <h1 className="text-2xl font-heading font-bold">Управление пользователями</h1>
      </div>
      
      <div className="mb-6">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground-muted" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => { setSearchQuery(e.target.value); setPage(1); }}
            className="input w-full pl-10"
            placeholder="Поиск по email или имени..."
          />
        </div>
      </div>

      <div className="bg-background-secondary rounded-card p-6">
        {isLoading ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Загрузка...</p></div>
        ) : !usersData || usersData.users.length === 0 ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Пользователи не найдены</p></div>
        ) : (
          <>
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-background-tertiary">
                    <th className="text-left py-3 px-4">Email</th>
                    <th className="text-left py-3 px-4">Имя</th>
                    <th className="text-left py-3 px-4">Роли</th>
                    <th className="text-left py-3 px-4">Статус</th>
                    <th className="text-right py-3 px-4">Действия</th>
                  </tr>
                </thead>
                <tbody>
                  {usersData.users.map((u) => (
                    <tr key={u.id} className="border-b border-background-tertiary hover:bg-background-hover">
                      <td className="py-3 px-4 font-mono text-sm">{u.email}</td>
                      <td className="py-3 px-4">{u.displayName}</td>
                      <td className="py-3 px-4">
                        <div className="flex gap-1 flex-wrap">
                          {(u.roles || []).map((r) => (
                            <span key={r} className="text-xs bg-background-tertiary px-2 py-0.5 rounded">{r}</span>
                          ))}
                        </div>
                      </td>
                      <td className="py-3 px-4">
                        {u.isBanned ? (
                          <span className="text-status-error">Заблокирован</span>
                        ) : (
                          <span className="text-status-success">Активен</span>
                        )}
                      </td>
                      <td className="py-3 px-4">
                        <div className="flex items-center justify-end gap-2">
                          {u.isBanned ? (
                            <button onClick={() => handleUnban(u.id)} className="btn-ghost p-2 text-status-success" title="Разбан" disabled={unbanUser.isPending}>
                              <Shield className="w-4 h-4" />
                            </button>
                          ) : (
                            <button onClick={() => handleBan(u.id, u.email)} className="btn-ghost p-2 text-status-error" title="Забанить" disabled={banUser.isPending}>
                              <Ban className="w-4 h-4" />
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            
            {usersData.totalCount > usersData.limit && (
              <div className="flex justify-between mt-6">
                <p className="text-sm text-foreground-secondary">Показано {usersData.users.length} из {usersData.totalCount}</p>
                <div className="flex gap-2">
                  <button onClick={() => setPage(p => Math.max(1, p - 1))} disabled={page === 1} className="btn-secondary">Назад</button>
                  <button onClick={() => setPage(p => p + 1)} disabled={page * usersData.limit >= usersData.totalCount} className="btn-secondary">Вперед</button>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
