import { useTranslations } from 'next-intl';
import Link from 'next/link';
import { ArrowRight, Vote, PlusCircle, Star, TrendingUp, Clock, Sparkles } from 'lucide-react';
import { NovelCard } from '@/components/novel/NovelCard';
import { HeroSlider } from '@/components/home/HeroSlider';

export default function HomePage() {
  const t = useTranslations('home');

  return (
    <div className="min-h-screen">
      {/* Hero Section with Slider */}
      <section className="relative">
        <HeroSlider />
      </section>

      {/* Action Blocks - Voting & Propose */}
      <section className="container-custom py-8">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Vote Block */}
          <Link
            href="/voting"
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
            href="/propose"
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
            href="/catalog?sort=updated_at"
            className="text-sm text-accent-primary hover:underline flex items-center gap-1"
          >
            {t('sections.latestUpdates')}
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
        <LatestUpdatesGrid />
      </section>

      {/* Popular */}
      <section className="section container-custom">
        <div className="section-title">
          <div className="flex items-center gap-2">
            <TrendingUp className="w-5 h-5 text-accent-secondary" />
            <h2>{t('sections.popular')}</h2>
          </div>
          <Link
            href="/catalog?sort=views"
            className="text-sm text-accent-primary hover:underline flex items-center gap-1"
          >
            Смотреть все
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
        <NovelCarousel />
      </section>

      {/* New Releases */}
      <section className="section container-custom">
        <div className="section-title">
          <div className="flex items-center gap-2">
            <Sparkles className="w-5 h-5 text-accent-warning" />
            <h2>{t('sections.newReleases')}</h2>
          </div>
          <Link
            href="/catalog?sort=created_at"
            className="text-sm text-accent-primary hover:underline flex items-center gap-1"
          >
            Смотреть все
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
        <NovelCarousel />
      </section>

      {/* Top Rated */}
      <section className="section container-custom">
        <div className="section-title">
          <div className="flex items-center gap-2">
            <Star className="w-5 h-5 text-accent-warning" />
            <h2>{t('sections.topRated')}</h2>
          </div>
          <Link
            href="/catalog?sort=rating"
            className="text-sm text-accent-primary hover:underline flex items-center gap-1"
          >
            Смотреть все
            <ArrowRight className="w-4 h-4" />
          </Link>
        </div>
        <NovelCarousel />
      </section>
    </div>
  );
}

// Компонент сетки последних обновлений
function LatestUpdatesGrid() {
  // TODO: Загружать данные через API
  const mockNovels = Array(12).fill(null).map((_, i) => ({
    id: `novel-${i}`,
    slug: `novel-${i}`,
    title: `Название новеллы ${i + 1}`,
    coverUrl: `/placeholder-cover.jpg`,
    rating: 4.5,
    latestChapter: 150 + i,
    updatedAt: new Date().toISOString(),
    isNew: i < 3,
  }));

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
      {mockNovels.map((novel) => (
        <NovelCard key={novel.id} novel={novel} />
      ))}
    </div>
  );
}

// Компонент карусели новелл
function NovelCarousel() {
  // TODO: Загружать данные через API
  const mockNovels = Array(10).fill(null).map((_, i) => ({
    id: `novel-${i}`,
    slug: `novel-${i}`,
    title: `Название новеллы ${i + 1}`,
    coverUrl: `/placeholder-cover.jpg`,
    rating: 4.5,
    latestChapter: 150 + i,
    updatedAt: new Date().toISOString(),
    isNew: i < 2,
  }));

  return (
    <div className="flex gap-4 overflow-x-auto pb-4 scrollbar-hide">
      {mockNovels.map((novel) => (
        <div key={novel.id} className="flex-shrink-0 w-[150px] md:w-[180px]">
          <NovelCard novel={novel} />
        </div>
      ))}
    </div>
  );
}
