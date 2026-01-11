import { Metadata } from 'next';
import { getTranslations, unstable_setRequestLocale } from 'next-intl/server';
import ProfilePageClient from './ProfilePageClient';

interface Props {
  params: Promise<{ locale: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: 'profile' });
  
  return {
    title: t('title'),
    description: t('title'),
  };
}

export default async function ProfilePage({ params }: Props) {
  const { locale } = await params;
  unstable_setRequestLocale(locale);
  
  return <ProfilePageClient />;
}
