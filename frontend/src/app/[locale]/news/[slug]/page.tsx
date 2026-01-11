import { Metadata } from 'next';
import { unstable_setRequestLocale } from 'next-intl/server';
import NewsDetailClient from './NewsDetailClient';

interface PageProps {
  params: { locale: string; slug: string };
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  return {
    title: 'Новость',
    description: 'Новость платформы Novels',
  };
}

export default function NewsDetailPage({ params }: PageProps) {
  unstable_setRequestLocale(params.locale);
  return <NewsDetailClient locale={params.locale} slug={params.slug} />;
}
