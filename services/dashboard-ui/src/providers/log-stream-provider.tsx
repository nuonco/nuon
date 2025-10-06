'use client'

import { createContext, useEffect, type ReactNode } from 'react'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useOrg } from '@/hooks/use-org'
import type { TLogStream } from '@/types'

type LogStreamContextValue = {
  logStream: TLogStream | null
  isLoading: boolean
  error: any
  refresh: () => void
}

export const LogStreamContext = createContext<
  LogStreamContextValue | undefined
>(undefined)

export function LogStreamProvider({
  children,
  initLogStream,
  pollInterval = 20000,
  shouldPoll = false,
}: {
  children: ReactNode
  initLogStream: TLogStream
} & IPollingProps) {
  const { org } = useOrg()
  const {
    data: logStream,
    error,
    isLoading,
    stopPolling,
  } = usePolling<TLogStream>({
    initData: initLogStream,
    path: `/api/orgs/${org.id}/log-streams/${initLogStream.id}`,
    pollInterval,
    shouldPoll,
  })

  useEffect(() => {
    if (!logStream.open) {
      stopPolling()
    }
  }, [logStream])

  return (
    <LogStreamContext.Provider
      value={{
        logStream,
        isLoading,
        error,
        refresh: () => {
          /* implement if needed */
        },
      }}
    >
      {children}
    </LogStreamContext.Provider>
  )
}
