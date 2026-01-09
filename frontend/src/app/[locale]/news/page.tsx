import { Metadata } from 'next';
import NewsPageClient from './NewsPageClient';

export const metadata: Metadata = {
  title: 'Новости',
  description: 'Новости платформы Novels',
};

interface PageProps {
  params: { locale: string };
  searchParams: { [key: string]: string | string[] | undefined };
}

export default function NewsPage({ params, searchParams }: PageProps) {
  return <NewsPageClient locale={params.locale} searchParams={searchParams} />;
}
