'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useTranslations, useLocale } from 'next-intl';
import { User, Mail, Lock, Save, ArrowLeft } from 'lucide-react';
import { useAuthStore } from '@/store/auth';
import Link from 'next/link';

export default function SettingsPageClient() {
  const t = useTranslations('profile');
  const locale = useLocale();
  const router = useRouter();
  const { isAuthenticated, user } = useAuthStore();
  
  const [mounted, setMounted] = useState(false);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  
  // Form state
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  
  // Handle mount to prevent hydration mismatch
  useEffect(() => {
    setMounted(true);
  }, []);
  
  // Redirect if not authenticated
  useEffect(() => {
    if (mounted && !isAuthenticated) {
      router.push(`/${locale}/login`);
    }
  }, [mounted, isAuthenticated, locale, router]);
  
  // Initialize form with user data
  useEffect(() => {
    if (user) {
      setDisplayName(user.displayName || '');
      setEmail(user.email || '');
    }
  }, [user]);
  
  // Show loading or redirect during initial mount
  if (!mounted || !isAuthenticated) {
    return <SettingsSkeleton />;
  }
  
  const handleProfileUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMessage(null);
    
    try {
      // TODO: Implement API call to update profile
      // const response = await fetch('/api/profile', {
      //   method: 'PUT',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({ displayName, email })
      // });
      
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      setMessage({ type: 'success', text: 'Профиль успешно обновлён' });
    } catch (error) {
      setMessage({ type: 'error', text: 'Ошибка при обновлении профиля' });
    } finally {
      setLoading(false);
    }
  };
  
  const handlePasswordChange = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMessage(null);
    
    if (newPassword !== confirmPassword) {
      setMessage({ type: 'error', text: 'Пароли не совпадают' });
      setLoading(false);
      return;
    }
    
    if (newPassword.length < 8) {
      setMessage({ type: 'error', text: 'Пароль должен быть не менее 8 символов' });
      setLoading(false);
      return;
    }
    
    try {
      // TODO: Implement API call to change password
      // const response = await fetch('/api/profile/password', {
      //   method: 'PUT',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify({ currentPassword, newPassword })
      // });
      
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      setMessage({ type: 'success', text: 'Пароль успешно изменён' });
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (error) {
      setMessage({ type: 'error', text: 'Ошибка при изменении пароля' });
    } finally {
      setLoading(false);
    }
  };
  
  return (
    <div className="container-custom py-6 max-w-4xl">
      {/* Header */}
      <div className="mb-6">
        <Link 
          href={`/${locale}/profile`}
          className="inline-flex items-center text-accent-primary hover:underline mb-3"
        >
          <ArrowLeft className="w-4 h-4 mr-1" />
          Вернуться к профилю
        </Link>
        <h1 className="text-3xl font-heading font-bold">Настройки профиля</h1>
      </div>
      
      {/* Message */}
      {message && (
        <div className={`mb-6 p-4 rounded-lg ${
          message.type === 'success' 
            ? 'bg-status-success/10 text-status-success border border-status-success/20' 
            : 'bg-status-error/10 text-status-error border border-status-error/20'
        }`}>
          {message.text}
        </div>
      )}
      
      {/* Profile Information */}
      <div className="bg-background-secondary rounded-card p-6 mb-6">
        <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
          <User className="w-5 h-5" />
          Основная информация
        </h2>
        
        <form onSubmit={handleProfileUpdate} className="space-y-4">
          <div>
            <label htmlFor="displayName" className="block text-sm font-medium mb-2">
              Имя пользователя
            </label>
            <input
              type="text"
              id="displayName"
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              className="w-full px-4 py-2 bg-background-primary border border-foreground-muted/20 rounded-lg focus:outline-none focus:border-accent-primary"
              required
            />
          </div>
          
          <div>
            <label htmlFor="email" className="block text-sm font-medium mb-2 flex items-center gap-2">
              <Mail className="w-4 h-4" />
              Email
            </label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-4 py-2 bg-background-primary border border-foreground-muted/20 rounded-lg focus:outline-none focus:border-accent-primary"
              required
            />
          </div>
          
          <div className="pt-4">
            <button
              type="submit"
              disabled={loading}
              className="btn-primary flex items-center gap-2"
            >
              <Save className="w-4 h-4" />
              {loading ? 'Сохранение...' : 'Сохранить изменения'}
            </button>
          </div>
        </form>
      </div>
      
      {/* Change Password */}
      <div className="bg-background-secondary rounded-card p-6">
        <h2 className="text-xl font-semibold mb-4 flex items-center gap-2">
          <Lock className="w-5 h-5" />
          Изменить пароль
        </h2>
        
        <form onSubmit={handlePasswordChange} className="space-y-4">
          <div>
            <label htmlFor="currentPassword" className="block text-sm font-medium mb-2">
              Текущий пароль
            </label>
            <input
              type="password"
              id="currentPassword"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              className="w-full px-4 py-2 bg-background-primary border border-foreground-muted/20 rounded-lg focus:outline-none focus:border-accent-primary"
              required
            />
          </div>
          
          <div>
            <label htmlFor="newPassword" className="block text-sm font-medium mb-2">
              Новый пароль
            </label>
            <input
              type="password"
              id="newPassword"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              className="w-full px-4 py-2 bg-background-primary border border-foreground-muted/20 rounded-lg focus:outline-none focus:border-accent-primary"
              required
              minLength={8}
            />
            <p className="text-xs text-foreground-muted mt-1">Минимум 8 символов</p>
          </div>
          
          <div>
            <label htmlFor="confirmPassword" className="block text-sm font-medium mb-2">
              Подтвердите новый пароль
            </label>
            <input
              type="password"
              id="confirmPassword"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="w-full px-4 py-2 bg-background-primary border border-foreground-muted/20 rounded-lg focus:outline-none focus:border-accent-primary"
              required
              minLength={8}
            />
          </div>
          
          <div className="pt-4">
            <button
              type="submit"
              disabled={loading}
              className="btn-primary flex items-center gap-2"
            >
              <Save className="w-4 h-4" />
              {loading ? 'Изменение...' : 'Изменить пароль'}
            </button>
          </div>
        </form>
      </div>
      
      {/* Preferences (Future) */}
      <div className="bg-background-secondary rounded-card p-6 mt-6">
        <h2 className="text-xl font-semibold mb-4">Настройки интерфейса</h2>
        <p className="text-foreground-secondary">
          Дополнительные настройки будут доступны в следующем обновлении
        </p>
      </div>
    </div>
  );
}

// Skeleton
function SettingsSkeleton() {
  return (
    <div className="container-custom py-6 max-w-4xl animate-pulse">
      <div className="mb-6">
        <div className="h-4 w-32 bg-background-hover rounded mb-3" />
        <div className="h-8 w-48 bg-background-hover rounded" />
      </div>
      
      <div className="bg-background-secondary rounded-card p-6 mb-6">
        <div className="h-6 w-48 bg-background-hover rounded mb-4" />
        <div className="space-y-4">
          <div className="h-10 bg-background-hover rounded" />
          <div className="h-10 bg-background-hover rounded" />
          <div className="h-10 w-32 bg-background-hover rounded" />
        </div>
      </div>
      
      <div className="bg-background-secondary rounded-card p-6">
        <div className="h-6 w-48 bg-background-hover rounded mb-4" />
        <div className="space-y-4">
          <div className="h-10 bg-background-hover rounded" />
          <div className="h-10 bg-background-hover rounded" />
          <div className="h-10 bg-background-hover rounded" />
          <div className="h-10 w-32 bg-background-hover rounded" />
        </div>
      </div>
    </div>
  );
}
