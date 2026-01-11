'use client';

import { useState, useEffect } from 'react';
import Image from 'next/image';
import Link from 'next/link';
import { useTranslations, useLocale } from 'next-intl';
import { 
  User, 
  BookOpen, 
  Clock, 
  MessageSquare, 
  Bookmark,
  Award,
  Settings,
  Crown,
  ChevronRight,
  TrendingUp
} from 'lucide-react';
import { useUserProfile, useCurrentUser } from '@/lib/api/hooks/useAuth';
import { useAuthStore } from '@/store/auth';
import { useRouter } from 'next/navigation';

// XP needed for each level (example progression)
const XP_PER_LEVEL = 1000;

function calculateLevel(xp: number): { level: number; currentXP: number; nextLevelXP: number; progress: number } {
  const level = Math.floor(xp / XP_PER_LEVEL) + 1;
  const currentXP = xp % XP_PER_LEVEL;
  const nextLevelXP = XP_PER_LEVEL;
  const progress = (currentXP / nextLevelXP) * 100;
  return { level, currentXP, nextLevelXP, progress };
}

export default function ProfilePageClient() {
  const t = useTranslations('profile');
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user: authUser, logout } = useAuthStore();
  
  const [activeTab, setActiveTab] = useState<'stats' | 'bookmarks' | 'achievements'>('stats');
  const [mounted, setMounted] = useState(false);
  
  const { data: profile, isLoading, error } = useUserProfile();
  
  // Handle mount to prevent hydration mismatch
  useEffect(() => {
    setMounted(true);
  }, []);
  
  // Redirect if not authenticated (only on client)
  useEffect(() => {
    if (mounted && !isAuthenticated) {
      router.push(`/${locale}/login`);
    }
  }, [mounted, isAuthenticated, locale, router]);
  
  // Handle 401 errors - logout and redirect
  useEffect(() => {
    if (error && (error as any)?.response?.status === 401) {
      logout();
      router.push(`/${locale}/login`);
    }
  }, [error, logout, locale, router]);
  
  // Show loading or redirect during initial mount
  if (!mounted || !isAuthenticated) {
    return <ProfileSkeleton />;
  }
  
  if (isLoading) {
    return <ProfileSkeleton />;
  }
  
  if (error || !profile) {
    // If 401, redirect will happen in useEffect
    if ((error as any)?.response?.status === 401) {
      return <ProfileSkeleton />;
    }
    
    return (
      <div className="container-custom py-12 text-center">
        <p className="text-foreground-secondary mb-4">Ошибка загрузки профиля</p>
        <button onClick={() => window.location.reload()} className="btn-primary">
          Попробовать снова
        </button>
      </div>
    );
  }
  
  const levelInfo = calculateLevel(profile.xp);
  
  return (
    <div className="container-custom py-6">
      {/* Profile Header */}
      <div className="bg-background-secondary rounded-card p-6 mb-6">
        <div className="flex flex-col md:flex-row items-center md:items-start gap-6">
          {/* Avatar */}
          <div className="relative">
            <div className="w-24 h-24 md:w-32 md:h-32 rounded-full overflow-hidden bg-background-tertiary">
              {profile.avatarUrl ? (
                <Image
                  src={profile.avatarUrl}
                  alt={profile.displayName}
                  fill
                  className="object-cover"
                />
              ) : (
                <div className="w-full h-full flex items-center justify-center">
                  <User className="w-12 h-12 text-foreground-muted" />
                </div>
              )}
            </div>
            {/* Level Badge */}
            <div className="absolute -bottom-2 left-1/2 -translate-x-1/2 bg-accent-primary text-white text-sm font-bold px-3 py-1 rounded-full">
              LVL {profile.level}
            </div>
          </div>
          
          {/* User Info */}
          <div className="flex-1 text-center md:text-left">
            <div className="flex items-center justify-center md:justify-start gap-2 mb-1">
              <h1 className="text-2xl font-heading font-bold">{profile.displayName}</h1>
              {profile.role === 'premium' && (
                <span title="Premium">
                  <Crown className="w-5 h-5 text-accent-warning" />
                </span>
              )}
              {profile.role === 'moderator' && (
                <span className="bg-status-info/20 text-status-info text-xs px-2 py-0.5 rounded">Модератор</span>
              )}
              {profile.role === 'admin' && (
                <span className="bg-status-error/20 text-status-error text-xs px-2 py-0.5 rounded">Админ</span>
              )}
            </div>
            <p className="text-foreground-secondary text-sm mb-3">{profile.email}</p>
            
            {/* XP Progress */}
            <div className="max-w-sm mx-auto md:mx-0 mb-4">
              <div className="flex justify-between text-sm mb-1">
                <span className="text-foreground-secondary">Level {levelInfo.level}</span>
                <span className="text-foreground-muted">
                  {levelInfo.currentXP} / {levelInfo.nextLevelXP} XP
                </span>
              </div>
              <div className="h-2 bg-background-tertiary rounded-full overflow-hidden">
                <div 
                  className="h-full bg-accent-primary transition-all"
                  style={{ width: `${levelInfo.progress}%` }}
                />
              </div>
            </div>
            
            {/* Quick Stats */}
            <div className="flex flex-wrap justify-center md:justify-start gap-6 text-sm">
              <div className="flex items-center gap-1">
                <BookOpen className="w-4 h-4 text-foreground-muted" />
                <span>{profile.readChaptersCount} глав прочитано</span>
              </div>
              <div className="flex items-center gap-1">
                <Clock className="w-4 h-4 text-foreground-muted" />
                <span>{formatReadingTime(profile.readingTime)}</span>
              </div>
              <div className="flex items-center gap-1">
                <MessageSquare className="w-4 h-4 text-foreground-muted" />
                <span>{profile.commentsCount} комментариев</span>
              </div>
            </div>
          </div>
          
          {/* Actions */}
          <div className="flex gap-2">
            <Link href={`/${locale}/profile/settings`} className="btn-secondary p-3">
              <Settings className="w-5 h-5" />
            </Link>
          </div>
        </div>
      </div>
      
      {/* Tabs */}
      <div className="flex gap-1 mb-6 overflow-x-auto">
        {(['stats', 'bookmarks', 'achievements'] as const).map(tab => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 text-sm font-medium rounded-lg whitespace-nowrap transition-colors ${
              activeTab === tab
                ? 'bg-accent-primary text-white'
                : 'bg-background-secondary text-foreground-secondary hover:text-foreground'
            }`}
          >
            {tab === 'stats' && 'Статистика'}
            {tab === 'bookmarks' && `Закладки (${profile.bookmarksCount})`}
            {tab === 'achievements' && 'Достижения'}
          </button>
        ))}
      </div>
      
      {/* Tab Content */}
      <div>
        {/* Stats Tab */}
        {activeTab === 'stats' && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <StatCard
              icon={<BookOpen className="w-6 h-6" />}
              label="Прочитано глав"
              value={profile.readChaptersCount}
              color="primary"
            />
            <StatCard
              icon={<Clock className="w-6 h-6" />}
              label="Время чтения"
              value={formatReadingTime(profile.readingTime)}
              color="secondary"
            />
            <StatCard
              icon={<MessageSquare className="w-6 h-6" />}
              label="Комментариев"
              value={profile.commentsCount}
              color="info"
            />
            <StatCard
              icon={<Bookmark className="w-6 h-6" />}
              label="В закладках"
              value={profile.bookmarksCount}
              color="warning"
            />
            <StatCard
              icon={<TrendingUp className="w-6 h-6" />}
              label="Уровень"
              value={`Level ${profile.level}`}
              subValue={`${profile.xp} XP`}
              color="success"
            />
            <StatCard
              icon={<Award className="w-6 h-6" />}
              label="Достижения"
              value={0}
              subValue="Скоро"
              color="muted"
            />
          </div>
        )}
        
        {/* Bookmarks Tab */}
        {activeTab === 'bookmarks' && (
          <div>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold">Мои закладки</h2>
              <Link 
                href={`/${locale}/bookmarks`}
                className="text-accent-primary text-sm hover:underline flex items-center"
              >
                Смотреть все <ChevronRight className="w-4 h-4" />
              </Link>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <BookmarkListCard code="reading" count={0} locale={locale} />
              <BookmarkListCard code="planned" count={0} locale={locale} />
              <BookmarkListCard code="completed" count={0} locale={locale} />
              <BookmarkListCard code="favorites" count={0} locale={locale} />
            </div>
          </div>
        )}
        
        {/* Achievements Tab */}
        {activeTab === 'achievements' && (
          <div className="text-center py-12">
            <Award className="w-16 h-16 mx-auto mb-4 text-foreground-muted" />
            <h2 className="text-lg font-semibold mb-2">Достижения</h2>
            <p className="text-foreground-secondary">
              Система достижений будет доступна в следующем обновлении
            </p>
          </div>
        )}
      </div>
    </div>
  );
}

// Stat Card Component
interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  subValue?: string;
  color: 'primary' | 'secondary' | 'info' | 'warning' | 'success' | 'muted';
}

function StatCard({ icon, label, value, subValue, color }: StatCardProps) {
  const colorClasses = {
    primary: 'text-accent-primary bg-accent-primary/10',
    secondary: 'text-accent-secondary bg-accent-secondary/10',
    info: 'text-status-info bg-status-info/10',
    warning: 'text-accent-warning bg-accent-warning/10',
    success: 'text-status-success bg-status-success/10',
    muted: 'text-foreground-muted bg-background-tertiary',
  };
  
  return (
    <div className="bg-background-secondary rounded-card p-4">
      <div className={`w-12 h-12 rounded-lg flex items-center justify-center mb-3 ${colorClasses[color]}`}>
        {icon}
      </div>
      <div className="text-foreground-muted text-sm mb-1">{label}</div>
      <div className="text-xl font-bold">{value}</div>
      {subValue && <div className="text-sm text-foreground-muted">{subValue}</div>}
    </div>
  );
}

// Bookmark List Card
interface BookmarkListCardProps {
  code: string;
  count: number;
  locale: string;
}

function BookmarkListCard({ code, count, locale }: BookmarkListCardProps) {
  const labels: Record<string, string> = {
    reading: 'Читаю',
    planned: 'В планах',
    completed: 'Прочитано',
    favorites: 'Избранное',
    dropped: 'Брошено',
  };
  
  const icons: Record<string, React.ReactNode> = {
    reading: <BookOpen className="w-6 h-6" />,
    planned: <Clock className="w-6 h-6" />,
    completed: <Award className="w-6 h-6" />,
    favorites: <Crown className="w-6 h-6" />,
    dropped: <MessageSquare className="w-6 h-6" />,
  };
  
  return (
    <Link
      href={`/${locale}/bookmarks?list=${code}`}
      className="bg-background-secondary rounded-card p-4 hover:bg-background-hover transition-colors"
    >
      <div className="flex items-center gap-3">
        <div className="text-foreground-muted">{icons[code]}</div>
        <div>
          <div className="font-medium">{labels[code]}</div>
          <div className="text-sm text-foreground-muted">{count} книг</div>
        </div>
      </div>
    </Link>
  );
}

// Skeleton
function ProfileSkeleton() {
  return (
    <div className="container-custom py-6 animate-pulse">
      <div className="bg-background-secondary rounded-card p-6 mb-6">
        <div className="flex flex-col md:flex-row items-center md:items-start gap-6">
          <div className="w-24 h-24 md:w-32 md:h-32 rounded-full bg-background-hover" />
          <div className="flex-1 space-y-3">
            <div className="h-6 bg-background-hover rounded w-48 mx-auto md:mx-0" />
            <div className="h-4 bg-background-hover rounded w-32 mx-auto md:mx-0" />
            <div className="h-2 bg-background-hover rounded w-full max-w-sm mx-auto md:mx-0" />
          </div>
        </div>
      </div>
      <div className="flex gap-2 mb-6">
        {[1, 2, 3].map(i => (
          <div key={i} className="h-10 w-24 bg-background-hover rounded-lg" />
        ))}
      </div>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {[1, 2, 3, 4, 5, 6].map(i => (
          <div key={i} className="h-32 bg-background-secondary rounded-card" />
        ))}
      </div>
    </div>
  );
}

// Helper function
function formatReadingTime(minutes: number): string {
  if (minutes < 60) return `${minutes} мин`;
  const hours = Math.floor(minutes / 60);
  const mins = minutes % 60;
  if (hours < 24) return `${hours}ч ${mins}м`;
  const days = Math.floor(hours / 24);
  const remainingHours = hours % 24;
  return `${days}д ${remainingHours}ч`;
}
