'use client';

import { useState, useMemo, useCallback } from 'react';
import { useSearchParams, useRouter, usePathname } from 'next/navigation';
import { useTranslations, useLocale } from 'next-intl';
import { Search, Filter, X, ChevronDown, Grid, List } from 'lucide-react';
import { NovelCard, NovelCardSkeleton } from '@/components/novel/NovelCard';
import { useNovels, useGenres, useTags } from '@/lib/api/hooks/useNovels';
import type { NovelFilters as NovelFiltersType } from '@/lib/api/types';

// Status options
const STATUS_OPTIONS = [
  { value: '', label: 'Все' },
  { value: 'ongoing', label: 'Продолжается' },
  { value: 'completed', label: 'Завершен' },
  { value: 'paused', label: 'Перерыв' },
  { value: 'dropped', label: 'Брошен' },
] as const;

// Sort options
const SORT_OPTIONS = [
  { value: 'popular', label: 'По популярности' },
  { value: 'rating', label: 'По рейтингу' },
  { value: 'views', label: 'По просмотрам' },
  { value: 'bookmarks', label: 'По закладкам' },
  { value: 'updated', label: 'По дате обновления' },
  { value: 'created', label: 'По дате добавления' },
] as const;

export default function CatalogPage() {
  const t = useTranslations('catalog');
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  
  // Local state for UI
  const [showFilters, setShowFilters] = useState(false);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [searchQuery, setSearchQuery] = useState(searchParams.get('search') || '');
  
  // Parse filters from URL
  const filters: NovelFiltersType = useMemo(() => ({
    search: searchParams.get('search') || undefined,
    genres: searchParams.get('genres')?.split(',').filter(Boolean) || undefined,
    tags: searchParams.get('tags')?.split(',').filter(Boolean) || undefined,
    status: (searchParams.get('status') as NovelFiltersType['status']) || undefined,
    sort: (searchParams.get('sort') as NovelFiltersType['sort']) || 'popular',
    order: (searchParams.get('order') as NovelFiltersType['order']) || 'desc',
    page: Number(searchParams.get('page')) || 1,
    limit: 24,
    lang: locale,
  }), [searchParams, locale]);
  
  // Fetch data
  const { data: novelsData, isLoading, error } = useNovels(filters);
  const { data: genres } = useGenres();
  const { data: tags } = useTags();
  
  // Update URL with new filters
  const updateFilters = useCallback((newFilters: Partial<NovelFiltersType>) => {
    const params = new URLSearchParams(searchParams.toString());
    
    Object.entries(newFilters).forEach(([key, value]) => {
      if (value === undefined || value === '' || (Array.isArray(value) && value.length === 0)) {
        params.delete(key);
      } else if (Array.isArray(value)) {
        params.set(key, value.join(','));
      } else {
        params.set(key, String(value));
      }
    });
    
    // Reset to page 1 when filters change (except when changing page)
    if (!('page' in newFilters)) {
      params.delete('page');
    }
    
    router.push(`${pathname}?${params.toString()}`);
  }, [router, pathname, searchParams]);
  
  // Handle search submit
  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    updateFilters({ search: searchQuery });
  };
  
  // Toggle genre filter
  const toggleGenre = (genreSlug: string) => {
    const currentGenres = filters.genres || [];
    const newGenres = currentGenres.includes(genreSlug)
      ? currentGenres.filter(g => g !== genreSlug)
      : [...currentGenres, genreSlug];
    updateFilters({ genres: newGenres });
  };
  
  // Toggle tag filter
  const toggleTag = (tagSlug: string) => {
    const currentTags = filters.tags || [];
    const newTags = currentTags.includes(tagSlug)
      ? currentTags.filter(t => t !== tagSlug)
      : [...currentTags, tagSlug];
    updateFilters({ tags: newTags });
  };
  
  // Clear all filters
  const clearFilters = () => {
    router.push(pathname);
    setSearchQuery('');
  };
  
  // Check if any filters are active
  const hasActiveFilters = Boolean(
    filters.search || 
    filters.genres?.length || 
    filters.tags?.length || 
    filters.status
  );
  
  const novels = novelsData?.data || [];
  const meta = novelsData?.meta;
  
  return (
    <div className="container-custom py-6">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
        <h1 className="text-2xl md:text-3xl font-heading font-bold">
          {t('title', { count: meta?.total || 0 })}
        </h1>
        
        {/* Search Form */}
        <form onSubmit={handleSearch} className="flex-1 max-w-md">
          <div className="relative">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder={t('searchPlaceholder')}
              className="input-primary w-full pl-10 pr-4"
            />
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground-muted" />
            {searchQuery && (
              <button
                type="button"
                onClick={() => {
                  setSearchQuery('');
                  updateFilters({ search: undefined });
                }}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-foreground-muted hover:text-foreground"
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </div>
        </form>
      </div>
      
      {/* Toolbar */}
      <div className="flex flex-wrap items-center gap-4 mb-6">
        {/* Filter Toggle */}
        <button
          onClick={() => setShowFilters(!showFilters)}
          className={`btn-secondary flex items-center gap-2 ${showFilters ? 'ring-2 ring-accent-primary' : ''}`}
        >
          <Filter className="w-4 h-4" />
          {t('filters')}
          {hasActiveFilters && (
            <span className="bg-accent-primary text-white text-xs px-1.5 py-0.5 rounded-full">
              {(filters.genres?.length || 0) + (filters.tags?.length || 0) + (filters.status ? 1 : 0)}
            </span>
          )}
        </button>
        
        {/* Sort Dropdown */}
        <div className="relative">
          <select
            value={filters.sort}
            onChange={(e) => updateFilters({ sort: e.target.value as NovelFiltersType['sort'] })}
            className="input-primary pr-8 appearance-none cursor-pointer"
          >
            {SORT_OPTIONS.map(option => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
          <ChevronDown className="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 pointer-events-none text-foreground-muted" />
        </div>
        
        {/* Status Filter */}
        <div className="relative">
          <select
            value={filters.status || ''}
            onChange={(e) => updateFilters({ status: e.target.value as NovelFiltersType['status'] || undefined })}
            className="input-primary pr-8 appearance-none cursor-pointer"
          >
            {STATUS_OPTIONS.map(option => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
          <ChevronDown className="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 pointer-events-none text-foreground-muted" />
        </div>
        
        {/* View Mode Toggle */}
        <div className="ml-auto flex gap-1">
          <button
            onClick={() => setViewMode('grid')}
            className={`p-2 rounded ${viewMode === 'grid' ? 'bg-accent-primary text-white' : 'bg-background-secondary text-foreground-secondary'}`}
            aria-label="Grid view"
          >
            <Grid className="w-5 h-5" />
          </button>
          <button
            onClick={() => setViewMode('list')}
            className={`p-2 rounded ${viewMode === 'list' ? 'bg-accent-primary text-white' : 'bg-background-secondary text-foreground-secondary'}`}
            aria-label="List view"
          >
            <List className="w-5 h-5" />
          </button>
        </div>
        
        {/* Clear Filters */}
        {hasActiveFilters && (
          <button
            onClick={clearFilters}
            className="text-sm text-accent-primary hover:underline flex items-center gap-1"
          >
            <X className="w-4 h-4" />
            Сбросить фильтры
          </button>
        )}
      </div>
      
      <div className="flex gap-6">
        {/* Filters Sidebar */}
        {showFilters && (
          <aside className="w-64 shrink-0 space-y-6">
            {/* Genres */}
            <div className="bg-background-secondary rounded-card p-4">
              <h3 className="font-semibold mb-3">Жанры</h3>
              <div className="space-y-2 max-h-64 overflow-y-auto">
                {genres?.map(genre => (
                  <label key={genre.id} className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={filters.genres?.includes(genre.slug) || false}
                      onChange={() => toggleGenre(genre.slug)}
                      className="checkbox"
                    />
                    <span className="text-sm">{genre.name}</span>
                  </label>
                ))}
              </div>
            </div>
            
            {/* Tags */}
            <div className="bg-background-secondary rounded-card p-4">
              <h3 className="font-semibold mb-3">Теги</h3>
              <div className="flex flex-wrap gap-2 max-h-64 overflow-y-auto">
                {tags?.map(tag => (
                  <button
                    key={tag.id}
                    onClick={() => toggleTag(tag.slug)}
                    className={`text-xs px-2 py-1 rounded-tag transition-colors ${
                      filters.tags?.includes(tag.slug)
                        ? 'bg-accent-primary text-white'
                        : 'bg-background-tertiary text-foreground-secondary hover:bg-background-hover'
                    }`}
                  >
                    {tag.name}
                  </button>
                ))}
              </div>
            </div>
          </aside>
        )}
        
        {/* Main Content */}
        <main className="flex-1">
          {/* Active Filters Tags */}
          {hasActiveFilters && (
            <div className="flex flex-wrap gap-2 mb-4">
              {filters.search && (
                <span className="filter-tag">
                  Поиск: {filters.search}
                  <button onClick={() => updateFilters({ search: undefined })}>
                    <X className="w-3 h-3" />
                  </button>
                </span>
              )}
              {filters.genres?.map(genreSlug => {
                const genre = genres?.find(g => g.slug === genreSlug);
                return (
                  <span key={genreSlug} className="filter-tag">
                    {genre?.name || genreSlug}
                    <button onClick={() => toggleGenre(genreSlug)}>
                      <X className="w-3 h-3" />
                    </button>
                  </span>
                );
              })}
              {filters.tags?.map(tagSlug => {
                const tag = tags?.find(t => t.slug === tagSlug);
                return (
                  <span key={tagSlug} className="filter-tag">
                    {tag?.name || tagSlug}
                    <button onClick={() => toggleTag(tagSlug)}>
                      <X className="w-3 h-3" />
                    </button>
                  </span>
                );
              })}
            </div>
          )}
          
          {/* Loading State */}
          {isLoading && (
            <div className={`grid gap-4 ${
              viewMode === 'grid' 
                ? 'grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6' 
                : 'grid-cols-1'
            }`}>
              {Array.from({ length: 12 }).map((_, i) => (
                <NovelCardSkeleton key={i} />
              ))}
            </div>
          )}
          
          {/* Error State */}
          {error && (
            <div className="text-center py-12">
              <p className="text-status-error mb-4">Ошибка загрузки каталога</p>
              <button 
                onClick={() => window.location.reload()}
                className="btn-primary"
              >
                Попробовать снова
              </button>
            </div>
          )}
          
          {/* Empty State */}
          {!isLoading && !error && novels.length === 0 && (
            <div className="text-center py-12">
              <p className="text-foreground-secondary mb-4">
                {hasActiveFilters 
                  ? 'По вашему запросу ничего не найдено' 
                  : 'Каталог пуст'}
              </p>
              {hasActiveFilters && (
                <button onClick={clearFilters} className="btn-secondary">
                  Сбросить фильтры
                </button>
              )}
            </div>
          )}
          
          {/* Novels Grid */}
          {!isLoading && !error && novels.length > 0 && (
            <>
              <div className={`grid gap-4 ${
                viewMode === 'grid' 
                  ? 'grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6' 
                  : 'grid-cols-1'
              }`}>
                {novels.map(novel => (
                  <NovelCard key={novel.id} novel={novel} />
                ))}
              </div>
              
              {/* Pagination */}
              {meta && meta.totalPages > 1 && (
                <div className="flex justify-center items-center gap-2 mt-8">
                  <button
                    onClick={() => updateFilters({ page: meta.page - 1 })}
                    disabled={meta.page <= 1}
                    className="btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Назад
                  </button>
                  
                  <span className="text-sm text-foreground-secondary">
                    Страница {meta.page} из {meta.totalPages}
                  </span>
                  
                  <button
                    onClick={() => updateFilters({ page: meta.page + 1 })}
                    disabled={meta.page >= meta.totalPages}
                    className="btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Вперед
                  </button>
                </div>
              )}
            </>
          )}
        </main>
      </div>
    </div>
  );
}
