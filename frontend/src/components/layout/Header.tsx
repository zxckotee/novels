'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { useTranslations, useLocale } from 'next-intl';
import { useState } from 'react';
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
} from 'lucide-react';
import { locales, localeNames, localeFlags, type Locale } from '@/i18n/config';
import { useAuthStore } from '@/store/auth';

export function Header() {
  const t = useTranslations('nav');
  const locale = useLocale() as Locale;
  const pathname = usePathname();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [langMenuOpen, setLangMenuOpen] = useState(false);
  
  const { user, isAuthenticated } = useAuthStore();

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

            {isAuthenticated ? (
              <>
                {/* Wallet */}
                <Link href={`/${locale}/wallet`} className="btn-ghost p-2">
                  <Wallet className="w-5 h-5" />
                </Link>

                {/* Bookmarks */}
                <Link href={`/${locale}/bookmarks`} className="btn-ghost p-2">
                  <Bookmark className="w-5 h-5" />
                </Link>

                {/* Notifications */}
                <button className="btn-ghost p-2 relative">
                  <Bell className="w-5 h-5" />
                  <span className="absolute -top-1 -right-1 w-4 h-4 bg-accent-danger rounded-full text-[10px] flex items-center justify-center">
                    3
                  </span>
                </button>

                {/* Profile */}
                <Link
                  href={`/${locale}/profile`}
                  className="flex items-center gap-2 ml-2"
                >
                  <div className="w-8 h-8 rounded-full bg-accent-primary/20 flex items-center justify-center">
                    <User className="w-4 h-4 text-accent-primary" />
                  </div>
                </Link>
              </>
            ) : (
              <>
                <Link href={`/${locale}/login`} className="btn-ghost">
                  <LogIn className="w-4 h-4 mr-2" />
                  {t('login')}
                </Link>
                <Link href={`/${locale}/register`} className="btn-primary hidden sm:flex">
                  {t('register')}
                </Link>
              </>
            )}

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
