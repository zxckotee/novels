'use client';

import React, { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { useRouter } from 'next/navigation';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, Play, RefreshCw, Pause, PlayCircle, X } from 'lucide-react';

import { useAuthStore, isAdmin } from '@/store/auth';
import { api } from '@/lib/api';
import { toast } from 'react-hot-toast';

type OpsTranslationTargetEntry = {
  targetId: string;
  status: 'voting' | 'waiting_release' | 'translating' | 'completed' | 'cancelled';
  score: number;
  novelId?: string | null;
  proposalId?: string | null;
  title: string;
  coverUrl?: string | null;
  updatedAt: string;
};

type ImportRun = {
  id: string;
  proposalId: string;
  novelId?: string | null;
  importer: string;
  status: 'running' | 'pause_requested' | 'paused' | 'succeeded' | 'failed' | 'cancelled';
  error?: string | null;
  cloudflareBlocked?: boolean;
  startedAt: string;
  finishedAt?: string | null;
  createdAt: string;
  updatedAt: string;
};

type ImportRunCookie = {
  runId: string;
  cookieHeader: string;
  createdAt: string;
  updatedAt: string;
};

const COOKIE_HINTS: Record<string, string> = {
  '101kks': 'cf_clearance=...; zh_choose=t',
  '69shuba': 'cf_clearance=...',
  'tadu': 'cf_clearance=... (если требуется)',
};

export default function AdminOpsPage() {
  const locale = useLocale();
  const router = useRouter();
  const queryClient = useQueryClient();
  const { isAuthenticated, user, isLoading } = useAuthStore();
  const hasAccess = isAuthenticated && isAdmin(user);
  const [forceWinner, setForceWinner] = useState(false);
  const [cloudflareFilter, setCloudflareFilter] = useState<boolean | null>(null);

  useEffect(() => {
    if (!isLoading && !hasAccess) router.replace(`/${locale}`);
  }, [isLoading, hasAccess, router, locale]);

  const { data: targetsData, isLoading: targetsLoading, refetch } = useQuery({
    queryKey: ['admin-ops', 'translation-targets'],
    queryFn: async () => {
      const res = await api.get<{ entries: OpsTranslationTargetEntry[] }>(`/admin/ops/translation-targets?limit=200`);
      return res.data;
    },
    enabled: hasAccess,
    refetchInterval: 30_000,
  });

  const targets = targetsData?.entries || [];

  const { data: importRunsData, isLoading: importRunsLoading, refetch: refetchRuns } = useQuery({
    queryKey: ['admin-ops', 'import-runs'],
    queryFn: async () => {
      const res = await api.get<{ runs: ImportRun[] }>(`/admin/ops/import-runs?limit=50`);
      return res.data;
    },
    enabled: hasAccess,
    refetchInterval: 10_000,
  });
  const runs = useMemo(() => {
    const allRuns = importRunsData?.runs || [];
    if (cloudflareFilter === null) return allRuns;
    return allRuns.filter((r) => (cloudflareFilter ? r.cloudflareBlocked : !r.cloudflareBlocked));
  }, [importRunsData?.runs, cloudflareFilter]);

  const runVotingWinner = useMutation({
    mutationFn: async () => {
      await api.post(`/admin/ops/jobs/voting-winner/run${forceWinner ? '?force=1' : ''}`, {});
    },
    onSuccess: async () => {
      toast.success('Запущено: выбор победителя (добавление)');
      await queryClient.invalidateQueries({ queryKey: ['voting-leaderboard'] });
    },
    onError: (e: any) => toast.error(e?.response?.data?.error?.message || 'Не удалось запустить'),
  });

  const runTranslationWinner = useMutation({
    mutationFn: async () => {
      await api.post(`/admin/ops/jobs/translation-winner/run${forceWinner ? '?force=1' : ''}`, {});
    },
    onSuccess: async () => {
      toast.success('Запущено: выбор победителя (перевод)');
      await queryClient.invalidateQueries({ queryKey: ['admin-ops', 'translation-targets'] });
      await queryClient.invalidateQueries({ queryKey: ['voting-leaderboard', 'translation'] });
    },
    onError: (e: any) => toast.error(e?.response?.data?.error?.message || 'Не удалось запустить'),
  });

  const setStatus = useMutation({
    mutationFn: async ({ id, status }: { id: string; status: OpsTranslationTargetEntry['status'] }) => {
      await api.post(`/admin/ops/translation-targets/${id}/status`, { status });
    },
    onSuccess: async () => {
      toast.success('Статус обновлен');
      await queryClient.invalidateQueries({ queryKey: ['admin-ops', 'translation-targets'] });
      await queryClient.invalidateQueries({ queryKey: ['voting-leaderboard', 'translation'] });
    },
    onError: (e: any) => toast.error(e?.response?.data?.error?.message || 'Не удалось обновить статус'),
  });

  const [statusFilter, setStatusFilter] = useState<'all' | OpsTranslationTargetEntry['status']>('all');
  const [proposalIdToRun, setProposalIdToRun] = useState('');
  const filtered = useMemo(() => {
    if (statusFilter === 'all') return targets;
    return targets.filter((t) => t.status === statusFilter);
  }, [targets, statusFilter]);

  const runImportNow = useMutation({
    mutationFn: async () => {
      const pid = proposalIdToRun.trim();
      if (!pid) throw new Error('proposalId required');
      const res = await api.post<{ runId: string }>(`/admin/ops/imports/run`, { proposalId: pid });
      return res.data;
    },
    onSuccess: async (d) => {
      toast.success(`Импорт запущен: ${d.runId.slice(0, 8)}`);
      setProposalIdToRun('');
      await queryClient.invalidateQueries({ queryKey: ['admin-ops', 'import-runs'] });
    },
    onError: (e: any) => toast.error(e?.response?.data?.error?.message || e?.message || 'Не удалось запустить импорт'),
  });

  const cancelRun = useMutation({
    mutationFn: async (runId: string) => {
      await api.post(`/admin/ops/import-runs/${runId}/cancel`, {});
    },
    onSuccess: async () => {
      toast.success('Отмена запрошена');
      await queryClient.invalidateQueries({ queryKey: ['admin-ops', 'import-runs'] });
    },
    onError: (e: any) => toast.error(e?.response?.data?.error?.message || 'Не удалось отменить'),
  });

  const pauseRun = useMutation({
    mutationFn: async (runId: string) => {
      await api.post(`/admin/ops/import-runs/${runId}/pause`, {});
    },
    onSuccess: async () => {
      toast.success('Пауза запрошена');
      await queryClient.invalidateQueries({ queryKey: ['admin-ops', 'import-runs'] });
    },
    onError: (e: any) => toast.error(e?.response?.data?.error?.message || 'Не удалось поставить на паузу'),
  });

  const resumeRun = useMutation({
    mutationFn: async (runId: string) => {
      await api.post(`/admin/ops/import-runs/${runId}/resume`, {});
    },
    onSuccess: async () => {
      toast.success('Продолжение запрошено');
      await queryClient.invalidateQueries({ queryKey: ['admin-ops', 'import-runs'] });
    },
    onError: (e: any) => toast.error(e?.response?.data?.error?.message || 'Не удалось продолжить'),
  });

  const [cookieInputs, setCookieInputs] = useState<Record<string, string>>({});
  const [expandedRuns, setExpandedRuns] = useState<Set<string>>(new Set());

  const getCookies = useMutation({
    mutationFn: async (runId: string) => {
      const res = await api.get<{ cookie: ImportRunCookie | null }>(`/admin/ops/import-runs/${runId}/cookies`);
      return res.data.cookie;
    },
    onSuccess: (cookie, runId) => {
      if (cookie) {
        setCookieInputs((prev) => ({ ...prev, [runId]: cookie.cookieHeader }));
      }
    },
  });

  const updateCookies = useMutation({
    mutationFn: async ({ runId, cookieHeader }: { runId: string; cookieHeader: string }) => {
      await api.put(`/admin/ops/import-runs/${runId}/cookies`, { cookieHeader });
    },
    onSuccess: async () => {
      toast.success('Куки сохранены');
      await queryClient.invalidateQueries({ queryKey: ['admin-ops', 'import-runs'] });
    },
    onError: (e: any) => toast.error(e?.response?.data?.error?.message || 'Не удалось сохранить куки'),
  });

  const retryRun = useMutation({
    mutationFn: async (runId: string) => {
      const res = await api.post<{ runId: string; message: string }>(`/admin/ops/import-runs/${runId}/retry`, {});
      return res.data;
    },
    onSuccess: async (d) => {
      toast.success(`Перезапуск запущен: ${d.runId.slice(0, 8)}`);
      await queryClient.invalidateQueries({ queryKey: ['admin-ops', 'import-runs'] });
    },
    onError: (e: any) => toast.error(e?.response?.data?.error?.message || 'Не удалось перезапустить'),
  });

  const handleToggleExpand = (runId: string) => {
    setExpandedRuns((prev) => {
      const next = new Set(prev);
      if (next.has(runId)) {
        next.delete(runId);
      } else {
        next.add(runId);
        if (!cookieInputs[runId]) {
          getCookies.mutate(runId);
        }
      }
      return next;
    });
  };

  const handleSaveAndRetry = (runId: string) => {
    const cookieHeader = cookieInputs[runId]?.trim();
    if (!cookieHeader) {
      toast.error('Введите куки');
      return;
    }
    updateCookies.mutate(
      { runId, cookieHeader },
      {
        onSuccess: () => {
          retryRun.mutate(runId);
        },
      }
    );
  };

  if (isLoading) return null;
  if (!hasAccess) return null;

  return (
    <div className="container-custom py-6">
      <div className="flex items-center gap-4 mb-6">
        <Link href={`/${locale}/admin`} className="btn-ghost p-2">
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <h1 className="text-2xl font-heading font-bold">Ops: джобы и очереди</h1>
        <div className="ml-auto flex gap-2">
          <button onClick={() => refetch()} className="btn-secondary flex items-center gap-2">
            <RefreshCw className="w-4 h-4" /> Обновить
          </button>
        </div>
      </div>

      {/* Job controls */}
      <div className="bg-background-secondary rounded-card p-6 mb-6">
        <h2 className="text-lg font-semibold mb-4">Ручной запуск</h2>
        <label className="flex items-center gap-2 mb-4 text-sm text-foreground-secondary">
          <input
            type="checkbox"
            checked={forceWinner}
            onChange={(e) => setForceWinner(e.target.checked)}
          />
          Force (для тестов): выбрать победителя даже при 0 голосов/билетов
        </label>
        <div className="flex flex-wrap gap-3">
          <button
            className="btn-primary flex items-center gap-2"
            onClick={() => runVotingWinner.mutate()}
            disabled={runVotingWinner.isPending}
          >
            <Play className="w-4 h-4" />
            Выбрать победителя (добавление)
          </button>
          <button
            className="btn-primary flex items-center gap-2"
            onClick={() => runTranslationWinner.mutate()}
            disabled={runTranslationWinner.isPending}
          >
            <Play className="w-4 h-4" />
            Выбрать победителя (перевод)
          </button>
        </div>
        <p className="text-sm text-foreground-secondary mt-3">
          Эти кнопки запускают выбор победителя сразу, без ожидания планировщика (каждые 6 часов).
        </p>
      </div>

      {/* Import runs */}
      <div className="bg-background-secondary rounded-card p-6 mb-6">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 mb-4">
          <h2 className="text-lg font-semibold">Парсинг/импорт: runs</h2>
          <div className="flex items-center gap-2">
            <select
              className="bg-background-tertiary rounded-lg px-3 py-2 text-sm"
              value={cloudflareFilter === null ? 'all' : cloudflareFilter ? 'blocked' : 'not-blocked'}
              onChange={(e) => {
                const val = e.target.value;
                setCloudflareFilter(val === 'all' ? null : val === 'blocked');
              }}
            >
              <option value="all">Все</option>
              <option value="blocked">Только Cloudflare</option>
              <option value="not-blocked">Без Cloudflare</option>
            </select>
            <button onClick={() => refetchRuns()} className="btn-secondary flex items-center gap-2">
              <RefreshCw className="w-4 h-4" /> Обновить
            </button>
          </div>
        </div>

        <div className="flex flex-col md:flex-row gap-3 mb-4">
          <input
            className="w-full md:w-[420px] bg-background-tertiary rounded-lg px-3 py-2 text-sm"
            placeholder="proposalId (uuid) чтобы запустить импорт вручную"
            value={proposalIdToRun}
            onChange={(e) => setProposalIdToRun(e.target.value)}
          />
          <button
            className="btn-primary flex items-center gap-2"
            onClick={() => runImportNow.mutate()}
            disabled={runImportNow.isPending}
          >
            <Play className="w-4 h-4" />
            Запустить импорт
          </button>
          <p className="text-sm text-foreground-secondary md:self-center">
            Для отладки/ручного контроля. Обычно импорт запускается после выбора победителя.
          </p>
        </div>

        {importRunsLoading ? (
          <div className="text-center py-10">
            <p className="text-foreground-secondary">Загрузка...</p>
          </div>
        ) : runs.length === 0 ? (
          <div className="text-center py-10">
            <p className="text-foreground-secondary">Пока нет запусков</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-background-tertiary">
                  <th className="text-left py-2 px-3">Run</th>
                  <th className="text-left py-2 px-3">Importer</th>
                  <th className="text-left py-2 px-3">Status</th>
                  <th className="text-left py-2 px-3">Started</th>
                  <th className="text-left py-2 px-3">Finished</th>
                  <th className="text-left py-2 px-3">Proposal</th>
                  <th className="text-left py-2 px-3">Error</th>
                  <th className="text-left py-2 px-3"></th>
                </tr>
              </thead>
              <tbody>
                {runs.map((r) => (
                  <React.Fragment key={r.id}>
                    <tr className="border-b border-background-tertiary hover:bg-background-hover">
                      <td className="py-2 px-3 font-mono text-xs text-foreground-secondary">{r.id.slice(0, 8)}</td>
                      <td className="py-2 px-3">
                        {r.importer}
                        {r.cloudflareBlocked && (
                          <span className="ml-2 px-2 py-0.5 bg-red-500/20 text-red-400 text-xs rounded">
                            Cloudflare
                          </span>
                        )}
                      </td>
                      <td className="py-2 px-3 font-mono">{r.status}</td>
                      <td className="py-2 px-3 text-foreground-secondary whitespace-nowrap">
                        {new Date(r.startedAt).toLocaleString('ru-RU')}
                      </td>
                      <td className="py-2 px-3 text-foreground-secondary whitespace-nowrap">
                        {r.finishedAt ? new Date(r.finishedAt).toLocaleString('ru-RU') : '-'}
                      </td>
                      <td className="py-2 px-3 font-mono text-xs text-foreground-secondary">
                        {r.proposalId.slice(0, 8)}
                      </td>
                      <td className="py-2 px-3 text-foreground-secondary max-w-[520px] truncate" title={r.error || ''}>
                        {r.error || '-'}
                      </td>
                      <td className="py-2 px-3">
                        <div className="flex items-center gap-2">
                          {r.status === 'running' ? (
                            <>
                              <button
                                className="btn-secondary text-sm flex items-center gap-1"
                                onClick={() => pauseRun.mutate(r.id)}
                                disabled={pauseRun.isPending}
                                title="Поставить на паузу (сохранить прогресс)"
                              >
                                <Pause className="w-3 h-3" /> Пауза
                              </button>
                              <button
                                className="btn-secondary text-sm flex items-center gap-1"
                                onClick={() => cancelRun.mutate(r.id)}
                                disabled={cancelRun.isPending}
                                title="Прервать (сохранить что успели)"
                              >
                                <X className="w-3 h-3" /> Отмена
                              </button>
                            </>
                          ) : r.status === 'failed' && r.cloudflareBlocked ? (
                            <>
                              <button
                                className="btn-secondary text-sm flex items-center gap-1"
                                onClick={() => handleToggleExpand(r.id)}
                                disabled={getCookies.isPending}
                              >
                                {expandedRuns.has(r.id) ? 'Скрыть' : 'Куки'}
                              </button>
                            </>
                          ) : (r.status === 'paused' || r.status === 'pause_requested' || r.status === 'cancelled' || (r.status === 'failed' && !r.cloudflareBlocked)) ? (
                            <button
                              className="btn-secondary text-sm flex items-center gap-1"
                              onClick={() => resumeRun.mutate(r.id)}
                              disabled={resumeRun.isPending}
                              title="Продолжить с места остановки"
                            >
                              <PlayCircle className="w-3 h-3" /> Продолжить
                            </button>
                          ) : null}
                        </div>
                      </td>
                    </tr>
                    {r.status === 'failed' && r.cloudflareBlocked && expandedRuns.has(r.id) && (
                      <tr key={`${r.id}-expand`} className="border-b border-background-tertiary bg-background-tertiary/50">
                        <td colSpan={8} className="py-4 px-3">
                          <div className="space-y-3">
                            <div className="text-sm">
                              <p className="text-foreground-secondary mb-2">
                                Для сайта <strong>{r.importer}</strong> нужны куки:
                              </p>
                              <code className="block bg-background-primary p-2 rounded text-xs mb-2">
                                {COOKIE_HINTS[r.importer] || 'cf_clearance=...'}
                              </code>
                              <p className="text-xs text-foreground-secondary mb-3">
                                Скопируйте значение из DevTools → Application → Cookies или из curl запроса
                              </p>
                            </div>
                            <textarea
                              className="w-full bg-background-primary rounded-lg px-3 py-2 text-sm font-mono text-xs min-h-[80px]"
                              placeholder="Вставьте cookie_header (например: cf_clearance=...; zh_choose=t)"
                              value={cookieInputs[r.id] || ''}
                              onChange={(e) => setCookieInputs((prev) => ({ ...prev, [r.id]: e.target.value }))}
                            />
                            <div className="flex gap-2">
                              <button
                                className="btn-primary text-sm"
                                onClick={() => handleSaveAndRetry(r.id)}
                                disabled={updateCookies.isPending || retryRun.isPending}
                              >
                                {updateCookies.isPending || retryRun.isPending ? 'Сохранение...' : 'Сохранить куки и перезапустить'}
                              </button>
                              <button
                                className="btn-secondary text-sm"
                                onClick={() => updateCookies.mutate({ runId: r.id, cookieHeader: cookieInputs[r.id] || '' })}
                                disabled={updateCookies.isPending || !cookieInputs[r.id]?.trim()}
                              >
                                Только сохранить
                              </button>
                            </div>
                          </div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Translation targets */}
      <div className="bg-background-secondary rounded-card p-6">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 mb-4">
          <h2 className="text-lg font-semibold">Очередь перевода (targets)</h2>
          <div className="flex items-center gap-2">
            <span className="text-sm text-foreground-secondary">Фильтр:</span>
            <select
              className="bg-background-tertiary rounded-lg px-3 py-2 text-sm"
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as any)}
            >
              <option value="all">Все</option>
              <option value="voting">voting</option>
              <option value="waiting_release">waiting_release</option>
              <option value="translating">translating</option>
              <option value="completed">completed</option>
              <option value="cancelled">cancelled</option>
            </select>
          </div>
        </div>

        {targetsLoading ? (
          <div className="text-center py-12">
            <p className="text-foreground-secondary">Загрузка...</p>
          </div>
        ) : filtered.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-foreground-secondary">Пусто</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-background-tertiary">
                  <th className="text-left py-2 px-3">Title</th>
                  <th className="text-left py-2 px-3">Tickets</th>
                  <th className="text-left py-2 px-3">Status</th>
                  <th className="text-left py-2 px-3">Updated</th>
                  <th className="text-left py-2 px-3">IDs</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((t) => (
                  <tr key={t.targetId} className="border-b border-background-tertiary hover:bg-background-hover">
                    <td className="py-2 px-3 font-medium">{t.title}</td>
                    <td className="py-2 px-3 text-foreground-secondary">{t.score}</td>
                    <td className="py-2 px-3">
                      <select
                        className="bg-background-tertiary rounded-lg px-2 py-1"
                        value={t.status}
                        onChange={(e) =>
                          setStatus.mutate({ id: t.targetId, status: e.target.value as OpsTranslationTargetEntry['status'] })
                        }
                        disabled={setStatus.isPending}
                      >
                        <option value="voting">voting</option>
                        <option value="waiting_release">waiting_release</option>
                        <option value="translating">translating</option>
                        <option value="completed">completed</option>
                        <option value="cancelled">cancelled</option>
                      </select>
                    </td>
                    <td className="py-2 px-3 text-foreground-secondary whitespace-nowrap">
                      {new Date(t.updatedAt).toLocaleString('ru-RU')}
                    </td>
                    <td className="py-2 px-3 font-mono text-xs text-foreground-secondary">
                      <div>target: {t.targetId.slice(0, 8)}</div>
                      <div>proposal: {t.proposalId ? t.proposalId.slice(0, 8) : '-'}</div>
                      <div>novel: {t.novelId ? t.novelId.slice(0, 8) : '-'}</div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

