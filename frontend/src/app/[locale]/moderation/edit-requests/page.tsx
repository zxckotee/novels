import { getTranslations } from 'next-intl/server';
import ModerationEditRequestsClient from './ModerationEditRequestsClient';

interface PageProps {
  params: { locale: string };
}

export async function generateMetadata({ params }: PageProps) {
  const t = await getTranslations({ locale: params.locale, namespace: 'moderation' });
  return {
    title: t('editRequests'),
  };
}

export default async function ModerationEditRequestsPage({ params }: PageProps) {
  return <ModerationEditRequestsClient locale={params.locale} />;
}
