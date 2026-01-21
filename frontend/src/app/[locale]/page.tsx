'use client';

import { useLocale, useTranslations } from 'next-intl';
import Link from 'next/link';
import { ArrowRight, Vote, PlusCircle, Star, TrendingUp, Clock, Sparkles } from 'lucide-react';
import { NovelCard, NovelCardSkeleton } from '@/components/novel/NovelCard';
import { HeroSlider } from '@/components/home/HeroSlider';
import { useLatestNovels, useTrendingNovels, useNewReleases, useTopRatedNovels } from '@/lib/api/hooks/useNovels';

export default function HomePage() {
  const t = useTranslations('home');
  const locale = useLocale();
  
  const { data: latest, isLoading: latestLoading } = useLatestNovels(12);
  const { data: trending, isLoading: trendingLoading } = useTrendingNovels(10);
  const { data: newReleases, isLoading: newLoading } = useNewReleases(10);
  const { data: topRated, isLoading: topLoading } = useTopRatedNovels(10);

  const heroSlides = (trending || []).slice(0, 3).map((n) => ({
    id: n.id,
    slug: n.slug,
    title: n.title,
    description: n.description,
    coverUrl: n.coverUrl,
  }));

  return (
    <div className="min-h-screen">
      {/* Hero Section with Slider */}
      <section className="relative">
        <HeroSlider slides={heroSlides} isLoading={trendingLoading} />
      </section>

      {/* Action Blocks - Voting & Propose */}
      <section className="container-custom py-8">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Vote Block */}
          <Link
            href={`/${locale}/voting`}
            className="card-hover p-6 flex items-center gap-4 group bg-gradient-to-br from-accent-primary/10 to-transparent"
          >
            <div className="w-14 h-14 rounded-full bg-accent-primary/20 flex items-center justify-center flex-shrink-0">
              <Vote className="w-7 h-7 text-accent-primary" />
            </div>
            <div className="flex-1">
              <h3 className="text-lg font-heading font-semibold mb-1">
                {t('voting.title')}
              </h3>
              <p className="text-foreground-secondary text-sm">
                {t('voting.description')}
              </p>
            </div>
            <ArrowRight className="w-5 h-5 text-foreground-muted group-hover:text-accent-primary group-hover:translate-x-1 transition-all" />
          </Link>

          {/* Propose Block */}
          <Link
            href={`/${locale}/proposals/new`}
            className="card-hover p-6 flex items-center gap-4 group bg-gradient-to-br from-accent-secondary/10 to-transparent"
          >
            <div className="w-14 h-14 rounded-full bg-accent-secondary/20 flex items-center justify-center flex-shrink-0">
              <PlusCircle className="w-7 h-7 text-accent-secondary" />
            </div>
            <div className="flex-1">
              <h3 className="text-lg font-heading font-semibold mb-1">
                {t('propose.title')}
              </h3>
              <p className="text-foreground-secondary text-sm">
                {t('propose.description')}
              </p>
            </div>
            <ArrowRight className="w-5 h-5 text-foreground-muted group-hover:text-accent-secondary group-hover:translate-x-1 transition-all" />
          </Link>
        </div>
      </section>

      {/* Latest Updates */}
      <section className="section container-custom">
        <div className="section-title">
          <div className="flex items-center gap-2">
            <Clock className="w-5 h-5 text-accent-primary" />
            <h2>{t('sections.latestUpdates')}</h2>
          </div>
          <Link
            href={`/${locale}/catalog?sort=updated`}
            className="text-sm text-accent-primary hover:underline flex items-center gap-1"
          >
            {t('sections.latestUpdates')}
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
        <LatestUpdatesGrid novels={latest || []} isLoading={latestLoading} />
      </section>

      {/* Popular */}
      <section className="section container-custom">
        <div className="section-title">
          <div className="flex items-center gap-2">
            <TrendingUp className="w-5 h-5 text-accent-secondary" />
            <h2>{t('sections.popular')}</h2>
          </div>
          <Link
            href={`/${locale}/catalog?sort=popular`}
            className="text-sm text-accent-primary hover:underline flex items-center gap-1"
          >
            Смотреть все
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
        <NovelCarousel novels={trending || []} isLoading={trendingLoading} />
      </section>

      {/* New Releases */}
      <section className="section container-custom">
        <div className="section-title">
          <div className="flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-accent-warning" />
            <h2>{t('sections.newReleases')}</h2>
          </div>
          <Link
            href={`/${locale}/catalog?sort=created`}
            className="text-sm text-accent-primary hover:underline flex items-center gap-1"
          >
            Смотреть все
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
        <NovelCarousel novels={newReleases || []} isLoading={newLoading} />
      </section>

      {/* Top Rated */}
      <section className="section container-custom">
        <div className="section-title">
          <div className="flex items-center gap-2">
            <Star className="w-5 h-5 text-accent-warning" />
            <h2>{t('sections.topRated')}</h2>
          </div>
          <Link
            href={`/${locale}/catalog?sort=rating`}
            className="text-sm text-accent-primary hover:underline flex items-center gap-1"
          >
            Смотреть все
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
        <NovelCarousel novels={topRated || []} isLoading={topLoading} />
      </section>
    </div>
  );
}

// Компонент сетки последних обновлений
function LatestUpdatesGrid({
  novels,
  isLoading,
}: {
  novels: Array<Parameters<typeof NovelCard>[0]['novel']>;
  isLoading?: boolean;
}) {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
      {isLoading
        ? Array.from({ length: 12 }).map((_, i) => <NovelCardSkeleton key={i} />)
        : novels.map((novel) => <NovelCard key={novel.id} novel={novel} />)}
    </div>
  );
}

// Компонент карусели новелл
function NovelCarousel({
  novels,
  isLoading,
}: {
  novels: Array<Parameters<typeof NovelCard>[0]['novel']>;
  isLoading?: boolean;
}) {
  return (
    <div className="flex gap-4 overflow-x-auto pb-4 scrollbar-hide">
      {isLoading
        ? Array.from({ length: 10 }).map((_, i) => (
            <div key={i} className="flex-shrink-0 w-[150px] md:w-[180px]">
              <NovelCardSkeleton />
            </div>
          ))
        : novels.map((novel) => (
            <div key={novel.id} className="flex-shrink-0 w-[150px] md:w-[180px]">
              <NovelCard novel={novel} />
            </div>
          ))}
    </div>
  );
}
