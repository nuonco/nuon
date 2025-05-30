'use client'

import React, {
  type FC,
  createContext,
  useContext,
  useEffect,
  useState,
} from 'react'
import type { TInstall } from '@/types'
import { POLL_DURATION } from '@/utils'

type TFetchError = Record<string | 'message', string>

interface IInstallContext {
  install?: TInstall
  error?: TFetchError
}

const InstallContext = createContext<IInstallContext>({})

interface IInstallProvider {
  children: React.ReactNode
  initInstall: TInstall
  shouldPoll?: boolean
  pollDuration?: number
}

export const InstallProvider: FC<IInstallProvider> = ({
  children,
  initInstall,
  shouldPoll = false,
  pollDuration = POLL_DURATION,
}) => {
  const [error, setError] = useState<TFetchError>()
  const [install, updateInstall] = useState<TInstall>(initInstall)

  useEffect(() => {
    const refreshInstall = () => {
      fetch(`/api/${initInstall?.id}`)
        .then((res) => res.json().then((o) => updateInstall(o)))
        .catch((err) => setError(err))
    }

    if (shouldPoll) {
      const pollInstall = setInterval(refreshInstall, pollDuration)

      return () => {
        clearInterval(pollInstall)
      }
    }
  }, [install])

  return (
    <InstallContext.Provider
      value={{
        error,
        install,
      }}
    >
      {children}
    </InstallContext.Provider>
  )
}

export const useInstall = (): IInstallContext => {
  return useContext(InstallContext)
}
