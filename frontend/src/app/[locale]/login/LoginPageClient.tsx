'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useTranslations, useLocale } from 'next-intl';
import { useLogin } from '@/lib/api/hooks/useAuth';
import { useAuthStore } from '@/store/auth';
import { toast } from 'react-hot-toast';
import { Mail, Lock, LogIn } from 'lucide-react';

export default function LoginPageClient() {
  const t = useTranslations('auth.login');
  const tErrors = useTranslations('auth.errors');
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  const loginMutation = useLogin();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [mounted, setMounted] = useState(false);

  // Handle mount to prevent hydration mismatch
  useEffect(() => {
    setMounted(true);
  }, []);

  // Redirect if already authenticated (only on client)
  useEffect(() => {
    if (mounted && isAuthenticated) {
      router.push(`/${locale}/profile`);
    }
  }, [mounted, isAuthenticated, locale, router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!email || !password) {
      toast.error(tErrors('invalidCredentials'));
      return;
    }

    setIsLoading(true);
    try {
      await loginMutation.mutateAsync({ email, password });
      toast.success(t('submit'));
      router.push(`/${locale}/profile`);
      router.refresh();
    } catch (error: any) {
      const errorMessage = error?.response?.data?.error || tErrors('invalidCredentials');
      toast.error(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="container-custom py-12">
      <div className="max-w-md mx-auto">
        <div className="bg-background-secondary rounded-card p-8">
          {/* Header */}
          <div className="text-center mb-8">
            <h1 className="text-3xl font-heading font-bold mb-2">{t('title')}</h1>
            <p className="text-foreground-secondary">
              {t('noAccount')}{' '}
              <Link 
                href={`/${locale}/register`} 
                className="text-accent-primary hover:underline"
              >
                {t('register')}
              </Link>
            </p>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Email Field */}
            <div>
              <label htmlFor="email" className="block text-sm font-medium mb-2">
                {t('email')}
              </label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground-muted" />
                <input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full pl-10 pr-4 py-3 bg-background-tertiary border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-accent-primary focus:border-transparent"
                  placeholder={t('email')}
                  required
                  disabled={isLoading}
                  autoComplete="email"
                />
              </div>
            </div>

            {/* Password Field */}
            <div>
              <label htmlFor="password" className="block text-sm font-medium mb-2">
                {t('password')}
              </label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-foreground-muted" />
                <input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full pl-10 pr-4 py-3 bg-background-tertiary border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-accent-primary focus:border-transparent"
                  placeholder={t('password')}
                  required
                  disabled={isLoading}
                  autoComplete="current-password"
                />
              </div>
            </div>

            {/* Forgot Password Link */}
            <div className="text-right">
              <Link 
                href={`/${locale}/forgot-password`} 
                className="text-sm text-accent-primary hover:underline"
              >
                {t('forgotPassword')}
              </Link>
            </div>

            {/* Submit Button */}
            <button
              type="submit"
              disabled={isLoading}
              className="w-full btn-primary flex items-center justify-center gap-2 py-3"
            >
              {isLoading ? (
                <>
                  <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  {t('submit')}
                </>
              ) : (
                <>
                  <LogIn className="w-5 h-5" />
                  {t('submit')}
                </>
              )}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
