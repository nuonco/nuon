import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getRunner, getRunnerLatestHeartbeat } from '@/lib'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TRunner } from '@/types'

type RunnerContextValue = {
  runner: TRunner
  isManaged: boolean
}

export const RunnerContext = createContext<RunnerContextValue | undefined>(
  undefined,
)

export function RunnerProvider({
  children,
  runnerId,
  pollInterval = 20000,
  shouldPoll = false,
}: {
  children: ReactNode
  runnerId: string
  shouldPoll?: boolean
  pollInterval?: number
}) {
  const { org } = useOrg()

  const { data: runner, isLoading, error } = useQuery({
    queryKey: ['runner', org.id!, runnerId],
    queryFn: () => getRunner({ orgId: org.id!, runnerId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!runnerId,
  })

  const { data: heartbeat } = useQuery({
    queryKey: ['runner-heartbeat-check', org.id, runnerId],
    queryFn: () => getRunnerLatestHeartbeat({ orgId: org.id!, runnerId }),
    enabled: !!org.id && !!runnerId,
  })

  if (error) return <ProviderError error={error} />

  if (isLoading || !runner) return <ProviderLoading />

  return (
    <RunnerContext.Provider value={{ runner, isManaged: Boolean(heartbeat?.mng) }}>
      {children}
    </RunnerContext.Provider>
  )
}
