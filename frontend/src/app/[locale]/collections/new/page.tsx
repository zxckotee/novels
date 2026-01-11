import { Metadata } from 'next';
import { unstable_setRequestLocale } from 'next-intl/server';
import CollectionFormClient from '../CollectionFormClient';

export const metadata: Metadata = {
  title: 'Создать коллекцию',
  description: 'Создание новой коллекции новелл',
};

interface PageProps {
  params: { locale: string };
}

export default function NewCollectionPage({ params }: PageProps) {
  unstable_setRequestLocale(params.locale);
  return <CollectionFormClient locale={params.locale} />;
}
