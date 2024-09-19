import { UserProvider } from '@auth0/nextjs-auth0/client'
import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import { InitDatadogLogs, InitDatadogRUM, InitSegmentAnalytics } from '@/utils'
import '../globals.css'

const inter = Inter({ subsets: ['latin'] })

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
      className="bg-slate-50 text-slate-950 dark:bg-slate-950 dark:text-slate-50"
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
        <body className={inter.className}>
          {children}
          {process.env.NEXT_PUBLIC_SEGMENT_WRITE_KEY && (
            <InitSegmentAnalytics />
          )}
        </body>
      </UserProvider>
    </html>
  )
}
