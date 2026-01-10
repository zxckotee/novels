import { Metadata } from 'next';
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
  return <CollectionDetailClient locale={params.locale} slug={params.slug} />;
}
