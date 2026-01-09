'use client';

import { useState, useEffect } from 'react';
import { useTranslations } from 'next-intl';
import { useAuth } from '@/contexts/AuthContext';

interface Novel {
  id: string;
  slug: string;
  title: string;
  description: string;
  altTitles: string[];
  author: string;
  releaseYear: number;
  originalChaptersCount: number;
  translationStatus: string;
  coverUrl: string;
}

interface EditChange {
  fieldType: string;
  lang?: string;
  oldValue: string;
  newValue: string;
}

interface WikiEditModalProps {
  novel: Novel;
  locale: string;
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export default function WikiEditModal({ novel, locale, isOpen, onClose, onSuccess }: WikiEditModalProps) {
  const t = useTranslations('community.wikiEdit');
  const { user, isAuthenticated } = useAuth();
  const [isPremium, setIsPremium] = useState(false);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  // Форма
  const [editReason, setEditReason] = useState('');
  const [changes, setChanges] = useState<EditChange[]>([]);
  const [values, setValues] = useState({
    title: novel.title,
    description: novel.description,
    altTitles: novel.altTitles?.join(', ') || '',
    author: novel.author || '',
    releaseYear: String(novel.releaseYear || ''),
    originalChaptersCount: String(novel.originalChaptersCount || ''),
    translationStatus: novel.translationStatus || '',
  });

  useEffect(() => {
    if (isAuthenticated) {
      checkPremium();
    }
  }, [isAuthenticated]);

  const checkPremium = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/v1/subscriptions/premium');
      if (response.ok) {
        const data = await response.json();
        setIsPremium(data.isPremium);
      }
    } catch (err) {
      console.error('Failed to check premium status:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleValueChange = (field: string, value: string) => {
    setValues(prev => ({ ...prev, [field]: value }));
  };

  const buildChanges = (): EditChange[] => {
    const newChanges: EditChange[] = [];

    if (values.title !== novel.title) {
      newChanges.push({
        fieldType: 'title',
        lang: locale,
        oldValue: novel.title,
        newValue: values.title,
      });
    }

    if (values.description !== novel.description) {
      newChanges.push({
        fieldType: 'description',
        lang: locale,
        oldValue: novel.description || '',
        newValue: values.description,
      });
    }

    const currentAltTitles = novel.altTitles?.join(', ') || '';
    if (values.altTitles !== currentAltTitles) {
      newChanges.push({
        fieldType: 'alt_titles',
        lang: locale,
        oldValue: currentAltTitles,
        newValue: values.altTitles,
      });
    }

    if (values.author !== (novel.author || '')) {
      newChanges.push({
        fieldType: 'author',
        oldValue: novel.author || '',
        newValue: values.author,
      });
    }

    const currentYear = String(novel.releaseYear || '');
    if (values.releaseYear !== currentYear) {
      newChanges.push({
        fieldType: 'release_year',
        oldValue: currentYear,
        newValue: values.releaseYear,
      });
    }

    const currentChapters = String(novel.originalChaptersCount || '');
    if (values.originalChaptersCount !== currentChapters) {
      newChanges.push({
        fieldType: 'original_chapters_count',
        oldValue: currentChapters,
        newValue: values.originalChaptersCount,
      });
    }

    if (values.translationStatus !== (novel.translationStatus || '')) {
      newChanges.push({
        fieldType: 'translation_status',
        oldValue: novel.translationStatus || '',
        newValue: values.translationStatus,
      });
    }

    return newChanges;
  };

  const handleSubmit = async () => {
    const changesList = buildChanges();
    
    if (changesList.length === 0) {
      setError('No changes made');
      return;
    }

    if (!editReason.trim()) {
      setError('Reason is required');
      return;
    }

    setSubmitting(true);
    setError(null);

    try {
      const response = await fetch(`/api/v1/novels/${novel.id}/edit-requests`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          editReason: editReason.trim(),
          changes: changesList,
        }),
      });

      if (!response.ok) {
        const data = await response.json();
        throw new Error(data.message || 'Failed to submit');
      }

      setSuccess(true);
      setTimeout(() => {
        onSuccess();
        onClose();
      }, 2000);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Failed to submit');
      }
    } finally {
      setSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Overlay */}
      <div 
        className="absolute inset-0 bg-black/70" 
        onClick={onClose}
      />

      {/* Modal */}
      <div className="relative bg-[#1a1a2e] rounded-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto mx-4">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-700">
          <h2 className="text-xl font-bold text-white">{t('editNovelInfo')}</h2>
          <button
            onClick={onClose}
            className="p-2 text-gray-400 hover:text-white rounded-lg hover:bg-gray-700"
          >
            ✕
          </button>
        </div>

        {/* Content */}
        <div className="p-6">
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="animate-spin rounded-full h-8 w-8 border-4 border-purple-500 border-t-transparent" />
            </div>
          ) : !isPremium ? (
            <div className="text-center py-12">
              <div className="text-6xl mb-4">👑</div>
              <p className="text-xl text-white mb-2">{t('premiumRequired')}</p>
              <p className="text-gray-400">
                Upgrade to Premium to edit novel descriptions
              </p>
            </div>
          ) : success ? (
            <div className="text-center py-12">
              <div className="text-6xl mb-4">✅</div>
              <p className="text-xl text-white">{t('requestSent')}</p>
            </div>
          ) : (
            <div className="space-y-6">
              {error && (
                <div className="p-4 bg-red-900/50 border border-red-500 rounded-lg text-red-200">
                  {error}
                </div>
              )}

              {/* Название */}
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  {t('fields.title')}
                </label>
                <input
                  type="text"
                  value={values.title}
                  onChange={(e) => handleValueChange('title', e.target.value)}
                  className="w-full px-4 py-3 bg-[#121212] border border-gray-700 rounded-lg text-white focus:outline-none focus:border-purple-500"
                />
              </div>

              {/* Описание */}
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  {t('fields.description')}
                </label>
                <textarea
                  value={values.description}
                  onChange={(e) => handleValueChange('description', e.target.value)}
                  rows={6}
                  className="w-full px-4 py-3 bg-[#121212] border border-gray-700 rounded-lg text-white focus:outline-none focus:border-purple-500 resize-none"
                />
              </div>

              {/* Альтернативные названия */}
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  {t('fields.altTitles')}
                </label>
                <input
                  type="text"
                  value={values.altTitles}
                  onChange={(e) => handleValueChange('altTitles', e.target.value)}
                  placeholder="Separate with commas"
                  className="w-full px-4 py-3 bg-[#121212] border border-gray-700 rounded-lg text-white focus:outline-none focus:border-purple-500"
                />
              </div>

              {/* Автор */}
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  {t('fields.author')}
                </label>
                <input
                  type="text"
                  value={values.author}
                  onChange={(e) => handleValueChange('author', e.target.value)}
                  className="w-full px-4 py-3 bg-[#121212] border border-gray-700 rounded-lg text-white focus:outline-none focus:border-purple-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                {/* Год выхода */}
                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-2">
                    {t('fields.releaseYear')}
                  </label>
                  <input
                    type="number"
                    value={values.releaseYear}
                    onChange={(e) => handleValueChange('releaseYear', e.target.value)}
                    className="w-full px-4 py-3 bg-[#121212] border border-gray-700 rounded-lg text-white focus:outline-none focus:border-purple-500"
                  />
                </div>

                {/* Глав в оригинале */}
                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-2">
                    {t('fields.originalChaptersCount')}
                  </label>
                  <input
                    type="number"
                    value={values.originalChaptersCount}
                    onChange={(e) => handleValueChange('originalChaptersCount', e.target.value)}
                    className="w-full px-4 py-3 bg-[#121212] border border-gray-700 rounded-lg text-white focus:outline-none focus:border-purple-500"
                  />
                </div>
              </div>

              {/* Статус перевода */}
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  {t('fields.translationStatus')}
                </label>
                <select
                  value={values.translationStatus}
                  onChange={(e) => handleValueChange('translationStatus', e.target.value)}
                  className="w-full px-4 py-3 bg-[#121212] border border-gray-700 rounded-lg text-white focus:outline-none focus:border-purple-500"
                >
                  <option value="">Select status</option>
                  <option value="ongoing">Ongoing</option>
                  <option value="completed">Completed</option>
                  <option value="paused">On Hiatus</option>
                  <option value="dropped">Dropped</option>
                </select>
              </div>

              {/* Причина изменения */}
              <div>
                <label className="block text-sm font-medium text-gray-300 mb-2">
                  {t('changeReason')} *
                </label>
                <textarea
                  value={editReason}
                  onChange={(e) => setEditReason(e.target.value)}
                  rows={3}
                  placeholder={t('enterReason')}
                  className="w-full px-4 py-3 bg-[#121212] border border-gray-700 rounded-lg text-white focus:outline-none focus:border-purple-500 resize-none"
                />
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        {!loading && isPremium && !success && (
          <div className="flex items-center justify-end gap-4 p-6 border-t border-gray-700">
            <button
              onClick={onClose}
              className="px-6 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleSubmit}
              disabled={submitting}
              className="px-6 py-2 bg-purple-600 hover:bg-purple-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50"
            >
              {submitting ? t('submitting') : t('submitForReview')}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
