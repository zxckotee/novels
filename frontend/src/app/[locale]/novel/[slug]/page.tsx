import { Metadata } from 'next';
import { notFound } from 'next/navigation';
import { getTranslations, unstable_setRequestLocale } from 'next-intl/server';
import NovelPageClient from './NovelPageClient';

// Generate metadata for SEO
export async function generateMetadata({
  params,
}: {
  params: { locale: string; slug: string };
}): Promise<Metadata> {
  const t = await getTranslations({ locale: params.locale, namespace: 'novel' });
  
  // TODO: Fetch novel data for meta tags
  // const novel = await fetchNovel(params.slug, params.locale);
  
  return {
    title: `Novel - ${params.slug}`,
    description: 'Reading platform',
    // TODO: Add OG image from novel cover
  };
}

// Server component wrapper
export default function NovelPage({
  params,
}: {
  params: { locale: string; slug: string };
}) {
  unstable_setRequestLocale(params.locale);
  return <NovelPageClient slug={params.slug} locale={params.locale} />;
}
