/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  env: {
    // API URL for the Go backend
    NEXT_PUBLIC_API_URL:
      process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  },
  // Ensure images from your domain are allowed
  images: {
    domains: ['localhost'],
  },
  // Enable CORS headers if needed
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          {
            key: 'Access-Control-Allow-Credentials',
            value: 'true',
          },
        ],
      },
    ];
  },
};

module.exports = nextConfig;
