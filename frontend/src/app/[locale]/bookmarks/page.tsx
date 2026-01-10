import { Metadata } from 'next';
import { getTranslations } from 'next-intl/server';
import BookmarksPageClient from './BookmarksPageClient';

interface Props {
  params: Promise<{ locale: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: 'bookmarks' });
  
  return {
    title: t('pageTitle'),
    description: t('pageDescription'),
  };
}

export default async function BookmarksPage({ params }: Props) {
  const { locale } = await params;
  
  return <BookmarksPageClient locale={locale} />;
}
