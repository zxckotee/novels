'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useTranslations, useLocale } from 'next-intl';
import { 
  ChevronLeft, 
  ChevronRight, 
  List, 
  Settings,
  BookOpen,
  MessageSquare,
  X,
  Minus,
  Plus
} from 'lucide-react';
import { useChapter, useSaveProgress } from '@/lib/api/hooks/useChapters';
import { useAuthStore } from '@/store/auth';

interface ReaderPageProps {
  params: {
    locale: string;
    slug: string;
    chapterId: string;
  };
}

// Reader settings interface
interface ReaderSettings {
  fontSize: number;
  lineHeight: number;
  maxWidth: 'narrow' | 'medium' | 'wide' | 'full';
  fontFamily: 'sans' | 'serif' | 'mono';
}

const DEFAULT_SETTINGS: ReaderSettings = {
  fontSize: 18,
  lineHeight: 1.8,
  maxWidth: 'medium',
  fontFamily: 'sans',
};

const MAX_WIDTH_CLASSES = {
  narrow: 'max-w-xl',
  medium: 'max-w-3xl',
  wide: 'max-w-5xl',
  full: 'max-w-none',
};

const FONT_FAMILY_CLASSES = {
  sans: 'font-sans',
  serif: 'font-serif',
  mono: 'font-mono',
};

export default function ReaderPage({ params }: ReaderPageProps) {
  const { slug, chapterId, locale } = params;
  const t = useTranslations('reader');
  const { isAuthenticated } = useAuthStore();
  
  // State
  const [showTOC, setShowTOC] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [settings, setSettings] = useState<ReaderSettings>(() => {
    // Load from localStorage on client
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('reader-settings');
      if (saved) {
        try {
          return { ...DEFAULT_SETTINGS, ...JSON.parse(saved) };
        } catch {
          return DEFAULT_SETTINGS;
        }
      }
    }
    return DEFAULT_SETTINGS;
  });
  
  // Fetch chapter data
  const { data: chapter, isLoading, error } = useChapter(slug, chapterId, locale);
  const { mutate: saveProgress } = useSaveProgress();
  
  // Save progress when chapter loads
  useEffect(() => {
    if (chapter && isAuthenticated) {
      saveProgress({
        novelId: chapter.novel.id,
        chapterId: chapter.id,
      });
    }
  }, [chapter, isAuthenticated, saveProgress]);
  
  // Save settings to localStorage
  useEffect(() => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('reader-settings', JSON.stringify(settings));
    }
  }, [settings]);
  
  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'ArrowLeft' && chapter?.prevChapter) {
        window.location.href = `/${locale}/novel/${slug}/chapter/${chapter.prevChapter.number}`;
      } else if (e.key === 'ArrowRight' && chapter?.nextChapter) {
        window.location.href = `/${locale}/novel/${slug}/chapter/${chapter.nextChapter.number}`;
      }
    };
    
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [chapter, locale, slug]);
  
  // Loading state
  if (isLoading) {
    return <ReaderSkeleton />;
  }
  
  // Error state
  if (error || !chapter) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold mb-4">Глава не найдена</h1>
          <Link href={`/${locale}/novel/${slug}`} className="btn-primary">
            Вернуться к новелле
          </Link>
        </div>
      </div>
    );
  }
  
  return (
    <div className="min-h-screen bg-background-primary">
      {/* Top Navigation Bar */}
      <header className="sticky top-0 z-30 bg-background-secondary/95 backdrop-blur-sm border-b border-background-tertiary">
        <div className="container-custom flex items-center justify-between h-14">
          {/* Back to novel */}
          <Link
            href={`/${locale}/novel/${slug}`}
            className="flex items-center gap-2 text-foreground-secondary hover:text-foreground transition-colors"
          >
            <BookOpen className="w-5 h-5" />
            <span className="hidden sm:inline truncate max-w-[200px]">
              {chapter.novel.title}
            </span>
          </Link>
          
          {/* Chapter info */}
          <div className="text-center">
            <div className="text-sm font-medium">
              Глава {chapter.number}
            </div>
            <div className="text-xs text-foreground-muted truncate max-w-[150px] sm:max-w-[300px]">
              {chapter.title}
            </div>
          </div>
          
          {/* Actions */}
          <div className="flex items-center gap-2">
            <button
              onClick={() => setShowTOC(true)}
              className="btn-ghost p-2"
              title="Оглавление"
            >
              <List className="w-5 h-5" />
            </button>
            <button
              onClick={() => setShowSettings(true)}
              className="btn-ghost p-2"
              title="Настройки"
            >
              <Settings className="w-5 h-5" />
            </button>
          </div>
        </div>
      </header>
      
      {/* Chapter Content */}
      <main 
        className={`container-custom py-8 ${MAX_WIDTH_CLASSES[settings.maxWidth]} mx-auto`}
      >
        {/* Chapter Title */}
        <h1 className="text-2xl font-heading font-bold mb-8 text-center">
          Глава {chapter.number}: {chapter.title}
        </h1>
        
        {/* Content */}
        <article 
          className={`${FONT_FAMILY_CLASSES[settings.fontFamily]} text-foreground`}
          style={{
            fontSize: `${settings.fontSize}px`,
            lineHeight: settings.lineHeight,
          }}
        >
          {/* Render content - in production, use proper HTML sanitization */}
          <div 
            className="reader-content prose prose-invert max-w-none"
            dangerouslySetInnerHTML={{ __html: formatContent(chapter.content) }}
          />
        </article>
        
        {/* Chapter End Navigation */}
        <div className="mt-12 pt-8 border-t border-background-tertiary">
          <div className="flex items-center justify-between gap-4">
            {chapter.prevChapter ? (
              <Link
                href={`/${locale}/novel/${slug}/chapter/${chapter.prevChapter.number}`}
                className="btn-secondary flex items-center gap-2 flex-1 justify-center max-w-[200px]"
              >
                <ChevronLeft className="w-5 h-5" />
                <span className="hidden sm:inline">Предыдущая</span>
              </Link>
            ) : (
              <div className="flex-1 max-w-[200px]" />
            )}
            
            <Link
              href={`/${locale}/novel/${slug}`}
              className="btn-ghost flex items-center gap-2"
            >
              <List className="w-5 h-5" />
              <span className="hidden sm:inline">Оглавление</span>
            </Link>
            
            {chapter.nextChapter ? (
              <Link
                href={`/${locale}/novel/${slug}/chapter/${chapter.nextChapter.number}`}
                className="btn-primary flex items-center gap-2 flex-1 justify-center max-w-[200px]"
              >
                <span className="hidden sm:inline">Следующая</span>
                <ChevronRight className="w-5 h-5" />
              </Link>
            ) : (
              <div className="text-center flex-1 max-w-[200px]">
                <span className="text-foreground-muted text-sm">Ждите обновления</span>
              </div>
            )}
          </div>
        </div>
        
        {/* Comments Section Placeholder */}
        <div className="mt-12 pt-8 border-t border-background-tertiary">
          <div className="flex items-center gap-2 mb-4">
            <MessageSquare className="w-5 h-5" />
            <h2 className="text-xl font-semibold">Комментарии</h2>
          </div>
          <div className="text-center py-8 text-foreground-secondary">
            Комментарии будут доступны в следующем обновлении
          </div>
        </div>
      </main>
      
      {/* Fixed Bottom Navigation (Mobile) */}
      <nav className="fixed bottom-0 left-0 right-0 bg-background-secondary/95 backdrop-blur-sm border-t border-background-tertiary md:hidden z-30">
        <div className="flex items-center justify-between h-14 px-4">
          <Link
            href={chapter.prevChapter ? `/${locale}/novel/${slug}/chapter/${chapter.prevChapter.number}` : '#'}
            className={`flex items-center gap-1 ${chapter.prevChapter ? '' : 'opacity-50 pointer-events-none'}`}
          >
            <ChevronLeft className="w-6 h-6" />
            <span>Назад</span>
          </Link>
          
          <button
            onClick={() => setShowTOC(true)}
            className="flex items-center gap-1"
          >
            <List className="w-5 h-5" />
            <span>Главы</span>
          </button>
          
          <Link
            href={chapter.nextChapter ? `/${locale}/novel/${slug}/chapter/${chapter.nextChapter.number}` : '#'}
            className={`flex items-center gap-1 ${chapter.nextChapter ? '' : 'opacity-50 pointer-events-none'}`}
          >
            <span>Далее</span>
            <ChevronRight className="w-6 h-6" />
          </Link>
        </div>
      </nav>
      
      {/* TOC Sidebar */}
      {showTOC && (
        <div className="fixed inset-0 z-50">
          <div 
            className="absolute inset-0 bg-black/50"
            onClick={() => setShowTOC(false)}
          />
          <aside className="absolute right-0 top-0 bottom-0 w-80 max-w-full bg-background-secondary p-4 overflow-y-auto">
            <div className="flex items-center justify-between mb-4">
              <h2 className="font-semibold">Оглавление</h2>
              <button 
                onClick={() => setShowTOC(false)}
                className="btn-ghost p-1"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <p className="text-foreground-muted text-center py-8">
              Загрузка оглавления...
            </p>
          </aside>
        </div>
      )}
      
      {/* Settings Modal */}
      {showSettings && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div 
            className="absolute inset-0 bg-black/50"
            onClick={() => setShowSettings(false)}
          />
          <div className="relative bg-background-secondary rounded-card p-6 w-full max-w-md">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-lg font-semibold">Настройки чтения</h2>
              <button 
                onClick={() => setShowSettings(false)}
                className="btn-ghost p-1"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            
            {/* Font Size */}
            <div className="mb-6">
              <label className="block text-sm font-medium mb-2">
                Размер шрифта: {settings.fontSize}px
              </label>
              <div className="flex items-center gap-4">
                <button
                  onClick={() => setSettings(s => ({ ...s, fontSize: Math.max(12, s.fontSize - 2) }))}
                  className="btn-secondary p-2"
                >
                  <Minus className="w-4 h-4" />
                </button>
                <input
                  type="range"
                  min="12"
                  max="32"
                  value={settings.fontSize}
                  onChange={(e) => setSettings(s => ({ ...s, fontSize: Number(e.target.value) }))}
                  className="flex-1"
                />
                <button
                  onClick={() => setSettings(s => ({ ...s, fontSize: Math.min(32, s.fontSize + 2) }))}
                  className="btn-secondary p-2"
                >
                  <Plus className="w-4 h-4" />
                </button>
              </div>
            </div>
            
            {/* Line Height */}
            <div className="mb-6">
              <label className="block text-sm font-medium mb-2">
                Межстрочный интервал: {settings.lineHeight}
              </label>
              <input
                type="range"
                min="1.2"
                max="2.5"
                step="0.1"
                value={settings.lineHeight}
                onChange={(e) => setSettings(s => ({ ...s, lineHeight: Number(e.target.value) }))}
                className="w-full"
              />
            </div>
            
            {/* Max Width */}
            <div className="mb-6">
              <label className="block text-sm font-medium mb-2">
                Ширина текста
              </label>
              <div className="grid grid-cols-4 gap-2">
                {(['narrow', 'medium', 'wide', 'full'] as const).map(width => (
                  <button
                    key={width}
                    onClick={() => setSettings(s => ({ ...s, maxWidth: width }))}
                    className={`py-2 px-3 text-sm rounded ${
                      settings.maxWidth === width 
                        ? 'bg-accent-primary text-white' 
                        : 'bg-background-tertiary hover:bg-background-hover'
                    }`}
                  >
                    {width === 'narrow' && 'Узко'}
                    {width === 'medium' && 'Средне'}
                    {width === 'wide' && 'Широко'}
                    {width === 'full' && 'Полная'}
                  </button>
                ))}
              </div>
            </div>
            
            {/* Font Family */}
            <div className="mb-6">
              <label className="block text-sm font-medium mb-2">
                Шрифт
              </label>
              <div className="grid grid-cols-3 gap-2">
                {(['sans', 'serif', 'mono'] as const).map(font => (
                  <button
                    key={font}
                    onClick={() => setSettings(s => ({ ...s, fontFamily: font }))}
                    className={`py-2 px-3 text-sm rounded ${FONT_FAMILY_CLASSES[font]} ${
                      settings.fontFamily === font 
                        ? 'bg-accent-primary text-white' 
                        : 'bg-background-tertiary hover:bg-background-hover'
                    }`}
                  >
                    {font === 'sans' && 'Sans'}
                    {font === 'serif' && 'Serif'}
                    {font === 'mono' && 'Mono'}
                  </button>
                ))}
              </div>
            </div>
            
            {/* Reset Button */}
            <button
              onClick={() => setSettings(DEFAULT_SETTINGS)}
              className="btn-secondary w-full"
            >
              Сбросить настройки
            </button>
          </div>
        </div>
      )}
      
      {/* Bottom padding for mobile nav */}
      <div className="h-14 md:hidden" />
    </div>
  );
}

// Skeleton loader
function ReaderSkeleton() {
  return (
    <div className="min-h-screen bg-background-primary">
      <header className="sticky top-0 z-30 bg-background-secondary border-b border-background-tertiary">
        <div className="container-custom flex items-center justify-between h-14">
          <div className="w-32 h-4 bg-background-hover rounded animate-pulse" />
          <div className="w-24 h-4 bg-background-hover rounded animate-pulse" />
          <div className="w-20 h-4 bg-background-hover rounded animate-pulse" />
        </div>
      </header>
      <main className="container-custom py-8 max-w-3xl mx-auto">
        <div className="h-8 bg-background-hover rounded w-2/3 mx-auto mb-8 animate-pulse" />
        <div className="space-y-4">
          {Array.from({ length: 15 }).map((_, i) => (
            <div 
              key={i} 
              className="h-4 bg-background-hover rounded animate-pulse"
              style={{ width: `${70 + Math.random() * 30}%` }}
            />
          ))}
        </div>
      </main>
    </div>
  );
}

// Format content with paragraphs
function formatContent(content: string): string {
  if (!content) return '';
  
  // Split by double newlines or single newlines and wrap in paragraphs
  return content
    .split(/\n\n+/)
    .map(paragraph => paragraph.trim())
    .filter(Boolean)
    .map(paragraph => `<p class="mb-4">${paragraph.replace(/\n/g, '<br />')}</p>`)
    .join('');
}
