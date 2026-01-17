'use client';

import { useEffect } from 'react';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { 
  BookOpen, 
  FileText, 
  Users, 
  Settings,
  Plus,
  BarChart3,
  MessageSquare,
  Flag,
  Layers
} from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import { useRouter } from 'next/navigation';

export default function AdminDashboard() {
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user, isLoading } = useAuthStore();
  
  const hasAccess = isAuthenticated && isAdmin(user);

  // Redirect must happen in an effect (not during render)
  useEffect(() => {
    if (!isLoading && !hasAccess) {
      router.replace(`/${locale}`);
    }
  }, [isLoading, hasAccess, router, locale]);

  if (isLoading) return null;
  if (!hasAccess) return null;
  
  return (
    <div className="container-custom py-6">
      <h1 className="text-2xl font-heading font-bold mb-6">Панель администратора</h1>
      
      {/* Quick Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        <StatCard label="Всего новелл" value="0" icon={<BookOpen className="w-5 h-5" />} />
        <StatCard label="Всего глав" value="0" icon={<FileText className="w-5 h-5" />} />
        <StatCard label="Пользователей" value="0" icon={<Users className="w-5 h-5" />} />
        <StatCard label="Комментариев" value="0" icon={<MessageSquare className="w-5 h-5" />} />
      </div>
      
      {/* Quick Actions */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
        <Link 
          href={`/${locale}/admin/novels/new`}
          className="flex items-center gap-4 p-4 bg-accent-primary/10 border border-accent-primary rounded-card hover:bg-accent-primary/20 transition-colors"
        >
          <div className="p-3 bg-accent-primary rounded-lg">
            <Plus className="w-6 h-6 text-white" />
          </div>
          <div>
            <div className="font-semibold">Добавить новеллу</div>
            <div className="text-sm text-foreground-secondary">Создать новый тайтл</div>
          </div>
        </Link>
        
        <Link 
          href={`/${locale}/admin/chapters/new`}
          className="flex items-center gap-4 p-4 bg-background-secondary border border-background-tertiary rounded-card hover:bg-background-hover transition-colors"
        >
          <div className="p-3 bg-background-tertiary rounded-lg">
            <FileText className="w-6 h-6" />
          </div>
          <div>
            <div className="font-semibold">Добавить главу</div>
            <div className="text-sm text-foreground-secondary">Загрузить новую главу</div>
          </div>
        </Link>
        
        <Link 
          href={`/${locale}/admin/moderation`}
          className="flex items-center gap-4 p-4 bg-background-secondary border border-background-tertiary rounded-card hover:bg-background-hover transition-colors"
        >
          <div className="p-3 bg-background-tertiary rounded-lg">
            <Flag className="w-6 h-6" />
          </div>
          <div>
            <div className="font-semibold">Модерация</div>
            <div className="text-sm text-foreground-secondary">Жалобы и правки</div>
          </div>
        </Link>
      </div>
      
      {/* Navigation */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Content Management */}
        <div className="bg-background-secondary rounded-card p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <BookOpen className="w-5 h-5" />
            Управление контентом
          </h2>
          <nav className="space-y-2">
            <AdminLink href={`/${locale}/admin/novels`} label="Все новеллы" count={0} />
            <AdminLink href={`/${locale}/admin/chapters`} label="Все главы" count={0} />
            <AdminLink href={`/${locale}/admin/collections`} label="Коллекции" />
            <AdminLink href={`/${locale}/admin/genres`} label="Жанры" />
            <AdminLink href={`/${locale}/admin/tags`} label="Теги" />
            <AdminLink href={`/${locale}/admin/authors`} label="Авторы" />
          </nav>
        </div>
        
        {/* User Management */}
        <div className="bg-background-secondary rounded-card p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <Users className="w-5 h-5" />
            Пользователи
          </h2>
          <nav className="space-y-2">
            <AdminLink href={`/${locale}/admin/users`} label="Все пользователи" count={0} />
            <AdminLink href={`/${locale}/admin/comments`} label="Комментарии" count={0} />
            <AdminLink href={`/${locale}/admin/reports`} label="Жалобы" badge="0" />
          </nav>
        </div>
        
        {/* System */}
        <div className="bg-background-secondary rounded-card p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <Settings className="w-5 h-5" />
            Система
          </h2>
          <nav className="space-y-2">
            <AdminLink href={`/${locale}/admin/news`} label="Новости" />
            <AdminLink href={`/${locale}/admin/settings`} label="Настройки" />
            <AdminLink href={`/${locale}/admin/logs`} label="Логи" />
          </nav>
        </div>
        
        {/* Analytics */}
        <div className="bg-background-secondary rounded-card p-6">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <BarChart3 className="w-5 h-5" />
            Аналитика
          </h2>
          <nav className="space-y-2">
            <AdminLink href={`/${locale}/admin/stats`} label="Статистика" />
            <AdminLink href={`/${locale}/admin/popular`} label="Популярное" />
          </nav>
        </div>
      </div>
    </div>
  );
}

// Stat Card
function StatCard({ label, value, icon }: { label: string; value: string; icon: React.ReactNode }) {
  return (
    <div className="bg-background-secondary rounded-card p-4">
      <div className="flex items-center justify-between mb-2">
        <span className="text-foreground-muted">{icon}</span>
      </div>
      <div className="text-2xl font-bold">{value}</div>
      <div className="text-sm text-foreground-secondary">{label}</div>
    </div>
  );
}

// Admin Link
function AdminLink({ 
  href, 
  label, 
  count, 
  badge 
}: { 
  href: string; 
  label: string; 
  count?: number; 
  badge?: string;
}) {
  return (
    <Link
      href={href}
      className="flex items-center justify-between p-2 rounded hover:bg-background-hover transition-colors"
    >
      <span>{label}</span>
      {count !== undefined && (
        <span className="text-sm text-foreground-muted">{count}</span>
      )}
      {badge && (
        <span className="bg-status-error text-white text-xs px-2 py-0.5 rounded-full">
          {badge}
        </span>
      )}
    </Link>
  );
}
