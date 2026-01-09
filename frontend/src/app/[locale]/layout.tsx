import { ReactNode } from 'react';
import { Inter, Manrope } from 'next/font/google';
import { notFound } from 'next/navigation';
import { NextIntlClientProvider, useMessages } from 'next-intl';
import { locales, type Locale } from '@/i18n/config';
import '@/styles/globals.css';
import { Providers } from './providers';
import { Header } from '@/components/layout/Header';
import { Footer } from '@/components/layout/Footer';

const inter = Inter({
  subsets: ['latin', 'cyrillic'],
  variable: '--font-inter',
});

const manrope = Manrope({
  subsets: ['latin', 'cyrillic'],
  variable: '--font-manrope',
});

interface LocaleLayoutProps {
  children: ReactNode;
  params: { locale: string };
}

export function generateStaticParams() {
  return locales.map((locale) => ({ locale }));
}

export async function generateMetadata({ params: { locale } }: LocaleLayoutProps) {
  const titles: Record<Locale, string> = {
    ru: 'Novels - Платформа для чтения новелл',
    en: 'Novels - Novel Reading Platform',
    zh: 'Novels - 小说阅读平台',
    ja: 'Novels - 小説閲覧プラットフォーム',
    ko: 'Novels - 소설 읽기 플랫폼',
    fr: 'Novels - Plateforme de lecture de romans',
    de: 'Novels - Roman-Leseplattform',
  };

  const descriptions: Record<Locale, string> = {
    ru: 'Читайте тысячи переведённых новелл на 7 языках. Голосуйте за любимые книги и предлагайте новые.',
    en: 'Read thousands of translated novels in 7 languages. Vote for your favorite books and suggest new ones.',
    zh: '阅读7种语言翻译的数千本小说。为您喜欢的书籍投票并提出新建议。',
    ja: '7か国語で翻訳された何千もの小説を読みましょう。お気に入りの本に投票し、新しい本を提案してください。',
    ko: '7개 언어로 번역된 수천 개의 소설을 읽으세요. 좋아하는 책에 투표하고 새로운 책을 제안하세요.',
    fr: 'Lisez des milliers de romans traduits en 7 langues. Votez pour vos livres préférés et proposez-en de nouveaux.',
    de: 'Lesen Sie Tausende von übersetzten Romanen in 7 Sprachen. Stimmen Sie für Ihre Lieblingsbücher und schlagen Sie neue vor.',
  };

  return {
    title: {
      template: `%s | ${titles[locale as Locale] || titles.en}`,
      default: titles[locale as Locale] || titles.en,
    },
    description: descriptions[locale as Locale] || descriptions.en,
    openGraph: {
      title: titles[locale as Locale] || titles.en,
      description: descriptions[locale as Locale] || descriptions.en,
      type: 'website',
      locale: locale,
    },
  };
}

export default function LocaleLayout({
  children,
  params: { locale },
}: LocaleLayoutProps) {
  // Validate locale
  if (!locales.includes(locale as Locale)) {
    notFound();
  }

  // Get messages for the locale
  const messages = useMessages();

  return (
    <html lang={locale} className="dark">
      <body className={`${inter.variable} ${manrope.variable} font-sans`}>
        <NextIntlClientProvider locale={locale} messages={messages}>
          <Providers>
            <div className="flex flex-col min-h-screen">
              <Header />
              <main className="flex-1">{children}</main>
              <Footer />
            </div>
          </Providers>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
