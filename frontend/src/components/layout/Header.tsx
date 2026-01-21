'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useTranslations, useLocale } from 'next-intl';
import { useState, useEffect, useRef } from 'react';
import {
  Search,
  Bookmark,
  Bell,
  User,
  Menu,
  X,
  Wallet,
  LogIn,
  Globe,
  Settings,
  LogOut,
  Shield,
} from 'lucide-react';
import { locales, localeNames, localeFlags, type Locale } from '@/i18n/config';
import { useAuthStore, isAdmin } from '@/store/auth';

export function Header() {
  const t = useTranslations('nav');
  const locale = useLocale() as Locale;
  const pathname = usePathname();
  const router = useRouter();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [langMenuOpen, setLangMenuOpen] = useState(false);
  const [notificationsOpen, setNotificationsOpen] = useState(false);
  const [profileMenuOpen, setProfileMenuOpen] = useState(false);
  const [mounted, setMounted] = useState(false);
  
  const notificationsRef = useRef<HTMLDivElement>(null);
  const profileRef = useRef<HTMLDivElement>(null);
  const { user, isAuthenticated, logout } = useAuthStore();

  // Handle mount to prevent hydration mismatch
  useEffect(() => {
    setMounted(true);
  }, []);

  // Close menus on outside click
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (notificationsRef.current && !notificationsRef.current.contains(event.target as Node)) {
        setNotificationsOpen(false);
      }
      if (profileRef.current && !profileRef.current.contains(event.target as Node)) {
        setProfileMenuOpen(false);
      }
    };

    if (notificationsOpen || profileMenuOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [notificationsOpen, profileMenuOpen]);

  const handleLogout = () => {
    logout();
    setProfileMenuOpen(false);
    router.push(`/${locale}/login`);
  };

  // Navigation links
  const navLinks = [
    { href: `/${locale}`, label: t('home') },
    { href: `/${locale}/catalog`, label: t('catalog') },
    { href: `/${locale}/voting`, label: t('voting') },
  ];

  // Change locale
  const changeLocale = (newLocale: Locale) => {
    const newPath = pathname.replace(`/${locale}`, `/${newLocale}`);
    window.location.href = newPath;
  };

  return (
    <header className="sticky top-0 z-50 bg-background-secondary border-b border-border-primary">
      <div className="container-custom">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <Link href={`/${locale}`} className="flex items-center gap-2">
            <span className="text-2xl font-heading font-bold gradient-text">
              Novels
            </span>
          </Link>

          {/* Desktop Navigation */}
          <nav className="hidden md:flex items-center gap-6">
            {navLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className={`text-sm font-medium transition-colors hover:text-foreground-primary ${
                  pathname === link.href
                    ? 'text-accent-primary'
                    : 'text-foreground-secondary'
                }`}
              >
                {link.label}
              </Link>
            ))}
          </nav>

          {/* Search Bar */}
          <div className="hidden lg:flex flex-1 max-w-md mx-6">
            <div className="relative w-full">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-foreground-muted" />
              <input
                type="text"
                placeholder={t('home') + '...'}
                className="input pl-10 py-2 text-sm"
              />
            </div>
          </div>

          {/* Right Actions */}
          <div className="flex items-center gap-2">
            {/* Language Switcher */}
            <div className="relative">
              <button
                onClick={() => setLangMenuOpen(!langMenuOpen)}
                className="btn-ghost p-2"
                title="Change language"
              >
                <Globe className="w-5 h-5" />
              </button>
              {langMenuOpen && (
                <div className="dropdown right-0 min-w-40">
                  {locales.map((loc) => (
                    <button
                      key={loc}
                      onClick={() => {
                        changeLocale(loc);
                        setLangMenuOpen(false);
                      }}
                      className={`dropdown-item w-full text-left flex items-center gap-2 ${
                        locale === loc ? 'text-accent-primary' : ''
                      }`}
                    >
                      <span>{localeFlags[loc]}</span>
                      <span>{localeNames[loc]}</span>
                    </button>
                  ))}
                </div>
              )}
            </div>

            {/* Auth-dependent links - render both and hide one to prevent hydration mismatch */}
            {/* Authenticated Links */}
            {mounted && isAuthenticated && (
              <>
                {/* Wallet */}
                <Link href={`/${locale}/profile#wallet`} className="btn-ghost p-2">
                  <Wallet className="w-5 h-5" />
                </Link>

                {/* Bookmarks */}
                <Link href={`/${locale}/bookmarks`} className="btn-ghost p-2">
                  <Bookmark className="w-5 h-5" />
                </Link>

                {/* Notifications */}
                <div className="relative" ref={notificationsRef}>
                  <button
                    onClick={() => setNotificationsOpen(!notificationsOpen)}
                    className="btn-ghost p-2 relative"
                  >
                    <Bell className="w-5 h-5" />
                    <span className="absolute -top-1 -right-1 w-4 h-4 bg-accent-danger rounded-full text-[10px] flex items-center justify-center">
                      3
                    </span>
                  </button>
                  {notificationsOpen && (
                    <div className="dropdown right-0 w-80">
                      <div className="p-4 border-b border-border-primary flex items-center justify-between">
                        <h3 className="font-semibold">Уведомления</h3>
                        <button
                          onClick={() => setNotificationsOpen(false)}
                          className="p-1 hover:bg-background-hover rounded transition-colors"
                        >
                          <X className="w-4 h-4" />
                        </button>
                      </div>
                      <div className="max-h-96 overflow-y-auto">
                        {/* Placeholder notifications */}
                        <div className="p-4 hover:bg-background-hover transition-colors border-b border-border-primary cursor-pointer">
                          <div className="flex gap-3">
                            <div className="w-2 h-2 bg-accent-primary rounded-full mt-2 flex-shrink-0" />
                            <div className="flex-1">
                              <p className="text-sm font-medium">Новая глава доступна</p>
                              <p className="text-xs text-foreground-secondary mt-1">
                                Вышла новая глава в "Название новеллы"
                              </p>
                              <p className="text-xs text-foreground-muted mt-1">2 часа назад</p>
                            </div>
                          </div>
                        </div>
                        <div className="p-4 hover:bg-background-hover transition-colors border-b border-border-primary cursor-pointer">
                          <div className="flex gap-3">
                            <div className="w-2 h-2 bg-accent-primary rounded-full mt-2 flex-shrink-0" />
                            <div className="flex-1">
                              <p className="text-sm font-medium">Ваша подписка активна</p>
                              <p className="text-xs text-foreground-secondary mt-1">
                                Подписка Premium успешно активирована
                              </p>
                              <p className="text-xs text-foreground-muted mt-1">5 часов назад</p>
                            </div>
                          </div>
                        </div>
                        <div className="p-4 hover:bg-background-hover transition-colors cursor-pointer">
                          <div className="flex gap-3">
                            <div className="w-2 h-2 bg-accent-primary rounded-full mt-2 flex-shrink-0" />
                            <div className="flex-1">
                              <p className="text-sm font-medium">Новый комментарий</p>
                              <p className="text-xs text-foreground-secondary mt-1">
                                Кто-то ответил на ваш комментарий
                              </p>
                              <p className="text-xs text-foreground-muted mt-1">1 день назад</p>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}
                </div>

                {/* Profile */}
                <div className="relative ml-2" ref={profileRef}>
                  <button
                    onClick={() => setProfileMenuOpen(!profileMenuOpen)}
                    className="flex items-center gap-2"
                  >
                    <div className="w-8 h-8 rounded-full bg-accent-primary/20 flex items-center justify-center hover:bg-accent-primary/30 transition-colors">
                      <User className="w-4 h-4 text-accent-primary" />
                    </div>
                  </button>
                  {profileMenuOpen && (
                    <div className="dropdown right-0 w-48">
                      <div className="p-2 border-b border-border-primary">
                        <p className="text-sm font-medium truncate">{user?.displayName}</p>
                        <p className="text-xs text-foreground-muted truncate">{user?.email}</p>
                      </div>
                      <div className="p-2">
                        <Link
                          href={`/${locale}/profile`}
                          onClick={() => setProfileMenuOpen(false)}
                          className="dropdown-item w-full text-left flex items-center gap-2"
                        >
                          <User className="w-4 h-4" />
                          Профиль
                        </Link>
                        <Link
                          href={`/${locale}/profile/settings`}
                          onClick={() => setProfileMenuOpen(false)}
                          className="dropdown-item w-full text-left flex items-center gap-2"
                        >
                          <Settings className="w-4 h-4" />
                          Настройки
                        </Link>
                        {isAdmin(user) && (
                          <Link
                            href={`/${locale}/admin`}
                            onClick={() => setProfileMenuOpen(false)}
                            className="dropdown-item w-full text-left flex items-center gap-2 text-accent-warning hover:bg-accent-warning/10"
                          >
                            <Shield className="w-4 h-4" />
                            Админка
                          </Link>
                        )}
                        <button
                          onClick={handleLogout}
                          className="dropdown-item w-full text-left flex items-center gap-2 text-red-500 hover:bg-red-500/10"
                        >
                          <LogOut className="w-4 h-4" />
                          Выйти
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              </>
            )}

            {/* Unauthenticated Links - always render on server, hide on client if authenticated */}
            <div suppressHydrationWarning style={{ display: mounted && isAuthenticated ? 'none' : 'contents' }}>
              <Link href={`/${locale}/login`} className="btn-ghost">
                <LogIn className="w-4 h-4 mr-2" />
                {t('login')}
              </Link>
              <Link href={`/${locale}/register`} className="btn-primary hidden sm:flex">
                {t('register')}
              </Link>
            </div>

            {/* Mobile Menu Toggle */}
            <button
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
              className="btn-ghost p-2 md:hidden"
            >
              {mobileMenuOpen ? (
                <X className="w-5 h-5" />
              ) : (
                <Menu className="w-5 h-5" />
              )}
            </button>
          </div>
        </div>

        {/* Mobile Navigation */}
        {mobileMenuOpen && (
          <div className="md:hidden py-4 border-t border-border-primary animate-slide-down">
            {/* Mobile Search */}
            <div className="relative mb-4">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-foreground-muted" />
              <input
                type="text"
                placeholder="Поиск..."
                className="input pl-10 py-2 text-sm"
              />
            </div>

            {/* Mobile Nav Links */}
            <nav className="flex flex-col gap-2">
              {navLinks.map((link) => (
                <Link
                  key={link.href}
                  href={link.href}
                  onClick={() => setMobileMenuOpen(false)}
                  className={`px-3 py-2 rounded-button transition-colors ${
                    pathname === link.href
                      ? 'bg-accent-primary/10 text-accent-primary'
                      : 'text-foreground-secondary hover:bg-background-hover'
                  }`}
                >
                  {link.label}
                </Link>
              ))}
            </nav>
          </div>
        )}
      </div>
    </header>
  );
}
