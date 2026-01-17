'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { useRouter } from 'next/navigation';
import { ArrowLeft, Star } from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import api from '@/lib/api/client';

interface CollectionCard {
  id: string;
  slug: string;
  title: string;
  isPublic: boolean;
  isFeatured: boolean;
  votesCount: number;
  itemsCount: number;
  createdAt: string;
  userId: string;
}

export default function AdminCollectionsPage() {
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user, isLoading } = useAuthStore();
  const hasAccess = isAuthenticated && isAdmin(user);

  const [collections, setCollections] = useState<CollectionCard[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!isLoading && !hasAccess) router.replace(`/${locale}`);
  }, [isLoading, hasAccess, router, locale]);

  useEffect(() => {
    if (hasAccess) load();
  }, [hasAccess]);

  const load = async () => {
    setLoading(true);
    setError(null);
    try {
      // Public list only (backend filters by is_public=true)
      const res = await api.get<{ collections: CollectionCard[]; total: number; page: number; limit: number }>(
        '/collections?sort=popular&limit=50&page=1'
      );
      setCollections(res.data.collections || []);
    } catch (e: any) {
      setError(e?.response?.data?.error?.message || e?.message || 'Failed to load collections');
    } finally {
      setLoading(false);
    }
  };

  const toggleFeatured = async (id: string, featured: boolean) => {
    try {
      await api.post(`/admin/collections/${id}/featured`, { featured });
      setCollections((prev) => prev.map((c) => (c.id === id ? { ...c, isFeatured: featured } : c)));
    } catch (e: any) {
      alert(e?.response?.data?.error?.message || e?.message || 'Failed to update featured status');
    }
  };

  if (isLoading) return null;
  if (!hasAccess) return null;

  return (
    <div className="container-custom py-6">
      <div className="flex items-center gap-4 mb-6">
        <Link href={`/${locale}/admin`} className="btn-ghost p-2">
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <h1 className="text-2xl font-heading font-bold">Коллекции</h1>
      </div>

      {error && (
        <div className="bg-status-error/10 border border-status-error text-status-error rounded-card p-4 mb-6">
          {error}
        </div>
      )}

      {loading ? (
        <div className="text-center py-12">
          <p className="text-foreground-secondary">Загрузка...</p>
        </div>
      ) : (
        <div className="bg-background-secondary rounded-card overflow-hidden">
          <div className="overflow-x-auto">
            <table className="min-w-full">
              <thead className="bg-background-tertiary">
                <tr>
                  <th className="text-left p-3 text-sm text-foreground-secondary">Название</th>
                  <th className="text-left p-3 text-sm text-foreground-secondary">ID</th>
                  <th className="text-left p-3 text-sm text-foreground-secondary">Статус</th>
                  <th className="text-left p-3 text-sm text-foreground-secondary">Голоса</th>
                  <th className="text-left p-3 text-sm text-foreground-secondary">Элементы</th>
                  <th className="text-left p-3 text-sm text-foreground-secondary">Featured</th>
                </tr>
              </thead>
              <tbody>
                {collections.map((c) => (
                  <tr key={c.id} className="border-t border-background-tertiary">
                    <td className="p-3">
                      <Link href={`/${locale}/collections/${c.id}`} className="hover:underline">
                        {c.title}
                      </Link>
                    </td>
                    <td className="p-3 text-xs text-foreground-muted font-mono">{c.id}</td>
                    <td className="p-3 text-sm">{c.isPublic ? 'public' : 'private'}</td>
                    <td className="p-3 text-sm">{c.votesCount}</td>
                    <td className="p-3 text-sm">{c.itemsCount}</td>
                    <td className="p-3">
                      <button
                        onClick={() => toggleFeatured(c.id, !c.isFeatured)}
                        className={`inline-flex items-center gap-2 px-3 py-1 rounded ${
                          c.isFeatured ? 'bg-yellow-500 text-black' : 'bg-background-tertiary'
                        }`}
                      >
                        <Star className="w-4 h-4" />
                        {c.isFeatured ? 'ON' : 'OFF'}
                      </button>
                    </td>
                  </tr>
                ))}
                {collections.length === 0 && (
                  <tr>
                    <td className="p-6 text-center text-foreground-secondary" colSpan={6}>
                      Коллекции не найдены
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}

