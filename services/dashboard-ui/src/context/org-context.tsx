'use client'

import React, {
  FunctionComponent,
  ReactNode,
  createContext,
  useContext,
  useEffect,
  useState,
} from 'react'
import type { TOrg } from '@/types'
import { POLL_DURATION } from '@/utils'
import { useUser } from '@auth0/nextjs-auth0/client'
import { AnalyticsBrowser } from '@segment/analytics-next'

type TOrgContext = {
  isFetching: boolean
  org: TOrg
}

export const OrgContext = createContext<TOrgContext | null>(null)

export const OrgProvider: FunctionComponent<{
  children?: ReactNode
  initOrg: TOrg
  shouldPoll?: boolean
}> = ({ children, initOrg, shouldPoll = false }) => {
  const [isFetching, setIsFetching] = useState<boolean>(false)
  const [org, setOrg] = useState<TOrg>(initOrg)
  const { user } = useUser()

  let analytics
  useEffect(() => {
    analytics = AnalyticsBrowser.load({
      writeKey: window.process.env.SEGMENT_WRITE_KEY!,
    })
  }, [])

  useEffect(() => {
    const fetchOrg = () => {
      setIsFetching(true)
      fetch(`/api/${org?.id}`)
        .then((res) =>
          res.json().then((org) => {
            setOrg(org)

            // Identify user org if we haven't already.
            if (analytics) {
              analytics.group(org.id, {
                userId: user.sub,
                name: org.name,
              })
            }

            setIsFetching(false)
          })
        )
        .catch(console.error)
    }

    const pollOrg = setInterval(fetchOrg, POLL_DURATION)
    return () => clearInterval(pollOrg)
  }, [org, shouldPoll])

  return (
    <OrgContext.Provider
      value={{
        org,
        isFetching,
      }}
    >
      {children}
    </OrgContext.Provider>
  )
}

export const useOrgContext = () => {
  const context = useContext(OrgContext)

  if (!context) {
    throw new Error(
      'useOrgContext() may only be used in the context of a <OrgProvider> component.'
    )
  }

  return context
}
