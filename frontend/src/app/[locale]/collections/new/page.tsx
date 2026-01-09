import { Metadata } from 'next';
import CollectionFormClient from '../CollectionFormClient';

export const metadata: Metadata = {
  title: 'Создать коллекцию',
  description: 'Создание новой коллекции новелл',
};

interface PageProps {
  params: { locale: string };
}

export default function NewCollectionPage({ params }: PageProps) {
  return <CollectionFormClient locale={params.locale} />;
}
