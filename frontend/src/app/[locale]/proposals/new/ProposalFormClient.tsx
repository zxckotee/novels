'use client';

import React, { useMemo, useState } from 'react';
import { useTranslations, useLocale } from 'next-intl';
import { useRouter } from 'next/navigation';
import { useMutation, useQuery } from '@tanstack/react-query';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';
import { toast } from 'react-hot-toast';
import Link from 'next/link';

type ProposalSource = 'tadu' | '69shuba' | '101kks';

const SOURCES: Array<{ id: ProposalSource; label: string; allowedHosts: string[] }> = [
  { id: 'tadu', label: 'www.tadu.com', allowedHosts: ['www.tadu.com', 'tadu.com', 'm.tadu.com'] },
  { id: '69shuba', label: 'www.69shuba.com', allowedHosts: ['www.69shuba.com', '69shuba.com'] },
  { id: '101kks', label: '101kks.com', allowedHosts: ['101kks.com'] },
];

const GENRES = [
  'fantasy', 'romance', 'action', 'adventure', 'drama', 'comedy',
  'horror', 'mystery', 'scifi', 'slice_of_life', 'thriller', 'tragedy',
  'martial_arts', 'xianxia', 'xuanhuan', 'wuxia', 'isekai', 'harem',
  'gamelit', 'litrpg', 'system', 'cultivation', 'reincarnation', 'modern'
];

const TAGS = [
  'male_lead', 'female_lead', 'strong_mc', 'weak_to_strong', 'anti_hero',
  'op_mc', 'smart_mc', 'ruthless_mc', 'kind_mc', 'revenge', 'kingdom_building',
  'magic', 'sword_and_sorcery', 'monsters', 'demons', 'gods', 'academy',
  'villainess', 'transmigration', 'regression', 'time_loop', 'slow_romance',
  'fast_paced', 'slow_burn', 'mature', 'no_romance', 'multiple_povs'
];

interface ProposalFormData {
  originalLink: string;
  title: string;
  altTitles: string[];
  description: string;
  genres: string[];
  tags: string[];
  coverUrl: string;
}

interface WalletInfo {
  userId: string;
  dailyVotes: number;
  novelRequests: number;
  translationTickets: number;
  nextDailyReset: string;
}

interface ProposalResponse {
  id: string;
}

export default function ProposalFormClient() {
  const t = useTranslations('proposals');
  const tGenres = useTranslations('genres');
  const tTags = useTranslations('tags');
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user } = useAuth();

  const [source, setSource] = useState<ProposalSource | ''>('');
  const [originalLinkTouched, setOriginalLinkTouched] = useState(false);
  
  const [formData, setFormData] = useState<ProposalFormData>({
    originalLink: '',
    title: '',
    altTitles: [],
    description: '',
    genres: [],
    tags: [],
    coverUrl: '',
  });
  const [altTitleInput, setAltTitleInput] = useState('');
  const [step, setStep] = useState(1);

  const selectedSource = useMemo(() => SOURCES.find((s) => s.id === source), [source]);

  const originalLinkValidation = useMemo(() => {
    const raw = (formData.originalLink || '').trim();
    if (!source) {
      return { valid: false, message: t('newProposal.selectSourceFirst') };
    }
    if (!raw) {
      return { valid: false, message: t('newProposal.originalLinkRequired') };
    }
    const withScheme = /^https?:\/\//i.test(raw) ? raw : `https://${raw}`;
    let u: URL;
    try {
      u = new URL(withScheme);
    } catch {
      return { valid: false, message: t('newProposal.invalidOriginalLink') };
    }
    const host = u.host.toLowerCase();
    const allowedHosts = selectedSource?.allowedHosts ?? [];
    if (!allowedHosts.includes(host)) {
      return {
        valid: false,
        message: t('newProposal.originalLinkHostError', { host: selectedSource?.label || '' }),
      };
    }
    return { valid: true, message: '' };
  }, [formData.originalLink, selectedSource?.allowedHosts, selectedSource?.label, source, t]);

  // Check wallet for Novel Request tickets
  const { data: wallet, isLoading: walletLoading } = useQuery<WalletInfo>({
    queryKey: ['wallet'],
    queryFn: async () => {
      const response = await api.get<WalletInfo>('/wallet');
      return response.data;
    },
    enabled: isAuthenticated,
  });

  // Backend enforces ticket balance; don't bypass with "premium" role here.
  const hasTicket = (wallet?.novelRequests ?? 0) > 0;

  // Submit proposal
  const submitMutation = useMutation<ProposalResponse, unknown, ProposalFormData>({
    mutationFn: async (data: ProposalFormData) => {
      const response = await api.post<ProposalResponse>('/proposals', data);
      return response.data;
    },
    onSuccess: (data) => {
      toast.success(t('newProposal.success'));
      router.push(`/${locale}/voting`);
    },
    onError: (error: any) => {
      const status = error?.response?.status;
      const backendMessage = error?.response?.data?.error?.message;
      if (status === 402) {
        toast.error(t('newProposal.noTicket'));
        return;
      }
      toast.error(backendMessage || t('newProposal.error'));
    },
  });

  const handleAddAltTitle = () => {
    if (altTitleInput.trim() && formData.altTitles.length < 5) {
      setFormData({
        ...formData,
        altTitles: [...formData.altTitles, altTitleInput.trim()],
      });
      setAltTitleInput('');
    }
  };

  const handleRemoveAltTitle = (index: number) => {
    setFormData({
      ...formData,
      altTitles: formData.altTitles.filter((_, i) => i !== index),
    });
  };

  const handleToggleGenre = (genre: string) => {
    if (formData.genres.includes(genre)) {
      setFormData({
        ...formData,
        genres: formData.genres.filter((g) => g !== genre),
      });
    } else if (formData.genres.length < 5) {
      setFormData({
        ...formData,
        genres: [...formData.genres, genre],
      });
    }
  };

  const handleToggleTag = (tag: string) => {
    if (formData.tags.includes(tag)) {
      setFormData({
        ...formData,
        tags: formData.tags.filter((t) => t !== tag),
      });
    } else if (formData.tags.length < 10) {
      setFormData({
        ...formData,
        tags: [...formData.tags, tag],
      });
    }
  };

  const handleSubmit = () => {
    if (!source) {
      toast.error(t('newProposal.selectSourceFirst'));
      return;
    }
    if (!formData.originalLink || !formData.title || !formData.description) {
      toast.error(t('newProposal.fillRequired'));
      return;
    }
    if (!originalLinkValidation.valid) {
      toast.error(originalLinkValidation.message || t('newProposal.invalidOriginalLink'));
      return;
    }
    if (formData.genres.length === 0) {
      toast.error(t('newProposal.selectGenre'));
      return;
    }
    submitMutation.mutate(formData);
  };

  const isStep1Valid = Boolean(source) && originalLinkValidation.valid && Boolean(formData.title);
  const isStep2Valid = formData.description.length >= 100;
  const isStep3Valid = formData.genres.length > 0;

  if (!isAuthenticated) {
    return (
      <div className="container mx-auto px-4 py-16 text-center">
        <h1 className="text-2xl font-bold text-foreground-primary mb-4">{t('newProposal.loginRequired')}</h1>
        <p className="text-foreground-secondary mb-8">{t('newProposal.loginMessage')}</p>
        <Link href={`/${locale}/login`} className="btn-primary">
          {t('newProposal.login')}
        </Link>
      </div>
    );
  }

  if (walletLoading) {
    return (
      <div className="container mx-auto px-4 py-16 text-center">
        <p className="text-foreground-secondary">{t('newProposal.submitting')}</p>
      </div>
    );
  }

  if (!hasTicket) {
    return (
      <div className="container mx-auto px-4 py-16 text-center">
        <div className="max-w-md mx-auto">
          <div className="w-20 h-20 bg-accent-warning/20 rounded-full flex items-center justify-center mx-auto mb-6">
            <svg className="w-10 h-10 text-accent-warning" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-foreground-primary mb-4">{t('newProposal.noTicket')}</h1>
          <p className="text-foreground-secondary mb-8">{t('newProposal.noTicketMessage')}</p>
          <div className="flex gap-4 justify-center">
            <Link href={`/${locale}/subscriptions`} className="btn-primary">
              {t('newProposal.getPremium')}
            </Link>
            <Link href={`/${locale}/voting`} className="btn-secondary">
              {t('newProposal.backToVoting')}
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-3xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground-primary">{t('newProposal.title')}</h1>
        <p className="text-foreground-secondary mt-2">{t('newProposal.subtitle')}</p>
      </div>

      {/* Progress Steps */}
      <div className="bg-background-secondary rounded-xl p-6 mb-8">
        <div className="flex items-center w-full gap-2">
          {[1, 2, 3, 4].map((s) => (
            <React.Fragment key={s}>
              <div
                className={`w-10 h-10 rounded-full flex items-center justify-center font-bold flex-shrink-0 ${
                  step >= s
                    ? 'bg-accent-primary text-white'
                    : 'bg-background-tertiary text-foreground-muted'
                }`}
              >
                {s}
              </div>
              {s < 4 && (
                <div
                  className={`flex-1 h-1 ${
                    step > s ? 'bg-accent-primary' : 'bg-background-tertiary'
                  }`}
                />
              )}
            </React.Fragment>
          ))}
        </div>
      </div>

      {/* Step 1: Basic Info */}
      {step === 1 && (
        <div className="bg-background-secondary rounded-xl p-6 space-y-6">
          <h2 className="text-xl font-semibold text-foreground-primary">{t('newProposal.step1Title')}</h2>

          <div>
            <label className="block text-sm font-medium text-foreground-secondary mb-2">
              {t('newProposal.source')} *
            </label>
            <select
              value={source}
              onChange={(e) => {
                const next = (e.target.value || '') as ProposalSource | '';
                setSource(next);
                // Reset link validation UX when switching sources
                setOriginalLinkTouched(false);
                // If link is present but doesn't match the new source, clear it to enforce correctness.
                if (formData.originalLink.trim()) {
                  setFormData({ ...formData, originalLink: '' });
                }
              }}
              className="input w-full"
            >
              <option value="">{t('newProposal.sourcePlaceholder')}</option>
              {SOURCES.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.label}
                </option>
              ))}
            </select>
            <p className="text-xs text-foreground-muted mt-1">
              Поддерживаются: www.tadu.com, www.69shuba.com, 101kks.com
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground-secondary mb-2">
              {t('newProposal.originalLink')} *
            </label>
            <input
              type="url"
              value={formData.originalLink}
              onChange={(e) => {
                setFormData({ ...formData, originalLink: e.target.value });
                setOriginalLinkTouched(true);
              }}
              onBlur={() => setOriginalLinkTouched(true)}
              placeholder="https://..."
              className="input w-full"
              disabled={!source}
            />
            {originalLinkTouched && !originalLinkValidation.valid ? (
              <p className="text-xs text-red-400 mt-1">{originalLinkValidation.message}</p>
            ) : (
              <p className="text-xs text-foreground-muted mt-1">{t('newProposal.originalLinkHint')}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground-secondary mb-2">
              {t('newProposal.novelTitle')} *
            </label>
            <input
              type="text"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              placeholder={t('newProposal.novelTitlePlaceholder')}
              className="input w-full"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground-secondary mb-2">
              {t('newProposal.altTitles')}
            </label>
            <div className="flex gap-2 mb-2">
              <input
                type="text"
                value={altTitleInput}
                onChange={(e) => setAltTitleInput(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), handleAddAltTitle())}
                placeholder={t('newProposal.altTitlePlaceholder')}
                className="input flex-1"
              />
              <button
                onClick={handleAddAltTitle}
                disabled={!altTitleInput.trim() || formData.altTitles.length >= 5}
                className="btn-primary"
              >
                {t('newProposal.add')}
              </button>
            </div>
            <div className="flex flex-wrap gap-2">
              {formData.altTitles.map((title, index) => (
                <span
                  key={index}
                  className="px-3 py-1 bg-background-tertiary rounded-full text-sm flex items-center gap-2"
                >
                  {title}
                  <button onClick={() => handleRemoveAltTitle(index)} className="text-foreground-muted hover:text-red-500">
                    ×
                  </button>
                </span>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground-secondary mb-2">
              {t('newProposal.coverUrl')}
            </label>
            <input
              type="url"
              value={formData.coverUrl}
              onChange={(e) => setFormData({ ...formData, coverUrl: e.target.value })}
              placeholder="https://..."
              className="input w-full"
            />
            {formData.coverUrl ? (
              <div className="mt-3 w-full max-w-[180px] aspect-[3/4] bg-background-tertiary rounded-lg overflow-hidden border border-background-tertiary">
                <img
                  src={formData.coverUrl}
                  alt=""
                  className="w-full h-full object-cover"
                  loading="lazy"
                  decoding="async"
                  onError={(e) => {
                    (e.currentTarget as HTMLImageElement).src = '/placeholder-cover.svg';
                  }}
                />
              </div>
            ) : null}
          </div>

          <div className="flex justify-end">
            <button
              onClick={() => setStep(2)}
              disabled={!isStep1Valid}
              className="btn-primary"
            >
              {t('newProposal.next')}
            </button>
          </div>
        </div>
      )}

      {/* Step 2: Description */}
      {step === 2 && (
        <div className="bg-background-secondary rounded-xl p-6 space-y-6">
          <h2 className="text-xl font-semibold text-foreground-primary">{t('newProposal.step2Title')}</h2>
          
          <div>
            <label className="block text-sm font-medium text-foreground-secondary mb-2">
              {t('newProposal.description')} *
            </label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder={t('newProposal.descriptionPlaceholder')}
              rows={8}
              className="input w-full resize-none"
            />
            <p className={`text-xs mt-1 ${formData.description.length >= 100 ? 'text-status-success' : 'text-foreground-muted'}`}>
              {formData.description.length}/100 {t('newProposal.minChars')}
            </p>
          </div>

          <div className="flex justify-between">
            <button
              onClick={() => setStep(1)}
              className="btn-secondary"
            >
              {t('newProposal.back')}
            </button>
            <button
              onClick={() => setStep(3)}
              disabled={!isStep2Valid}
              className="btn-primary"
            >
              {t('newProposal.next')}
            </button>
          </div>
        </div>
      )}

      {/* Step 3: Genres & Tags */}
      {step === 3 && (
        <div className="bg-background-secondary rounded-xl p-6 space-y-6">
          <h2 className="text-xl font-semibold text-foreground-primary">{t('newProposal.step3Title')}</h2>
          
          <div>
            <label className="block text-sm font-medium text-foreground-secondary mb-3">
              {t('newProposal.genres')} * ({formData.genres.length}/5)
            </label>
            <div className="flex flex-wrap gap-2">
              {GENRES.map((genre) => (
                <button
                  key={genre}
                  onClick={() => handleToggleGenre(genre)}
                  className={`px-3 py-1.5 rounded-full text-sm transition-colors ${
                    formData.genres.includes(genre)
                      ? 'bg-accent-primary text-white'
                      : 'bg-background-tertiary text-foreground-secondary hover:bg-background-hover'
                  }`}
                >
                  {tGenres(genre)}
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground-secondary mb-3">
              {t('newProposal.tags')} ({formData.tags.length}/10)
            </label>
            <div className="flex flex-wrap gap-2">
              {TAGS.map((tag) => (
                <button
                  key={tag}
                  onClick={() => handleToggleTag(tag)}
                  className={`px-3 py-1.5 rounded-full text-sm transition-colors ${
                    formData.tags.includes(tag)
                      ? 'bg-accent-secondary text-white'
                      : 'bg-background-tertiary text-foreground-secondary hover:bg-background-hover'
                  }`}
                >
                  {tTags(tag)}
                </button>
              ))}
            </div>
          </div>

          <div className="flex justify-between">
            <button
              onClick={() => setStep(2)}
              className="btn-secondary"
            >
              {t('newProposal.back')}
            </button>
            <button
              onClick={() => setStep(4)}
              disabled={!isStep3Valid}
              className="btn-primary"
            >
              {t('newProposal.next')}
            </button>
          </div>
        </div>
      )}

      {/* Step 4: Review */}
      {step === 4 && (
        <div className="bg-background-secondary rounded-xl p-6 space-y-6">
          <h2 className="text-xl font-semibold text-foreground-primary">{t('newProposal.step4Title')}</h2>
          
          <div className="space-y-4">
            <div className="flex gap-4">
              {formData.coverUrl ? (
                <img src={formData.coverUrl} alt="" className="w-24 h-32 object-cover rounded-lg" />
              ) : (
                <div className="w-24 h-32 bg-background-tertiary rounded-lg flex items-center justify-center">
                  <span className="text-foreground-muted text-xs">{t('newProposal.noCover')}</span>
                </div>
              )}
              <div>
                <h3 className="text-lg font-semibold text-foreground-primary">{formData.title}</h3>
                <a href={formData.originalLink} target="_blank" rel="noopener noreferrer" className="text-accent-primary text-sm hover:underline">
                  {t('newProposal.viewOriginal')}
                </a>
              </div>
            </div>

            {formData.altTitles.length > 0 && (
              <div>
                <p className="text-sm text-foreground-muted">{t('newProposal.altTitles')}:</p>
                <p className="text-foreground-secondary">{formData.altTitles.join(', ')}</p>
              </div>
            )}

            <div>
              <p className="text-sm text-foreground-muted">{t('newProposal.description')}:</p>
              <p className="text-foreground-secondary line-clamp-4">{formData.description}</p>
            </div>

            <div>
              <p className="text-sm text-foreground-muted mb-2">{t('newProposal.genres')}:</p>
              <div className="flex flex-wrap gap-2">
                {formData.genres.map((genre) => (
                  <span key={genre} className="px-2 py-1 bg-accent-primary/20 text-accent-primary rounded-full text-xs">
                    {tGenres(genre)}
                  </span>
                ))}
              </div>
            </div>

            {formData.tags.length > 0 && (
              <div>
                <p className="text-sm text-foreground-muted mb-2">{t('newProposal.tags')}:</p>
                <div className="flex flex-wrap gap-2">
                  {formData.tags.map((tag) => (
                    <span key={tag} className="px-2 py-1 bg-accent-secondary/20 text-accent-secondary rounded-full text-xs">
                      {tTags(tag)}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>

          <div className="bg-accent-warning/10 border border-accent-warning/20 rounded-lg p-4">
            <p className="text-sm text-accent-warning">
              {t('newProposal.ticketWarning')}
            </p>
          </div>

          <div className="flex justify-between">
            <button
              onClick={() => setStep(3)}
              className="btn-secondary"
            >
              {t('newProposal.back')}
            </button>
            <button
              onClick={handleSubmit}
              disabled={submitMutation.isPending}
              className="btn-primary"
            >
              {submitMutation.isPending ? t('newProposal.submitting') : t('newProposal.submit')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
