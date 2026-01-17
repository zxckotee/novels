'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { ArrowLeft, BookOpen, FileText, Users, MessageSquare, TrendingUp } from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import { useAdminStats } from '@/lib/api/hooks/useAdminSystem';

export default function AdminStatsPage() {
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user, isLoading: authLoading } = useAuthStore();
  const hasAccess = isAuthenticated && isAdmin(user);

  useEffect(() => {
    if (!authLoading && !hasAccess) router.replace(`/${locale}`);
  }, [authLoading, hasAccess, router, locale]);

  const { data: stats, isLoading } = useAdminStats();

  if (authLoading) return null;
  if (!hasAccess) return null;
  
  return (
    <div className="container-custom py-6">
      <div className="flex items-center gap-4 mb-6">
        <Link href={`/${locale}/admin`} className="btn-ghost p-2"><ArrowLeft className="w-5 h-5" /></Link>
        <h1 className="text-2xl font-heading font-bold">Статистика платформы</h1>
      </div>
      
      {isLoading ? (
        <div className="text-center py-12"><p className="text-foreground-secondary">Загрузка...</p></div>
      ) : !stats ? (
        <div className="text-center py-12"><p className="text-foreground-secondary">Нет данных</p></div>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <StatCard label="Всего новелл" value={stats.totalNovels} icon={<BookOpen className="w-5 h-5" />} />
            <StatCard label="Всего глав" value={stats.totalChapters} icon={<FileText className="w-5 h-5" />} />
            <StatCard label="Пользователей" value={stats.totalUsers} icon={<Users className="w-5 h-5" />} />
            <StatCard label="Комментариев" value={stats.totalComments} icon={<MessageSquare className="w-5 h-5" />} />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="bg-background-secondary rounded-card p-6">
              <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
                <TrendingUp className="w-5 h-5" />
                За неделю
              </h2>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-foreground-secondary">Новых пользователей</span>
                  <span className="font-semibold">{stats.newUsersThisWeek}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-foreground-secondary">Новых новелл</span>
                  <span className="font-semibold">{stats.newNovelsThisWeek}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-foreground-secondary">Новых глав</span>
                  <span className="font-semibold">{stats.newChaptersThisWeek}</span>
                </div>
              </div>
            </div>

            <div className="bg-background-secondary rounded-card p-6">
              <h2 className="text-lg font-semibold mb-4">Средние показатели</h2>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-foreground-secondary">Глав в день</span>
                  <span className="font-semibold">{stats.avgChaptersPerDay.toFixed(1)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-foreground-secondary">Комментариев в день</span>
                  <span className="font-semibold">{stats.avgCommentsPerDay.toFixed(1)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-foreground-secondary">Жалоб на рассмотрении</span>
                  <span className="fontsemibold text-status-warning">{stats.pendingReports}</span>
                </div>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
}

function StatCard({ label, value, icon }: { label: string; value: number; icon: React.ReactNode }) {
  return (
    <div className="bg-background-secondary rounded-card p-4">
      <div className="flex items-center justify-between mb-2">
        <span className="text-foreground-muted">{icon}</span>
      </div>
      <div className="text-2xl font-bold">{value.toLocaleString('ru-RU')}</div>
      <div className="text-sm text-foreground-secondary">{label}</div>
    </div>
  );
}
