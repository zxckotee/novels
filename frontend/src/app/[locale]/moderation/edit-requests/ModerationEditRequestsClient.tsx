'use client';

import { useState, useEffect } from 'react';
import { useTranslations } from 'next-intl';
import Link from 'next/link';
import { useAuth } from '@/contexts/AuthContext';

interface EditChange {
  fieldType: string;
  lang?: string;
  oldValue: string;
  newValue: string;
}

interface EditRequest {
  id: string;
  novelId: string;
  novelTitle: string;
  novelSlug: string;
  userId: string;
  userName: string;
  status: 'pending' | 'approved' | 'rejected' | 'cancelled';
  editReason: string;
  changes: EditChange[];
  createdAt: string;
}

interface ModerationEditRequestsClientProps {
  locale: string;
}

export default function ModerationEditRequestsClient({ locale }: ModerationEditRequestsClientProps) {
  const t = useTranslations('moderation');
  const tw = useTranslations('community.wikiEdit');
  const { user } = useAuth();
  
  const [requests, setRequests] = useState<EditRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [processingId, setProcessingId] = useState<string | null>(null);
  const [moderatorComment, setModeratorComment] = useState('');
  const [selectedRequest, setSelectedRequest] = useState<EditRequest | null>(null);
  const [actionType, setActionType] = useState<'approve' | 'reject' | null>(null);

  // Check if user is moderator/admin
  const isModerator = user?.role === 'moderator' || user?.role === 'admin';

  useEffect(() => {
    if (isModerator) {
      fetchRequests();
    }
  }, [isModerator]);

  const fetchRequests = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/v1/moderation/edit-requests');
      if (!response.ok) throw new Error('Failed to load');
      const data = await response.json();
      setRequests(data.requests || []);
    } catch (err) {
      setError('Failed to load edit requests');
    } finally {
      setLoading(false);
    }
  };

  const handleAction = async (requestId: string, action: 'approve' | 'reject') => {
    setProcessingId(requestId);
    try {
      const response = await fetch(`/api/v1/moderation/edit-requests/${requestId}/${action}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ comment: moderatorComment }),
      });

      if (!response.ok) throw new Error('Action failed');
      
      setSelectedRequest(null);
      setActionType(null);
      setModeratorComment('');
      fetchRequests();
    } catch (err) {
      console.error('Failed to process:', err);
    } finally {
      setProcessingId(null);
    }
  };

  const openActionModal = (request: EditRequest, action: 'approve' | 'reject') => {
    setSelectedRequest(request);
    setActionType(action);
    setModeratorComment('');
  };

  if (!isModerator) {
    return (
      <div className="min-h-screen bg-[#121212] flex items-center justify-center">
        <div className="text-center">
          <div className="text-6xl mb-4">🔒</div>
          <p className="text-xl text-white">Access denied</p>
          <p className="text-gray-400 mt-2">Moderator or admin access required</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#121212] py-8">
      <div className="max-w-5xl mx-auto px-4">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-2xl font-bold text-white mb-2">{t('editRequests')}</h1>
            <p className="text-gray-400">Review and manage wiki edit requests</p>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-gray-400">Pending:</span>
            <span className="bg-yellow-500/20 text-yellow-400 px-3 py-1 rounded-lg font-medium">
              {requests.filter(r => r.status === 'pending').length}
            </span>
          </div>
        </div>

        {/* Content */}
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <div className="animate-spin rounded-full h-8 w-8 border-4 border-purple-500 border-t-transparent" />
          </div>
        ) : error ? (
          <div className="text-center py-12">
            <p className="text-red-400 mb-4">{error}</p>
            <button 
              onClick={fetchRequests}
              className="px-4 py-2 bg-purple-600 text-white rounded-lg"
            >
              Retry
            </button>
          </div>
        ) : requests.length === 0 ? (
          <div className="text-center py-12">
            <div className="text-6xl mb-4">✅</div>
            <p className="text-xl text-white">No pending requests</p>
            <p className="text-gray-400 mt-2">All edit requests have been processed</p>
          </div>
        ) : (
          <div className="space-y-4">
            {requests.map(request => (
              <div
                key={request.id}
                className="bg-[#1a1a2e] rounded-xl overflow-hidden"
              >
                {/* Request Header */}
                <div className="p-4 border-b border-gray-700 bg-[#16162a]">
                  <div className="flex items-start justify-between">
                    <div>
                      <Link 
                        href={`/${locale}/novel/${request.novelSlug}`}
                        className="text-lg font-medium text-white hover:text-purple-400"
                      >
                        {request.novelTitle}
                      </Link>
                      <div className="flex items-center gap-4 text-sm text-gray-400 mt-1">
                        <span>By: {request.userName}</span>
                        <span>{new Date(request.createdAt).toLocaleString()}</span>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <button
                        onClick={() => openActionModal(request, 'approve')}
                        disabled={processingId === request.id}
                        className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg transition-colors disabled:opacity-50"
                      >
                        ✓ Approve
                      </button>
                      <button
                        onClick={() => openActionModal(request, 'reject')}
                        disabled={processingId === request.id}
                        className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors disabled:opacity-50"
                      >
                        ✕ Reject
                      </button>
                    </div>
                  </div>
                </div>

                {/* Edit Reason */}
                <div className="p-4 border-b border-gray-700">
                  <p className="text-sm text-gray-400">
                    <span className="font-medium text-gray-300">Reason:</span> {request.editReason}
                  </p>
                </div>

                {/* Changes */}
                <div className="p-4">
                  <p className="text-sm font-medium text-gray-300 mb-3">Changes:</p>
                  <div className="space-y-3">
                    {request.changes.map((change, idx) => (
                      <div key={idx} className="bg-[#121212] rounded-lg p-4">
                        <p className="text-purple-400 font-medium mb-3">
                          {tw(`fields.${change.fieldType}`) || change.fieldType}
                          {change.lang && ` (${change.lang})`}
                        </p>
                        <div className="grid grid-cols-2 gap-4">
                          <div className="bg-red-900/20 border border-red-900/30 rounded-lg p-3">
                            <p className="text-red-400 text-xs font-medium mb-2">OLD VALUE</p>
                            <p className="text-gray-300 text-sm whitespace-pre-wrap">
                              {change.oldValue || '(empty)'}
                            </p>
                          </div>
                          <div className="bg-green-900/20 border border-green-900/30 rounded-lg p-3">
                            <p className="text-green-400 text-xs font-medium mb-2">NEW VALUE</p>
                            <p className="text-white text-sm whitespace-pre-wrap">
                              {change.newValue || '(empty)'}
                            </p>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Action Modal */}
      {selectedRequest && actionType && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div 
            className="absolute inset-0 bg-black/70" 
            onClick={() => {
              setSelectedRequest(null);
              setActionType(null);
            }}
          />
          <div className="relative bg-[#1a1a2e] rounded-xl max-w-md w-full mx-4 p-6">
            <h3 className="text-xl font-bold text-white mb-4">
              {actionType === 'approve' ? 'Approve Edit Request' : 'Reject Edit Request'}
            </h3>
            
            <p className="text-gray-400 mb-4">
              {actionType === 'approve' 
                ? 'This will apply the changes to the novel.'
                : 'The user will be notified that their edit was rejected.'}
            </p>

            <div className="mb-4">
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Moderator Comment (optional)
              </label>
              <textarea
                value={moderatorComment}
                onChange={(e) => setModeratorComment(e.target.value)}
                rows={3}
                placeholder="Add a comment for the user..."
                className="w-full px-4 py-3 bg-[#121212] border border-gray-700 rounded-lg text-white focus:outline-none focus:border-purple-500 resize-none"
              />
            </div>

            <div className="flex items-center justify-end gap-3">
              <button
                onClick={() => {
                  setSelectedRequest(null);
                  setActionType(null);
                }}
                className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg"
              >
                Cancel
              </button>
              <button
                onClick={() => handleAction(selectedRequest.id, actionType)}
                disabled={processingId === selectedRequest.id}
                className={`px-4 py-2 text-white rounded-lg transition-colors disabled:opacity-50 ${
                  actionType === 'approve'
                    ? 'bg-green-600 hover:bg-green-700'
                    : 'bg-red-600 hover:bg-red-700'
                }`}
              >
                {processingId === selectedRequest.id ? 'Processing...' : 'Confirm'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
