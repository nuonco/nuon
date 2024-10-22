/** @type {import('next').NextConfig} */
const nextConfig = {
  swcMinify: true,
  fastRefresh: true,
  concurrentFeatures: true,
  optimizeFonts: false,
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
}

export default nextConfig
