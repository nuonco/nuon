'use client'

import { AnalyticsBrowser } from '@segment/analytics-next'
import React, { type FC, useEffect } from 'react'
import { useUser } from '@auth0/nextjs-auth0/client'
import { usePathname, useSearchParams } from 'next/navigation'

export const InitSegmentAnalytics: FC = () => {
  // Identify user if we haven't already.
  const { user, error, isLoading } = useUser()
  let analytics
  useEffect(() => {
    analytics = AnalyticsBrowser.load({
      writeKey: window.process.env.SEGMENT_WRITE_KEY!,
    })
  }, [])

  useEffect(() => {
    if (analytics && user && !isLoading) {
      analytics.identify(user.sub, {
        email: user.email,
        userId: user.sub,
        name: user.name,
      })
    }
  }, [user, error, isLoading])

  // Track page load.
  const pathname = usePathname()
  const searchParams = useSearchParams()
  useEffect(() => {
    if (analytics) analytics.page(pathname)
  }, [pathname, searchParams])

  return <></>
}
