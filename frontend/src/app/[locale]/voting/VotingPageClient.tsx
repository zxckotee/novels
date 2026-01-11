'use client';

import { useState } from 'react';
import { useLocale, useTranslations } from 'next-intl';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import Link from 'next/link';
import Image from 'next/image';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';
import { toast } from 'react-hot-toast';

interface Proposal {
  id: string;
  userId: string;
  originalLink: string;
  status: string;
  title: string;
  altTitles: string[];
  author: string;
  description: string;
  coverUrl?: string;
  genres: string[];
  tags: string[];
  voteScore: number;
  votesCount: number;
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
  const { isAuthenticated } = useAuth();
  const queryClient = useQueryClient();
  const [voteAmount, setVoteAmount] = useState<{ [key: string]: number }>({});
  const [selectedTicketType, setSelectedTicketType] = useState<'daily_vote' | 'translation_ticket'>('daily_vote');

  // Fetch leaderboard
  const { data: leaderboard, isLoading } = useQuery<VotingLeaderboard>({
    queryKey: ['voting-leaderboard'],
    queryFn: async () => {
      const response = await api.get<VotingLeaderboard>('/voting/leaderboard?limit=20');
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
    mutationFn: async ({ proposalId, amount }: { proposalId: string; amount: number }) => {
      const response = await api.post('/votes', {
        proposalId,
        ticketType: selectedTicketType,
        amount,
      });
      return response.data;
    },
    onSuccess: () => {
      toast.success(t('voteSuccess'));
      queryClient.invalidateQueries({ queryKey: ['voting-leaderboard'] });
      queryClient.invalidateQueries({ queryKey: ['wallet'] });
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || t('voteError'));
    },
  });

  const handleVote = (proposalId: string) => {
    const amount = voteAmount[proposalId] || 1;
    voteMutation.mutate({ proposalId, amount });
  };

  const formatTimeRemaining = (endTime: string) => {
    const end = new Date(endTime);
    const now = new Date();
    const diff = end.getTime() - now.getTime();

    if (diff <= 0) return t('pollEnded');

    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

    return `${hours}h ${minutes}m`;
  };

  const getTicketBalance = () => {
    if (!wallet) return 0;
    return selectedTicketType === 'daily_vote' ? wallet.dailyVotes : wallet.translationTickets;
  };

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
        <div>
          <h1 className="text-3xl font-bold text-text-primary">{t('title')}</h1>
          <p className="text-text-secondary mt-2">{t('description')}</p>
        </div>
        
        <div className="flex gap-3">
          <Link
            href={`/${locale}/proposals`}
            className="px-4 py-2 bg-surface-elevated text-text-primary rounded-lg hover:bg-surface-muted transition-colors"
          >
            {t('viewAll')}
          </Link>
          {isAuthenticated && (
            <Link
              href={`/${locale}/proposals/new`}
              className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-hover transition-colors"
            >
              {t('proposeNovel')}
            </Link>
          )}
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
        <div className="bg-surface-elevated rounded-xl p-4">
          <p className="text-text-muted text-sm">{t('stats.totalProposals')}</p>
          <p className="text-2xl font-bold text-text-primary">{stats?.totalProposals ?? '-'}</p>
        </div>
        <div className="bg-surface-elevated rounded-xl p-4">
          <p className="text-text-muted text-sm">{t('stats.activeVoting')}</p>
          <p className="text-2xl font-bold text-text-primary">{stats?.activeProposals ?? '-'}</p>
        </div>
        <div className="bg-surface-elevated rounded-xl p-4">
          <p className="text-text-muted text-sm">{t('stats.totalVotes')}</p>
          <p className="text-2xl font-bold text-text-primary">{stats?.totalVotesCast ?? '-'}</p>
        </div>
        <div className="bg-surface-elevated rounded-xl p-4">
          <p className="text-text-muted text-sm">{t('stats.translated')}</p>
          <p className="text-2xl font-bold text-primary">{stats?.proposalsTranslated ?? '-'}</p>
        </div>
      </div>

      {/* Timer */}
      {leaderboard?.nextReset && (
        <div className="bg-gradient-to-r from-primary/20 to-purple-500/20 rounded-xl p-4 mb-8">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-text-secondary">{t('nextWinner')}</p>
              <p className="text-2xl font-bold text-text-primary">
                {formatTimeRemaining(leaderboard.nextReset)}
              </p>
            </div>
            <div className="text-right">
              <p className="text-text-secondary">{t('topNovel')}</p>
              <p className="text-lg font-medium text-primary">
                {leaderboard.entries[0]?.proposal?.title || '-'}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Ticket Type Selector */}
      {isAuthenticated && (
        <div className="bg-surface-elevated rounded-xl p-4 mb-6">
          <div className="flex items-center justify-between">
            <div className="flex gap-4">
              <button
                onClick={() => setSelectedTicketType('daily_vote')}
                className={`px-4 py-2 rounded-lg transition-colors ${
                  selectedTicketType === 'daily_vote'
                    ? 'bg-blue-500 text-white'
                    : 'bg-surface text-text-secondary hover:text-text-primary'
                }`}
              >
                {t('dailyVotes')} ({wallet?.dailyVotes ?? 0})
              </button>
              <button
                onClick={() => setSelectedTicketType('translation_ticket')}
                className={`px-4 py-2 rounded-lg transition-colors ${
                  selectedTicketType === 'translation_ticket'
                    ? 'bg-green-500 text-white'
                    : 'bg-surface text-text-secondary hover:text-text-primary'
                }`}
              >
                {t('translationTickets')} ({wallet?.translationTickets ?? 0})
              </button>
            </div>
            <p className="text-text-muted text-sm">
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
            <div key={i} className="bg-surface-elevated rounded-xl p-4 animate-pulse">
              <div className="flex gap-4">
                <div className="w-20 h-28 bg-surface-muted rounded-lg" />
                <div className="flex-1 space-y-3">
                  <div className="w-3/4 h-5 bg-surface-muted rounded" />
                  <div className="w-1/2 h-4 bg-surface-muted rounded" />
                  <div className="w-full h-16 bg-surface-muted rounded" />
                </div>
              </div>
            </div>
          ))
        ) : leaderboard?.entries.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-text-muted">{t('noProposals')}</p>
            <Link
              href={`/${locale}/proposals/new`}
              className="inline-block mt-4 px-6 py-2 bg-primary text-white rounded-lg hover:bg-primary-hover transition-colors"
            >
              {t('beFirst')}
            </Link>
          </div>
        ) : (
          leaderboard?.entries.map((entry, index) => (
            <div
              key={entry.proposal.id}
              className={`bg-surface-elevated rounded-xl p-4 border-2 transition-colors ${
                index === 0
                  ? 'border-yellow-500/50'
                  : index === 1
                  ? 'border-gray-400/50'
                  : index === 2
                  ? 'border-amber-600/50'
                  : 'border-transparent'
              }`}
            >
              <div className="flex gap-4">
                {/* Rank */}
                <div className="flex-shrink-0 w-12 h-12 rounded-full bg-surface flex items-center justify-center">
                  <span
                    className={`text-xl font-bold ${
                      index === 0
                        ? 'text-yellow-500'
                        : index === 1
                        ? 'text-gray-400'
                        : index === 2
                        ? 'text-amber-600'
                        : 'text-text-secondary'
                    }`}
                  >
                    #{index + 1}
                  </span>
                </div>

                {/* Cover */}
                <div className="flex-shrink-0">
                  {entry.proposal.coverUrl ? (
                    <Image
                      src={entry.proposal.coverUrl}
                      alt={entry.proposal.title}
                      width={80}
                      height={112}
                      className="rounded-lg object-cover"
                    />
                  ) : (
                    <div className="w-20 h-28 bg-surface-muted rounded-lg flex items-center justify-center">
                      <svg className="w-8 h-8 text-text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                      </svg>
                    </div>
                  )}
                </div>

                {/* Info */}
                <div className="flex-1 min-w-0">
                  <h3 className="text-lg font-semibold text-text-primary truncate">
                    {entry.proposal.title}
                  </h3>
                  <p className="text-sm text-text-secondary">
                    {t('by')} {entry.proposal.author || t('unknownAuthor')}
                  </p>
                  <p className="text-sm text-text-muted mt-1 line-clamp-2">
                    {entry.proposal.description}
                  </p>
                  <div className="flex flex-wrap gap-2 mt-2">
                    {entry.proposal.genres.slice(0, 3).map((genre) => (
                      <span key={genre} className="px-2 py-0.5 text-xs bg-primary/20 text-primary rounded-full">
                        {genre}
                      </span>
                    ))}
                  </div>
                </div>

                {/* Votes and Action */}
                <div className="flex-shrink-0 flex flex-col items-center justify-between">
                  <div className="text-center">
                    <p className="text-2xl font-bold text-primary">{entry.score}</p>
                    <p className="text-xs text-text-muted">{t('votes')}</p>
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
                        className="w-16 px-2 py-1 text-center bg-surface border border-border rounded-lg text-text-primary"
                      />
                      <button
                        onClick={() => handleVote(entry.proposal.id)}
                        disabled={voteMutation.isPending || getTicketBalance() < 1}
                        className="px-4 py-2 bg-primary text-white rounded-lg hover:bg-primary-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {voteMutation.isPending ? '...' : t('vote')}
                      </button>
                    </div>
                  )}
                </div>
              </div>

              {/* Proposer info */}
              {entry.proposal.user && (
                <div className="flex items-center gap-2 mt-3 pt-3 border-t border-border">
                  <p className="text-xs text-text-muted">
                    {t('proposedBy')}
                  </p>
                  <Link
                    href={`/profile/${entry.proposal.user.id}`}
                    className="flex items-center gap-2 text-sm text-text-secondary hover:text-text-primary"
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
                      <div className="w-5 h-5 bg-primary/20 rounded-full" />
                    )}
                    <span>{entry.proposal.user.displayName}</span>
                    <span className="text-xs text-text-muted">Lv.{entry.proposal.user.level}</span>
                  </Link>
                </div>
              )}
            </div>
          ))
        )}
      </div>

      {/* How it works */}
      <div className="mt-12 bg-surface-elevated rounded-xl p-6">
        <h2 className="text-xl font-semibold text-text-primary mb-4">{t('howItWorks')}</h2>
        <div className="grid md:grid-cols-3 gap-6">
          <div className="text-center">
            <div className="w-12 h-12 bg-purple-500/20 rounded-full flex items-center justify-center mx-auto mb-3">
              <span className="text-xl font-bold text-purple-500">1</span>
            </div>
            <h3 className="font-medium text-text-primary mb-1">{t('step1Title')}</h3>
            <p className="text-sm text-text-secondary">{t('step1Description')}</p>
          </div>
          <div className="text-center">
            <div className="w-12 h-12 bg-blue-500/20 rounded-full flex items-center justify-center mx-auto mb-3">
              <span className="text-xl font-bold text-blue-500">2</span>
            </div>
            <h3 className="font-medium text-text-primary mb-1">{t('step2Title')}</h3>
            <p className="text-sm text-text-secondary">{t('step2Description')}</p>
          </div>
          <div className="text-center">
            <div className="w-12 h-12 bg-green-500/20 rounded-full flex items-center justify-center mx-auto mb-3">
              <span className="text-xl font-bold text-green-500">3</span>
            </div>
            <h3 className="font-medium text-text-primary mb-1">{t('step3Title')}</h3>
            <p className="text-sm text-text-secondary">{t('step3Description')}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
