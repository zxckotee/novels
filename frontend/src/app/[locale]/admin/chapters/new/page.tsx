'use client';

import { useState, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { useTranslations, useLocale } from 'next-intl';
import { ArrowLeft, Save } from 'lucide-react';
import { useAuthStore, isModerator } from '@/store/auth';
import api from '@/lib/api/client';

interface ChapterFormData {
  novelId: string;
  number: number;
  title: string;
  contentRu: string;
  contentEn: string;
  publishNow: boolean;
}

const initialFormData: ChapterFormData = {
  novelId: '',
  number: 1,
  title: '',
  contentRu: '',
  contentEn: '',
  publishNow: true,
};

function NewChapterPageContent() {
  const t = useTranslations('admin');
  const locale = useLocale();
  const router = useRouter();
  const searchParams = useSearchParams();
  const { user } = useAuthStore();
  
  const [formData, setFormData] = useState<ChapterFormData>({
    ...initialFormData,
    novelId: searchParams.get('novel') || '',
  });
  const [novels, setNovels] = useState<Array<{ id: string; title: string; slug: string }>>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isLoadingNovels, setIsLoadingNovels] = useState(true);
  
  // Check admin access
  if (!user || !isModerator(user)) {
    if (typeof window !== 'undefined') {
      router.push(`/${locale}`);
    }
    return null;
  }
  
  // Load novels list
  useEffect(() => {
    const loadNovels = async () => {
      try {
        const response = await api.get<Array<{ id: string; title: string; slug: string }>>('/admin/novels?limit=100');
        setNovels(response.data || []);
        
        // If novel is specified but not in list, try to get next chapter number
        if (formData.novelId) {
          const chaptersResponse = await api.get<Array<{ number: number }>>(`/novels/${formData.novelId}/chapters?limit=1&sort=number&order=desc`);
          const lastChapter = chaptersResponse.data?.[0];
          if (lastChapter) {
            setFormData(prev => ({ ...prev, number: lastChapter.number + 1 }));
          }
        }
      } catch (err) {
        console.error('Failed to load novels:', err);
      } finally {
        setIsLoadingNovels(false);
      }
    };
    
    loadNovels();
  }, [formData.novelId]);
  
  // Update form field
  const updateField = <K extends keyof ChapterFormData>(
    field: K, 
    value: ChapterFormData[K]
  ) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };
  
  // Submit form
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsSubmitting(true);
    
    try {
      // Validate required fields
      if (!formData.novelId) {
        throw new Error('Выберите новеллу');
      }
      if (!formData.title.trim()) {
        throw new Error('Укажите название главы');
      }
      if (!formData.contentRu.trim()) {
        throw new Error('Введите содержимое главы');
      }
      
      // Submit to API
      await api.post('/admin/chapters', {
        novelId: formData.novelId,
        number: formData.number,
        title: formData.title,
        publishNow: formData.publishNow,
        localizations: [
          {
            lang: 'ru',
            content: formData.contentRu,
          },
          formData.contentEn && {
            lang: 'en',
            content: formData.contentEn,
          },
        ].filter(Boolean),
      });
      
      // Redirect back
      const novel = novels.find(n => n.id === formData.novelId);
      if (novel) {
        router.push(`/${locale}/admin/novels/${novel.slug}`);
      } else {
        router.push(`/${locale}/admin/chapters`);
      }
    } catch (err: any) {
      setError(err.response?.data?.error?.message || err.message || 'Ошибка при создании главы');
    } finally {
      setIsSubmitting(false);
    }
  };
  
  // Calculate word count
  const wordCount = formData.contentRu.trim().split(/\s+/).filter(Boolean).length;
  
  return (
    <div className="container-custom py-6">
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <Link
          href={`/${locale}/admin/chapters`}
          className="btn-ghost p-2"
        >
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <h1 className="text-2xl font-heading font-bold">Добавить главу</h1>
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
            {/* Chapter Content */}
            <div className="bg-background-secondary rounded-card p-6">
              <h2 className="text-lg font-semibold mb-4">Содержимое главы</h2>
              
              {/* Title */}
              <div className="mb-4">
                <label className="block text-sm font-medium mb-2">
                  Название главы <span className="text-status-error">*</span>
                </label>
                <input
                  type="text"
                  value={formData.title}
                  onChange={(e) => updateField('title', e.target.value)}
                  className="input-primary w-full"
                  placeholder="Глава 1. Начало"
                  required
                />
              </div>
              
              {/* Russian Content */}
              <div className="mb-4">
                <label className="block text-sm font-medium mb-2">
                  Текст (RU) <span className="text-status-error">*</span>
                </label>
                <textarea
                  value={formData.contentRu}
                  onChange={(e) => updateField('contentRu', e.target.value)}
                  className="input-primary w-full font-mono text-sm resize-y"
                  style={{ minHeight: '400px' }}
                  placeholder="Введите текст главы..."
                  required
                />
                <p className="text-xs text-foreground-muted mt-1">
                  Слов: {wordCount}
                </p>
              </div>
              
              {/* English Content (Optional) */}
              <div>
                <label className="block text-sm font-medium mb-2">
                  Текст (EN) <span className="text-foreground-muted">(опционально)</span>
                </label>
                <textarea
                  value={formData.contentEn}
                  onChange={(e) => updateField('contentEn', e.target.value)}
                  className="input-primary w-full font-mono text-sm resize-y"
                  style={{ minHeight: '200px' }}
                  placeholder="Enter English translation..."
                />
              </div>
            </div>
          </div>
          
          {/* Sidebar */}
          <div className="space-y-6">
            {/* Novel Selection */}
            <div className="bg-background-secondary rounded-card p-6">
              <h2 className="text-lg font-semibold mb-4">Новелла</h2>
              
              {/* Novel Select */}
              <div className="mb-4">
                <label className="block text-sm font-medium mb-2">
                  Выберите новеллу <span className="text-status-error">*</span>
                </label>
                {isLoadingNovels ? (
                  <div className="h-10 bg-background-hover rounded animate-pulse" />
                ) : (
                  <select
                    value={formData.novelId}
                    onChange={(e) => updateField('novelId', e.target.value)}
                    className="input-primary w-full"
                    required
                  >
                    <option value="">-- Выберите новеллу --</option>
                    {novels.map(novel => (
                      <option key={novel.id} value={novel.id}>
                        {novel.title}
                      </option>
                    ))}
                  </select>
                )}
              </div>
              
              {/* Chapter Number */}
              <div className="mb-4">
                <label className="block text-sm font-medium mb-2">
                  Номер главы
                </label>
                <input
                  type="number"
                  value={formData.number}
                  onChange={(e) => updateField('number', parseInt(e.target.value) || 1)}
                  className="input-primary w-full"
                  min="1"
                />
              </div>
              
              {/* Publish Now */}
              <div>
                <label className="flex items-center gap-2 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={formData.publishNow}
                    onChange={(e) => updateField('publishNow', e.target.checked)}
                    className="checkbox"
                  />
                  <span className="text-sm">Опубликовать сразу</span>
                </label>
              </div>
            </div>
            
            {/* Tips */}
            <div className="bg-background-secondary rounded-card p-6">
              <h2 className="text-lg font-semibold mb-4">Советы</h2>
              <ul className="text-sm text-foreground-secondary space-y-2">
                <li>• Используйте пустые строки для разделения абзацев</li>
                <li>• Глава должна содержать не менее 100 слов</li>
                <li>• Проверьте текст на опечатки перед публикацией</li>
              </ul>
            </div>
            
            {/* Submit Button */}
            <button
              type="submit"
              disabled={isSubmitting || isLoadingNovels}
              className="btn-primary w-full py-3 flex items-center justify-center gap-2"
            >
              <Save className="w-5 h-5" />
              {isSubmitting ? 'Сохранение...' : 'Опубликовать главу'}
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}

export default function NewChapterPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <NewChapterPageContent />
    </Suspense>
  );
}
