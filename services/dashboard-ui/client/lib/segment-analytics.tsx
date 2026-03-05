import { useEffect } from 'react'
import { useLocation, useSearchParams } from 'react-router'
import { useAuth } from '@/hooks/use-auth'
import { AnalyticsBrowser } from '@segment/analytics-next'
import type { TOrg, IUser } from '@/types'

export const SegmentAnalyticsIdentify = () => {
  const { user, isLoading } = useAuth()

  useEffect(() => {
    if (window['analytics'] && user && !isLoading) {
      window['analytics']?.identify(user.sub, {
        email: user.email,
        userId: user.sub,
        name: user.name,
      })
    }
  }, [user, isLoading])

  const { pathname } = useLocation()
  const [searchParams] = useSearchParams()
  useEffect(() => {
    if (window['analytics']) window['analytics']?.page(pathname)
  }, [pathname, searchParams])

  return <></>
}

export const SegmentAnalyticsSetOrg = ({ org }: { org: TOrg }) => {
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

export const InitSegmentAnalytics = ({ writeKey }: { writeKey: string }) => {
  useEffect(() => {
    window['analytics'] = AnalyticsBrowser.load({
      writeKey,
    })
  }, [])

  return null
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
