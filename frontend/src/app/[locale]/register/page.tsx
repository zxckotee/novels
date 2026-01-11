import { Metadata } from 'next';
import { getTranslations, unstable_setRequestLocale } from 'next-intl/server';
import RegisterPageClient from './RegisterPageClient';

export async function generateMetadata({ params: { locale } }: { params: { locale: string } }): Promise<Metadata> {
  const t = await getTranslations({ locale, namespace: 'auth' });
  
  return {
    title: t('register.title'),
    description: t('register.title'),
  };
}

export default async function RegisterPage({ params: { locale } }: { params: { locale: string } }) {
  unstable_setRequestLocale(locale);
  return <RegisterPageClient />;
}
