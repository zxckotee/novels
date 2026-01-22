'use client';

import { useState, useCallback } from 'react';
import { useTranslations } from 'next-intl';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { MessageCircle, Loader2 } from 'lucide-react';
import { CommentItem } from './CommentItem';
import { apiClient } from '@/lib/api/client';
import { useAuthStore } from '@/store/auth';

interface Comment {
  id: string;
  parentId?: string;
  depth: number;
  body: string;
  isSpoiler: boolean;
  isDeleted: boolean;
  likesCount: number;
  dislikesCount: number;
  repliesCount: number;
  createdAt: string;
  user?: {
    id: string;
    displayName: string;
    avatarUrl?: string;
    level: number;
    role: string;
  };
  userVote?: number;
  replies?: Comment[];
}

interface CommentsResponse {
  comments: Comment[];
  totalCount: number;
  page: number;
  limit: number;
}

function updateCommentTree(
  comments: Comment[],
  commentId: string,
  updater: (c: Comment) => Comment
): Comment[] {
  return comments.map((c) => {
    if (c.id === commentId) return updater(c);
    if (c.replies?.length) {
      return { ...c, replies: updateCommentTree(c.replies, commentId, updater) };
    }
    return c;
  });
}

interface CommentListProps {
  targetType: 'novel' | 'chapter' | 'news' | 'profile';
  targetId: string;
  locale: string;
  anchor?: string;
}

export function CommentList({ targetType, targetId, locale, anchor }: CommentListProps) {
  const t = useTranslations('comments');
  const { isAuthenticated } = useAuthStore();
  const queryClient = useQueryClient();
  
  const [replyingTo, setReplyingTo] = useState<string | null>(null);
  const [newComment, setNewComment] = useState('');
  const [replyText, setReplyText] = useState('');
  const [isSpoiler, setIsSpoiler] = useState(false);
  const [page, setPage] = useState(1);
  const [sort, setSort] = useState<'newest' | 'oldest' | 'top'>('newest');

  // Fetch comments
  const { data, isLoading, error } = useQuery<CommentsResponse>({
    queryKey: ['comments', targetType, targetId, anchor, page, sort],
    queryFn: async () => {
      const response = await apiClient.get('/comments', {
        params: {
          target_type: targetType,
          target_id: targetId,
          anchor,
          page,
          limit: 20,
          sort,
        },
      });
      // Backend wraps responses as { data: ..., meta: ... }
      return (response.data as any)?.data ?? response.data;
    },
  });

  // Create comment mutation
  const createMutation = useMutation({
    mutationFn: async (data: { body: string; parentId?: string; isSpoiler: boolean }) => {
      return apiClient.post('/comments', {
        targetType,
        targetId,
        parentId: data.parentId,
        anchor,
        body: data.body,
        isSpoiler: data.isSpoiler,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['comments', targetType, targetId] });
      setNewComment('');
      setReplyText('');
      setReplyingTo(null);
      setIsSpoiler(false);
    },
  });

  // Vote mutation  
  const voteMutation = useMutation({
    mutationFn: async ({ commentId, value }: { commentId: string; value: number }) => {
      return apiClient.post(`/comments/${commentId}/vote`, { value });
    },
    onMutate: async ({ commentId, value }) => {
      // Keep replies visible: update cache instead of invalidating/refetching.
      await queryClient.cancelQueries({ queryKey: ['comments', targetType, targetId] });

      const snapshots = queryClient.getQueriesData<CommentsResponse>({
        queryKey: ['comments', targetType, targetId],
      });

      for (const [key, old] of snapshots) {
        if (!old) continue;

        const next = {
          ...old,
          comments: updateCommentTree(old.comments, commentId, (c) => {
            const prevVote = c.userVote ?? 0;
            const nextVote = prevVote === value ? 0 : value;

            let likes = c.likesCount;
            let dislikes = c.dislikesCount;

            // Remove previous vote effect
            if (prevVote === 1) likes = Math.max(0, likes - 1);
            if (prevVote === -1) dislikes = Math.max(0, dislikes - 1);

            // Apply next vote effect
            if (nextVote === 1) likes += 1;
            if (nextVote === -1) dislikes += 1;

            const updated: Comment = { ...c, likesCount: likes, dislikesCount: dislikes };
            if (nextVote === 0) {
              delete (updated as any).userVote;
            } else {
              (updated as any).userVote = nextVote;
            }
            return updated;
          }),
        };

        queryClient.setQueryData(key, next);
      }

      return { snapshots };
    },
    onError: (_err, _vars, ctx) => {
      // Roll back optimistic update
      if (ctx?.snapshots) {
        for (const [key, data] of ctx.snapshots) {
          queryClient.setQueryData(key, data);
        }
      }
    },
  });

  // Edit mutation
  const editMutation = useMutation({
    mutationFn: async ({ commentId, body, isSpoiler }: { commentId: string; body: string; isSpoiler: boolean }) => {
      return apiClient.put(`/comments/${commentId}`, { body, isSpoiler });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['comments', targetType, targetId] });
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: async (commentId: string) => {
      return apiClient.delete(`/comments/${commentId}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['comments', targetType, targetId] });
    },
  });

  // Report mutation
  const reportMutation = useMutation({
    mutationFn: async ({ commentId, reason }: { commentId: string; reason: string }) => {
      return apiClient.post(`/comments/${commentId}/report`, { reason });
    },
  });

  const handleSubmitComment = useCallback(() => {
    if (!newComment.trim()) return;
    createMutation.mutate({ body: newComment, isSpoiler });
  }, [newComment, isSpoiler, createMutation]);

  const handleSubmitReply = useCallback(() => {
    if (!replyText.trim() || !replyingTo) return;
    createMutation.mutate({ body: replyText, parentId: replyingTo, isSpoiler: false });
  }, [replyText, replyingTo, createMutation]);

  const handleVote = useCallback((commentId: string, value: number) => {
    voteMutation.mutate({ commentId, value });
  }, [voteMutation]);

  const handleEdit = useCallback((commentId: string, body: string, isSpoiler: boolean) => {
    editMutation.mutate({ commentId, body, isSpoiler });
  }, [editMutation]);

  const handleDelete = useCallback((commentId: string) => {
    if (confirm(t('confirmDelete'))) {
      deleteMutation.mutate(commentId);
    }
  }, [deleteMutation, t]);

  const handleReport = useCallback((commentId: string) => {
    const reason = prompt(t('reportReason'));
    if (reason && reason.length >= 10) {
      reportMutation.mutate({ commentId, reason });
    }
  }, [reportMutation, t]);

  const handleLoadReplies = useCallback(async (parentId: string) => {
    // Load replies for a specific comment
    const response = await apiClient.get(`/comments/${parentId}/replies`, {
      params: { limit: 50 },
    });
    const replies = (response.data as any)?.data ?? response.data;
    // Update the comment in cache with replies
    queryClient.setQueryData<CommentsResponse>(
      ['comments', targetType, targetId, anchor, page, sort],
      (old) => {
        if (!old) return old;
        const updateReplies = (comments: Comment[]): Comment[] => {
          return comments.map((c) => {
            if (c.id === parentId) {
              return { ...c, replies };
            }
            if (c.replies) {
              return { ...c, replies: updateReplies(c.replies) };
            }
            return c;
          });
        };
        return { ...old, comments: updateReplies(old.comments) };
      }
    );
  }, [queryClient, targetType, targetId, anchor, page, sort]);

  return (
    <div className="comments-section">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <h3 className="flex items-center gap-2 text-xl font-bold text-foreground-primary">
          <MessageCircle className="w-6 h-6" />
          {t('title')} {data?.totalCount ? `(${data.totalCount})` : ''}
        </h3>
        
        {/* Sort */}
        <select
          value={sort}
          onChange={(e) => setSort(e.target.value as typeof sort)}
          className="input w-auto text-sm"
        >
          <option value="newest">{t('sortNewest')}</option>
          <option value="oldest">{t('sortOldest')}</option>
          <option value="top">{t('sortTop')}</option>
        </select>
      </div>

      {/* New comment form */}
      {isAuthenticated ? (
        <div className="mb-8">
          <textarea
            value={newComment}
            onChange={(e) => setNewComment(e.target.value)}
            placeholder={t('placeholder')}
            className="input min-h-[96px] resize-none"
            rows={3}
          />
          <div className="flex items-center justify-between mt-2">
            <label className="flex items-center gap-2 text-sm text-foreground-secondary cursor-pointer">
              <input
                type="checkbox"
                checked={isSpoiler}
                onChange={(e) => setIsSpoiler(e.target.checked)}
                className="checkbox"
              />
              {t('markSpoiler')}
            </label>
            <button
              onClick={handleSubmitComment}
              disabled={!newComment.trim() || createMutation.isPending}
              className="btn-primary flex items-center gap-2"
            >
              {createMutation.isPending && <Loader2 className="w-4 h-4 animate-spin" />}
              {t('submit')}
            </button>
          </div>
        </div>
      ) : (
        <div className="mb-8 p-4 bg-background-secondary rounded-lg text-center text-foreground-secondary">
          {t('loginToComment')}
        </div>
      )}

      {/* Reply form */}
      {replyingTo && (
        <div className="mb-8 p-4 bg-background-secondary rounded-lg">
          <div className="flex justify-between items-center mb-2">
            <span className="text-sm text-foreground-secondary">{t('replyingTo')}</span>
            <button
              onClick={() => setReplyingTo(null)}
              className="text-sm text-foreground-muted hover:text-foreground-primary"
            >
              {t('cancel')}
            </button>
          </div>
          <textarea
            value={replyText}
            onChange={(e) => setReplyText(e.target.value)}
            placeholder={t('replyPlaceholder')}
            className="input resize-none"
            rows={2}
            autoFocus
          />
          <div className="flex justify-end mt-2">
            <button
              onClick={handleSubmitReply}
              disabled={!replyText.trim() || createMutation.isPending}
              className="btn-primary flex items-center gap-2"
            >
              {createMutation.isPending && <Loader2 className="w-4 h-4 animate-spin" />}
              {t('submitReply')}
            </button>
          </div>
        </div>
      )}

      {/* Comments list */}
      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="w-8 h-8 animate-spin text-accent-primary" />
        </div>
      ) : error ? (
        <div className="text-center py-12 text-red-500">
          {t('loadError')}
        </div>
      ) : (data?.comments?.length ?? 0) === 0 ? (
        <div className="text-center py-12 text-foreground-muted">
          {t('noComments')}
        </div>
      ) : (
        <div className="space-y-6">
          {data?.comments.map((comment) => (
            <CommentItem
              key={comment.id}
              comment={comment}
              locale={locale}
              onReply={setReplyingTo}
              onVote={handleVote}
              onEdit={handleEdit}
              onDelete={handleDelete}
              onReport={handleReport}
              onLoadReplies={handleLoadReplies}
            />
          ))}
        </div>
      )}

      {/* Pagination */}
      {data && data.totalCount > 20 && (
        <div className="flex justify-center gap-2 mt-8">
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page === 1}
            className="px-4 py-2 bg-bg-secondary rounded-lg text-text-primary hover:bg-bg-tertiary disabled:opacity-50"
          >
            {t('prevPage')}
          </button>
          <span className="px-4 py-2 text-text-secondary">
            {page} / {Math.ceil(data.totalCount / 20)}
          </span>
          <button
            onClick={() => setPage((p) => p + 1)}
            disabled={page >= Math.ceil(data.totalCount / 20)}
            className="px-4 py-2 bg-bg-secondary rounded-lg text-text-primary hover:bg-bg-tertiary disabled:opacity-50"
          >
            {t('nextPage')}
          </button>
        </div>
      )}
    </div>
  );
}

export default CommentList;
