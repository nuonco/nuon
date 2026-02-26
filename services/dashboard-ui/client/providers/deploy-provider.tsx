'use client'

import { createContext, type ReactNode } from 'react'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useOrg } from '@/hooks/use-org'
import type { TDeploy } from '@/types'

type DeployContextValue = {
  deploy: TDeploy | null
  isLoading: boolean
  error: any
}

export const DeployContext = createContext<DeployContextValue | undefined>(
  undefined
)

export function DeployProvider({
  children,
  initDeploy,
  pollInterval = 10000,
  shouldPoll = true,
}: {
  children: ReactNode
  initDeploy: TDeploy
} & IPollingProps) {
  const { org } = useOrg()
  const {
    data: deploy,
    error,
    isLoading,
  } = usePolling<TDeploy>({
    dependencies: [initDeploy],
    initData: initDeploy,
    path: `/api/orgs/${org.id}/installs/${initDeploy?.install_id}/deploys/${initDeploy.id}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <DeployContext.Provider
      value={{
        deploy,
        isLoading,
        error,
      }}
    >
      {children}
    </DeployContext.Provider>
  )
}