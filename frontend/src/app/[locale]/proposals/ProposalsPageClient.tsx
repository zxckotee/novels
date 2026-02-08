'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useTranslations, useLocale } from 'next-intl';
import { useQuery } from '@tanstack/react-query';
import { Plus, Search, Filter } from 'lucide-react';
import { api } from '@/lib/api';
import { useAuth } from '@/hooks/useAuth';

interface Proposal {
  id: string;
  title: string;
  description: string;
  coverUrl?: string;
  genres: string[];
  status: string;
  voteScore: number;
  createdAt: string;
}

export default function ProposalsPageClient() {
  const t = useTranslations('proposals');
  const locale = useLocale();
  const { isAuthenticated } = useAuth();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');

  // Fetch proposals
  const { data: proposals, isLoading, error } = useQuery<Proposal[]>({
    queryKey: ['proposals', statusFilter, searchQuery],
    queryFn: async () => {
      const params: Record<string, string> = {};
      if (statusFilter !== 'all') params.status = statusFilter;
      if (searchQuery) params.search = searchQuery;
      
      const response = await api.get('/proposals', { params });
      // API может возвращать { proposals: [...] } или просто массив
      const data = response.data;
      if (Array.isArray(data)) {
        return data;
      } else if (data && typeof data === 'object' && 'proposals' in data) {
        return (data as any).proposals || [];
      }
      return [];
    },
  });
  
  // Debug logs removed

  return (
    <div className="container-custom py-8">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
        <div>
          <h1 className="text-3xl font-bold text-foreground-primary">Все предложения</h1>
          <p className="text-foreground-secondary mt-2">Новеллы, предложенные сообществом</p>
        </div>
        
        {isAuthenticated && (
          <Link
            href={`/${locale}/proposals/new`}
            className="btn-primary flex items-center gap-2"
          >
            <Plus className="w-4 h-4" />
            Предложить новеллу
          </Link>
        )}
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4 mb-6">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground-muted" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="input w-full pl-10"
            placeholder="Поиск предложений..."
          />
        </div>
        
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="input w-full sm:w-48"
        >
          <option value="all">Все статусы</option>
          <option value="moderation">На рассмотрении</option>
          <option value="voting">На голосовании</option>
          <option value="accepted">Одобрено</option>
          <option value="rejected">Отклонено</option>
        </select>
      </div>

      {/* Proposals List */}
      <div className="space-y-4">
        {isLoading ? (
          // Skeleton
          Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="bg-background-secondary rounded-card p-6 animate-pulse">
              <div className="flex gap-4">
                <div className="w-20 h-28 bg-background-hover rounded" />
                <div className="flex-1 space-y-3">
                  <div className="h-5 bg-background-hover rounded w-3/4" />
                  <div className="h-4 bg-background-hover rounded w-1/2" />
                  <div className="h-16 bg-background-hover rounded" />
                </div>
              </div>
            </div>
          ))
        ) : error ? (
          <div className="text-center py-12">
            <p className="text-status-error mb-4">Ошибка загрузки предложений</p>
            <button onClick={() => window.location.reload()} className="btn-secondary">
              Попробовать снова
            </button>
          </div>
        ) : proposals && proposals.length > 0 ? (
          proposals.map((proposal) => (
            <div key={proposal.id} className="bg-background-secondary rounded-card p-6 hover:bg-background-hover transition-colors">
              <div className="flex gap-4">
                {/* Cover */}
                {proposal.coverUrl ? (
                  <img
                    src={proposal.coverUrl}
                    alt={proposal.title}
                    className="w-20 h-28 object-cover rounded"
                  />
                ) : (
                  <div className="w-20 h-28 bg-background-tertiary rounded flex items-center justify-center">
                    <span className="text-foreground-muted text-xs">Нет обложки</span>
                  </div>
                )}
                
                {/* Info */}
                <div className="flex-1 min-w-0">
                  <h3 className="text-lg font-semibold text-foreground-primary truncate">
                    {proposal.title}
                  </h3>
                  <p className="text-sm text-foreground-muted line-clamp-2 mb-2">
                    {proposal.description}
                  </p>
                  <div className="flex flex-wrap gap-2">
                    {proposal.genres?.slice(0, 3).map((genre) => (
                      <span key={genre} className="px-2 py-0.5 text-xs bg-accent-primary/20 text-accent-primary rounded-full">
                        {genre}
                      </span>
                    ))}
                  </div>
                </div>
                
                {/* Stats */}
                <div className="flex flex-col items-end justify-between">
                  <div className="text-center">
                    <p className="text-2xl font-bold text-accent-primary">{proposal.voteScore || 0}</p>
                    <p className="text-xs text-foreground-muted">голосов</p>
                  </div>
                  <span className={`px-3 py-1 rounded-full text-xs ${
                    proposal.status === 'voting'
                      ? 'bg-status-info/20 text-status-info'
                      : proposal.status === 'accepted'
                      ? 'bg-status-success/20 text-status-success'
                      : proposal.status === 'rejected'
                      ? 'bg-status-error/20 text-status-error'
                      : proposal.status === 'moderation'
                      ? 'bg-accent-warning/20 text-accent-warning'
                      : 'bg-background-tertiary text-foreground-muted'
                  }`}>
                    {proposal.status === 'voting' && 'На голосовании'}
                    {proposal.status === 'accepted' && 'Одобрено'}
                    {proposal.status === 'rejected' && 'Отклонено'}
                    {proposal.status === 'moderation' && 'На рассмотрении'}
                  </span>
                </div>
              </div>
            </div>
          ))
        ) : (
          <div className="text-center py-12">
            <p className="text-foreground-muted mb-4">Предложений пока нет</p>
            {isAuthenticated && (
              <Link
                href={`/${locale}/proposals/new`}
                className="btn-primary inline-flex items-center gap-2"
              >
                <Plus className="w-4 h-4" />
                Предложить первую новеллу
              </Link>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
