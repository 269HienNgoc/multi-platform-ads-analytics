import type { NextConfig } from "next";

const backendURL = (process.env.BACKEND_API_URL ?? "http://127.0.0.1:8080").replace(/\/$/, "");

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/backend/:path*",
        destination: `${backendURL}/:path*`,
      },
    ];
  },
};

export default nextConfig;
