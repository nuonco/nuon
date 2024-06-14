'use client'

import React, {
  FunctionComponent,
  ReactNode,
  createContext,
  useContext,
  useEffect,
  useState,
} from 'react'
import type { TInstall } from '@/types'
import { POLL_DURATION } from '@/utils'

type TInstallContext = {
  isFetching: boolean
  install: TInstall
}

export const InstallContext = createContext<TInstallContext | null>(null)

export const InstallProvider: FunctionComponent<{
  children?: ReactNode
  initInstall: TInstall
  shouldPoll?: boolean
}> = ({ children, initInstall, shouldPoll = false }) => {
  const [isFetching, setIsFetching] = useState<boolean>(false)
  const [install, setInstall] = useState<TInstall>(initInstall)

  useEffect(() => {
    if (!shouldPoll) {
      setInstall(initInstall)
    }
  }, [initInstall, shouldPoll])

  useEffect(() => {
    const fetchInstall = () => {
      setIsFetching(true)
      fetch(`/api/${install?.org_id}/installs/${install?.id}`)
        .then((res) =>
          res.json().then((ins) => {
            setInstall(ins)
            setIsFetching(false)
          })
        )
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollInstall = setInterval(fetchInstall, POLL_DURATION)
      return () => clearInterval(pollInstall)
    }
  }, [install, shouldPoll])

  return (
    <InstallContext.Provider
      value={{
        install,
        isFetching,
      }}
    >
      {children}
    </InstallContext.Provider>
  )
}

export const useInstallContext = () => {
  const context = useContext(InstallContext)

  if (!context) {
    throw new Error(
      'useInstallContext() may only be used in the context of a <InstallProvider> component.'
    )
  }

  return context
}
