import path from "path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Vite config for building the SPA bundle served by the Go BFF server.
// Separate from vite.config.ts which is used by Ladle/Vitest.
export default defineConfig({
  root: __dirname,
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "next/navigation": path.resolve(__dirname, "./src/shims/next-navigation"),
      "next/link": path.resolve(__dirname, "./src/shims/next-link"),
      "next/image": path.resolve(__dirname, "./src/shims/next-image"),
      "next/headers": path.resolve(__dirname, "./src/shims/next-headers"),
      "next/cache": path.resolve(__dirname, "./src/shims/next-cache"),
      "@auth0/nextjs-auth0/client": path.resolve(
        __dirname,
        "./src/shims/auth0-client"
      ),
      "@auth0/nextjs-auth0/server": path.resolve(
        __dirname,
        "./src/shims/auth0-server"
      ),
    },
  },
  define: {
    "process.env": JSON.stringify({
      NODE_ENV: process.env.NODE_ENV || "production",
      NEXT_PUBLIC_DASHBOARD_MODE: "go",
      NEXT_PUBLIC_DATADOG_ENV: process.env.NEXT_PUBLIC_DATADOG_ENV || "",
      VERSION: process.env.VERSION || "development",
      GITHUB_APP_NAME: process.env.GITHUB_APP_NAME || "",
      SEGMENT_WRITE_KEY: process.env.SEGMENT_WRITE_KEY || "",
    }),
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
    sourcemap: true,
    rollupOptions: {
      input: path.resolve(__dirname, "index.html"),
    },
  },
  server: {
    port: 5173,
    strictPort: true,
    // Proxy API requests to the Go BFF server
    proxy: {
      "/api": {
        target: "http://localhost:4000",
        changeOrigin: true,
      },
      "/livez": {
        target: "http://localhost:4000",
        changeOrigin: true,
      },
      "/readyz": {
        target: "http://localhost:4000",
        changeOrigin: true,
      },
    },
    // HMR connects directly to Vite, not through the Go BFF proxy
    hmr: {
      port: 5173,
    },
  },
});
