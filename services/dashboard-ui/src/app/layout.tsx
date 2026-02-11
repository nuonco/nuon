import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import localFont from 'next/font/local'
import { Suspense } from 'react'
import { IS_BYOC } from '@/configs/app'
import { API_URL } from '@/configs/api'
import { USE_AUTH_SERVICE, AUTH_SERVICE_URL } from '@/configs/auth'
import { PYLON_APP_ID } from "@/configs/pylon"
import { getUserProfile } from '@/lib/auth-server'
import { InitDatadogLogs } from '@/lib/datadog-logs'
import { InitDatadogRUM } from '@/lib/datadog-rum'
import {
  InitSegmentAnalytics,
  SegmentAnalyticsIdentify,
} from '@/lib/segment-analytics'
import { InitPylonChat } from '@/lib/pylon-chat'
import { AccountProvider } from '@/providers/account-provider'
import { AuthProvider } from '@/providers/auth-provider'
import { UserJourneyProvider } from '@/providers/user-journey-provider'
import './old-styles.css'
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

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  const initialUser = USE_AUTH_SERVICE ? await getUserProfile() : null

  return (
    <html
      className="bg-light text-cool-grey-950 dark:bg-dark-grey-100 dark:text-cool-grey-50 overflow-hidden"
      lang="en"
    >
      <AuthProvider
        useAuthService={USE_AUTH_SERVICE}
        initialUser={initialUser}
        authServiceUrl={AUTH_SERVICE_URL}
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
            className={`${inter.variable} ${hack.variable} font-sans overflow-hidden disable-ligatures antialiased`}
          >
            <div id="ui-version" className="hidden">
              Version: {process.env.VERSION || 'development'}
            </div>
            <EnvScript
              env={process?.env?.NEXT_PUBLIC_DATADOG_ENV}
              githubAppName={process.env.GITHUB_APP_NAME}
              tfBackendUrl={API_URL}
            />

            <AccountProvider shouldPoll>
              <UserJourneyProvider isBYOC={IS_BYOC}>
                {children}
              </UserJourneyProvider>
            </AccountProvider>

           {PYLON_APP_ID ?  <InitPylonChat PYLON_APP_ID={PYLON_APP_ID} /> : null}

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
      </AuthProvider>
    </html>
  )
}

const EnvScript = ({ env, githubAppName, tfBackendUrl }) => {
  return (
    <div
      dangerouslySetInnerHTML={{
        __html: `<script id="client-env">
          window.env = "${env}";
          window.GITHUB_APP_NAME = "${githubAppName}";
          window.TF_BACKEND_URL = "${tfBackendUrl}"
        </script>`,
      }}
    />
  )
}
