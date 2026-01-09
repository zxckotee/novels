import type { Metadata } from 'next';

const SITE_NAME = 'Novels';
const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL || 'https://novels.example.com';

export interface SEOProps {
  title?: string;
  description?: string;
  image?: string;
  url?: string;
  type?: 'website' | 'article' | 'book';
  locale?: string;
  publishedTime?: string;
  modifiedTime?: string;
  author?: string;
  tags?: string[];
  noindex?: boolean;
}

export function generateMetadata({
  title,
  description,
  image,
  url,
  type = 'website',
  locale = 'ru',
  publishedTime,
  modifiedTime,
  author,
  tags,
  noindex = false,
}: SEOProps): Metadata {
  const fullTitle = title ? `${title} | ${SITE_NAME}` : SITE_NAME;
  const fullUrl = url ? `${SITE_URL}${url}` : SITE_URL;
  const defaultDescription = 'Читайте новеллы онлайн на русском языке. Огромная библиотека веб-романов с удобным ридером.';
  const imageUrl = image || `${SITE_URL}/og-image.png`;

  const metadata: Metadata = {
    title: fullTitle,
    description: description || defaultDescription,
    metadataBase: new URL(SITE_URL),
    alternates: {
      canonical: fullUrl,
      languages: {
        'ru': `${SITE_URL}/ru${url || ''}`,
        'en': `${SITE_URL}/en${url || ''}`,
        'zh': `${SITE_URL}/zh${url || ''}`,
        'ja': `${SITE_URL}/ja${url || ''}`,
        'ko': `${SITE_URL}/ko${url || ''}`,
        'fr': `${SITE_URL}/fr${url || ''}`,
        'de': `${SITE_URL}/de${url || ''}`,
      },
    },
    openGraph: {
      title: fullTitle,
      description: description || defaultDescription,
      url: fullUrl,
      siteName: SITE_NAME,
      images: [
        {
          url: imageUrl,
          width: 1200,
          height: 630,
          alt: title || SITE_NAME,
        },
      ],
      locale: locale,
      type: type,
    },
    twitter: {
      card: 'summary_large_image',
      title: fullTitle,
      description: description || defaultDescription,
      images: [imageUrl],
    },
    robots: noindex
      ? { index: false, follow: false }
      : { index: true, follow: true },
  };

  // Add article-specific metadata
  if (type === 'article' && metadata.openGraph) {
    const articleMeta: Record<string, string | string[] | undefined> = {};
    if (publishedTime) articleMeta.publishedTime = publishedTime;
    if (modifiedTime) articleMeta.modifiedTime = modifiedTime;
    if (author) articleMeta.authors = [author];
    if (tags) articleMeta.tags = tags;
    
    (metadata.openGraph as Record<string, unknown>).article = articleMeta;
  }

  return metadata;
}

// JSON-LD structured data generators
export function generateNovelJsonLd(novel: {
  title: string;
  description: string;
  author: string;
  coverUrl: string;
  rating?: number;
  ratingCount?: number;
  genres?: string[];
  url: string;
}) {
  return {
    '@context': 'https://schema.org',
    '@type': 'Book',
    name: novel.title,
    description: novel.description,
    author: {
      '@type': 'Person',
      name: novel.author,
    },
    image: novel.coverUrl,
    genre: novel.genres,
    url: `${SITE_URL}${novel.url}`,
    aggregateRating: novel.rating
      ? {
          '@type': 'AggregateRating',
          ratingValue: novel.rating,
          ratingCount: novel.ratingCount || 0,
          bestRating: 5,
          worstRating: 1,
        }
      : undefined,
  };
}

export function generateChapterJsonLd(chapter: {
  title: string;
  novelTitle: string;
  chapterNumber: number;
  content?: string;
  url: string;
}) {
  return {
    '@context': 'https://schema.org',
    '@type': 'Chapter',
    name: chapter.title,
    isPartOf: {
      '@type': 'Book',
      name: chapter.novelTitle,
    },
    position: chapter.chapterNumber,
    url: `${SITE_URL}${chapter.url}`,
  };
}

export function generateBreadcrumbJsonLd(items: { name: string; url: string }[]) {
  return {
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: items.map((item, index) => ({
      '@type': 'ListItem',
      position: index + 1,
      name: item.name,
      item: `${SITE_URL}${item.url}`,
    })),
  };
}

export function generateOrganizationJsonLd() {
  return {
    '@context': 'https://schema.org',
    '@type': 'Organization',
    name: SITE_NAME,
    url: SITE_URL,
    logo: `${SITE_URL}/logo.png`,
    sameAs: [
      // Add social media links here
    ],
  };
}

export function generateWebsiteJsonLd() {
  return {
    '@context': 'https://schema.org',
    '@type': 'WebSite',
    name: SITE_NAME,
    url: SITE_URL,
    potentialAction: {
      '@type': 'SearchAction',
      target: {
        '@type': 'EntryPoint',
        urlTemplate: `${SITE_URL}/ru/catalog?search={search_term_string}`,
      },
      'query-input': 'required name=search_term_string',
    },
  };
}
