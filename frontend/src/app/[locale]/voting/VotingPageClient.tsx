'use client';

import { useEffect, useMemo, useState } from 'react';
import { useLocale, useTranslations } from 'next-intl';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import Image from 'next/image';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';
import { toast } from 'react-hot-toast';
import { isModerator } from '@/store/auth';

interface Proposal {
  id: string;
  userId: string;
  originalLink: string;
  status: string;
  title: string;
  altTitles: string[];
  description: string;
  coverUrl?: string;
  genres: string[];
  tags: string[];
  voteScore: number;
  votesCount: number;
  translationTicketsInvested?: number;
  createdAt: string;
  user?: {
    id: string;
    displayName: string;
    avatarUrl?: string;
    level: number;
  };
  userVote?: number;
}

interface VotingLeaderboard {
  poll?: {
    id: string;
    status: string;
    endsAt: string;
  };
  entries: {
    novelId: string;
    score: number;
    proposal: Proposal;
  }[];
  nextReset: string;
}

interface TranslationLeaderboard {
  entries: {
    targetId: string;
    status: string;
    score: number;
    novelId?: string;
    proposalId?: string;
    title: string;
    coverUrl?: string;
  }[];
}

interface VotingStats {
  totalProposals: number;
  activeProposals: number;
  totalVotesCast: number;
  proposalsTranslated: number;
}

interface WalletInfo {
  userId: string;
  dailyVotes: number;
  novelRequests: number;
  translationTickets: number;
  nextDailyReset: string;
}

export default function VotingPageClient() {
  const t = useTranslations('voting');
  const locale = useLocale();
  const { isAuthenticated, user } = useAuth();
  const queryClient = useQueryClient();
  const [voteAmount, setVoteAmount] = useState<{ [key: string]: number }>({});
  const [selectedBoard, setSelectedBoard] = useState<'daily' | 'translation'>('daily');
  const [nowMs, setNowMs] = useState(() => Date.now());

  // Re-render periodically so countdown timers tick down.
  useEffect(() => {
    const id = window.setInterval(() => setNowMs(Date.now()), 30_000);
    return () => window.clearInterval(id);
  }, []);

  // Fetch daily vote leaderboard
  const { data: dailyLeaderboard, isLoading: isDailyLoading } = useQuery<VotingLeaderboard>({
    queryKey: ['voting-leaderboard', 'daily'],
    queryFn: async () => {
      const response = await api.get<VotingLeaderboard>('/voting/leaderboard?limit=20');
      return response.data;
    },
  });

  // Fetch translation vote leaderboard
  const { data: translationLeaderboard, isLoading: isTranslationLoading } = useQuery<TranslationLeaderboard>({
    queryKey: ['voting-leaderboard', 'translation'],
    queryFn: async () => {
      const response = await api.get<TranslationLeaderboard>('/translation/leaderboard?limit=20');
      return response.data;
    },
  });

  // Fetch stats
  const { data: stats } = useQuery<VotingStats>({
    queryKey: ['voting-stats'],
    queryFn: async () => {
      const response = await api.get<VotingStats>('/voting/stats');
      return response.data;
    },
  });

  // Fetch wallet
  const { data: wallet } = useQuery<WalletInfo>({
    queryKey: ['wallet'],
    queryFn: async () => {
      const response = await api.get<WalletInfo>('/wallet');
      return response.data;
    },
    enabled: isAuthenticated,
  });

  // Vote mutation
  const voteMutation = useMutation({
    mutationFn: async ({ id, amount }: { id: string; amount: number }) => {
      if (selectedBoard === 'daily') {
      const response = await api.post('/votes', {
          proposalId: id,
          ticketType: 'daily_vote',
          amount,
        });
        return response.data;
      }

      const response = await api.post('/translation-votes', {
        targetId: id,
        amount,
      });
      return response.data;
    },
    onSuccess: () => {
      toast.success(selectedBoard === 'translation' ? t('investSuccess') : t('voteSuccess'));
      queryClient.invalidateQueries({ queryKey: ['voting-leaderboard', 'daily'] });
      queryClient.invalidateQueries({ queryKey: ['voting-leaderboard', 'translation'] });
      queryClient.invalidateQueries({ queryKey: ['wallet'] });
    },
    onError: (error: any) => {
      const backendMessage = error?.response?.data?.error?.message;
      toast.error(backendMessage || t('voteError'));
    },
  });

  // Moderator action: remove a proposal from voting (force-reject).
  const forceRejectMutation = useMutation({
    mutationFn: async (proposalId: string) => {
      return api.post(`/moderation/proposals/${proposalId}/force-reject`, { reason: 'bad_cover' });
    },
    onSuccess: () => {
      toast.success('Убрано из голосования');
      queryClient.invalidateQueries({ queryKey: ['voting-leaderboard'] });
      queryClient.invalidateQueries({ queryKey: ['proposals'] });
      queryClient.invalidateQueries({ queryKey: ['voting-stats'] });
    },
    onError: (error: any) => {
      const backendMessage = error?.response?.data?.error?.message;
      toast.error(backendMessage || 'Не удалось убрать из голосования');
    },
  });

  const handleVote = (id: string) => {
    const amount = voteAmount[id] || 1;
    voteMutation.mutate({ id, amount });
  };

  const parseApiDateMs = (s: string): number => {
    // Backend uses RFC3339 with up to nanoseconds (e.g. 2026-02-04T22:32:13.259165357Z).
    // Some JS engines fail to parse >3 fractional digits, so normalize to milliseconds.
    const raw = (s || '').trim();
    if (!raw) return NaN;
    const normalized = raw.replace(/(\.\d{3})\d+Z$/i, '$1Z');
    const ms = Date.parse(normalized);
    return Number.isFinite(ms) ? ms : NaN;
  };

  const formatTimeRemaining = (endTime: string, now: number) => {
    const endMs = parseApiDateMs(endTime);
    if (!Number.isFinite(endMs)) return '-';
    const diff = endMs - now;

    if (diff <= 0) return t('pollEnded');

    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

    return `${hours}h ${minutes}m`;
  };

  const getTicketBalance = () => {
    if (!wallet) return 0;
    return selectedBoard === 'daily' ? wallet.dailyVotes : wallet.translationTickets;
  };

  const isLoading = selectedBoard === 'daily' ? isDailyLoading : isTranslationLoading;
  const nextReset = dailyLeaderboard?.nextReset;
  const nextResetLabel = useMemo(() => {
    if (!nextReset) return null;
    return formatTimeRemaining(nextReset, nowMs);
  }, [nextReset, nowMs]);
  const topTitle =
    selectedBoard === 'daily'
      ? dailyLeaderboard?.entries?.[0]?.proposal?.title
      : translationLeaderboard?.entries?.[0]?.title;

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
        <div>
          <h1 className="text-3xl font-bold text-foreground-primary">{t('title')}</h1>
          <p className="text-foreground-secondary mt-2">{t('description')}</p>
        </div>
        
        <div className="flex gap-3">
          <Link
            href={`/${locale}/proposals`}
            className="btn-secondary"
          >
            {t('viewAll')}
          </Link>
          {isAuthenticated && (
            <Link
              href={`/${locale}/proposals/new`}
              className="btn-primary"
            >
              {t('proposeNovel')}
            </Link>
          )}
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        <div className="bg-background-secondary rounded-xl p-4">
          <p className="text-foreground-muted text-sm">{t('stats.totalProposals')}</p>
          <p className="text-2xl font-bold text-foreground-primary">{stats?.totalProposals ?? '-'}</p>
        </div>
        <div className="bg-background-secondary rounded-xl p-4">
          <p className="text-foreground-muted text-sm">{t('stats.activeVoting')}</p>
          <p className="text-2xl font-bold text-foreground-primary">{stats?.activeProposals ?? '-'}</p>
        </div>
        <div className="bg-background-secondary rounded-xl p-4">
          <p className="text-foreground-muted text-sm">{t('stats.totalVotes')}</p>
          <p className="text-2xl font-bold text-foreground-primary">{stats?.totalVotesCast ?? '-'}</p>
        </div>
        <div className="bg-background-secondary rounded-xl p-4">
          <p className="text-foreground-muted text-sm">{t('stats.translated')}</p>
          <p className="text-2xl font-bold text-accent-primary">{stats?.proposalsTranslated ?? '-'}</p>
        </div>
      </div>

      {/* Timer */}
      {nextReset && (
        <div className="bg-gradient-to-r from-accent-primary/20 to-accent-secondary/20 rounded-xl p-4 mb-8">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-foreground-secondary">{t('nextWinner')}</p>
              <p className="text-2xl font-bold text-foreground-primary">
                {nextResetLabel}
              </p>
            </div>
            <div className="text-right">
              <p className="text-foreground-secondary">{t('topNovel')}</p>
              <p className="text-lg font-medium text-accent-primary">
                {topTitle || '-'}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Board Selector */}
      {isAuthenticated && (
        <div className="bg-background-secondary rounded-xl p-4 mb-6">
          <div className="flex items-center justify-between">
            <div className="flex gap-4">
              <button
                onClick={() => setSelectedBoard('daily')}
                className={`px-4 py-2 rounded-lg transition-colors ${
                  selectedBoard === 'daily'
                    ? 'bg-status-info text-white'
                    : 'bg-background-tertiary text-foreground-secondary hover:text-foreground-primary'
                }`}
              >
                {t('dailyVotes')} ({wallet?.dailyVotes ?? 0})
              </button>
              <button
                onClick={() => setSelectedBoard('translation')}
                className={`px-4 py-2 rounded-lg transition-colors ${
                  selectedBoard === 'translation'
                    ? 'bg-status-success text-white'
                    : 'bg-background-tertiary text-foreground-secondary hover:text-foreground-primary'
                }`}
              >
                {t('translationTickets')} ({wallet?.translationTickets ?? 0})
              </button>
            </div>
            <p className="text-foreground-muted text-sm">
              {t('ticketBalance', { count: getTicketBalance() })}
            </p>
          </div>
        </div>
      )}

      {/* Leaderboard */}
      <div className="space-y-4">
        {isLoading ? (
          // Skeleton
          Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="bg-background-secondary rounded-xl p-4 animate-pulse">
              <div className="flex gap-4">
                <div className="w-20 h-28 bg-background-hover rounded-lg" />
                <div className="flex-1 space-y-3">
                  <div className="w-3/4 h-5 bg-background-hover rounded" />
                  <div className="w-1/2 h-4 bg-background-hover rounded" />
                  <div className="w-full h-16 bg-background-hover rounded" />
                </div>
              </div>
            </div>
          ))
        ) : selectedBoard === 'daily' && (dailyLeaderboard?.entries.length ?? 0) === 0 ? (
          <div className="text-center py-12">
            <p className="text-foreground-muted">{t('noProposals')}</p>
            <Link
              href={`/${locale}/proposals/new`}
              className="inline-block mt-4 btn-primary"
            >
              {t('beFirst')}
            </Link>
          </div>
        ) : selectedBoard === 'daily' ? (
          dailyLeaderboard?.entries.map((entry, index) => (
            <div
              key={entry.proposal.id}
              className={`bg-background-secondary rounded-xl p-4 border-2 transition-colors ${
                index === 0
                  ? 'border-accent-warning/50'
                  : index === 1
                  ? 'border-foreground-muted/30'
                  : index === 2
                  ? 'border-accent-warning/30'
                  : 'border-transparent'
              }`}
            >
              <div className="flex gap-4">
                {/* Rank */}
                <div className="flex-shrink-0 w-12 h-12 rounded-full bg-background-tertiary flex items-center justify-center">
                  <span
                    className={`text-xl font-bold ${
                      index === 0
                        ? 'text-accent-warning'
                        : index === 1
                        ? 'text-foreground-muted'
                        : index === 2
                        ? 'text-accent-warning/70'
                        : 'text-foreground-secondary'
                    }`}
                  >
                    #{index + 1}
                  </span>
                </div>

                {/* Cover */}
                <div className="flex-shrink-0">
                  {entry.proposal.coverUrl ? (
                    <img
                      src={entry.proposal.coverUrl}
                      alt={entry.proposal.title}
                      width={80}
                      height={112}
                      className="rounded-lg object-cover"
                      loading="lazy"
                      decoding="async"
                      onError={(e) => {
                        (e.currentTarget as HTMLImageElement).src = '/placeholder-cover.svg';
                      }}
                    />
                  ) : (
                    <div className="w-20 h-28 bg-background-tertiary rounded-lg flex items-center justify-center">
                      <svg className="w-8 h-8 text-foreground-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                      </svg>
                    </div>
                  )}
                </div>

                {/* Info */}
                <div className="flex-1 min-w-0">
                  <h3 className="text-lg font-semibold text-foreground-primary truncate">
                    {entry.proposal.title}
                  </h3>
                  <p className="text-sm text-foreground-muted mt-1 line-clamp-2">
                    {entry.proposal.description}
                  </p>
                  <div className="flex flex-wrap gap-2 mt-2">
                    {entry.proposal.genres.slice(0, 3).map((genre) => (
                      <span key={genre} className="px-2 py-0.5 text-xs bg-accent-primary/20 text-accent-primary rounded-full">
                        {genre}
                      </span>
                    ))}
                  </div>
                </div>

                {/* Votes and Action */}
                <div className="flex-shrink-0 flex flex-col items-center justify-between">
                  <div className="text-center">
                    <p className="text-2xl font-bold text-accent-primary">{entry.score}</p>
                    <p className="text-xs text-foreground-muted">{t('dailyVotesLabel')}</p>
                    <p className="text-xs text-foreground-muted mt-1">
                      {t('translationInvested')}: {entry.proposal.translationTicketsInvested ?? 0}
                    </p>
                  </div>

                  {isAuthenticated && (
                    <div className="flex items-center gap-2 mt-2">
                      <input
                        type="number"
                        min="1"
                        max={getTicketBalance()}
                        value={voteAmount[entry.proposal.id] || 1}
                        onChange={(e) =>
                          setVoteAmount({
                            ...voteAmount,
                            [entry.proposal.id]: Math.max(1, parseInt(e.target.value) || 1),
                          })
                        }
                        className="w-16 px-2 py-1 text-center input"
                      />
                      <button
                        onClick={() => handleVote(entry.proposal.id)}
                        disabled={voteMutation.isPending || getTicketBalance() < 1}
                        className="btn-primary"
                      >
                        {voteMutation.isPending ? '...' : t('vote')}
                      </button>
                    </div>
                  )}

                  {isModerator(user) && (
                    <button
                      type="button"
                      onClick={() => forceRejectMutation.mutate(entry.proposal.id)}
                      disabled={forceRejectMutation.isPending}
                      className="btn-secondary mt-2"
                      title="Убрать из голосования"
                    >
                      Убрать
                    </button>
                  )}
                </div>
              </div>

              {/* Proposer info */}
              {entry.proposal.user && (
                <div className="flex items-center gap-2 mt-3 pt-3 border-t border-border-primary">
                  <p className="text-xs text-foreground-muted">
                    {t('proposedBy')}
                  </p>
                  <Link
                    href={`/profile/${entry.proposal.user.id}`}
                    className="flex items-center gap-2 text-sm text-foreground-secondary hover:text-foreground-primary"
                  >
                    {entry.proposal.user.avatarUrl ? (
                      <Image
                        src={entry.proposal.user.avatarUrl}
                        alt=""
                        width={20}
                        height={20}
                        className="rounded-full"
                      />
                    ) : (
                      <div className="w-5 h-5 bg-accent-primary/20 rounded-full" />
                    )}
                    <span>{entry.proposal.user.displayName}</span>
                    <span className="text-xs text-foreground-muted">Lv.{entry.proposal.user.level}</span>
                  </Link>
                </div>
              )}
            </div>
          ))
        ) : (translationLeaderboard?.entries.length ?? 0) === 0 ? (
          <div className="text-center py-12">
            <p className="text-foreground-muted">{t('noProposals')}</p>
          </div>
        ) : (
          translationLeaderboard?.entries.map((entry, index) => (
            <div
              key={entry.targetId}
              className={`bg-background-secondary rounded-xl p-4 border-2 transition-colors ${
                index === 0
                  ? 'border-accent-warning/50'
                  : index === 1
                  ? 'border-foreground-muted/30'
                  : index === 2
                  ? 'border-accent-warning/30'
                  : 'border-transparent'
              }`}
            >
              <div className="flex gap-4">
                {/* Rank */}
                <div className="flex-shrink-0 w-12 h-12 rounded-full bg-background-tertiary flex items-center justify-center">
                  <span className="text-xl font-bold text-foreground-secondary">#{index + 1}</span>
                </div>

                {/* Cover */}
                <div className="flex-shrink-0">
                  {entry.coverUrl ? (
                    <Image
                      src={entry.coverUrl}
                      alt={entry.title}
                      width={80}
                      height={112}
                      className="rounded-lg object-cover"
                    />
                  ) : (
                    <div className="w-20 h-28 bg-background-tertiary rounded-lg flex items-center justify-center">
                      <svg className="w-8 h-8 text-foreground-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                      </svg>
                    </div>
                  )}
                </div>

                {/* Info */}
                <div className="flex-1 min-w-0">
                  <h3 className="text-lg font-semibold text-foreground-primary truncate">{entry.title}</h3>
                  <p className="text-sm text-foreground-muted mt-1">
                    {t('translationTickets')}: {entry.score}
                  </p>
                </div>

                {/* Action */}
                <div className="flex-shrink-0 flex flex-col items-center justify-between">
                  {isAuthenticated && (
                    <div className="flex items-center gap-2 mt-2">
                      <input
                        type="number"
                        min="1"
                        max={getTicketBalance()}
                        value={voteAmount[entry.targetId] || 1}
                        onChange={(e) =>
                          setVoteAmount({
                            ...voteAmount,
                            [entry.targetId]: Math.max(1, parseInt(e.target.value) || 1),
                          })
                        }
                        className="w-16 px-2 py-1 text-center input"
                      />
                      <button
                        onClick={() => handleVote(entry.targetId)}
                        disabled={voteMutation.isPending || getTicketBalance() < 1}
                        className="btn-primary"
                      >
                        {voteMutation.isPending ? '...' : t('invest')}
                      </button>
                    </div>
                  )}
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {/* How it works */}
      <div className="mt-12 bg-background-secondary rounded-xl p-6">
        <h2 className="text-xl font-semibold text-foreground-primary mb-4">{t('howItWorks')}</h2>
        <div className="grid md:grid-cols-3 gap-6">
          <div className="text-center">
            <div className="w-12 h-12 bg-accent-secondary/20 rounded-full flex items-center justify-center mx-auto mb-3">
              <span className="text-xl font-bold text-accent-secondary">1</span>
            </div>
            <h3 className="font-medium text-foreground-primary mb-1">{t('step1Title')}</h3>
            <p className="text-sm text-foreground-secondary">{t('step1Description')}</p>
          </div>
          <div className="text-center">
            <div className="w-12 h-12 bg-status-info/20 rounded-full flex items-center justify-center mx-auto mb-3">
              <span className="text-xl font-bold text-status-info">2</span>
            </div>
            <h3 className="font-medium text-foreground-primary mb-1">{t('step2Title')}</h3>
            <p className="text-sm text-foreground-secondary">{t('step2Description')}</p>
          </div>
          <div className="text-center">
            <div className="w-12 h-12 bg-status-success/20 rounded-full flex items-center justify-center mx-auto mb-3">
              <span className="text-xl font-bold text-status-success">3</span>
            </div>
            <h3 className="font-medium text-foreground-primary mb-1">{t('step3Title')}</h3>
            <p className="text-sm text-foreground-secondary">{t('step3Description')}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
