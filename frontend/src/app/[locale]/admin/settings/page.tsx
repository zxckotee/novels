'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useLocale } from 'next-intl';
import { ArrowLeft, Settings as SettingsIcon } from 'lucide-react';
import { useAuthStore, isAdmin } from '@/store/auth';
import { useAdminSettings } from '@/lib/api/hooks/useAdminSystem';

export default function AdminSettingsPage() {
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user, isLoading: authLoading } = useAuthStore();
  const hasAccess = isAuthenticated && isAdmin(user);

  useEffect(() => {
    if (!authLoading && !hasAccess) router.replace(`/${locale}`);
  }, [authLoading, hasAccess, router, locale]);

  const { data: settings, isLoading } = useAdminSettings();

  if (authLoading) return null;
  if (!hasAccess) return null;
  
  return (
    <div className="container-custom py-6">
      <div className="flex items-center gap-4 mb-6">
        <Link href={`/${locale}/admin`} className="btn-ghost p-2"><ArrowLeft className="w-5 h-5" /></Link>
        <h1 className="text-2xl font-heading font-bold">Настройки приложения</h1>
      </div>
      
      <div className="bg-background-secondary rounded-card p-6">
        {isLoading ? (
          <div className="text-center py-12"><p className="textforeground-secondary">Загрузка...</p></div>
        ) : !settings || settings.length === 0 ? (
          <div className="text-center py-12"><p className="text-foreground-secondary">Настройки не найдены</p></div>
        ) : (
          <div className="space-y-4">
            {settings.map((setting) => (
              <div key={setting.key} className="p-4 border border-background-tertiary rounded">
                <div className="flex justify-between items-start mb-2">
                  <div>
                    <h3 className="font-semibold font-mono text-sm">{setting.key}</h3>
                    {setting.description && (
                      <p className="text-sm text-foreground-secondary mt-1">{setting.description}</p>
                    )}
                  </div>
                  <div className="text-xs text-foreground-muted">
                    {new Date(setting.updatedAt).toLocaleDateString('ru-RU')}
                  </div>
                </div>
                <pre className="bg-background-tertiary p-2 rounded text-xs overflow-x-auto">
                  {JSON.stringify(setting.value, null, 2)}
                </pre>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
