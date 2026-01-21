import Link from 'next/link';
import Image from 'next/image';
import { Star } from 'lucide-react';
import { useLocale } from 'next-intl';

interface NovelCardProps {
  novel: {
    id: string;
    slug: string;
    title: string;
    coverUrl?: string;
    rating?: number;
    latestChapter?: number;
    updatedAt?: string;
    isNew?: boolean;
  };
  showRating?: boolean;
  showChapter?: boolean;
}

export function NovelCard({ novel, showRating = true, showChapter = true }: NovelCardProps) {
  const locale = useLocale();

  // Format relative time
  const formatRelativeTime = (dateString?: string) => {
    if (!dateString) return '';
    
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 60) return `${diffMins} мин. назад`;
    if (diffHours < 24) return `${diffHours} ч. назад`;
    if (diffDays < 7) return `${diffDays} дн. назад`;
    return date.toLocaleDateString('ru');
  };

  return (
    <Link
      href={`/${locale}/novel/${novel.slug}`}
      className="group block"
    >
      <div className="relative aspect-cover rounded-card overflow-hidden bg-background-tertiary shadow-card group-hover:shadow-card-hover transition-shadow">
        {/* Cover Image */}
        {novel.coverUrl ? (
          <Image
            src={novel.coverUrl}
            alt={novel.title}
            fill
            sizes="(max-width: 640px) 50vw, (max-width: 1024px) 33vw, 180px"
            className="object-cover group-hover:scale-105 transition-transform duration-300"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-foreground-muted">
            <span className="text-4xl">📚</span>
          </div>
        )}

        {/* Gradient Overlay */}
        <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-transparent" />

        {/* New Badge */}
        {novel.isNew && (
          <div className="absolute top-2 left-2">
            <span className="badge-new">NEW</span>
          </div>
        )}

        {/* Rating */}
        {showRating && typeof novel.rating === 'number' && novel.rating > 0 && (
          <div className="absolute top-2 right-2 flex items-center gap-1 bg-black/50 backdrop-blur-sm rounded-tag px-1.5 py-0.5">
            <Star className="w-3 h-3 text-accent-warning fill-current" />
            <span className="text-xs font-medium">{novel.rating.toFixed(1)}</span>
          </div>
        )}

        {/* Bottom Info */}
        <div className="absolute bottom-0 left-0 right-0 p-3">
          {/* Latest Chapter */}
          {showChapter && typeof novel.latestChapter === 'number' && novel.latestChapter > 0 && (
            <div className="flex items-center justify-between text-xs text-foreground-secondary mb-1">
              <span>Глава {novel.latestChapter}</span>
              {novel.updatedAt && (
                <span>{formatRelativeTime(novel.updatedAt)}</span>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Title */}
      <h3 className="mt-2 text-sm font-medium line-clamp-2 group-hover:text-accent-primary transition-colors">
        {novel.title}
      </h3>
    </Link>
  );
}

// Скелетон для загрузки
export function NovelCardSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="aspect-cover rounded-card bg-background-hover" />
      <div className="mt-2 h-4 bg-background-hover rounded w-3/4" />
      <div className="mt-1 h-4 bg-background-hover rounded w-1/2" />
    </div>
  );
}
