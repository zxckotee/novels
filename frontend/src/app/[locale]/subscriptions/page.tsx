import { Metadata } from 'next';
import { getTranslations } from 'next-intl/server';
import SubscriptionsPageClient from './SubscriptionsPageClient';

export async function generateMetadata({ params: { locale } }: { params: { locale: string } }): Promise<Metadata> {
  const t = await getTranslations({ locale, namespace: 'subscriptions' });
  
  return {
    title: t('pageTitle'),
    description: t('pageDescription'),
  };
}

export default async function SubscriptionsPage() {
  return <SubscriptionsPageClient />;
}
