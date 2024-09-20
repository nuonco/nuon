'use client'

import React, { type FC, useEffect } from 'react'
import { useUser } from '@auth0/nextjs-auth0/client'
import { usePathname, useSearchParams } from 'next/navigation'

export const InitSegmentAnalytics: FC = () => {
  // Identify user if we haven't already.
  const { user, error, isLoading } = useUser()

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
