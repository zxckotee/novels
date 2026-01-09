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
  onEdit: (commentId: string, body: string) => void;
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

  const isOwner = user?.id === comment.user?.id;
  const canReply = isAuthenticated && comment.depth < maxDepth && !comment.isDeleted;
  const dateLocale = locale === 'ru' ? ru : enUS;

  const handleVote = (value: number) => {
    if (!isAuthenticated) return;
    onVote(comment.id, value);
  };

  const handleSaveEdit = () => {
    onEdit(comment.id, editBody);
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
    <div className={`comment-item ${comment.depth > 0 ? 'ml-4 md:ml-8 border-l-2 border-bg-secondary pl-4' : ''}`}>
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
              <span className="text-lg font-bold text-text-secondary">
                {comment.user?.displayName?.charAt(0).toUpperCase() || '?'}
              </span>
            </div>
          )}
        </div>

        {/* Content */}
        <div className="flex-1 min-w-0">
          {/* Header */}
          <div className="flex items-center gap-2 flex-wrap">
            <span className="font-medium text-text-primary">
              {comment.user?.displayName || t('deletedUser')}
            </span>
            
            {comment.user && (
              <span className={`px-1.5 py-0.5 text-xs text-white rounded ${getLevelBadgeColor(comment.user.level)}`}>
                Lv.{comment.user.level}
              </span>
            )}
            
            {comment.user && getRoleBadge(comment.user.role)}
            
            <span className="text-sm text-text-tertiary">
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
                className="w-full p-3 bg-bg-secondary border border-bg-tertiary rounded-lg resize-none focus:outline-none focus:border-accent"
                rows={3}
              />
              <div className="flex gap-2 mt-2">
                <button
                  onClick={handleSaveEdit}
                  className="px-3 py-1 bg-accent text-white rounded-lg text-sm hover:bg-accent-hover"
                >
                  {t('save')}
                </button>
                <button
                  onClick={() => setIsEditing(false)}
                  className="px-3 py-1 bg-bg-secondary text-text-secondary rounded-lg text-sm hover:bg-bg-tertiary"
                >
                  {t('cancel')}
                </button>
              </div>
            </div>
          ) : (
            <div className="mt-1">
              {comment.isDeleted ? (
                <p className="text-text-tertiary italic">{t('deletedComment')}</p>
              ) : comment.isSpoiler && !showSpoiler ? (
                <button
                  onClick={() => setShowSpoiler(true)}
                  className="flex items-center gap-2 px-3 py-2 bg-bg-secondary rounded-lg text-text-secondary hover:bg-bg-tertiary"
                >
                  <AlertTriangle className="w-4 h-4" />
                  {t('spoilerWarning')}
                </button>
              ) : (
                <p className="text-text-primary whitespace-pre-wrap break-words">
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
                  className={`p-1 rounded hover:bg-bg-secondary transition-colors ${
                    comment.userVote === 1 ? 'text-green-500' : 'text-text-tertiary'
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
                      : 'text-text-tertiary'
                }`}>
                  {comment.likesCount - comment.dislikesCount}
                </span>
                <button
                  onClick={() => handleVote(-1)}
                  className={`p-1 rounded hover:bg-bg-secondary transition-colors ${
                    comment.userVote === -1 ? 'text-red-500' : 'text-text-tertiary'
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
                  className="flex items-center gap-1 text-sm text-text-tertiary hover:text-text-primary"
                >
                  <MessageCircle className="w-4 h-4" />
                  {t('reply')}
                </button>
              )}

              {/* More menu */}
              <div className="relative">
                <button
                  onClick={() => setShowMenu(!showMenu)}
                  className="p-1 text-text-tertiary hover:text-text-primary rounded hover:bg-bg-secondary"
                >
                  <MoreHorizontal className="w-4 h-4" />
                </button>

                {showMenu && (
                  <div className="absolute left-0 top-full mt-1 bg-bg-secondary border border-bg-tertiary rounded-lg shadow-lg py-1 z-10 min-w-[120px]">
                    {isOwner && (
                      <>
                        <button
                          onClick={() => {
                            setIsEditing(true);
                            setShowMenu(false);
                          }}
                          className="w-full flex items-center gap-2 px-3 py-2 text-sm text-text-primary hover:bg-bg-tertiary"
                        >
                          <Edit className="w-4 h-4" />
                          {t('edit')}
                        </button>
                        <button
                          onClick={() => {
                            onDelete(comment.id);
                            setShowMenu(false);
                          }}
                          className="w-full flex items-center gap-2 px-3 py-2 text-sm text-red-500 hover:bg-bg-tertiary"
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
                        className="w-full flex items-center gap-2 px-3 py-2 text-sm text-text-primary hover:bg-bg-tertiary"
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
              className="flex items-center gap-1 mt-2 text-sm text-accent hover:text-accent-hover"
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
