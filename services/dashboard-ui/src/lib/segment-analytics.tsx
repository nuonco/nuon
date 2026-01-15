// @ts-nocheck
'use client'

import React, { type FC, useEffect } from 'react'
import { usePathname, useSearchParams } from 'next/navigation'
import Script from 'next/script'
import { useAuth } from '@/hooks/use-auth'
import { AnalyticsBrowser } from '@segment/analytics-next'
import type { TOrg } from '@/types'
import type { IUser } from '@/types/dashboard.types'

export const SegmentAnalyticsIdentify: FC = () => {
  // Identify user if we haven't already.
  const { user, error, isLoading } = useAuth()

  useEffect(() => {
    if (window['analytics'] && user && !isLoading) {
      window['analytics']?.identify(user.sub, {
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
    if (window['analytics']) window['analytics']?.page(pathname)
  }, [pathname, searchParams])

  return <></>
}

export const SegmentAnalyticsSetOrg: FC<{ org: TOrg }> = ({ org }) => {
  const { user, isLoading } = useAuth()

  useEffect(() => {
    if (window['analytics'] && user && !isLoading) {
      window['analytics']?.group(org.id, {
        userId: user?.sub,
        name: org.name,
      })
    }
  }, [])

  return <></>
}

export const InitSegmentAnalytics: FC<{ writeKey: string }> = ({
  writeKey,
}) => {
  useEffect(() => {
    window.analytics = AnalyticsBrowser.load({
      writeKey,
    })
  }, [])

  // eslint-disable-next-line
  return <Script id="load-env"></Script>
}

interface ITrackEvent {
  event: string
  props?: Record<string, unknown>
  status: 'ok' | 'error'
  user: IUser
}

export function trackEvent({ event, user, status, props = {} }: ITrackEvent) {
  if (window['analytics'] && user) {
    window['analytics']?.track(event, {
      userId: user?.sub,
      userEmail: user?.email,
      userName: user?.name,
      status,
      ...props,
    })
  }
}
