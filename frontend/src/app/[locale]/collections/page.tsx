import { Metadata } from 'next';
import CollectionsPageClient from './CollectionsPageClient';

export const metadata: Metadata = {
  title: 'Коллекции',
  description: 'Пользовательские коллекции новелл',
};

interface PageProps {
  params: { locale: string };
  searchParams: { [key: string]: string | string[] | undefined };
}

export default function CollectionsPage({ params, searchParams }: PageProps) {
  return <CollectionsPageClient locale={params.locale} searchParams={searchParams} />;
}
