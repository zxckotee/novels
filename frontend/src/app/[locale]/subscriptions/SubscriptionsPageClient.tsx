'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';
import { toast } from 'react-hot-toast';
import Link from 'next/link';

interface SubscriptionPlan {
  id: string;
  code: string;
  title: string;
  description: string;
  price: number;
  currency: string;
  period: string;
  features: {
    noAds: boolean;
    dailyVoteMultiplier: number;
    monthlyNovelRequests: number;
    monthlyTranslationTickets: number;
    canEditDescriptions: boolean;
    canRequestRetranslate: boolean;
  };
  isActive: boolean;
}

interface UserSubscriptionInfo {
  hasActiveSubscription: boolean;
  subscription?: {
    id: string;
    userId: string;
    planId: string;
    status: string;
    startsAt: string;
    endsAt: string;
    autoRenew: boolean;
  };
  plan?: SubscriptionPlan;
  features?: {
    noAds: boolean;
    dailyVoteMultiplier: number;
    monthlyNovelRequests: number;
    monthlyTranslationTickets: number;
    canEditDescriptions: boolean;
    canRequestRetranslate: boolean;
  };
  daysRemaining?: number;
}

export default function SubscriptionsPageClient() {
  const t = useTranslations('subscriptions');
  const { isAuthenticated, user } = useAuth();
  const queryClient = useQueryClient();
  const [selectedPlan, setSelectedPlan] = useState<string | null>(null);
  const [billingPeriod, setBillingPeriod] = useState<'monthly' | 'yearly'>('monthly');

  // Fetch plans
  const { data: plans, isLoading: plansLoading } = useQuery<SubscriptionPlan[]>({
    queryKey: ['subscription-plans'],
    queryFn: async () => {
      const response = await api.get<SubscriptionPlan[]>('/subscriptions/plans');
      return response.data;
    },
  });

  // Fetch current subscription
  const { data: subInfo, isLoading: subLoading } = useQuery<UserSubscriptionInfo>({
    queryKey: ['my-subscription'],
    queryFn: async () => {
      const response = await api.get<UserSubscriptionInfo>('/subscriptions/me');
      return response.data;
    },
    enabled: isAuthenticated,
  });

  // Subscribe mutation
  const subscribeMutation = useMutation({
    mutationFn: async (planId: string) => {
      const response = await api.post('/subscriptions', { planId });
      return response.data;
    },
    onSuccess: () => {
      toast.success(t('subscribeSuccess'));
      queryClient.invalidateQueries({ queryKey: ['my-subscription'] });
      queryClient.invalidateQueries({ queryKey: ['wallet'] });
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || t('subscribeError'));
    },
  });

  // Cancel subscription mutation
  const cancelMutation = useMutation({
    mutationFn: async () => {
      const response = await api.post('/subscriptions/cancel');
      return response.data;
    },
    onSuccess: () => {
      toast.success(t('cancelSuccess'));
      queryClient.invalidateQueries({ queryKey: ['my-subscription'] });
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || t('cancelError'));
    },
  });

  const handleSubscribe = (planId: string) => {
    if (!isAuthenticated) {
      toast.error(t('loginRequired'));
      return;
    }
    subscribeMutation.mutate(planId);
  };

  const formatPrice = (price: number, currency: string) => {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency: currency || 'RUB',
    }).format(price);
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  // Determine current plan code
  const currentPlanCode = subInfo?.hasActiveSubscription && subInfo.plan?.code ? subInfo.plan.code : 'free';

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold text-text-primary mb-4">{t('title')}</h1>
        <p className="text-text-secondary text-lg max-w-2xl mx-auto">{t('subtitle')}</p>
      </div>

      {/* Current Subscription Banner */}
      {subInfo?.hasActiveSubscription && subInfo.subscription && subInfo.plan && (
        <div className="bg-gradient-to-r from-primary/20 to-purple-500/20 rounded-xl p-6 mb-8">
          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
            <div>
              <p className="text-sm text-text-secondary">{t('currentPlan')}</p>
              <h3 className="text-2xl font-bold text-text-primary">{subInfo.plan.title}</h3>
              <p className="text-text-secondary">
                {t('validUntil', { date: formatDate(subInfo.subscription.endsAt) })}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Billing Period Toggle */}
      <div className="flex justify-center mb-8">
        <div className="bg-surface-elevated rounded-full p-1 flex">
          <button
            onClick={() => setBillingPeriod('monthly')}
            className={`px-6 py-2 rounded-full transition-colors ${
              billingPeriod === 'monthly'
                ? 'bg-primary text-white'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            {t('monthly')}
          </button>
          <button
            onClick={() => setBillingPeriod('yearly')}
            className={`px-6 py-2 rounded-full transition-colors ${
              billingPeriod === 'yearly'
                ? 'bg-primary text-white'
                : 'text-text-secondary hover:text-text-primary'
            }`}
          >
            {t('yearly')}
            <span className="ml-2 text-xs text-green-500">{t('save20')}</span>
          </button>
        </div>
      </div>

      {/* Plans Grid */}
      <div className="grid md:grid-cols-3 gap-6 mb-12">
        {/* Free Plan */}
        <div className="bg-surface-elevated rounded-xl p-6 border-2 border-transparent">
          <div className="mb-6">
            <h3 className="text-xl font-bold text-text-primary">{t('plans.free.title')}</h3>
            <p className="text-text-secondary text-sm mt-1">{t('plans.free.description')}</p>
          </div>
          <div className="mb-6">
            <span className="text-4xl font-bold text-text-primary">{t('plans.free.price')}</span>
          </div>
          <ul className="space-y-3 mb-8">
            <li className="flex items-center gap-2 text-text-secondary">
              <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              {t('features.readAll')}
            </li>
            <li className="flex items-center gap-2 text-text-secondary">
              <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              {t('features.dailyVote1')}
            </li>
            <li className="flex items-center gap-2 text-text-secondary">
              <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              {t('features.comments')}
            </li>
            <li className="flex items-center gap-2 text-text-muted">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
              {t('features.noAds')}
            </li>
            <li className="flex items-center gap-2 text-text-muted">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
              {t('features.novelRequests')}
            </li>
          </ul>
          <button
            disabled
            className="w-full py-3 bg-surface text-text-muted rounded-lg cursor-not-allowed"
          >
            {currentPlanCode === 'free' ? t('currentPlanButton') : t('plans.free.title')}
          </button>
        </div>

        {/* Premium Plan */}
        <div className="bg-surface-elevated rounded-xl p-6 border-2 border-primary relative">
          <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
            <span className="px-3 py-1 bg-primary text-white text-sm font-medium rounded-full">
              {t('popular')}
            </span>
          </div>
          <div className="mb-6">
            <h3 className="text-xl font-bold text-text-primary">{t('plans.premium.title')}</h3>
            <p className="text-text-secondary text-sm mt-1">{t('plans.premium.description')}</p>
          </div>
          <div className="mb-6">
            <span className="text-4xl font-bold text-text-primary">
              {billingPeriod === 'monthly' ? '299 ₽' : '2 390 ₽'}
            </span>
            <span className="text-text-secondary">/{billingPeriod === 'monthly' ? t('month') : t('year')}</span>
          </div>
          <ul className="space-y-3 mb-8">
            <li className="flex items-center gap-2 text-text-secondary">
              <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              {t('features.readAll')}
            </li>
            <li className="flex items-center gap-2 text-text-secondary">
              <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              {t('features.noAds')}
            </li>
            <li className="flex items-center gap-2 text-text-secondary">
              <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              {t('features.dailyVote2')}
            </li>
            <li className="flex items-center gap-2 text-text-secondary">
              <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              {t('features.novelRequests2')}
            </li>
            <li className="flex items-center gap-2 text-text-secondary">
              <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              {t('features.translationTickets5')}
            </li>
          </ul>
          <button
            onClick={() => handleSubscribe('premium')}
            disabled={subscribeMutation.isPending || currentPlanCode === 'premium'}
            className={`w-full py-3 rounded-lg transition-colors ${
              currentPlanCode === 'premium'
                ? 'bg-surface text-text-muted cursor-not-allowed'
                : 'bg-primary text-white hover:bg-primary-hover disabled:opacity-50'
            }`}
          >
            {currentPlanCode === 'premium'
              ? t('currentPlanButton')
              : subscribeMutation.isPending ? '...' : t('subscribe')}
          </button>
        </div>

        {/* VIP Plan */}
        <div className="bg-surface-elevated rounded-xl p-6 border-2 border-yellow-500/50">
          <div className="mb-6">
            <h3 className="text-xl font-bold text-text-primary flex items-center gap-2">
              {t('plans.vip.title')}
              <span className="text-yellow-500">👑</span>
            </h3>
            <p className="text-text-secondary text-sm mt-1">{t('plans.vip.description')}</p>
          </div>
          <div className="mb-6">
            <span className="text-4xl font-bold text-text-primary">
              {billingPeriod === 'monthly' ? '799 ₽' : '6 390 ₽'}
            </span>
            <span className="text-text-secondary">/{billingPeriod === 'monthly' ? t('month') : t('year')}</span>
          </div>
          <ul className="space-y-3 mb-8">
            <li className="flex items-center gap-2 text-text-secondary">
              <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              {t('features.allPremium')}
            </li>
            <li className="flex items-center gap-2 text-text-secondary">
              <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              {t('features.dailyVote5')}
            </li>
            <li className="flex items-center gap-2 text-text-secondary">
              <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              {t('features.novelRequests5')}
            </li>
            <li className="flex items-center gap-2 text-text-secondary">
              <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              {t('features.translationTickets15')}
            </li>
            <li className="flex items-center gap-2 text-text-secondary">
              <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              {t('features.editDescriptions')}
            </li>
            <li className="flex items-center gap-2 text-text-secondary">
              <svg className="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              {t('features.retranslate')}
            </li>
          </ul>
          <button
            onClick={() => handleSubscribe('vip')}
            disabled={subscribeMutation.isPending || currentPlanCode === 'vip'}
            className={`w-full py-3 rounded-lg transition-colors ${
              currentPlanCode === 'vip'
                ? 'bg-surface text-text-muted cursor-not-allowed'
                : 'bg-gradient-to-r from-yellow-500 to-orange-500 text-white hover:opacity-90 disabled:opacity-50'
            }`}
          >
            {currentPlanCode === 'vip'
              ? t('currentPlanButton')
              : subscribeMutation.isPending ? '...' : t('subscribe')}
          </button>
        </div>
      </div>

      {/* Features Comparison */}
      <div className="bg-surface-elevated rounded-xl p-6 mb-12">
        <h2 className="text-2xl font-bold text-text-primary mb-6 text-center">{t('comparison.title')}</h2>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border">
                <th className="text-left py-4 pr-4 text-text-secondary font-medium">{t('comparison.feature')}</th>
                <th className="text-center py-4 px-4 text-text-secondary font-medium">{t('plans.free.title')}</th>
                <th className="text-center py-4 px-4 text-primary font-medium">{t('plans.premium.title')}</th>
                <th className="text-center py-4 px-4 text-yellow-500 font-medium">{t('plans.vip.title')}</th>
              </tr>
            </thead>
            <tbody>
              <tr className="border-b border-border/50">
                <td className="py-4 pr-4 text-text-secondary">{t('comparison.dailyVotes')}</td>
                <td className="text-center py-4 px-4 text-text-primary">1</td>
                <td className="text-center py-4 px-4 text-text-primary">2</td>
                <td className="text-center py-4 px-4 text-text-primary">5</td>
              </tr>
              <tr className="border-b border-border/50">
                <td className="py-4 pr-4 text-text-secondary">{t('comparison.novelRequests')}</td>
                <td className="text-center py-4 px-4 text-text-muted">—</td>
                <td className="text-center py-4 px-4 text-text-primary">2/{t('week')}</td>
                <td className="text-center py-4 px-4 text-text-primary">5/{t('week')}</td>
              </tr>
              <tr className="border-b border-border/50">
                <td className="py-4 pr-4 text-text-secondary">{t('comparison.translationTickets')}</td>
                <td className="text-center py-4 px-4 text-text-muted">—</td>
                <td className="text-center py-4 px-4 text-text-primary">5/{t('week')}</td>
                <td className="text-center py-4 px-4 text-text-primary">15/{t('week')}</td>
              </tr>
              <tr className="border-b border-border/50">
                <td className="py-4 pr-4 text-text-secondary">{t('comparison.noAds')}</td>
                <td className="text-center py-4 px-4">
                  <span className="text-red-500">✕</span>
                </td>
                <td className="text-center py-4 px-4">
                  <span className="text-green-500">✓</span>
                </td>
                <td className="text-center py-4 px-4">
                  <span className="text-green-500">✓</span>
                </td>
              </tr>
              <tr className="border-b border-border/50">
                <td className="py-4 pr-4 text-text-secondary">{t('comparison.editDescriptions')}</td>
                <td className="text-center py-4 px-4">
                  <span className="text-red-500">✕</span>
                </td>
                <td className="text-center py-4 px-4">
                  <span className="text-red-500">✕</span>
                </td>
                <td className="text-center py-4 px-4">
                  <span className="text-green-500">✓</span>
                </td>
              </tr>
              <tr>
                <td className="py-4 pr-4 text-text-secondary">{t('comparison.retranslate')}</td>
                <td className="text-center py-4 px-4">
                  <span className="text-red-500">✕</span>
                </td>
                <td className="text-center py-4 px-4">
                  <span className="text-red-500">✕</span>
                </td>
                <td className="text-center py-4 px-4">
                  <span className="text-green-500">✓</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      {/* FAQ */}
      <div className="max-w-3xl mx-auto">
        <h2 className="text-2xl font-bold text-text-primary mb-6 text-center">{t('faq.title')}</h2>
        <div className="space-y-4">
          {['q1', 'q2', 'q3', 'q4'].map((q) => (
            <div key={q} className="bg-surface-elevated rounded-xl p-6">
              <h3 className="font-semibold text-text-primary mb-2">{t(`faq.${q}.question`)}</h3>
              <p className="text-text-secondary">{t(`faq.${q}.answer`)}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
