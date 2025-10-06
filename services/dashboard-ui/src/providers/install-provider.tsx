'use client'

import { createContext, type ReactNode } from 'react'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useOrg } from '@/hooks/use-org'
import type { TInstall } from '@/types'

type InstallContextValue = {
  install: TInstall | null
  isLoading: boolean
  error: any
  refresh: () => void
}

export const InstallContext = createContext<InstallContextValue | undefined>(
  undefined
)

export function InstallProvider({
  children,
  initInstall,
  pollInterval = 20000,
  shouldPoll = false,
}: {
  children: ReactNode
  initInstall: TInstall
} & IPollingProps) {
  const { org } = useOrg()
  const {
    data: install,
    error,
    isLoading,
  } = usePolling<TInstall>({
    dependencies: [initInstall],
    initData: initInstall,
    path: `/api/orgs/${org.id}/installs/${initInstall.id}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <InstallContext.Provider
      value={{
        install,
        isLoading,
        error,
        refresh: () => {
          /* implement if needed */
        },
      }}
    >
      {children}
    </InstallContext.Provider>
  )
}
