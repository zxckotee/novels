import { Metadata } from 'next';
import { getTranslations, unstable_setRequestLocale } from 'next-intl/server';
import ProposalsPageClient from './ProposalsPageClient';

interface Props {
  params: Promise<{ locale: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: 'proposals' });
  
  return {
    title: t('title'),
    description: t('description'),
  };
}

export default async function ProposalsPage({ params }: Props) {
  const { locale } = await params;
  unstable_setRequestLocale(locale);
  
  return <ProposalsPageClient />;
}
