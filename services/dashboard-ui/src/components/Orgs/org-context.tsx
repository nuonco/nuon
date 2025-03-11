'use client'

import React, {
  type FC,
  createContext,
  useContext,
  useEffect,
  useState,
} from 'react'
import { setOrgSessionCookie } from '@/app/actions'
import type { TOrg } from '@/types'
import { POLL_DURATION } from '@/utils'

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
    async function setSession() {
      await setOrgSessionCookie(initOrg.id)
    }

    setSession()
  }, [])

  useEffect(() => {
    const refreshOrg = () => {
      fetch(`/api/${initOrg?.id}`)
        .then((res) => res.json().then((o) => updateOrg(o)))
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
