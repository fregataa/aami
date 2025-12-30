import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'export',  // Static export for nginx serving
  trailingSlash: true,
  images: {
    unoptimized: true,  // Required for static export
  },
};

export default nextConfig;
