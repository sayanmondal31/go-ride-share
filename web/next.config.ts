import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  images: {
    // Load images from randomuser.me for Lego driver profile pictures
    // domains: ["https://randomuser.me"],
    remotePatterns: [
      {
        protocol: "https",
        hostname: "randomuser.me",
        port: "",
        pathname: "/api/portraits/**",
      },
    ],
  },
  reactStrictMode: false,
};

export default nextConfig;
