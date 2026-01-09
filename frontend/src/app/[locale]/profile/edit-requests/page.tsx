import { getTranslations } from 'next-intl/server';
import EditRequestsClient from './EditRequestsClient';

interface PageProps {
  params: { locale: string };
}

export async function generateMetadata({ params }: PageProps) {
  const t = await getTranslations({ locale: params.locale, namespace: 'community.wikiEdit' });
  return {
    title: t('editRequests'),
  };
}

export default async function EditRequestsPage({ params }: PageProps) {
  return <EditRequestsClient locale={params.locale} />;
}
