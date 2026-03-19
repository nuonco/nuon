import path from "path";
import { defineConfig } from "vite";

export default defineConfig({
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./client"),
    },
  },
  define: {
    global: "globalThis",
    "process.env": JSON.stringify({
      NODE_ENV: "development",
      ...process.env,
    }),
  },
  optimizeDeps: {
    include: [
      // React ecosystem
      "react",
      "react-dom",
      "react/jsx-runtime",
      "react-dom/client",
      
      // UI libraries
      "@phosphor-icons/react",
      "react-icons",
      "react-icons/bs",
      "react-icons/fa",
      "react-icons/fi",
      "react-icons/go",
      "react-icons/hi",
      "react-icons/md",
      "react-icons/ri",
      "react-icons/si",
      "react-icons/tb",
      "react-icons/vsc",
      
      // Table library
      "@tanstack/react-table",
      
      // Utility libraries
      "classnames",
      "luxon",
      "uuid",
      "yaml",
      "showdown",
      
      // Syntax highlighting
      "react-syntax-highlighter",
      "react-syntax-highlighter/dist/esm/styles/prism",
      "react-syntax-highlighter/dist/esm/languages/prism/javascript",
      "react-syntax-highlighter/dist/esm/languages/prism/typescript",
      "react-syntax-highlighter/dist/esm/languages/prism/json",
      "react-syntax-highlighter/dist/esm/languages/prism/yaml",
      "react-syntax-highlighter/dist/esm/languages/prism/bash",
    ],
  },
  server: {
    fs: {
      allow: [".."],
    },
  },
});
