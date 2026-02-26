'use client'

import { createContext, type ReactNode } from 'react'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useOrg } from '@/hooks/use-org'
import type { TSandboxRun } from '@/types'

type SandboxRunContextValue = {
  sandboxRun: TSandboxRun | null
  isLoading: boolean
  error: any
}

export const SandboxRunContext = createContext<SandboxRunContextValue | undefined>(
  undefined
)

export function SandboxRunProvider({
  children,
  initSandboxRun,
  pollInterval = 10000,
  shouldPoll = true,
}: {
  children: ReactNode
  initSandboxRun: TSandboxRun
} & IPollingProps) {
  const { org } = useOrg()
  const {
    data: sandboxRun,
    error,
    isLoading,
  } = usePolling<TSandboxRun>({
    dependencies: [initSandboxRun],
    initData: initSandboxRun,
    path: `/api/orgs/${org.id}/installs/${initSandboxRun?.install_id}/sandbox/runs/${initSandboxRun.id}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <SandboxRunContext.Provider
      value={{
        sandboxRun,
        isLoading,
        error,
      }}
    >
      {children}
    </SandboxRunContext.Provider>
  )
}