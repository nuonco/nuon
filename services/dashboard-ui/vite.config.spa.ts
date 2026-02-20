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
      external: [
        // Server-only Next.js modules — never bundled into the SPA
        "@auth0/nextjs-auth0",
        "next/server",
        "next/headers",
        "next/cache",
      ],
    },
  },
  server: {
    port: 5173,
    strictPort: true,
  },
});
