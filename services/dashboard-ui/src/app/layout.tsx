import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import localFont from 'next/font/local'
import { Suspense } from 'react'
import { InitDatadogLogs } from '@/lib/datadog-logs'
import { InitDatadogRUM } from '@/lib/datadog-rum'
import {
  InitSegmentAnalytics,
  SegmentAnalyticsIdentify,
} from '@/lib/segment-analytics'
import { AccountProvider } from '@/components/AccountProvider'
import { GlobalUserJourneyProvider } from '@/components/GlobalUserJourneyProvider'
import './globals.css'

const inter = Inter({
  subsets: ['latin'],
  variable: '--font-inter',
  display: 'swap',
})
const hack = localFont({
  src: [
    {
      path: '../../public/fonts/hack-regular.woff2',
      weight: '400',
      style: 'normal',
    },
  ],
  variable: '--font-hack',
})

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
      className="bg-light text-cool-grey-950 dark:bg-dark-grey-100 dark:text-cool-grey-50 overflow-hidden"
      lang="en"
    >
      <>
        {process?.env?.NEXT_PUBLIC_DATADOG_ENV === 'prod' ||
        process?.env?.NEXT_PUBLIC_DATADOG_ENV === 'stage' ||
        process?.env?.NEXT_PUBLIC_DATADOG_ENV === 'local' ? (
          <>
            <InitDatadogLogs env={process?.env?.NEXT_PUBLIC_DATADOG_ENV} />
            <InitDatadogRUM env={process?.env?.NEXT_PUBLIC_DATADOG_ENV} />
          </>
        ) : null}
        <body
          className={`${inter.variable} ${hack.variable} font-sans overflow-hidden disable-ligatures`}
        >
          <EnvScript
            env={process?.env?.NEXT_PUBLIC_DATADOG_ENV}
            githubAppName={process.env.GITHUB_APP_NAME}
          />
          <AccountProvider>
            <GlobalUserJourneyProvider>{children}</GlobalUserJourneyProvider>
          </AccountProvider>
          {process.env.SEGMENT_WRITE_KEY && (
            <Suspense>
              <InitSegmentAnalytics writeKey={process.env.SEGMENT_WRITE_KEY} />
              <SegmentAnalyticsIdentify />
            </Suspense>
          )}
        </body>
      </>
    </html>
  )
}

const EnvScript = ({ env, githubAppName }) => {
  return (
    <div
      dangerouslySetInnerHTML={{
        __html: `<script id="client-env">
          window.env = "${env}";
          window.GITHUB_APP_NAME = "${githubAppName}";
        </script>`,
      }}
    />
  )
}
