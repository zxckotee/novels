import { Metadata } from 'next';
import { getTranslations, unstable_setRequestLocale } from 'next-intl/server';
import ProposalFormClient from './ProposalFormClient';

export async function generateMetadata({ params: { locale } }: { params: { locale: string } }): Promise<Metadata> {
  const t = await getTranslations({ locale, namespace: 'proposals' });
  
  return {
    title: t('newProposal.pageTitle'),
    description: t('newProposal.pageDescription'),
  };
}

export default async function NewProposalPage({ params: { locale } }: { params: { locale: string } }) {
  unstable_setRequestLocale(locale);
  return <ProposalFormClient />;
}
