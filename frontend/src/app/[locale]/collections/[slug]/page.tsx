import { Metadata } from 'next';
import { unstable_setRequestLocale } from 'next-intl/server';
import CollectionDetailClient from './CollectionDetailClient';

interface PageProps {
  params: { locale: string; slug: string };
}

export async function generateMetadata({ params }: PageProps): Promise<Metadata> {
  return {
    title: 'Коллекция',
    description: 'Коллекция новелл',
  };
}

export default function CollectionDetailPage({ params }: PageProps) {
  unstable_setRequestLocale(params.locale);
  return <CollectionDetailClient locale={params.locale} slug={params.slug} />;
}
