'use client'

import { createContext, type ReactNode } from 'react'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TInstallActionRun } from '@/types'

type InstallActionRunContextValue = {
  installActionRun: TInstallActionRun | null
  isLoading: boolean
  error: any
  refresh: () => void
}

export const InstallActionRunContext = createContext<
  InstallActionRunContextValue | undefined
>(undefined)

export function InstallActionRunProvider({
  children,
  initInstallActionRun,
  pollInterval = 3000,
  shouldPoll = false,
}: {
  children: ReactNode
  initInstallActionRun: TInstallActionRun
} & IPollingProps) {
  const { org } = useOrg()
  const { install } = useInstall()
  const {
    data: installActionRun,
    error,
    isLoading,
  } = usePolling<TInstallActionRun>({
    initData: initInstallActionRun,
    path: `/api/orgs/${org.id}/installs/${install.id}/actions/runs/${initInstallActionRun.id}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <InstallActionRunContext.Provider
      value={{
        installActionRun,
        isLoading,
        error,
        refresh: () => {
          /* implement if needed */
        },
      }}
    >
      {children}
    </InstallActionRunContext.Provider>
  )
}
