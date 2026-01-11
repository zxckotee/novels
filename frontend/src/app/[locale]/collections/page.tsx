import { Metadata } from 'next';
import { unstable_setRequestLocale } from 'next-intl/server';
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
  unstable_setRequestLocale(params.locale);
  return <CollectionsPageClient locale={params.locale} searchParams={searchParams} />;
}
