import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Dynamic routes require server-side rendering
  // Deploy with Node.js server or use nginx proxy
  trailingSlash: true,
  images: {
    unoptimized: true,
  },
};

export default nextConfig;
