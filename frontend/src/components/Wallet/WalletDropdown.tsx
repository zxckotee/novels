'use client';

import { useState, useRef, useEffect } from 'react';
import { useTranslations } from 'next-intl';
import { useQuery } from '@tanstack/react-query';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';

interface WalletInfo {
  userId: string;
  dailyVotes: number;
  novelRequests: number;
  translationTickets: number;
  nextDailyReset: string;
}

export function WalletDropdown() {
  const t = useTranslations('wallet');
  const { user, isAuthenticated } = useAuth();
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const { data: wallet, isLoading } = useQuery<WalletInfo>({
    queryKey: ['wallet'],
    queryFn: async () => {
      const response = await api.get('/wallet');
      return response.data;
    },
    enabled: isAuthenticated,
    refetchInterval: 60000, // Refresh every minute
  });

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  if (!isAuthenticated) {
    return null;
  }

  const totalTickets = wallet
    ? wallet.dailyVotes + wallet.novelRequests + wallet.translationTickets
    : 0;

  const formatTimeUntilReset = (resetTime: string) => {
    const reset = new Date(resetTime);
    const now = new Date();
    const diff = reset.getTime() - now.getTime();

    if (diff <= 0) return t('resetNow');

    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

    if (hours > 0) {
      return t('resetIn', { hours, minutes });
    }
    return t('resetInMinutes', { minutes });
  };

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-2 px-3 py-2 rounded-lg bg-surface-elevated hover:bg-surface-muted transition-colors"
        aria-label={t('title')}
      >
        {/* Wallet Icon */}
        <svg
          className="w-5 h-5 text-primary"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z"
          />
        </svg>

        {/* Total count */}
        {isLoading ? (
          <span className="w-8 h-4 bg-surface-muted animate-pulse rounded" />
        ) : (
          <span className="font-medium text-text-primary">{totalTickets}</span>
        )}

        {/* Dropdown arrow */}
        <svg
          className={`w-4 h-4 text-text-secondary transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {/* Dropdown menu */}
      {isOpen && (
        <div className="absolute right-0 mt-2 w-72 bg-surface-elevated rounded-xl shadow-lg border border-border z-50">
          <div className="p-4">
            <h3 className="text-lg font-semibold text-text-primary mb-3">{t('title')}</h3>

            {/* Ticket balances */}
            <div className="space-y-3">
              {/* Daily Votes */}
              <div className="flex items-center justify-between p-3 bg-surface rounded-lg">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-blue-500/20 flex items-center justify-center">
                    <svg className="w-5 h-5 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M2 10.5a1.5 1.5 0 113 0v6a1.5 1.5 0 01-3 0v-6zM6 10.333v5.43a2 2 0 001.106 1.79l.05.025A4 4 0 008.943 18h5.416a2 2 0 001.962-1.608l1.2-6A2 2 0 0015.56 8H12V4a2 2 0 00-2-2 1 1 0 00-1 1v.667a4 4 0 01-.8 2.4L6.8 7.933a4 4 0 00-.8 2.4z" />
                    </svg>
                  </div>
                  <div>
                    <p className="text-sm text-text-secondary">{t('dailyVotes')}</p>
                    <p className="text-lg font-bold text-text-primary">{wallet?.dailyVotes ?? 0}</p>
                  </div>
                </div>
                <Link
                  href="/voting"
                  className="px-3 py-1 text-sm bg-blue-500 text-white rounded-full hover:bg-blue-600 transition-colors"
                  onClick={() => setIsOpen(false)}
                >
                  {t('vote')}
                </Link>
              </div>

              {/* Novel Requests */}
              <div className="flex items-center justify-between p-3 bg-surface rounded-lg">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-purple-500/20 flex items-center justify-center">
                    <svg className="w-5 h-5 text-purple-500" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-11a1 1 0 10-2 0v2H7a1 1 0 100 2h2v2a1 1 0 102 0v-2h2a1 1 0 100-2h-2V7z" clipRule="evenodd" />
                    </svg>
                  </div>
                  <div>
                    <p className="text-sm text-text-secondary">{t('novelRequests')}</p>
                    <p className="text-lg font-bold text-text-primary">{wallet?.novelRequests ?? 0}</p>
                  </div>
                </div>
                <Link
                  href="/proposals/new"
                  className="px-3 py-1 text-sm bg-purple-500 text-white rounded-full hover:bg-purple-600 transition-colors"
                  onClick={() => setIsOpen(false)}
                >
                  {t('propose')}
                </Link>
              </div>

              {/* Translation Tickets */}
              <div className="flex items-center justify-between p-3 bg-surface rounded-lg">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-full bg-green-500/20 flex items-center justify-center">
                    <svg className="w-5 h-5 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M7 2a1 1 0 011 1v1h3a1 1 0 110 2H9.578a18.87 18.87 0 01-1.724 4.78c.29.354.596.696.914 1.026a1 1 0 11-1.44 1.389c-.188-.196-.373-.396-.554-.6a19.098 19.098 0 01-3.107 3.567 1 1 0 01-1.334-1.49 17.087 17.087 0 003.13-3.733 18.992 18.992 0 01-1.487-2.494 1 1 0 111.79-.89c.234.47.489.928.764 1.372.417-.934.752-1.913.997-2.927H3a1 1 0 110-2h3V3a1 1 0 011-1zm6 6a1 1 0 01.894.553l2.991 5.982a.869.869 0 01.02.037l.99 1.98a1 1 0 11-1.79.895L15.383 16h-4.764l-.724 1.447a1 1 0 11-1.788-.894l.99-1.98.019-.038 2.99-5.982A1 1 0 0113 8zm-1.382 6h2.764L13 11.236 11.618 14z" clipRule="evenodd" />
                    </svg>
                  </div>
                  <div>
                    <p className="text-sm text-text-secondary">{t('translationTickets')}</p>
                    <p className="text-lg font-bold text-text-primary">{wallet?.translationTickets ?? 0}</p>
                  </div>
                </div>
              </div>
            </div>

            {/* Reset timer */}
            {wallet?.nextDailyReset && (
              <div className="mt-4 p-3 bg-surface rounded-lg">
                <p className="text-xs text-text-muted text-center">
                  {t('nextReset')}: {formatTimeUntilReset(wallet.nextDailyReset)}
                </p>
              </div>
            )}

            {/* Links */}
            <div className="mt-4 pt-4 border-t border-border space-y-2">
              <Link
                href="/wallet/transactions"
                className="flex items-center gap-2 px-3 py-2 text-sm text-text-secondary hover:text-text-primary hover:bg-surface rounded-lg transition-colors"
                onClick={() => setIsOpen(false)}
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                </svg>
                {t('transactionHistory')}
              </Link>
              <Link
                href="/subscriptions"
                className="flex items-center gap-2 px-3 py-2 text-sm text-text-secondary hover:text-text-primary hover:bg-surface rounded-lg transition-colors"
                onClick={() => setIsOpen(false)}
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z" />
                </svg>
                {t('getPremium')}
              </Link>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
