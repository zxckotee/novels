import type { Metadata } from 'next';
import { getTranslations, unstable_setRequestLocale } from 'next-intl/server';
import WalletPageClient from './WalletPageClient';

interface Props {
  params: { locale: string };
}

export async function generateMetadata({ params: { locale } }: Props): Promise<Metadata> {
  const t = await getTranslations({ locale, namespace: 'wallet' });
  return {
    title: t('title'),
    description: t('title'),
  };
}

export default function WalletPage({ params: { locale } }: Props) {
  unstable_setRequestLocale(locale);
  return <WalletPageClient />;
}

