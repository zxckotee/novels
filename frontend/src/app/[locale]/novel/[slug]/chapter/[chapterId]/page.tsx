'use client';

import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import Link from 'next/link';
import { useTranslations } from 'next-intl';
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
import { useSaveProgress, useChapters } from '@/lib/api/hooks/useChapters';
import { useAuthStore } from '@/store/auth';
import { CommentList } from '@/components/Comments/CommentList';
import { useSearchParams } from 'next/navigation';
import { useQueryClient } from '@tanstack/react-query';
import api from '@/lib/api/client';
import type { ChapterWithContent } from '@/lib/api/types';

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
  const searchParams = useSearchParams();
  const commentAnchor = searchParams.get('comment_anchor');
  const queryClient = useQueryClient();

  // Temporarily disable per-paragraph comments UI (💬 + modal) until we need it again.
  const ENABLE_PARAGRAPH_COMMENTS = false;
  
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
  
  const { mutate: saveProgress } = useSaveProgress();
  const { data: tocData } = useChapters(slug, 1, 200);
  const tocChapters = tocData?.data || [];

  // Infinite reader state
  const [loaded, setLoaded] = useState<ChapterWithContent[]>([]);
  const [activeChapterId, setActiveChapterId] = useState<string>(chapterId);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);
  const [loadingNext, setLoadingNext] = useState(false);
  const [loadingPrev, setLoadingPrev] = useState(false);
  const [suspendAutoLoad, setSuspendAutoLoad] = useState(true);
  const didInitialJumpRef = useRef<string>('');
  const suspendAutoLoadRef = useRef<boolean>(true);
  const topObserverRef = useRef<IntersectionObserver | null>(null);
  const bottomObserverRef = useRef<IntersectionObserver | null>(null);

  const bottomSentinelRef = useRef<HTMLDivElement | null>(null);
  const topSentinelRef = useRef<HTMLDivElement | null>(null);
  const sectionTopRefs = useRef<Map<string, HTMLDivElement>>(new Map());

  // Backend shape for /chapters/:id
  type BackendChapterWithContent = {
    id: string;
    novel_id: string;
    number: number;
    slug?: string | null;
    title?: string | null;
    views: number;
    published_at?: string | null;
    created_at: string;
    updated_at: string;

    content: string;
    word_count: number;
    reading_time_minutes?: number;
    source?: string;

    prev_chapter?: { id: string; number: number; title?: string | null } | null;
    next_chapter?: { id: string; number: number; title?: string | null } | null;
    novel_slug: string;
    novel_title: string;
  };

  const mapBackendChapterWithContent = useCallback((ch: BackendChapterWithContent): ChapterWithContent => {
    return {
      id: ch.id,
      novelId: ch.novel_id,
      number: ch.number,
      slug: ch.slug ?? undefined,
      title: ch.title ?? `Глава ${ch.number}`,
      wordCount: ch.word_count ?? 0,
      publishedAt: ch.published_at ?? ch.created_at,
      createdAt: ch.created_at,
      updatedAt: ch.updated_at,
      content: ch.content,
      novel: {
        id: ch.novel_id,
        slug: ch.novel_slug,
        title: ch.novel_title,
      },
      prevChapter: ch.prev_chapter
        ? { id: ch.prev_chapter.id, number: ch.prev_chapter.number, title: ch.prev_chapter.title ?? `Глава ${ch.prev_chapter.number}` }
        : undefined,
      nextChapter: ch.next_chapter
        ? { id: ch.next_chapter.id, number: ch.next_chapter.number, title: ch.next_chapter.title ?? `Глава ${ch.next_chapter.number}` }
        : undefined,
    };
  }, []);

  const fetchChapter = useCallback(
    async (id: string): Promise<ChapterWithContent> => {
      const lang = locale || 'ru';
      return await queryClient.fetchQuery({
        queryKey: ['chapter', id, lang],
        queryFn: async () => {
          const res = await api.get<BackendChapterWithContent>(`/chapters/${id}?lang=${lang}`);
          return mapBackendChapterWithContent(res.data);
        },
        staleTime: 10 * 60 * 1000,
      });
    },
    [locale, mapBackendChapterWithContent, queryClient]
  );

  // Reset and load initial chapter when route param changes
  useEffect(() => {
    let cancelled = false;
    setIsLoading(true);
    setError(null);
    setLoaded([]);
    setActiveChapterId(chapterId);
    setSuspendAutoLoad(true);
    suspendAutoLoadRef.current = true;
    didInitialJumpRef.current = '';
    // Immediately stop any previous observers to avoid "scroll fighting" during navigation.
    topObserverRef.current?.disconnect();
    bottomObserverRef.current?.disconnect();
    topObserverRef.current = null;
    bottomObserverRef.current = null;

    (async () => {
      try {
        const ch = await fetchChapter(chapterId);
        if (!cancelled) setLoaded([ch]);
      } catch (e) {
        if (!cancelled) setError(e);
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [chapterId, fetchChapter]);

  // Prefer manual scroll restoration to avoid browser trying to restore previous scroll position between chapters.
  useEffect(() => {
    if (typeof window === 'undefined') return;
    if (!('scrollRestoration' in window.history)) return;
    const prev = window.history.scrollRestoration;
    window.history.scrollRestoration = 'manual';
    return () => {
      window.history.scrollRestoration = prev;
    };
  }, []);

  // Keep ref in sync for observer callbacks (so old observers can't keep loading).
  useEffect(() => {
    suspendAutoLoadRef.current = suspendAutoLoad;
  }, [suspendAutoLoad]);

  // When navigating to a specific chapter route (e.g. from chapter list),
  // jump immediately to the top to avoid "scroll fighting" from the infinite reader logic.
  useLayoutEffect(() => {
    if (typeof window === 'undefined') return;
    window.scrollTo({ top: 0, left: 0, behavior: 'auto' });
  }, [chapterId]);

  const activeChapter = useMemo(() => {
    return loaded.find((c) => c.id === activeChapterId) ?? loaded[0];
  }, [activeChapterId, loaded]);

  const scrollToChapter = useCallback((id: string) => {
    const el = sectionTopRefs.current.get(id);
    if (el) {
      el.scrollIntoView({ behavior: 'auto', block: 'start' });
      return true;
    }
    return false;
  }, []);

  // After initial chapter is loaded and rendered, ensure we're positioned exactly at it,
  // then re-enable autoload observers.
  useEffect(() => {
    if (isLoading) return;
    if (!loaded.length) return;
    if (didInitialJumpRef.current === chapterId) return;
    didInitialJumpRef.current = chapterId;
    const id = chapterId;
    // wait a frame for refs to be registered
    requestAnimationFrame(() => {
      scrollToChapter(id);
      setSuspendAutoLoad(false);
      suspendAutoLoadRef.current = false;
    });
  }, [chapterId, isLoading, loaded.length, scrollToChapter]);
  
  // Save progress when chapter loads
  useEffect(() => {
    if (activeChapter && isAuthenticated) {
      saveProgress({
        chapterId: activeChapter.id,
        position: 0,
      });
    }
  }, [activeChapter?.id, isAuthenticated, saveProgress]);
  
  // Save settings to localStorage
  useEffect(() => {
    if (typeof window !== 'undefined') {
      localStorage.setItem('reader-settings', JSON.stringify(settings));
    }
  }, [settings]);
  
  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'ArrowLeft' && activeChapter?.prevChapter) {
        const id = activeChapter.prevChapter.id;
        const el = sectionTopRefs.current.get(id);
        if (el) el.scrollIntoView({ behavior: 'auto', block: 'start' });
      } else if (e.key === 'ArrowRight' && activeChapter?.nextChapter) {
        const id = activeChapter.nextChapter.id;
        const el = sectionTopRefs.current.get(id);
        if (el) el.scrollIntoView({ behavior: 'auto', block: 'start' });
      }
    };
    
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [activeChapter?.nextChapter, activeChapter?.prevChapter]);

  const loadNextChapter = useCallback(async () => {
    const last = loaded[loaded.length - 1];
    const nextId = last?.nextChapter?.id;
    if (!nextId) return;
    if (loaded.some((c) => c.id === nextId)) return;
    if (loadingNext) return;
    try {
      setLoadingNext(true);
      const ch = await fetchChapter(nextId);
      setLoaded((prev) => [...prev, ch]);
    } finally {
      setLoadingNext(false);
    }
  }, [fetchChapter, loaded, loadingNext]);

  const loadPrevChapter = useCallback(async () => {
    const first = loaded[0];
    const prevId = first?.prevChapter?.id;
    if (!prevId) return;
    if (loaded.some((c) => c.id === prevId)) return;
    if (loadingPrev) return;

    const beforeHeight = document.body.scrollHeight;
    const beforeY = window.scrollY;
    try {
      setLoadingPrev(true);
      const ch = await fetchChapter(prevId);
      setLoaded((prev) => [ch, ...prev]);
      // Compensate scroll so content doesn't "jump" when prepending
      requestAnimationFrame(() => {
        const afterHeight = document.body.scrollHeight;
        const delta = afterHeight - beforeHeight;
        window.scrollTo({ top: beforeY + delta, behavior: 'auto' });
      });
    } finally {
      setLoadingPrev(false);
    }
  }, [fetchChapter, loaded, loadingPrev]);

  const goPrev = useCallback(async () => {
    const prevId = activeChapter?.prevChapter?.id;
    if (!prevId) return;
    if (!scrollToChapter(prevId)) {
      await loadPrevChapter();
      requestAnimationFrame(() => {
        scrollToChapter(prevId);
      });
    }
  }, [activeChapter?.prevChapter?.id, loadPrevChapter, scrollToChapter]);

  const goNext = useCallback(async () => {
    const nextId = activeChapter?.nextChapter?.id;
    if (!nextId) return;
    if (!scrollToChapter(nextId)) {
      await loadNextChapter();
      requestAnimationFrame(() => {
        scrollToChapter(nextId);
      });
    }
  }, [activeChapter?.nextChapter?.id, loadNextChapter, scrollToChapter]);

  // Observe bottom/top sentinels to auto-load next/prev
  useEffect(() => {
    if (!bottomSentinelRef.current || !topSentinelRef.current) return;

    const bottomObserver = new IntersectionObserver(
      (entries) => {
        if (suspendAutoLoadRef.current) return;
        const e = entries[0];
        if (e.isIntersecting) void loadNextChapter();
      },
      { root: null, rootMargin: '800px 0px 800px 0px', threshold: 0 }
    );

    const topObserver = new IntersectionObserver(
      (entries) => {
        if (suspendAutoLoadRef.current) return;
        const e = entries[0];
        if (e.isIntersecting) void loadPrevChapter();
      },
      { root: null, rootMargin: '800px 0px 800px 0px', threshold: 0 }
    );

    bottomObserver.observe(bottomSentinelRef.current);
    topObserver.observe(topSentinelRef.current);
    bottomObserverRef.current = bottomObserver;
    topObserverRef.current = topObserver;

    return () => {
      bottomObserver.disconnect();
      topObserver.disconnect();
      if (bottomObserverRef.current === bottomObserver) bottomObserverRef.current = null;
      if (topObserverRef.current === topObserver) topObserverRef.current = null;
    };
  }, [loadNextChapter, loadPrevChapter]);

  // Observe chapter section tops to drive sticky header + URL (without Next navigation)
  useEffect(() => {
    if (loaded.length === 0) return;
    const els = Array.from(sectionTopRefs.current.entries());
    if (els.length === 0) return;

    const io = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((e) => e.isIntersecting)
          .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top);
        const first = visible[0];
        const id = first?.target?.getAttribute('data-chapter-id');
        if (id) {
          setActiveChapterId(id);
        }
      },
      { root: null, rootMargin: '-40% 0px -55% 0px', threshold: 0 }
    );

    for (const [, el] of els) io.observe(el);
    return () => io.disconnect();
  }, [loaded, locale, slug]);
  
  // Loading state
  if (isLoading) {
    return <ReaderSkeleton />;
  }
  
  // Error state
  if (error || loaded.length === 0) {
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
    <div className="min-h-screen bg-background-primary pt-14">
      {/* Top Navigation Bar */}
      <header className="fixed top-0 left-0 right-0 z-30 bg-background-secondary/95 backdrop-blur-sm border-b border-background-tertiary">
        <div className="container-custom flex items-center justify-between h-14">
          {/* Back to novel */}
          <Link
            href={`/${locale}/novel/${slug}`}
            className="flex items-center gap-2 text-foreground-secondary hover:text-foreground transition-colors"
          >
            <BookOpen className="w-5 h-5" />
            <span className="hidden sm:inline truncate max-w-[200px]">
              {activeChapter?.novel.title ?? loaded[0]?.novel.title}
            </span>
          </Link>
          
          {/* Chapter info */}
          <div className="text-center">
            <div className="text-sm font-medium">
              Глава {activeChapter?.number ?? loaded[0]?.number}
            </div>
            <div className="text-xs text-foreground-muted truncate max-w-[150px] sm:max-w-[300px]">
              {activeChapter?.title ?? loaded[0]?.title}
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

      {/* Visual badge: always shows current reading position */}
      <div className="fixed top-16 left-1/2 -translate-x-1/2 z-40 pointer-events-none">
        <div className="px-3 py-1.5 rounded-full bg-background-secondary/90 backdrop-blur-sm border border-background-tertiary shadow-card-hover max-w-[92vw]">
          <div className="text-xs text-foreground-muted text-center truncate">
            <span className="font-medium text-foreground">Глава {activeChapter?.number ?? loaded[0]?.number}</span>
            <span className="mx-2 text-foreground-muted">•</span>
            <span className="truncate">{activeChapter?.title ?? loaded[0]?.title}</span>
          </div>
        </div>
      </div>
      
      {/* Chapter Content */}
      <main 
        className={`container-custom py-8 pb-24 md:pb-8 ${MAX_WIDTH_CLASSES[settings.maxWidth]} mx-auto`}
      >
        {/* Top sentinel (for loading previous chapters) */}
        <div ref={topSentinelRef} />

        {/* Webtoon-style stream */}
        <div className="space-y-16">
          {loaded.map((ch, idx) => (
            <section key={ch.id} className="scroll-mt-16">
              <div
                data-chapter-id={ch.id}
                ref={(el) => {
                  if (!el) return;
                  sectionTopRefs.current.set(ch.id, el);
                }}
              />

              {/* Chapter separator (helps user notice a new chapter) */}
              {idx > 0 && (
                <div className="flex items-center gap-3 -mt-4">
                  <div className="h-px flex-1 bg-background-tertiary" />
                  <span className="text-xs text-foreground-muted uppercase tracking-wider">
                    Новая глава
                  </span>
                  <div className="h-px flex-1 bg-background-tertiary" />
                </div>
              )}

        {/* Chapter Title */}
              <h2 className="text-2xl font-heading font-bold mb-8 text-center">
                Глава {ch.number}: {ch.title}
              </h2>
        
        {/* Content */}
        <article 
          className={`${FONT_FAMILY_CLASSES[settings.fontFamily]} text-foreground`}
          style={{
            fontSize: `${settings.fontSize}px`,
            lineHeight: settings.lineHeight,
          }}
        >
          <div 
            className="reader-content prose prose-invert max-w-none"
                  dangerouslySetInnerHTML={{ __html: formatContent(ch.content, { enableParagraphComments: ENABLE_PARAGRAPH_COMMENTS }) }}
          />
        </article>
            </section>
          ))}
        </div>

        {/* Bottom sentinel (for loading next chapters) */}
        <div ref={bottomSentinelRef} className="h-16" />

        {(loadingNext || loadingPrev) && (
          <div className="mt-8 text-center text-foreground-muted text-sm">
            Загрузка…
          </div>
        )}
        
        {/* Comments Section */}
        <div id="comments" className="mt-12 pt-8 border-t border-background-tertiary">
          <div className="flex items-center gap-2 mb-4">
            <MessageSquare className="w-5 h-5" />
            <h2 className="text-xl font-semibold">Комментарии</h2>
          </div>
          <CommentList targetType="chapter" targetId={(activeChapter?.id ?? loaded[0]?.id) as string} locale={locale} />
        </div>
      </main>

      {/* Per-paragraph comments modal (disabled for now) */}
      {ENABLE_PARAGRAPH_COMMENTS && commentAnchor && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/60" onClick={() => {
            const url = new URL(window.location.href);
            url.searchParams.delete('comment_anchor');
            window.history.replaceState({}, '', url.toString());
          }} />
          <div className="relative bg-background-secondary rounded-card w-full max-w-3xl max-h-[80vh] overflow-y-auto p-4">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-lg font-semibold">Комментарии к абзацу</h3>
              <button
                className="btn-ghost p-2"
                onClick={() => {
                  const url = new URL(window.location.href);
                  url.searchParams.delete('comment_anchor');
                  window.history.replaceState({}, '', url.toString());
                }}
                aria-label="Закрыть"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
            <CommentList
              targetType="chapter"
              targetId={(activeChapter?.id ?? loaded[0]?.id) as string}
              locale={locale}
              anchor={commentAnchor}
            />
          </div>
        </div>
      )}
      
      {/* Fixed Bottom Navigation (Mobile) */}
      <nav className="fixed bottom-0 left-0 right-0 bg-background-secondary/95 backdrop-blur-sm border-t border-background-tertiary md:hidden z-30">
        <div className="flex items-center justify-between h-14 px-4">
          <button
            type="button"
            onClick={goPrev}
            disabled={!activeChapter?.prevChapter}
            className={`flex items-center gap-1 ${activeChapter?.prevChapter ? '' : 'opacity-50'}`}
          >
            <ChevronLeft className="w-6 h-6" />
            <span>Назад</span>
          </button>
          
          <button
            onClick={() => setShowTOC(true)}
            className="flex items-center gap-1"
          >
            <List className="w-5 h-5" />
            <span>Главы</span>
          </button>
          
          <button
            type="button"
            onClick={goNext}
            disabled={!activeChapter?.nextChapter}
            className={`flex items-center gap-1 ${activeChapter?.nextChapter ? '' : 'opacity-50'}`}
          >
            <span>Далее</span>
            <ChevronRight className="w-6 h-6" />
          </button>
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
            {tocChapters.length === 0 ? (
              <p className="text-foreground-muted text-center py-8">Главы не найдены</p>
            ) : (
              <div className="space-y-1">
                {tocChapters.map((ch) => {
                  const isActive = ch.id === activeChapterId;
                  return (
                    <Link
                      key={ch.id}
                      href={`/${locale}/novel/${slug}/chapter/${ch.id}`}
                      onClick={() => setShowTOC(false)}
                      className={`block px-3 py-2 rounded transition-colors ${
                        isActive ? 'bg-accent-primary text-white' : 'hover:bg-background-hover'
                      }`}
                    >
                      <div className="text-sm font-medium">Глава {ch.number}</div>
                      {ch.title && <div className="text-xs opacity-80 line-clamp-1">{ch.title}</div>}
                    </Link>
                  );
                })}
              </div>
            )}
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
function formatContent(content: string, opts?: { enableParagraphComments?: boolean }): string {
  if (!content) return '';
  const enableParagraphComments = opts?.enableParagraphComments ?? false;
  
  // Split by double newlines or single newlines and wrap in paragraphs
  let paragraphIndex = 0;
  return content
    .split(/\n\n+/)
    .map(paragraph => paragraph.trim())
    .filter(Boolean)
    .map(paragraph => {
      const anchor = `p:${paragraphIndex++}`;
      const link = enableParagraphComments
        ? `<a href="?comment_anchor=${encodeURIComponent(anchor)}#comments" class="ml-2 text-sm text-accent-primary hover:underline" title="Комментарии к абзацу">💬</a>`
        : '';
      return `<p id="${anchor.replace(':', '-')}" class="mb-4">${paragraph.replace(/\n/g, '<br />')}${link}</p>`;
    })
    .join('');
}
