'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { ArrowLeft, Save } from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import api from '@/lib/api/client';

interface NewsFormData {
  title: string;
  summary: string;
  content: string;
  category: string;
  publish: boolean;
}

export default function AdminNewNewsPage() {
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user, isLoading } = useAuthStore();
  const [formData, setFormData] = useState<NewsFormData>({
    title: '',
    summary: '',
    content: '',
    category: 'announcement',
    publish: true,
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const hasAccess = isAuthenticated && isAdmin(user);

  useEffect(() => {
    if (!isLoading && !hasAccess) {
      router.replace(`/${locale}`);
    }
  }, [isLoading, hasAccess, router, locale]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      if (!formData.title.trim()) {
        throw new Error('Укажите заголовок');
      }
      if (!formData.content.trim()) {
        throw new Error('Введите текст новости');
      }

      await api.post('/admin/news', {
        title: formData.title,
        summary: formData.summary,
        content: formData.content,
        category: formData.category,
        publish: formData.publish,
      });

      router.push(`/${locale}/admin/news`);
    } catch (err: any) {
      setError(err.response?.data?.error?.message || err.message || 'Ошибка при создании');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isLoading) return null;
  if (!hasAccess) return null;

  return (
    <div className="container-custom py-6">
      <div className="flex items-center gap-4 mb-6">
        <Link href={`/${locale}/admin/news`} className="btn-ghost p-2">
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <h1 className="text-2xl font-heading font-bold">Добавить новость</h1>
      </div>

      {error && (
        <div className="bg-status-error/10 border border-status-error text-status-error rounded-card p-4 mb-6">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="bg-background-secondary rounded-card p-6">
          <div className="mb-4">
            <label className="block text-sm font-medium mb-2">
              Заголовок <span className="text-status-error">*</span>
            </label>
            <input
              type="text"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              className="input w-full"
              placeholder="Заголовок новости"
              required
            />
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium mb-2">Краткое описание</label>
            <textarea
              value={formData.summary}
              onChange={(e) => setFormData({ ...formData, summary: e.target.value })}
              className="input w-full h-20 resize-y"
              placeholder="Краткое описание (опционально)..."
            />
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium mb-2">
              Текст новости <span className="text-status-error">*</span>
            </label>
            <textarea
              value={formData.content}
              onChange={(e) => setFormData({ ...formData, content: e.target.value })}
              className="input w-full h-64 resize-y"
              placeholder="Полный текст новости (минимум 50 символов)..."
              required
            />
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium mb-2">
              Категория <span className="text-status-error">*</span>
            </label>
            <select
              value={formData.category}
              onChange={(e) => setFormData({ ...formData, category: e.target.value })}
              className="input w-full"
            >
              <option value="announcement">Объявление</option>
              <option value="update">Обновление</option>
              <option value="event">Событие</option>
              <option value="community">Сообщество</option>
              <option value="translation">Перевод</option>
            </select>
          </div>

          <div className="mb-4">
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={formData.publish}
                onChange={(e) => setFormData({ ...formData, publish: e.target.checked })}
                className="checkbox"
              />
              <span className="text-sm">Опубликовать сразу</span>
            </label>
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="btn-primary w-full py-3 flex items-center justify-center gap-2"
          >
            <Save className="w-5 h-5" />
            {isSubmitting ? 'Сохранение...' : 'Создать новость'}
          </button>
        </div>
      </form>
    </div>
  );
}
