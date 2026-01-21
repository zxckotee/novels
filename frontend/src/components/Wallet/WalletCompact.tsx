'use client';

import Link from 'next/link';
import { useLocale, useTranslations } from 'next-intl';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { toast } from 'react-hot-toast';

interface WalletInfo {
  userId: string;
  dailyVotes: number;
  novelRequests: number;
  translationTickets: number;
  nextDailyReset: string;
}

export function WalletCompact() {
  const t = useTranslations('wallet');
  const locale = useLocale();

  const { data: wallet, isLoading } = useQuery<WalletInfo>({
    queryKey: ['wallet'],
    queryFn: async () => {
      const response = await api.get<WalletInfo>('/wallet');
      return response.data;
    },
    retry: false,
  });

  return (
    <div id="wallet" className="bg-background-secondary rounded-card p-6 mb-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
        <h3 className="text-lg font-semibold">{t('title')}</h3>
        <div className="flex items-center gap-2">
          <Link
            href={`/${locale}/voting`}
            className="btn-secondary"
            onClick={(e) => {
              if (!isLoading && (wallet?.dailyVotes ?? 0) < 1) {
                e.preventDefault();
                toast.error(t('noDailyVotesToast'));
              }
            }}
          >
            {t('goToVoting')}
          </Link>
          <Link
            href={`/${locale}/proposals/new`}
            className="btn-primary"
            onClick={(e) => {
              if (!isLoading && (wallet?.novelRequests ?? 0) < 1) {
                e.preventDefault();
                toast.error(t('noNovelRequestsToast'));
              }
            }}
          >
            {t('goToProposal')}
          </Link>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <div className="bg-background-tertiary rounded-card p-4">
          <div className="text-sm text-foreground-muted mb-1" title={t('dailyVotesHint')}>{t('dailyVotes')}</div>
          <div className="text-2xl font-bold">{isLoading ? '—' : wallet?.dailyVotes ?? 0}</div>
        </div>
        <div className="bg-background-tertiary rounded-card p-4">
          <div className="text-sm text-foreground-muted mb-1" title={t('novelRequestsHint')}>{t('novelRequests')}</div>
          <div className="text-2xl font-bold">{isLoading ? '—' : wallet?.novelRequests ?? 0}</div>
        </div>
        <div className="bg-background-tertiary rounded-card p-4">
          <div className="text-sm text-foreground-muted mb-1" title={t('translationTicketsHint')}>{t('translationTickets')}</div>
          <div className="text-2xl font-bold">{isLoading ? '—' : wallet?.translationTickets ?? 0}</div>
        </div>
      </div>
    </div>
  );
}

