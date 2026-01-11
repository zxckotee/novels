'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { useRouter } from 'next/navigation';
import { useMutation, useQuery } from '@tanstack/react-query';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';
import { toast } from 'react-hot-toast';
import Link from 'next/link';

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
  author: string;
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
  const router = useRouter();
  const { isAuthenticated, user } = useAuth();
  
  const [formData, setFormData] = useState<ProposalFormData>({
    originalLink: '',
    title: '',
    altTitles: [],
    author: '',
    description: '',
    genres: [],
    tags: [],
    coverUrl: '',
  });
  const [altTitleInput, setAltTitleInput] = useState('');
  const [step, setStep] = useState(1);

  // Check wallet for Novel Request tickets
  const { data: wallet } = useQuery<WalletInfo>({
    queryKey: ['wallet'],
    queryFn: async () => {
      const response = await api.get<WalletInfo>('/wallet');
      return response.data;
    },
    enabled: isAuthenticated,
  });

  const hasTicket = (wallet?.novelRequests ?? 0) > 0;

  // Submit proposal
  const submitMutation = useMutation<ProposalResponse, unknown, ProposalFormData>({
    mutationFn: async (data: ProposalFormData) => {
      const response = await api.post<ProposalResponse>('/proposals', data);
      return response.data;
    },
    onSuccess: (data) => {
      toast.success(t('newProposal.success'));
      router.push(`/proposals/${data.id}`);
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.message || t('newProposal.error'));
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
    if (!formData.originalLink || !formData.title || !formData.description) {
      toast.error(t('newProposal.fillRequired'));
      return;
    }
    if (formData.genres.length === 0) {
      toast.error(t('newProposal.selectGenre'));
      return;
    }
    submitMutation.mutate(formData);
  };

  const isStep1Valid = formData.originalLink && formData.title && formData.author;
  const isStep2Valid = formData.description.length >= 100;
  const isStep3Valid = formData.genres.length > 0;

  if (!isAuthenticated) {
    return (
      <div className="container mx-auto px-4 py-16 text-center">
        <h1 className="text-2xl font-bold text-text-primary mb-4">{t('newProposal.loginRequired')}</h1>
        <p className="text-text-secondary mb-8">{t('newProposal.loginMessage')}</p>
        <Link href="/auth/login" className="px-6 py-3 bg-primary text-white rounded-lg">
          {t('newProposal.login')}
        </Link>
      </div>
    );
  }

  if (!hasTicket) {
    return (
      <div className="container mx-auto px-4 py-16 text-center">
        <div className="max-w-md mx-auto">
          <div className="w-20 h-20 bg-yellow-500/20 rounded-full flex items-center justify-center mx-auto mb-6">
            <svg className="w-10 h-10 text-yellow-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <h1 className="text-2xl font-bold text-text-primary mb-4">{t('newProposal.noTicket')}</h1>
          <p className="text-text-secondary mb-8">{t('newProposal.noTicketMessage')}</p>
          <div className="flex gap-4 justify-center">
            <Link href="/subscriptions" className="px-6 py-3 bg-primary text-white rounded-lg">
              {t('newProposal.getPremium')}
            </Link>
            <Link href="/voting" className="px-6 py-3 bg-surface-elevated text-text-primary rounded-lg">
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
        <h1 className="text-3xl font-bold text-text-primary">{t('newProposal.title')}</h1>
        <p className="text-text-secondary mt-2">{t('newProposal.subtitle')}</p>
      </div>

      {/* Progress Steps */}
      <div className="flex items-center justify-between mb-8">
        {[1, 2, 3, 4].map((s) => (
          <div key={s} className="flex items-center">
            <div
              className={`w-10 h-10 rounded-full flex items-center justify-center font-bold ${
                step >= s
                  ? 'bg-primary text-white'
                  : 'bg-surface-elevated text-text-muted'
              }`}
            >
              {s}
            </div>
            {s < 4 && (
              <div
                className={`w-16 md:w-24 h-1 ${
                  step > s ? 'bg-primary' : 'bg-surface-elevated'
                }`}
              />
            )}
          </div>
        ))}
      </div>

      {/* Step 1: Basic Info */}
      {step === 1 && (
        <div className="bg-surface-elevated rounded-xl p-6 space-y-6">
          <h2 className="text-xl font-semibold text-text-primary">{t('newProposal.step1Title')}</h2>
          
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              {t('newProposal.originalLink')} *
            </label>
            <input
              type="url"
              value={formData.originalLink}
              onChange={(e) => setFormData({ ...formData, originalLink: e.target.value })}
              placeholder="https://..."
              className="w-full px-4 py-3 bg-surface border border-border rounded-lg text-text-primary focus:outline-none focus:border-primary"
            />
            <p className="text-xs text-text-muted mt-1">{t('newProposal.originalLinkHint')}</p>
          </div>

          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              {t('newProposal.novelTitle')} *
            </label>
            <input
              type="text"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              placeholder={t('newProposal.novelTitlePlaceholder')}
              className="w-full px-4 py-3 bg-surface border border-border rounded-lg text-text-primary focus:outline-none focus:border-primary"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              {t('newProposal.altTitles')}
            </label>
            <div className="flex gap-2 mb-2">
              <input
                type="text"
                value={altTitleInput}
                onChange={(e) => setAltTitleInput(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), handleAddAltTitle())}
                placeholder={t('newProposal.altTitlePlaceholder')}
                className="flex-1 px-4 py-3 bg-surface border border-border rounded-lg text-text-primary focus:outline-none focus:border-primary"
              />
              <button
                onClick={handleAddAltTitle}
                disabled={!altTitleInput.trim() || formData.altTitles.length >= 5}
                className="px-4 py-3 bg-primary text-white rounded-lg disabled:opacity-50"
              >
                {t('newProposal.add')}
              </button>
            </div>
            <div className="flex flex-wrap gap-2">
              {formData.altTitles.map((title, index) => (
                <span
                  key={index}
                  className="px-3 py-1 bg-surface rounded-full text-sm flex items-center gap-2"
                >
                  {title}
                  <button onClick={() => handleRemoveAltTitle(index)} className="text-text-muted hover:text-red-500">
                    ×
                  </button>
                </span>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              {t('newProposal.author')} *
            </label>
            <input
              type="text"
              value={formData.author}
              onChange={(e) => setFormData({ ...formData, author: e.target.value })}
              placeholder={t('newProposal.authorPlaceholder')}
              className="w-full px-4 py-3 bg-surface border border-border rounded-lg text-text-primary focus:outline-none focus:border-primary"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              {t('newProposal.coverUrl')}
            </label>
            <input
              type="url"
              value={formData.coverUrl}
              onChange={(e) => setFormData({ ...formData, coverUrl: e.target.value })}
              placeholder="https://..."
              className="w-full px-4 py-3 bg-surface border border-border rounded-lg text-text-primary focus:outline-none focus:border-primary"
            />
          </div>

          <div className="flex justify-end">
            <button
              onClick={() => setStep(2)}
              disabled={!isStep1Valid}
              className="px-6 py-3 bg-primary text-white rounded-lg disabled:opacity-50"
            >
              {t('newProposal.next')}
            </button>
          </div>
        </div>
      )}

      {/* Step 2: Description */}
      {step === 2 && (
        <div className="bg-surface-elevated rounded-xl p-6 space-y-6">
          <h2 className="text-xl font-semibold text-text-primary">{t('newProposal.step2Title')}</h2>
          
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              {t('newProposal.description')} *
            </label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder={t('newProposal.descriptionPlaceholder')}
              rows={8}
              className="w-full px-4 py-3 bg-surface border border-border rounded-lg text-text-primary focus:outline-none focus:border-primary resize-none"
            />
            <p className={`text-xs mt-1 ${formData.description.length >= 100 ? 'text-green-500' : 'text-text-muted'}`}>
              {formData.description.length}/100 {t('newProposal.minChars')}
            </p>
          </div>

          <div className="flex justify-between">
            <button
              onClick={() => setStep(1)}
              className="px-6 py-3 bg-surface text-text-primary rounded-lg"
            >
              {t('newProposal.back')}
            </button>
            <button
              onClick={() => setStep(3)}
              disabled={!isStep2Valid}
              className="px-6 py-3 bg-primary text-white rounded-lg disabled:opacity-50"
            >
              {t('newProposal.next')}
            </button>
          </div>
        </div>
      )}

      {/* Step 3: Genres & Tags */}
      {step === 3 && (
        <div className="bg-surface-elevated rounded-xl p-6 space-y-6">
          <h2 className="text-xl font-semibold text-text-primary">{t('newProposal.step3Title')}</h2>
          
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-3">
              {t('newProposal.genres')} * ({formData.genres.length}/5)
            </label>
            <div className="flex flex-wrap gap-2">
              {GENRES.map((genre) => (
                <button
                  key={genre}
                  onClick={() => handleToggleGenre(genre)}
                  className={`px-3 py-1.5 rounded-full text-sm transition-colors ${
                    formData.genres.includes(genre)
                      ? 'bg-primary text-white'
                      : 'bg-surface text-text-secondary hover:bg-surface-muted'
                  }`}
                >
                  {t(`genres.${genre}`)}
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-text-secondary mb-3">
              {t('newProposal.tags')} ({formData.tags.length}/10)
            </label>
            <div className="flex flex-wrap gap-2">
              {TAGS.map((tag) => (
                <button
                  key={tag}
                  onClick={() => handleToggleTag(tag)}
                  className={`px-3 py-1.5 rounded-full text-sm transition-colors ${
                    formData.tags.includes(tag)
                      ? 'bg-purple-500 text-white'
                      : 'bg-surface text-text-secondary hover:bg-surface-muted'
                  }`}
                >
                  {t(`tags.${tag}`)}
                </button>
              ))}
            </div>
          </div>

          <div className="flex justify-between">
            <button
              onClick={() => setStep(2)}
              className="px-6 py-3 bg-surface text-text-primary rounded-lg"
            >
              {t('newProposal.back')}
            </button>
            <button
              onClick={() => setStep(4)}
              disabled={!isStep3Valid}
              className="px-6 py-3 bg-primary text-white rounded-lg disabled:opacity-50"
            >
              {t('newProposal.next')}
            </button>
          </div>
        </div>
      )}

      {/* Step 4: Review */}
      {step === 4 && (
        <div className="bg-surface-elevated rounded-xl p-6 space-y-6">
          <h2 className="text-xl font-semibold text-text-primary">{t('newProposal.step4Title')}</h2>
          
          <div className="space-y-4">
            <div className="flex gap-4">
              {formData.coverUrl ? (
                <img src={formData.coverUrl} alt="" className="w-24 h-32 object-cover rounded-lg" />
              ) : (
                <div className="w-24 h-32 bg-surface rounded-lg flex items-center justify-center">
                  <span className="text-text-muted text-xs">{t('newProposal.noCover')}</span>
                </div>
              )}
              <div>
                <h3 className="text-lg font-semibold text-text-primary">{formData.title}</h3>
                <p className="text-text-secondary">{formData.author}</p>
                <a href={formData.originalLink} target="_blank" rel="noopener noreferrer" className="text-primary text-sm hover:underline">
                  {t('newProposal.viewOriginal')}
                </a>
              </div>
            </div>

            {formData.altTitles.length > 0 && (
              <div>
                <p className="text-sm text-text-muted">{t('newProposal.altTitles')}:</p>
                <p className="text-text-secondary">{formData.altTitles.join(', ')}</p>
              </div>
            )}

            <div>
              <p className="text-sm text-text-muted">{t('newProposal.description')}:</p>
              <p className="text-text-secondary line-clamp-4">{formData.description}</p>
            </div>

            <div>
              <p className="text-sm text-text-muted mb-2">{t('newProposal.genres')}:</p>
              <div className="flex flex-wrap gap-2">
                {formData.genres.map((genre) => (
                  <span key={genre} className="px-2 py-1 bg-primary/20 text-primary rounded-full text-xs">
                    {t(`genres.${genre}`)}
                  </span>
                ))}
              </div>
            </div>

            {formData.tags.length > 0 && (
              <div>
                <p className="text-sm text-text-muted mb-2">{t('newProposal.tags')}:</p>
                <div className="flex flex-wrap gap-2">
                  {formData.tags.map((tag) => (
                    <span key={tag} className="px-2 py-1 bg-purple-500/20 text-purple-400 rounded-full text-xs">
                      {t(`tags.${tag}`)}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>

          <div className="bg-yellow-500/10 border border-yellow-500/20 rounded-lg p-4">
            <p className="text-sm text-yellow-500">
              {t('newProposal.ticketWarning')}
            </p>
          </div>

          <div className="flex justify-between">
            <button
              onClick={() => setStep(3)}
              className="px-6 py-3 bg-surface text-text-primary rounded-lg"
            >
              {t('newProposal.back')}
            </button>
            <button
              onClick={handleSubmit}
              disabled={submitMutation.isPending}
              className="px-6 py-3 bg-primary text-white rounded-lg disabled:opacity-50"
            >
              {submitMutation.isPending ? t('newProposal.submitting') : t('newProposal.submit')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
