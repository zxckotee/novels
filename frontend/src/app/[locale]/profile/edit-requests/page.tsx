import { getTranslations, unstable_setRequestLocale } from 'next-intl/server';
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
  unstable_setRequestLocale(params.locale);
  return <EditRequestsClient locale={params.locale} />;
}
