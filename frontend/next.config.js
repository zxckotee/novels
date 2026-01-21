const createNextIntlPlugin = require('next-intl/plugin');

const withNextIntl = createNextIntlPlugin();

// Normalize API origins (strip trailing /api/v1 if present)
// - rawApiUrl: what the browser should use (public)
// - rawInternalApiUrl: what the Next.js server (inside Docker) should use for rewrites / image proxying
const rawApiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
const rawInternalApiUrl = process.env.INTERNAL_API_URL || rawApiUrl;

const apiOrigin = rawApiUrl.replace(/\/api\/v1\/?$/, '');
const internalApiOrigin = rawInternalApiUrl.replace(/\/api\/v1\/?$/, '');

/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  
  // Standalone output for Docker
  output: 'standalone',
  
  // Webpack configuration to suppress warnings for dynamic imports
  webpack: (config, { isServer }) => {
    // Suppress module not found warnings for dynamic imports in messages directory
    config.module.exprContextCritical = false;
    return config;
  },
  
  // Настройки изображений
  images: {
    remotePatterns: [
      {
        protocol: 'http',
        hostname: 'localhost',
        port: '8080',
        pathname: '/uploads/**',
      },
      // Docker internal backend (dev/prod compose)
      {
        protocol: 'http',
        hostname: 'backend',
        port: '8080',
        pathname: '/uploads/**',
      },
      {
        protocol: 'http',
        hostname: 'api',
        port: '8080',
        pathname: '/uploads/**',
      },
      {
        protocol: 'https',
        hostname: 'api.novels.app',
        pathname: '/uploads/**',
      },
      {
        protocol: 'https',
        hostname: 'via.placeholder.com',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: 'remanga.org',
        pathname: '/**',
      },
    ],
    formats: ['image/avif', 'image/webp'],
  },

  // Переменные окружения для клиента
  env: {
    NEXT_PUBLIC_API_URL: rawApiUrl,
  },

  // Rewrites для API
  async rewrites() {
    return [
      // Preferred: proxy v1 API when frontend uses same-origin calls
      {
        source: '/api/v1/:path*',
        destination: `${internalApiOrigin}/api/v1/:path*`,
      },
      {
        source: '/api/:path*',
        destination: `${internalApiOrigin}/api/:path*`,
      },
      {
        source: '/uploads/:path*',
        destination: `${internalApiOrigin}/uploads/:path*`,
      },
    ];
  },

  // Headers для безопасности
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          {
            key: 'X-DNS-Prefetch-Control',
            value: 'on',
          },
          {
            key: 'X-Frame-Options',
            value: 'SAMEORIGIN',
          },
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
        ],
      },
    ];
  },
};

module.exports = withNextIntl(nextConfig);
