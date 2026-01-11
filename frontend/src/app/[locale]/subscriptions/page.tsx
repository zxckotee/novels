import { Metadata } from 'next';
import { getTranslations, unstable_setRequestLocale } from 'next-intl/server';
import SubscriptionsPageClient from './SubscriptionsPageClient';

export async function generateMetadata({ params: { locale } }: { params: { locale: string } }): Promise<Metadata> {
  const t = await getTranslations({ locale, namespace: 'subscriptions' });
  
  return {
    title: t('pageTitle'),
    description: t('pageDescription'),
  };
}

export default async function SubscriptionsPage({ params: { locale } }: { params: { locale: string } }) {
  unstable_setRequestLocale(locale);
  return <SubscriptionsPageClient />;
}
