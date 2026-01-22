'use client';

import { useState } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { useTranslations } from 'next-intl';
import { 
  Star, 
  BookOpen, 
  Bookmark, 
  Eye, 
  Clock, 
  User, 
  Calendar,
  ChevronRight,
  MessageSquare,
  ThumbsUp,
  Edit,
  Share2
} from 'lucide-react';
import { useNovel } from '@/lib/api/hooks/useNovels';
import { useChapters } from '@/lib/api/hooks/useChapters';
import { useAuthStore } from '@/store/auth';
import { CommentList } from '@/components/Comments/CommentList';

interface NovelPageClientProps {
  slug: string;
  locale: string;
}

// Status badge styles
const STATUS_STYLES = {
  ongoing: 'bg-status-success text-white',
  completed: 'bg-accent-primary text-white',
  paused: 'bg-accent-warning text-black',
  dropped: 'bg-status-error text-white',
} as const;

const STATUS_LABELS = {
  ongoing: 'Продолжается',
  completed: 'Завершен',
  paused: 'Перерыв',
  dropped: 'Брошен',
} as const;

export default function NovelPageClient({ slug, locale }: NovelPageClientProps) {
  const t = useTranslations('novel');
  const [activeTab, setActiveTab] = useState<'description' | 'chapters' | 'comments'>('description');
  const { isAuthenticated } = useAuthStore();
  
  const { data: novel, isLoading, error } = useNovel(slug, locale);
  const { data: chaptersData } = useChapters(slug);
  
  const chapters = chaptersData?.data || [];
  const chaptersCount = chapters.length;
  const firstChapterId = chapters[0]?.id;
  
  if (isLoading) {
    return <NovelPageSkeleton />;
  }
  
  if (error || !novel) {
    return (
      <div className="container-custom py-12 text-center">
        <h1 className="text-2xl font-bold mb-4">Новелла не найдена</h1>
        <Link href={`/${locale}/catalog`} className="btn-primary">
          Перейти в каталог
        </Link>
      </div>
    );
  }
  
  return (
    <div>
      {/* Hero Section with blurred background */}
      <div className="relative">
        {/* Background */}
        <div className="absolute inset-0 h-[400px] overflow-hidden">
          {novel.coverUrl && (
            <Image
              src={novel.coverUrl}
              alt=""
              fill
              sizes="100vw"
              className="object-cover blur-xl scale-110 opacity-30"
            />
          )}
          <div className="absolute inset-0 bg-gradient-to-b from-transparent via-background-primary/80 to-background-primary" />
        </div>
        
        {/* Content */}
        <div className="container-custom relative pt-8 pb-6">
          <div className="flex flex-col md:flex-row gap-6">
            {/* Cover */}
            <div className="shrink-0 mx-auto md:mx-0">
              <div className="relative w-[200px] aspect-cover rounded-card overflow-hidden shadow-card-hover">
                {novel.coverUrl ? (
                  <Image
                    src={novel.coverUrl}
                    alt={novel.title}
                    fill
                    sizes="200px"
                    className="object-cover"
                    priority
                  />
                ) : (
                  <div className="w-full h-full bg-background-tertiary flex items-center justify-center">
                    <BookOpen className="w-12 h-12 text-foreground-muted" />
                  </div>
                )}
                
                {/* Status Badge */}
                <div className={`absolute top-2 left-2 px-2 py-1 text-xs font-medium rounded ${STATUS_STYLES[novel.translationStatus]}`}>
                  {STATUS_LABELS[novel.translationStatus]}
                </div>
              </div>
            </div>
            
            {/* Info */}
            <div className="flex-1 text-center md:text-left">
              {/* Title */}
              <h1 className="text-2xl md:text-3xl lg:text-4xl font-heading font-bold mb-2">
                {novel.title}
              </h1>
              
              {/* Alt titles */}
              {novel.altTitles && novel.altTitles.length > 0 && (
                <p className="text-foreground-secondary text-sm mb-4">
                  {novel.altTitles.join(' / ')}
                </p>
              )}
              
              {/* Meta info */}
              <div className="flex flex-wrap justify-center md:justify-start gap-4 text-sm text-foreground-secondary mb-4">
                {novel.author && (
                  <span className="flex items-center gap-1">
                    <User className="w-4 h-4" />
                    {novel.author.name}
                  </span>
                )}
                {novel.releaseYear && (
                  <span className="flex items-center gap-1">
                    <Calendar className="w-4 h-4" />
                    {novel.releaseYear}
                  </span>
                )}
                <span className="flex items-center gap-1">
                  <BookOpen className="w-4 h-4" />
                  {chaptersCount} / {novel.originalChaptersCount} глав
                </span>
              </div>
              
              {/* Stats */}
              <div className="flex flex-wrap justify-center md:justify-start gap-6 mb-6">
                {/* Rating */}
                <div className="flex items-center gap-2">
                  <div className="flex items-center gap-1 text-accent-warning">
                    <Star className="w-5 h-5 fill-current" />
                    <span className="font-bold text-lg">{novel.rating.toFixed(1)}</span>
                  </div>
                  <span className="text-sm text-foreground-secondary">
                    ({novel.ratingsCount} оценок)
                  </span>
                </div>
                
                {/* Views */}
                <div className="flex items-center gap-1 text-foreground-secondary">
                  <Eye className="w-5 h-5" />
                  <span>{formatNumber(novel.viewsCount)}</span>
                </div>
                
                {/* Bookmarks */}
                <div className="flex items-center gap-1 text-foreground-secondary">
                  <Bookmark className="w-5 h-5" />
                  <span>{formatNumber(novel.bookmarksCount)}</span>
                </div>
              </div>
              
              {/* Genres */}
              {novel.genres && novel.genres.length > 0 && (
                <div className="flex flex-wrap justify-center md:justify-start gap-2 mb-6">
                  {novel.genres.map(genre => (
                    <Link
                      key={genre.id}
                      href={`/${locale}/catalog?genres=${genre.slug}`}
                      className="px-3 py-1 bg-accent-primary/20 text-accent-primary rounded-tag text-sm hover:bg-accent-primary/30 transition-colors"
                    >
                      {genre.name}
                    </Link>
                  ))}
                </div>
              )}
              
              {/* Action Buttons */}
              <div className="flex flex-wrap justify-center md:justify-start gap-3">
                <Link
                  href={firstChapterId ? `/${locale}/novel/${slug}/chapter/${firstChapterId}` : `/${locale}/novel/${slug}`}
                  className={`btn-primary text-base px-6 py-3 flex items-center gap-2 ${firstChapterId ? '' : 'opacity-50 pointer-events-none'}`}
                >
                  <BookOpen className="w-5 h-5" />
                  Читать
                </Link>
                
                <button className="btn-secondary text-base px-6 py-3 flex items-center gap-2">
                  <Bookmark className="w-5 h-5" />
                  В закладки
                </button>
                
                {isAuthenticated && (
                  <Link
                    href={`/${locale}/voting`}
                    className="btn-ghost text-base px-4 py-3 flex items-center gap-2"
                  >
                    <ThumbsUp className="w-5 h-5" />
                    Голосовать
                  </Link>
                )}
                
                <button className="btn-ghost text-base p-3">
                  <Share2 className="w-5 h-5" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      {/* Tabs */}
      <div className="sticky top-16 bg-background-primary border-b border-background-tertiary z-20">
        <div className="container-custom">
          <div className="flex gap-1">
            {(['description', 'chapters', 'comments'] as const).map(tab => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={`px-4 py-3 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === tab
                    ? 'border-accent-primary text-accent-primary'
                    : 'border-transparent text-foreground-secondary hover:text-foreground'
                }`}
              >
                {tab === 'description' && 'Описание'}
                {tab === 'chapters' && `Главы (${chaptersCount})`}
                {tab === 'comments' && 'Комментарии'}
              </button>
            ))}
          </div>
        </div>
      </div>
      
      {/* Tab Content */}
      <div className="container-custom py-6">
        {/* Description Tab */}
        {activeTab === 'description' && (
          <div className="max-w-3xl">
            {/* Description */}
            <div className="prose prose-invert max-w-none mb-8">
              <h2 className="text-xl font-semibold mb-4">Описание</h2>
              <p className="whitespace-pre-line text-foreground-secondary">
                {novel.description || 'Описание отсутствует'}
              </p>
            </div>
            
            {/* Tags */}
            {novel.tags && novel.tags.length > 0 && (
              <div className="mb-8">
                <h3 className="text-lg font-semibold mb-3">Теги</h3>
                <div className="flex flex-wrap gap-2">
                  {novel.tags.map(tag => (
                    <Link
                      key={tag.id}
                      href={`/${locale}/catalog?tags=${tag.slug}`}
                      className="px-2 py-1 bg-background-secondary text-foreground-secondary rounded-tag text-sm hover:bg-background-hover transition-colors"
                    >
                      {tag.name}
                    </Link>
                  ))}
                </div>
              </div>
            )}
            
            {/* Details Table */}
            <div className="bg-background-secondary rounded-card p-4">
              <h3 className="text-lg font-semibold mb-3">Информация</h3>
              <dl className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm">
                <div className="flex justify-between sm:block">
                  <dt className="text-foreground-muted">Автор</dt>
                  <dd className="font-medium">{novel.author?.name || '-'}</dd>
                </div>
                <div className="flex justify-between sm:block">
                  <dt className="text-foreground-muted">Год выпуска</dt>
                  <dd className="font-medium">{novel.releaseYear || '-'}</dd>
                </div>
                <div className="flex justify-between sm:block">
                  <dt className="text-foreground-muted">Глав в оригинале</dt>
                  <dd className="font-medium">{novel.originalChaptersCount}</dd>
                </div>
                <div className="flex justify-between sm:block">
                  <dt className="text-foreground-muted">Переведено глав</dt>
                  <dd className="font-medium">{novel.chaptersCount}</dd>
                </div>
                <div className="flex justify-between sm:block">
                  <dt className="text-foreground-muted">Статус перевода</dt>
                  <dd className="font-medium">{STATUS_LABELS[novel.translationStatus]}</dd>
                </div>
              </dl>
            </div>
          </div>
        )}
        
        {/* Chapters Tab */}
        {activeTab === 'chapters' && (
          <div>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-xl font-semibold">Оглавление</h2>
              {/* Sort toggle could go here */}
            </div>
            
            {chapters.length === 0 ? (
              <p className="text-foreground-secondary">Главы еще не добавлены</p>
            ) : (
              <div className="space-y-1">
                {chapters.map(chapter => (
                  <Link
                    key={chapter.id}
                    href={`/${locale}/novel/${slug}/chapter/${chapter.id}`}
                    className="flex items-center justify-between p-3 bg-background-secondary rounded hover:bg-background-hover transition-colors group"
                  >
                    <div className="flex items-center gap-3">
                      <span className="text-foreground-muted text-sm w-12">
                        #{chapter.number}
                      </span>
                      <span className="group-hover:text-accent-primary transition-colors">
                        {chapter.title}
                      </span>
                    </div>
                    <div className="flex items-center gap-4 text-sm text-foreground-muted">
                      <span>{chapter.publishedAt ? formatDate(chapter.publishedAt) : '—'}</span>
                      <ChevronRight className="w-4 h-4" />
                    </div>
                  </Link>
                ))}
              </div>
            )}
          </div>
        )}
        
        {/* Comments Tab */}
        {activeTab === 'comments' && (
          <div>
            <h2 className="text-xl font-semibold mb-4">Комментарии</h2>
            <CommentList targetType="novel" targetId={novel.id} locale={locale} />
          </div>
        )}
      </div>
    </div>
  );
}

// Skeleton loader
function NovelPageSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="container-custom py-8">
        <div className="flex flex-col md:flex-row gap-6">
          <div className="w-[200px] aspect-cover bg-background-hover rounded-card mx-auto md:mx-0" />
          <div className="flex-1 space-y-4">
            <div className="h-8 bg-background-hover rounded w-3/4 mx-auto md:mx-0" />
            <div className="h-4 bg-background-hover rounded w-1/2 mx-auto md:mx-0" />
            <div className="h-4 bg-background-hover rounded w-1/3 mx-auto md:mx-0" />
            <div className="flex gap-3 justify-center md:justify-start">
              <div className="h-12 w-32 bg-background-hover rounded" />
              <div className="h-12 w-32 bg-background-hover rounded" />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

// Helper functions
function formatNumber(num?: number | null): string {
  if (typeof num !== 'number' || Number.isNaN(num)) return '0';
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
  return num.toString();
}

function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('ru', { day: 'numeric', month: 'short', year: 'numeric' });
}
