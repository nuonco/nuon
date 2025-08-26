'use client'

import React, {
  type FC,
  createContext,
  useContext,
  useEffect,
  useState,
} from 'react'
import type { TOrg } from '@/types'
import { POLL_DURATION } from '@/utils'
import { setOrgSessionCookie } from '../org-actions'

type TFetchError = Record<string | 'message', string>

interface IOrgContext {
  org?: TOrg
  error?: TFetchError
}

const OrgContext = createContext<IOrgContext>({})

interface IOrgProvider {
  children: React.ReactNode
  initOrg: TOrg
  shouldPoll?: boolean
}

export const OrgProvider: FC<IOrgProvider> = ({
  children,
  initOrg,
  shouldPoll = false,
}) => {
  const [error, setError] = useState<TFetchError>()
  const [org, updateOrg] = useState<TOrg>(initOrg)

  useEffect(() => {
    setOrgSessionCookie(org.id)
  }, [org])

  useEffect(() => {
    const refreshOrg = () => {
      fetch(`/api/${initOrg?.id}`)
        .then((res) =>
          res.json().then((o) => {
            updateOrg(o)
          })
        )
        .catch((err) => setError(err))
    }

    if (shouldPoll) {
      const pollOrg = setInterval(refreshOrg, POLL_DURATION)

      return () => {
        clearInterval(pollOrg)
      }
    }
  }, [org])

  return (
    <OrgContext.Provider
      value={{
        error,
        org,
      }}
    >
      {children}
    </OrgContext.Provider>
  )
}

export const useOrg = (): IOrgContext => {
  return useContext(OrgContext)
}
