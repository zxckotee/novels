'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { formatDistanceToNow } from 'date-fns';
import { ru, enUS } from 'date-fns/locale';
import { 
  ThumbsUp, 
  ThumbsDown, 
  MessageCircle, 
  MoreHorizontal,
  Flag,
  Edit,
  Trash,
  ChevronDown,
  ChevronUp,
  AlertTriangle
} from 'lucide-react';
import { useAuthStore } from '@/store/auth';

interface CommentUser {
  id: string;
  displayName: string;
  avatarUrl?: string;
  level: number;
  role: string;
}

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
  user?: CommentUser;
  userVote?: number;
  replies?: Comment[];
}

interface CommentItemProps {
  comment: Comment;
  locale: string;
  onReply: (commentId: string) => void;
  onVote: (commentId: string, value: number) => void;
  onEdit: (commentId: string, body: string, isSpoiler: boolean) => void;
  onDelete: (commentId: string) => void;
  onReport: (commentId: string) => void;
  onLoadReplies: (commentId: string) => void;
  maxDepth?: number;
}

export function CommentItem({
  comment,
  locale,
  onReply,
  onVote,
  onEdit,
  onDelete,
  onReport,
  onLoadReplies,
  maxDepth = 5,
}: CommentItemProps) {
  const t = useTranslations('comments');
  const { user, isAuthenticated } = useAuthStore();
  const [showSpoiler, setShowSpoiler] = useState(false);
  const [showReplies, setShowReplies] = useState(true);
  const [showMenu, setShowMenu] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editBody, setEditBody] = useState(comment.body);
  const [editIsSpoiler, setEditIsSpoiler] = useState(comment.isSpoiler);

  const isOwner = user?.id === comment.user?.id;
  const canReply = isAuthenticated && comment.depth < maxDepth && !comment.isDeleted;
  const dateLocale = locale === 'ru' ? ru : enUS;

  const handleVote = (value: number) => {
    if (!isAuthenticated) return;
    onVote(comment.id, value);
  };

  const handleSaveEdit = () => {
    onEdit(comment.id, editBody, editIsSpoiler);
    setIsEditing(false);
  };

  const getLevelBadgeColor = (level: number) => {
    if (level >= 50) return 'bg-gradient-to-r from-yellow-500 to-orange-500';
    if (level >= 30) return 'bg-gradient-to-r from-purple-500 to-pink-500';
    if (level >= 20) return 'bg-gradient-to-r from-blue-500 to-cyan-500';
    if (level >= 10) return 'bg-green-500';
    return 'bg-gray-500';
  };

  const getRoleBadge = (role: string) => {
    switch (role) {
      case 'admin':
        return <span className="ml-2 px-1.5 py-0.5 text-xs bg-red-500 text-white rounded">{t('roleAdmin')}</span>;
      case 'moderator':
        return <span className="ml-2 px-1.5 py-0.5 text-xs bg-blue-500 text-white rounded">{t('roleMod')}</span>;
      case 'premium':
        return <span className="ml-2 px-1.5 py-0.5 text-xs bg-yellow-500 text-black rounded">{t('rolePremium')}</span>;
      default:
        return null;
    }
  };

  return (
    <div className={`comment-item ${comment.depth > 0 ? 'ml-4 md:ml-8 border-l-2 border-background-secondary pl-4' : ''}`}>
      <div className="flex gap-3">
        {/* Avatar */}
        <div className="flex-shrink-0">
          {comment.user?.avatarUrl ? (
            <img 
              src={comment.user.avatarUrl} 
              alt={comment.user.displayName}
              className="w-10 h-10 rounded-full object-cover"
            />
          ) : (
            <div className="w-10 h-10 rounded-full bg-bg-secondary flex items-center justify-center">
              <span className="text-lg font-bold text-foreground-secondary">
                {comment.user?.displayName?.charAt(0).toUpperCase() || '?'}
              </span>
            </div>
          )}
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          {/* Header */}
          <div className="flex items-center gap-2 flex-wrap">
            <span className="font-medium text-foreground-primary">
              {comment.user?.displayName || t('deletedUser')}
            </span>
            
            {comment.user && (
              <span className={`px-1.5 py-0.5 text-xs text-white rounded ${getLevelBadgeColor(comment.user.level)}`}>
                Lv.{comment.user.level}
              </span>
            )}
            
            {comment.user && getRoleBadge(comment.user.role)}
            
            <span className="text-sm text-foreground-muted">
              {formatDistanceToNow(new Date(comment.createdAt), { 
                addSuffix: true, 
                locale: dateLocale 
              })}
            </span>
          </div>

          {/* Body */}
          {isEditing ? (
            <div className="mt-2">
              <textarea
                value={editBody}
                onChange={(e) => setEditBody(e.target.value)}
                className="input resize-none"
                rows={3}
              />
              <label className="mt-2 flex items-center gap-2 text-sm text-text-secondary cursor-pointer">
                <input
                  type="checkbox"
                  checked={editIsSpoiler}
                  onChange={(e) => setEditIsSpoiler(e.target.checked)}
                  className="checkbox"
                />
                {t('markSpoiler')}
              </label>
              <div className="flex gap-2 mt-2">
                <button
                  onClick={handleSaveEdit}
                  className="btn-primary px-3 py-1 text-sm"
                >
                  {t('save')}
                </button>
                <button
                  onClick={() => setIsEditing(false)}
                  className="btn-secondary px-3 py-1 text-sm"
                >
                  {t('cancel')}
                </button>
              </div>
            </div>
          ) : (
            <div className="mt-1">
              {comment.isDeleted ? (
                <p className="text-foreground-muted italic">{t('deletedComment')}</p>
              ) : comment.isSpoiler && !showSpoiler ? (
                <button
                  onClick={() => setShowSpoiler(true)}
                  className="flex items-center gap-2 px-3 py-2 bg-background-secondary rounded-lg text-foreground-secondary hover:bg-background-tertiary"
                >
                  <AlertTriangle className="w-4 h-4" />
                  {t('spoilerWarning')}
                </button>
              ) : (
                <p className="text-foreground-primary whitespace-pre-wrap break-words">
                  {comment.body}
                </p>
              )}
            </div>
          )}

          {/* Actions */}
          {!comment.isDeleted && (
            <div className="flex items-center gap-4 mt-2">
              {/* Votes */}
              <div className="flex items-center gap-1">
                <button
                  onClick={() => handleVote(1)}
                  className={`p-1 rounded hover:bg-background-secondary transition-colors ${
                    comment.userVote === 1 ? 'text-green-500' : 'text-foreground-muted'
                  }`}
                  disabled={!isAuthenticated}
                >
                  <ThumbsUp className="w-4 h-4" />
                </button>
                <span className={`text-sm min-w-[2rem] text-center ${
                  comment.likesCount - comment.dislikesCount > 0 
                    ? 'text-green-500' 
                    : comment.likesCount - comment.dislikesCount < 0 
                      ? 'text-red-500' 
                      : 'text-foreground-muted'
                }`}>
                  {comment.likesCount - comment.dislikesCount}
                </span>
                <button
                  onClick={() => handleVote(-1)}
                  className={`p-1 rounded hover:bg-background-secondary transition-colors ${
                    comment.userVote === -1 ? 'text-red-500' : 'text-foreground-muted'
                  }`}
                  disabled={!isAuthenticated}
                >
                  <ThumbsDown className="w-4 h-4" />
                </button>
              </div>

              {/* Reply */}
              {canReply && (
                <button
                  onClick={() => onReply(comment.id)}
                  className="flex items-center gap-1 text-sm text-foreground-muted hover:text-foreground-primary"
                >
                  <MessageCircle className="w-4 h-4" />
                  {t('reply')}
                </button>
              )}

              {/* More menu */}
              <div className="relative">
                <button
                  onClick={() => setShowMenu(!showMenu)}
                  className="p-1 text-foreground-muted hover:text-foreground-primary rounded hover:bg-background-secondary"
                >
                  <MoreHorizontal className="w-4 h-4" />
                </button>

                {showMenu && (
                  <div className="absolute left-0 top-full mt-1 bg-background-secondary border border-border-primary rounded-lg shadow-lg py-1 z-10 min-w-[120px]">
                    {isOwner && (
                      <>
                        <button
                          onClick={() => {
                            setIsEditing(true);
                            setShowMenu(false);
                          }}
                          className="w-full flex items-center gap-2 px-3 py-2 text-sm text-foreground-primary hover:bg-background-tertiary"
                        >
                          <Edit className="w-4 h-4" />
                          {t('edit')}
                        </button>
                        <button
                          onClick={() => {
                            onDelete(comment.id);
                            setShowMenu(false);
                          }}
                          className="w-full flex items-center gap-2 px-3 py-2 text-sm text-red-500 hover:bg-background-tertiary"
                        >
                          <Trash className="w-4 h-4" />
                          {t('delete')}
                        </button>
                      </>
                    )}
                    {!isOwner && isAuthenticated && (
                      <button
                        onClick={() => {
                          onReport(comment.id);
                          setShowMenu(false);
                        }}
                        className="w-full flex items-center gap-2 px-3 py-2 text-sm text-foreground-primary hover:bg-background-tertiary"
                      >
                        <Flag className="w-4 h-4" />
                        {t('report')}
                      </button>
                    )}
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Replies toggle */}
          {comment.repliesCount > 0 && (
            <button
              onClick={() => {
                if (!comment.replies?.length) {
                  onLoadReplies(comment.id);
                }
                setShowReplies(!showReplies);
              }}
              className="flex items-center gap-1 mt-2 text-sm text-accent-primary hover:text-accent-primary/80"
            >
              {showReplies ? (
                <>
                  <ChevronUp className="w-4 h-4" />
                  {t('hideReplies')}
                </>
              ) : (
                <>
                  <ChevronDown className="w-4 h-4" />
                  {t('showReplies', { count: comment.repliesCount })}
                </>
              )}
            </button>
          )}

          {/* Nested replies */}
          {showReplies && comment.replies && comment.replies.length > 0 && (
            <div className="mt-4 space-y-4">
              {comment.replies.map((reply) => (
                <CommentItem
                  key={reply.id}
                  comment={reply}
                  locale={locale}
                  onReply={onReply}
                  onVote={onVote}
                  onEdit={onEdit}
                  onDelete={onDelete}
                  onReport={onReport}
                  onLoadReplies={onLoadReplies}
                  maxDepth={maxDepth}
                />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
