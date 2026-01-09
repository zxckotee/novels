import Link from 'next/link';
import { useLocale, useTranslations } from 'next-intl';
import { Github, Twitter, MessageCircle } from 'lucide-react';

export function Footer() {
  const t = useTranslations('nav');
  const locale = useLocale();
  const currentYear = new Date().getFullYear();

  const footerLinks = {
    platform: [
      { href: `/${locale}/catalog`, label: t('catalog') },
      { href: `/${locale}/voting`, label: t('voting') },
      { href: `/${locale}/news`, label: 'Новости' },
    ],
    support: [
      { href: `/${locale}/faq`, label: 'FAQ' },
      { href: `/${locale}/contact`, label: 'Контакты' },
      { href: `/${locale}/donate`, label: 'Поддержать' },
    ],
    legal: [
      { href: `/${locale}/terms`, label: 'Условия' },
      { href: `/${locale}/privacy`, label: 'Конфиденциальность' },
    ],
  };

  return (
    <footer className="bg-background-secondary border-t border-border-primary mt-auto">
      <div className="container-custom py-12">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          {/* Logo & Description */}
          <div className="col-span-1">
            <Link href={`/${locale}`} className="inline-block mb-4">
              <span className="text-2xl font-heading font-bold gradient-text">
                Novels
              </span>
            </Link>
            <p className="text-foreground-secondary text-sm mb-4">
              Платформа для чтения переведённых новелл с поддержкой 7 языков.
            </p>
            {/* Social Links */}
            <div className="flex items-center gap-3">
              <a
                href="https://github.com"
                target="_blank"
                rel="noopener noreferrer"
                className="text-foreground-muted hover:text-foreground-primary transition-colors"
              >
                <Github className="w-5 h-5" />
              </a>
              <a
                href="https://twitter.com"
                target="_blank"
                rel="noopener noreferrer"
                className="text-foreground-muted hover:text-foreground-primary transition-colors"
              >
                <Twitter className="w-5 h-5" />
              </a>
              <a
                href="https://discord.gg"
                target="_blank"
                rel="noopener noreferrer"
                className="text-foreground-muted hover:text-foreground-primary transition-colors"
              >
                <MessageCircle className="w-5 h-5" />
              </a>
            </div>
          </div>

          {/* Platform Links */}
          <div>
            <h4 className="font-heading font-semibold mb-4">Платформа</h4>
            <ul className="space-y-2">
              {footerLinks.platform.map((link) => (
                <li key={link.href}>
                  <Link
                    href={link.href}
                    className="text-foreground-secondary text-sm hover:text-foreground-primary transition-colors"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* Support Links */}
          <div>
            <h4 className="font-heading font-semibold mb-4">Поддержка</h4>
            <ul className="space-y-2">
              {footerLinks.support.map((link) => (
                <li key={link.href}>
                  <Link
                    href={link.href}
                    className="text-foreground-secondary text-sm hover:text-foreground-primary transition-colors"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* Legal Links */}
          <div>
            <h4 className="font-heading font-semibold mb-4">Правовая информация</h4>
            <ul className="space-y-2">
              {footerLinks.legal.map((link) => (
                <li key={link.href}>
                  <Link
                    href={link.href}
                    className="text-foreground-secondary text-sm hover:text-foreground-primary transition-colors"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="border-t border-border-primary mt-8 pt-8 flex flex-col md:flex-row justify-between items-center gap-4">
          <p className="text-foreground-muted text-sm">
            © {currentYear} Novels. Все права защищены.
          </p>
          <p className="text-foreground-muted text-sm">
            Сделано с ❤️ для читателей
          </p>
        </div>
      </div>
    </footer>
  );
}
