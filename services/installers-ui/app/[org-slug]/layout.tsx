import { headers } from 'next/headers'
import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import React from 'react'
import { getInstaller } from '@/app/actions'
import { Link, PoweredByNuon } from '@/components'
import { Markdown } from '@/components/Markdown'
import '../globals.css'
import theme from '@/theme'

const inter = Inter({ subsets: ['latin'] })

export async function generateMetadata(): Promise<Metadata> {
  const headerList = headers()
  const orgId = headerList.get('X-Nuon-Org-Id')
  const installer = await getInstaller(orgId)
  const { metadata } = installer

  return {
    title: metadata.name,
    description: metadata.description,
    icons: {
      icon: metadata.favicon_url,
      shortcut: metadata.favicon_url,
    },
    openGraph: {
      title: metadata.name,
      description: metadata.description,
      type: 'website',
      images: [
        {
          url: metadata.logo_url,
        },
      ],
    },
    twitter: {
      title: metadata.name,
      description: metadata.description,
      images: [
        {
          url: metadata.logo_url,
        },
      ],
    },
  }
}

const missingData = {
  orgName: 'Nuon',
}

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  const headerList = headers()
  const orgId = headerList.get('X-Nuon-Org-Id')
  const { metadata } = await getInstaller(orgId)

  return (
    <>
      {children}

      <footer className="flex items-center justify-between">
        <div className="flex gap-2 items-center">
          {metadata.copyright_markdown ? (
            <Markdown content={metadata.copyright_markdown} />
          ) : (
            <>
              <span className="text-xs">&copy; {new Date().getFullYear()}</span>
              <Link
                href={metadata.homepage_url}
                className="text-xs"
                target="_blank"
                rel="noreferrer"
              >
                {missingData.orgName}
              </Link>
            </>
          )}
        </div>
        <div className="flex gap-6 items-center">
          {metadata.footer_markdown ? (
            <Markdown content={metadata.footer_markdown} />
          ) : (
            <PoweredByNuon />
          )}
        </div>
      </footer>
    </>
  )
}
