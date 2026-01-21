'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import { useTranslations, useLocale } from 'next-intl';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  AlertCircle,
  FileEdit,
  MessageSquare,
  Flag,
  ArrowLeft,
  Check,
  X,
  Eye,
  ExternalLink
} from 'lucide-react';
import { useAuthStore, isModerator } from '@/store/auth';
import { useRouter } from 'next/navigation';
import { api } from '@/lib/api';
import { toast } from 'react-hot-toast';

type TabType = 'reports' | 'proposals' | 'edits' | 'comments';

interface Proposal {
  id: string;
  title: string;
  author: string;
  description: string;
  coverUrl?: string;
  originalLink: string;
  genres: string[];
  status: string;
  createdAt: string;
  user?: {
    displayName: string;
  };
}

export default function AdminModerationPage() {
  const t = useTranslations('admin');
  const locale = useLocale();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { isAuthenticated, user, isLoading: authLoading } = useAuthStore();
  const [activeTab, setActiveTab] = useState<TabType>('proposals');
  
  // Fetch proposals pending moderation
  const { data: pendingProposals, isLoading: pendingProposalsLoading } = useQuery<Proposal[]>({
    queryKey: ['pending-proposals'],
    queryFn: async () => {
      const response = await api.get('/moderation/proposals');
      const data = response.data;
      // API может возвращать массив или объект с proposals
      if (Array.isArray(data)) {
        return data as Proposal[];
      } else if (data && typeof data === 'object' && 'proposals' in data) {
        return ((data as any).proposals || []) as Proposal[];
      }
      return [] as Proposal[];
    },
    enabled: isAuthenticated && activeTab === 'proposals',
  });
  
  // Moderate proposal mutation
  const moderateMutation = useMutation({
    mutationFn: async ({ proposalId, action, reason }: { proposalId: string; action: 'approve' | 'reject'; reason?: string }) => {
      return api.post(`/moderation/proposals/${proposalId}`, { action, reason });
    },
    onSuccess: (_, variables) => {
      toast.success(variables.action === 'approve' ? 'Предложение одобрено' : 'Предложение отклонено');
      queryClient.invalidateQueries({ queryKey: ['pending-proposals'] });
      // Keep public pages in sync (react-query cache has 60s staleTime).
      queryClient.invalidateQueries({ queryKey: ['voting-leaderboard'] });
      queryClient.invalidateQueries({ queryKey: ['voting-stats'] });
      queryClient.invalidateQueries({ queryKey: ['proposals'] });
    },
    onError: () => {
      toast.error('Ошибка модерации');
    },
  });
  
  // Check admin access with useEffect
  useEffect(() => {
    if (!authLoading && (!isAuthenticated || !isModerator(user))) {
      router.push(`/${locale}`);
    }
  }, [authLoading, isAuthenticated, user, router, locale]);
  
  // Don't render if not authorized
  if (authLoading) {
    return (
      <div className="container-custom py-12 text-center">
        <p className="text-foreground-muted">Загрузка...</p>
      </div>
    );
  }
  if (!isAuthenticated || !isModerator(user)) {
    return (
      <div className="container-custom py-12 text-center">
        <p className="text-foreground-muted">Загрузка...</p>
      </div>
    );
  }
  
  return (
    <div className="container-custom py-6">
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <Link
          href={`/${locale}/admin`}
          className="btn-ghost p-2"
        >
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <h1 className="text-2xl font-heading font-bold">Модерация</h1>
      </div>
      
      {/* Stats */}
      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="bg-background-secondary rounded-card p-4">
          <div className="flex items-center justify-between mb-2">
            <Flag className="w-5 h-5 text-status-error" />
            <span className="text-2xl font-bold">0</span>
          </div>
          <div className="text-sm text-foreground-secondary">Жалоб на рассмотрении</div>
        </div>
        
        <div className="bg-background-secondary rounded-card p-4">
          <div className="flex items-center justify-between mb-2">
            <FileEdit className="w-5 h-5 text-accent-primary" />
            <span className="text-2xl font-bold">{pendingProposals?.length || 0}</span>
          </div>
          <div className="text-sm text-foreground-secondary">Предложений новелл</div>
        </div>
        
        <div className="bg-background-secondary rounded-card p-4">
          <div className="flex items-center justify-between mb-2">
            <MessageSquare className="w-5 h-5 text-accent-warning" />
            <span className="text-2xl font-bold">0</span>
          </div>
          <div className="text-sm text-foreground-secondary">Жалоб на комментарии</div>
        </div>
      </div>
      
      {/* Tabs */}
      <div className="flex gap-1 mb-6 border-b border-border-primary">
        <button
          onClick={() => setActiveTab('proposals')}
          className={`px-4 py-2 text-sm font-medium transition-colors ${
            activeTab === 'proposals'
              ? 'text-accent-primary border-b-2 border-accent-primary -mb-px'
              : 'text-foreground-secondary hover:text-foreground-primary'
          }`}
        >
          <span className="flex items-center gap-2">
            <FileEdit className="w-4 h-4" />
            Предложения новелл ({pendingProposals?.length || 0})
          </span>
        </button>
        
        <button
          onClick={() => setActiveTab('edits')}
          className={`px-4 py-2 text-sm font-medium transition-colors ${
            activeTab === 'edits'
              ? 'text-accent-primary border-b-2 border-accent-primary -mb-px'
              : 'text-foreground-secondary hover:text-foreground-primary'
          }`}
        >
          <span className="flex items-center gap-2">
            <FileEdit className="w-4 h-4" />
            Правки описаний
          </span>
        </button>
        
        <button
          onClick={() => setActiveTab('reports')}
          className={`px-4 py-2 text-sm font-medium transition-colors ${
            activeTab === 'reports'
              ? 'text-accent-primary border-b-2 border-accent-primary -mb-px'
              : 'text-foreground-secondary hover:text-foreground-primary'
          }`}
        >
          <span className="flex items-center gap-2">
            <Flag className="w-4 h-4" />
            Жалобы
          </span>
        </button>
        
        <button
          onClick={() => setActiveTab('comments')}
          className={`px-4 py-2 text-sm font-medium transition-colors ${
            activeTab === 'comments'
              ? 'text-accent-primary border-b-2 border-accent-primary -mb-px'
              : 'text-foreground-secondary hover:text-foreground-primary'
          }`}
        >
          <span className="flex items-center gap-2">
            <MessageSquare className="w-4 h-4" />
            Жалобы на комментарии
          </span>
        </button>
      </div>
      
      {/* Content */}
      <div className="bg-background-secondary rounded-card p-6">
        {/* Proposals Tab */}
        {activeTab === 'proposals' && (
          <div>
            {pendingProposalsLoading ? (
              <div className="text-center py-12">
                <div className="animate-spin w-8 h-8 border-2 border-accent-primary border-t-transparent rounded-full mx-auto" />
              </div>
            ) : pendingProposals && pendingProposals.length > 0 ? (
              <div className="space-y-4">
                {pendingProposals.map((proposal) => (
                  <div key={proposal.id} className="bg-background-tertiary rounded-lg p-4">
                    <div className="flex gap-4">
                      {/* Cover */}
                      {proposal.coverUrl ? (
                        <img src={proposal.coverUrl} alt="" className="w-20 h-28 object-cover rounded" />
                      ) : (
                        <div className="w-20 h-28 bg-background-hover rounded flex items-center justify-center">
                          <span className="text-xs text-foreground-muted">Нет обложки</span>
                        </div>
                      )}
                      
                      {/* Info */}
                      <div className="flex-1 min-w-0">
                        <h3 className="text-lg font-semibold text-foreground-primary mb-1">{proposal.title}</h3>
                        <p className="text-sm text-foreground-secondary mb-2">Автор: {proposal.author}</p>
                        <p className="text-sm text-foreground-muted line-clamp-2 mb-2">{proposal.description}</p>
                        
                        <div className="flex flex-wrap gap-2 mb-2">
                          {proposal.genres?.map((genre) => (
                            <span key={genre} className="px-2 py-0.5 text-xs bg-accent-primary/20 text-accent-primary rounded-full">
                              {genre}
                            </span>
                          ))}
                        </div>
                        
                        <div className="flex items-center gap-2 text-xs text-foreground-muted">
                          <span>Предложил: {proposal.user?.displayName}</span>
                          <span>•</span>
                          <span>{new Date(proposal.createdAt).toLocaleDateString('ru-RU')}</span>
                          <a
                            href={proposal.originalLink}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="flex items-center gap-1 text-accent-primary hover:underline"
                          >
                            <ExternalLink className="w-3 h-3" />
                            Оригинал
                          </a>
                        </div>
                      </div>
                      
                      {/* Actions */}
                      <div className="flex flex-col gap-2">
                        <button
                          onClick={() => moderateMutation.mutate({ proposalId: proposal.id, action: 'approve' })}
                          disabled={moderateMutation.isPending}
                          className="btn-primary flex items-center gap-2 whitespace-nowrap"
                        >
                          <Check className="w-4 h-4" />
                          Одобрить
                        </button>
                        <button
                          onClick={() => {
                            const reason = prompt('Причина отклонения (опционально):');
                            moderateMutation.mutate({ proposalId: proposal.id, action: 'reject', reason: reason || undefined });
                          }}
                          disabled={moderateMutation.isPending}
                          className="btn-danger flex items-center gap-2 whitespace-nowrap"
                        >
                          <X className="w-4 h-4" />
                          Отклонить
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-12">
                <FileEdit className="w-16 h-16 mx-auto mb-4 text-foreground-muted" />
                <h3 className="text-lg font-semibold mb-2">Нет предложений</h3>
                <p className="text-foreground-secondary">
                  Все предложения обработаны
                </p>
              </div>
            )}
          </div>
        )}
        
        {/* Edits Tab */}
        {activeTab === 'edits' && (
          <div>
            <div className="flex items-center justify-between mb-4">
              <h2 className="font-semibold">Запросы на правки описаний</h2>
              <Link
                href={`/${locale}/moderation/edit-requests`}
                className="text-sm text-accent-primary hover:underline"
              >
                Открыть в режиме модерации
              </Link>
            </div>
            
            <div className="text-center py-12">
              <FileEdit className="w-16 h-16 mx-auto mb-4 text-foreground-muted" />
              <h3 className="text-lg font-semibold mb-2">Нет запросов</h3>
              <p className="text-foreground-secondary">
                Все запросы на правки обработаны
              </p>
            </div>
          </div>
        )}
        
        {/* Reports Tab */}
        {activeTab === 'reports' && (
          <div className="text-center py-12">
            <Flag className="w-16 h-16 mx-auto mb-4 text-foreground-muted" />
            <h3 className="text-lg font-semibold mb-2">Нет жалоб</h3>
            <p className="text-foreground-secondary">
              Все жалобы обработаны
            </p>
          </div>
        )}
        
        {/* Comments Tab */}
        {activeTab === 'comments' && (
          <div className="text-center py-12">
            <MessageSquare className="w-16 h-16 mx-auto mb-4 text-foreground-muted" />
            <h3 className="text-lg font-semibold mb-2">Нет жалоб на комментарии</h3>
            <p className="text-foreground-secondary">
              Все жалобы обработаны
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
