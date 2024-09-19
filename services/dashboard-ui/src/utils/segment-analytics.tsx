'use client'

import { AnalyticsBrowser } from '@segment/analytics-next'
import React, { type FC, useEffect } from 'react'
import { useUser } from '@auth0/nextjs-auth0/client'
import { usePathname, useSearchParams } from 'next/navigation'

// Init segment.
export const analytics = AnalyticsBrowser.load({
  writeKey: process.env.NEXT_PUBLIC_SEGMENT_WRITE_KEY!,
})

export const InitSegmentAnalytics: FC = () => {
  // Identify user if we haven't already.
  const { user, error, isLoading } = useUser()
  useEffect(() => {
    if (user && !isLoading) {
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
    analytics.page(pathname)
  }, [pathname, searchParams])

  return <></>
}
