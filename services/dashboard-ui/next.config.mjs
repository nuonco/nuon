/** @type {import('next').NextConfig} */
const nextConfig = {
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
        destination: `${process.env.NUON_TEMPORAL_UI_URL || 'http://localhost:8234'}/admin/temporal/:path*`,
      },
      {
        source: '/admin/swagger/docs/:path*',
        destination: `${process.env.NUON_CTL_API_ADMIN_URL || 'http://localhost:8082'}/docs/:path*`,
      },
    ]
  },
}

export default nextConfig
