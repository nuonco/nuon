'use client'

import { createContext, type ReactNode } from 'react'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useOrg } from '@/hooks/use-org'
import type { TRunner } from '@/types'

type RunnerContextValue = {
  runner: TRunner | null
  isLoading: boolean
  error: any
  refresh: () => void
}

export const RunnerContext = createContext<RunnerContextValue | undefined>(
  undefined,
)

export function RunnerProvider({
  children,
  initRunner,
  pollInterval = 20000,
  shouldPoll = false,
}: {
  children: ReactNode
  initRunner: TRunner
} & IPollingProps) {
  const { org } = useOrg()

  const {
    data: runner,
    error,
    isLoading,
  } = usePolling<TRunner>({
    dependencies: [initRunner],
    initData: initRunner,
    path: `/api/orgs/${org.id}/runners/${initRunner.id}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <RunnerContext.Provider
      value={{
        runner,
        isLoading,
        error,
        refresh: () => {
          // Placeholder for manual refresh if needed
        },
      }}
    >
      {children}
    </RunnerContext.Provider>
  )
}
