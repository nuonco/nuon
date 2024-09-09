import { UserProvider } from '@auth0/nextjs-auth0/client'
import type { Metadata } from 'next'
import { GeistSans } from 'geist/font/sans';
import { GeistMono } from 'geist/font/mono';
import { InitDatadogLogs, InitDatadogRUM, InitPosthogAnalytics } from '@/utils'
import './globals.css'

export const metadata: Metadata = {
  title: 'Nuon',
  description: 'Bring your own cloud with Nuon',
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html
      className="bg-light text-slate-950 dark:bg-dark dark:text-slate-50"
      lang="en"
    >
      {process?.env?.NEXT_PUBLIC_DATADOG_ENV === 'prod' ||
      process?.env?.NEXT_PUBLIC_DATADOG_ENV === 'stage' ||
      process?.env?.NEXT_PUBLIC_DATADOG_ENV === 'local-test' ? (
        <>
          <InitDatadogLogs />
          <InitDatadogRUM />
        </>
      ) : null}
      <UserProvider>
        <body className={`${GeistMono.className} ${GeistSans.className}`}>
          {children}
          {process?.env?.NEXT_PUBLIC_POSTHOG_TOKEN && <InitPosthogAnalytics />}
        </body>
      </UserProvider>
    </html>
  )
}
