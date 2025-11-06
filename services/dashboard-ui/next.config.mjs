/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    optimizePackageImports: ['@/components', '@/stratus', '@/utils'],
  },
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'lh3.googleusercontent.com',
      },
      {
        protocol: 'https',
        hostname: 'avatars.githubusercontent.com',
      },
      {
        protocol: 'https',
        hostname: 'app.nuon.co',
      },
    ],
  },
  async rewrites() {
    return [
      {
        source: '/admin/temporal',
        destination: '/api/admin/temporal/ui',
      },
      {
        source: '/admin/temporal/:path*',
        destination: '/api/admin/temporal/ui/:path*',
      },
      {
        source: '/_app/:path*',
        destination: '/api/admin/temporal/ui/_app/:path*',
      },
      {
        source: '/admin/temporal-codec/decode',
        destination: '/api/admin/temporal/decode',
      },
      {
        source: '/public/swagger/:path*',
        destination: '/api/ctl-api/:path*',
      },
      {
        source: '/public/:path*',
        destination: '/api/ctl-api/docs/:path*',
      },
      {
        source: '/admin/swagger/:path*',
        destination: '/api/admin/ctl-api/:path*',
      },
      {
        source: '/admin/:path*',
        destination: '/api/admin/ctl-api/docs/:path*',
      },
    ]
  },
  async redirects() {
    return [
      {
        source: '/:orgId/installs/:installId/history',
        destination: '/:orgId/installs/:installId/workflows',
        permanent: true, // This sends a 308 status code
      },
    ]
  },
  onDemandEntries: {
    // period (in ms) where the server will keep pages in the buffer
    maxInactiveAge: 15 * 60 * 1000, // 15 minutes
    // number of pages that should be kept simultaneously without being disposed
    pagesBufferLength: 4,
  },
}

export default nextConfig
