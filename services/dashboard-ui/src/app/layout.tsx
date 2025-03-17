// @ts-nocheck
import { UserProvider } from '@auth0/nextjs-auth0/client'
import type { Metadata } from 'next'
import { Suspense } from 'react'
import { GeistSans } from 'geist/font/sans'
import { GeistMono } from 'geist/font/mono'
import {
  InitDatadogLogs,
  InitDatadogRUM,
  InitSegmentAnalytics,
  SegmentAnalyticsIdentify,
} from '@/utils'
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
      className="bg-light text-cool-grey-950 dark:bg-dark-grey-100 dark:text-cool-grey-50"
      lang="en"
    >
      <UserProvider>
        <>
          {process?.env?.NEXT_PUBLIC_DATADOG_ENV === 'prod' ||
          process?.env?.NEXT_PUBLIC_DATADOG_ENV === 'stage' ||
          process?.env?.NEXT_PUBLIC_DATADOG_ENV === 'local-test' ? (
            <>
              <InitDatadogLogs env={process?.env?.NEXT_PUBLIC_DATADOG_ENV} />
              <InitDatadogRUM env={process?.env?.NEXT_PUBLIC_DATADOG_ENV} />
            </>
          ) : null}
          <body
            className={`${GeistMono.variable} ${GeistSans.variable} font-sans`}
          >
            {children}
            {process.env.SEGMENT_WRITE_KEY && (
              <Suspense>
                <InitSegmentAnalytics
                  writeKey={process.env.SEGMENT_WRITE_KEY}
                />
                <SegmentAnalyticsIdentify />
              </Suspense>
            )}
          </body>
        </>
      </UserProvider>
    </html>
  )
}
