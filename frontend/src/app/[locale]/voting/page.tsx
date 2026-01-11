import { Metadata } from 'next';
import { getTranslations, unstable_setRequestLocale } from 'next-intl/server';
import VotingPageClient from './VotingPageClient';

export async function generateMetadata({ params: { locale } }: { params: { locale: string } }): Promise<Metadata> {
  const t = await getTranslations({ locale, namespace: 'voting' });
  
  return {
    title: t('pageTitle'),
    description: t('pageDescription'),
  };
}

export default async function VotingPage({ params: { locale } }: { params: { locale: string } }) {
  unstable_setRequestLocale(locale);
  return <VotingPageClient />;
}
