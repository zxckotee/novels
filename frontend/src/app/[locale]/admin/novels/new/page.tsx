'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useTranslations, useLocale } from 'next-intl';
import { 
  ArrowLeft, 
  Upload, 
  Plus, 
  X,
  Save,
  Image as ImageIcon
} from 'lucide-react';
import { useAuthStore } from '@/store/auth';
import api from '@/lib/api/client';

interface NovelFormData {
  slug: string;
  coverUrl: string;
  translationStatus: 'ongoing' | 'completed' | 'paused' | 'dropped';
  originalChaptersCount: number;
  releaseYear: number;
  authorName: string;
  // Localized content
  titleRu: string;
  titleEn: string;
  descriptionRu: string;
  descriptionEn: string;
  altTitles: string[];
  // Tags and genres
  genreIds: string[];
  tagIds: string[];
}

const initialFormData: NovelFormData = {
  slug: '',
  coverUrl: '',
  translationStatus: 'ongoing',
  originalChaptersCount: 0,
  releaseYear: new Date().getFullYear(),
  authorName: '',
  titleRu: '',
  titleEn: '',
  descriptionRu: '',
  descriptionEn: '',
  altTitles: [],
  genreIds: [],
  tagIds: [],
};

const STATUS_OPTIONS = [
  { value: 'ongoing', label: 'Продолжается' },
  { value: 'completed', label: 'Завершен' },
  { value: 'paused', label: 'Перерыв' },
  { value: 'dropped', label: 'Брошен' },
];

export default function NewNovelPage() {
  const t = useTranslations('admin');
  const locale = useLocale();
  const router = useRouter();
  const { user } = useAuthStore();
  
  const [formData, setFormData] = useState<NovelFormData>(initialFormData);
  const [altTitleInput, setAltTitleInput] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  // Check admin access
  if (!user || (user.role !== 'admin' && user.role !== 'moderator')) {
    if (typeof window !== 'undefined') {
      router.push(`/${locale}`);
    }
    return null;
  }
  
  // Generate slug from Russian title
  const generateSlug = (title: string) => {
    return title
      .toLowerCase()
      .replace(/[а-яё]/g, (char) => {
        const map: Record<string, string> = {
          'а': 'a', 'б': 'b', 'в': 'v', 'г': 'g', 'д': 'd', 'е': 'e', 'ё': 'yo',
          'ж': 'zh', 'з': 'z', 'и': 'i', 'й': 'y', 'к': 'k', 'л': 'l', 'м': 'm',
          'н': 'n', 'о': 'o', 'п': 'p', 'р': 'r', 'с': 's', 'т': 't', 'у': 'u',
          'ф': 'f', 'х': 'h', 'ц': 'ts', 'ч': 'ch', 'ш': 'sh', 'щ': 'sch', 'ъ': '',
          'ы': 'y', 'ь': '', 'э': 'e', 'ю': 'yu', 'я': 'ya'
        };
        return map[char] || char;
      })
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '');
  };
  
  // Update form field
  const updateField = <K extends keyof NovelFormData>(
    field: K, 
    value: NovelFormData[K]
  ) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    
    // Auto-generate slug when Russian title changes
    if (field === 'titleRu' && typeof value === 'string') {
      setFormData(prev => ({ ...prev, slug: generateSlug(value) }));
    }
  };
  
  // Add alternative title
  const addAltTitle = () => {
    if (altTitleInput.trim() && !formData.altTitles.includes(altTitleInput.trim())) {
      updateField('altTitles', [...formData.altTitles, altTitleInput.trim()]);
      setAltTitleInput('');
    }
  };
  
  // Remove alternative title
  const removeAltTitle = (title: string) => {
    updateField('altTitles', formData.altTitles.filter(t => t !== title));
  };
  
  // Submit form
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsSubmitting(true);
    
    try {
      // Validate required fields
      if (!formData.titleRu.trim()) {
        throw new Error('Укажите название на русском');
      }
      if (!formData.slug.trim()) {
        throw new Error('Укажите slug');
      }
      
      // Submit to API
      await api.post('/admin/novels', {
        slug: formData.slug,
        coverUrl: formData.coverUrl || null,
        translationStatus: formData.translationStatus,
        originalChaptersCount: formData.originalChaptersCount,
        releaseYear: formData.releaseYear,
        authorName: formData.authorName || null,
        localizations: [
          {
            lang: 'ru',
            title: formData.titleRu,
            description: formData.descriptionRu,
            altTitles: formData.altTitles,
          },
          formData.titleEn && {
            lang: 'en',
            title: formData.titleEn,
            description: formData.descriptionEn,
            altTitles: [],
          },
        ].filter(Boolean),
        genreIds: formData.genreIds,
        tagIds: formData.tagIds,
      });
      
      // Redirect to novels list
      router.push(`/${locale}/admin/novels`);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || err.message || 'Ошибка при создании новеллы');
    } finally {
      setIsSubmitting(false);
    }
  };
  
  return (
    <div className="container-custom py-6">
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <Link
          href={`/${locale}/admin/novels`}
          className="btn-ghost p-2"
        >
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <h1 className="text-2xl font-heading font-bold">Добавить новеллу</h1>
      </div>
      
      {/* Error Message */}
      {error && (
        <div className="bg-status-error/10 border border-status-error text-status-error rounded-card p-4 mb-6">
          {error}
        </div>
      )}
      
      {/* Form */}
      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-6">
            {/* Basic Info */}
            <div className="bg-background-secondary rounded-card p-6">
              <h2 className="text-lg font-semibold mb-4">Основная информация</h2>
              
              {/* Russian Title */}
              <div className="mb-4">
                <label className="block text-sm font-medium mb-2">
                  Название (RU) <span className="text-status-error">*</span>
                </label>
                <input
                  type="text"
                  value={formData.titleRu}
                  onChange={(e) => updateField('titleRu', e.target.value)}
                  className="input-primary w-full"
                  placeholder="Введите название на русском"
                  required
                />
              </div>
              
              {/* English Title */}
              <div className="mb-4">
                <label className="block text-sm font-medium mb-2">
                  Название (EN)
                </label>
                <input
                  type="text"
                  value={formData.titleEn}
                  onChange={(e) => updateField('titleEn', e.target.value)}
                  className="input-primary w-full"
                  placeholder="Enter English title"
                />
              </div>
              
              {/* Slug */}
              <div className="mb-4">
                <label className="block text-sm font-medium mb-2">
                  Slug (URL) <span className="text-status-error">*</span>
                </label>
                <input
                  type="text"
                  value={formData.slug}
                  onChange={(e) => updateField('slug', e.target.value)}
                  className="input-primary w-full font-mono"
                  placeholder="novel-slug"
                  required
                />
                <p className="text-xs text-foreground-muted mt-1">
                  URL: /{locale}/novel/{formData.slug || 'slug'}
                </p>
              </div>
              
              {/* Alternative Titles */}
              <div className="mb-4">
                <label className="block text-sm font-medium mb-2">
                  Альтернативные названия
                </label>
                <div className="flex gap-2 mb-2">
                  <input
                    type="text"
                    value={altTitleInput}
                    onChange={(e) => setAltTitleInput(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && (e.preventDefault(), addAltTitle())}
                    className="input-primary flex-1"
                    placeholder="Добавить альтернативное название"
                  />
                  <button
                    type="button"
                    onClick={addAltTitle}
                    className="btn-secondary p-2"
                  >
                    <Plus className="w-5 h-5" />
                  </button>
                </div>
                {formData.altTitles.length > 0 && (
                  <div className="flex flex-wrap gap-2">
                    {formData.altTitles.map((title, idx) => (
                      <span 
                        key={idx}
                        className="bg-background-tertiary px-2 py-1 rounded-tag text-sm flex items-center gap-1"
                      >
                        {title}
                        <button
                          type="button"
                          onClick={() => removeAltTitle(title)}
                          className="hover:text-status-error"
                        >
                          <X className="w-3 h-3" />
                        </button>
                      </span>
                    ))}
                  </div>
                )}
              </div>
              
              {/* Author */}
              <div className="mb-4">
                <label className="block text-sm font-medium mb-2">
                  Автор
                </label>
                <input
                  type="text"
                  value={formData.authorName}
                  onChange={(e) => updateField('authorName', e.target.value)}
                  className="input-primary w-full"
                  placeholder="Имя автора"
                />
              </div>
            </div>
            
            {/* Description */}
            <div className="bg-background-secondary rounded-card p-6">
              <h2 className="text-lg font-semibold mb-4">Описание</h2>
              
              {/* Russian Description */}
              <div className="mb-4">
                <label className="block text-sm font-medium mb-2">
                  Описание (RU)
                </label>
                <textarea
                  value={formData.descriptionRu}
                  onChange={(e) => updateField('descriptionRu', e.target.value)}
                  className="input-primary w-full h-40 resize-y"
                  placeholder="Введите описание новеллы на русском..."
                />
              </div>
              
              {/* English Description */}
              <div>
                <label className="block text-sm font-medium mb-2">
                  Описание (EN)
                </label>
                <textarea
                  value={formData.descriptionEn}
                  onChange={(e) => updateField('descriptionEn', e.target.value)}
                  className="input-primary w-full h-40 resize-y"
                  placeholder="Enter English description..."
                />
              </div>
            </div>
          </div>
          
          {/* Sidebar */}
          <div className="space-y-6">
            {/* Cover Image */}
            <div className="bg-background-secondary rounded-card p-6">
              <h2 className="text-lg font-semibold mb-4">Обложка</h2>
              
              <div className="aspect-cover bg-background-tertiary rounded-card mb-4 flex items-center justify-center overflow-hidden">
                {formData.coverUrl ? (
                  <img 
                    src={formData.coverUrl} 
                    alt="Cover preview" 
                    className="w-full h-full object-cover"
                  />
                ) : (
                  <div className="text-center text-foreground-muted">
                    <ImageIcon className="w-12 h-12 mx-auto mb-2" />
                    <p className="text-sm">Нет обложки</p>
                  </div>
                )}
              </div>
              
              <input
                type="url"
                value={formData.coverUrl}
                onChange={(e) => updateField('coverUrl', e.target.value)}
                className="input-primary w-full"
                placeholder="URL обложки"
              />
              
              <p className="text-xs text-foreground-muted mt-2">
                Загрузка файлов будет доступна позже
              </p>
            </div>
            
            {/* Meta Info */}
            <div className="bg-background-secondary rounded-card p-6">
              <h2 className="text-lg font-semibold mb-4">Информация</h2>
              
              {/* Status */}
              <div className="mb-4">
                <label className="block text-sm font-medium mb-2">
                  Статус перевода
                </label>
                <select
                  value={formData.translationStatus}
                  onChange={(e) => updateField('translationStatus', e.target.value as any)}
                  className="input-primary w-full"
                >
                  {STATUS_OPTIONS.map(option => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </div>
              
              {/* Release Year */}
              <div className="mb-4">
                <label className="block text-sm font-medium mb-2">
                  Год выпуска
                </label>
                <input
                  type="number"
                  value={formData.releaseYear}
                  onChange={(e) => updateField('releaseYear', parseInt(e.target.value) || 0)}
                  className="input-primary w-full"
                  min="1900"
                  max="2100"
                />
              </div>
              
              {/* Original Chapters Count */}
              <div>
                <label className="block text-sm font-medium mb-2">
                  Глав в оригинале
                </label>
                <input
                  type="number"
                  value={formData.originalChaptersCount}
                  onChange={(e) => updateField('originalChaptersCount', parseInt(e.target.value) || 0)}
                  className="input-primary w-full"
                  min="0"
                />
              </div>
            </div>
            
            {/* Submit Button */}
            <button
              type="submit"
              disabled={isSubmitting}
              className="btn-primary w-full py-3 flex items-center justify-center gap-2"
            >
              <Save className="w-5 h-5" />
              {isSubmitting ? 'Сохранение...' : 'Создать новеллу'}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}
