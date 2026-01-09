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
  status: 'pending' | 'approved' | 'rejected' | 'cancelled';
  editReason: string;
  moderatorComment?: string;
  changes: EditChange[];
  createdAt: string;
  updatedAt: string;
}

interface EditRequestsClientProps {
  locale: string;
}

export default function EditRequestsClient({ locale }: EditRequestsClientProps) {
  const t = useTranslations('community.wikiEdit');
  const tc = useTranslations('common');
  const { isAuthenticated } = useAuth();
  
  const [requests, setRequests] = useState<EditRequest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<string>('all');

  useEffect(() => {
    if (isAuthenticated) {
      fetchRequests();
    }
  }, [isAuthenticated]);

  const fetchRequests = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/v1/me/edit-requests');
      if (!response.ok) throw new Error('Failed to load');
      const data = await response.json();
      setRequests(data.requests || []);
    } catch (err) {
      setError('Failed to load edit requests');
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = async (requestId: string) => {
    try {
      await fetch(`/api/v1/edit-requests/${requestId}/cancel`, {
        method: 'POST',
      });
      fetchRequests();
    } catch (err) {
      console.error('Failed to cancel:', err);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'pending': return 'bg-yellow-500/20 text-yellow-400';
      case 'approved': return 'bg-green-500/20 text-green-400';
      case 'rejected': return 'bg-red-500/20 text-red-400';
      case 'cancelled': return 'bg-gray-500/20 text-gray-400';
      default: return 'bg-gray-500/20 text-gray-400';
    }
  };

  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'pending': return t('pending');
      case 'approved': return t('approved');
      case 'rejected': return t('rejected');
      case 'cancelled': return t('cancelled');
      default: return status;
    }
  };

  const filteredRequests = filter === 'all' 
    ? requests 
    : requests.filter(r => r.status === filter);

  if (!isAuthenticated) {
    return (
      <div className="min-h-screen bg-[#121212] flex items-center justify-center">
        <div className="text-center">
          <p className="text-xl text-white mb-4">{tc('loginRequired')}</p>
          <Link 
            href={`/${locale}/login`}
            className="px-6 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg"
          >
            {tc('login')}
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#121212] py-8">
      <div className="max-w-4xl mx-auto px-4">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-white mb-2">{t('editRequests')}</h1>
          <p className="text-gray-400">View and manage your edit requests</p>
        </div>

        {/* Filter */}
        <div className="flex gap-2 mb-6">
          {['all', 'pending', 'approved', 'rejected'].map(status => (
            <button
              key={status}
              onClick={() => setFilter(status)}
              className={`px-4 py-2 rounded-lg transition-colors ${
                filter === status
                  ? 'bg-purple-600 text-white'
                  : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
              }`}
            >
              {status === 'all' ? 'All' : getStatusLabel(status)}
            </button>
          ))}
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
        ) : filteredRequests.length === 0 ? (
          <div className="text-center py-12">
            <div className="text-6xl mb-4">📝</div>
            <p className="text-xl text-white mb-2">{t('noRequests')}</p>
            <p className="text-gray-400">Your edit requests will appear here</p>
          </div>
        ) : (
          <div className="space-y-4">
            {filteredRequests.map(request => (
              <div
                key={request.id}
                className="bg-[#1a1a2e] rounded-xl overflow-hidden"
              >
                {/* Request Header */}
                <div className="p-4 border-b border-gray-700">
                  <div className="flex items-start justify-between">
                    <div>
                      <Link 
                        href={`/${locale}/novel/${request.novelSlug}`}
                        className="text-lg font-medium text-white hover:text-purple-400"
                      >
                        {request.novelTitle}
                      </Link>
                      <p className="text-sm text-gray-400 mt-1">
                        {new Date(request.createdAt).toLocaleDateString()}
                      </p>
                    </div>
                    <span className={`px-3 py-1 text-sm rounded-lg ${getStatusColor(request.status)}`}>
                      {getStatusLabel(request.status)}
                    </span>
                  </div>
                </div>

                {/* Changes */}
                <div className="p-4 space-y-3">
                  <p className="text-sm text-gray-400">
                    <span className="font-medium">Reason:</span> {request.editReason}
                  </p>

                  <div className="space-y-2">
                    <p className="text-sm font-medium text-gray-300">Changes:</p>
                    {request.changes.map((change, idx) => (
                      <div key={idx} className="bg-[#121212] rounded-lg p-3 text-sm">
                        <p className="text-purple-400 font-medium mb-2">
                          {t(`fields.${change.fieldType}`) || change.fieldType}
                          {change.lang && ` (${change.lang})`}
                        </p>
                        <div className="grid grid-cols-2 gap-4">
                          <div>
                            <p className="text-gray-500 text-xs mb-1">Before:</p>
                            <p className="text-gray-400 line-clamp-3">
                              {change.oldValue || '(empty)'}
                            </p>
                          </div>
                          <div>
                            <p className="text-gray-500 text-xs mb-1">After:</p>
                            <p className="text-white line-clamp-3">
                              {change.newValue || '(empty)'}
                            </p>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>

                  {/* Moderator Comment */}
                  {request.moderatorComment && (
                    <div className="bg-gray-800 rounded-lg p-3">
                      <p className="text-sm text-gray-400">
                        <span className="font-medium">Moderator:</span> {request.moderatorComment}
                      </p>
                    </div>
                  )}
                </div>

                {/* Actions */}
                {request.status === 'pending' && (
                  <div className="p-4 border-t border-gray-700">
                    <button
                      onClick={() => handleCancel(request.id)}
                      className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded-lg text-sm"
                    >
                      {t('cancelRequest')}
                    </button>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
