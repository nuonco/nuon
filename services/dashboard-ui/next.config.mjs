/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    optimizePackageImports: ['@/components', '@/stratus', "@/utils"],
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
    ],
  },
  async rewrites() {
    return [
      {
        source: '/admin/temporal/:path*',
        destination: `${
          process.env.NUON_TEMPORAL_UI_URL ||
          'http://temporal-web.temporal.svc.cluster.local:8080'
        }/admin/temporal/:path*`,
      },
      {
        source: '/admin/swagger/docs/:path*',
        destination: `${
          process.env.NUON_CTL_API_ADMIN_URL ||
          'http://ctl-api-admin.ctl-api.svc.cluster.local:8082'
        }/docs/:path*`,
      },
      {
        source: '/admin/temporal-codec/decode',
        destination: `${
          process.env.NUON_CTL_API_ADMIN_URL ||
          'http://ctl.nuon.us-west-2.stage.nuon.cloud'
        }/v1/general/temporal-codec/decode`,
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
