import type { Metadata } from "next";
import { Inter } from "next/font/google";
import React from "react";
import { Link, PoweredByNuon } from "@/components";
import { Markdown } from "@/components/Markdown";
import "./globals.css";
import theme from "@/theme";

const inter = Inter({ subsets: ["latin"] });

export async function generateMetadata(): Promise<Metadata> {
  return {
    title: "Nuon Hosted Installers",
    description: "Nuon Hosted Installers",
  };
}

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      className={`${theme.forceDarkMode ? "dark" : ""} bg-white dark:bg-black text-black dark:text-white`}
      lang="en"
    >
      <body className={`${inter.className} w-full h-dvh`}>
        <div className="flex flex-col w-full max-w-5xl mx-auto p-6 py-12 gap-6 md:gap-12">
          {children}
        </div>
      </body>
    </html>
  );
}
