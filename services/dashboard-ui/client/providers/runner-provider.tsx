import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getRunner } from '@/lib'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TRunner } from '@/types'

type RunnerContextValue = {
  runner: TRunner
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

  if (error) return <ProviderError error={error} />

  if (isLoading || !runner) return <ProviderLoading />

  return (
    <RunnerContext.Provider value={{ runner }}>
      {children}
    </RunnerContext.Provider>
  )
}
