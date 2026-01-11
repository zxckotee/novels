'use client';

import { useTranslations, useLocale } from 'next-intl';
import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { useAuthStore } from '@/store/auth';

interface WalletInfo {
  userId: string;
  dailyVotes: number;
  novelRequests: number;
  translationTickets: number;
  nextDailyReset: string;
}

export default function WalletPageClient() {
  const t = useTranslations('wallet');
  const tCommon = useTranslations('common');
  const locale = useLocale();
  const { isAuthenticated } = useAuthStore();

  const { data: wallet, isLoading } = useQuery<WalletInfo>({
    queryKey: ['wallet'],
    queryFn: async () => {
      const response = await api.get<WalletInfo>('/wallet');
      return response.data;
    },
    enabled: isAuthenticated,
    retry: false,
  });

  return (
    <div className="container-custom py-10">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-heading font-bold">{t('title')}</h1>
        <div className="flex items-center gap-2">
          <Link href={`/${locale}/voting`} className="btn-secondary">
            {t('dailyVotes')}
          </Link>
          <Link href={`/${locale}/proposals/new`} className="btn-primary">
            {t('novelRequests')}
          </Link>
        </div>
      </div>

      {!isAuthenticated ? (
        <div className="bg-background-secondary rounded-card p-6">
          <p className="text-foreground-secondary mb-4">{tCommon('loginRequired')}</p>
          <Link href={`/${locale}/login`} className="btn-primary">
            {tCommon('login')}
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-background-secondary rounded-card p-5">
            <div className="text-sm text-foreground-muted mb-1">{t('dailyVotes')}</div>
            <div className="text-3xl font-bold">{isLoading ? '—' : wallet?.dailyVotes ?? 0}</div>
          </div>
          <div className="bg-background-secondary rounded-card p-5">
            <div className="text-sm text-foreground-muted mb-1">{t('novelRequests')}</div>
            <div className="text-3xl font-bold">{isLoading ? '—' : wallet?.novelRequests ?? 0}</div>
          </div>
          <div className="bg-background-secondary rounded-card p-5">
            <div className="text-sm text-foreground-muted mb-1">{t('translationTickets')}</div>
            <div className="text-3xl font-bold">{isLoading ? '—' : wallet?.translationTickets ?? 0}</div>
          </div>
        </div>
      )}
    </div>
  );
}

