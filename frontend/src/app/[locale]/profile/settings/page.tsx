import { Metadata } from 'next';
import { getTranslations, unstable_setRequestLocale } from 'next-intl/server';
import SettingsPageClient from './SettingsPageClient';

interface Props {
  params: Promise<{ locale: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: 'profile' });
  
  return {
    title: t('tabs.settings'),
    description: t('tabs.settings'),
  };
}

export default async function SettingsPage({ params }: Props) {
  const { locale } = await params;
  unstable_setRequestLocale(locale);
  
  return <SettingsPageClient />;
}
