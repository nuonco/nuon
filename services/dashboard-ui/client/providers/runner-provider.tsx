import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getRunner } from '@/lib'
import type { TRunner } from '@/types'

type RunnerContextValue = {
  runner: TRunner | null
  isLoading: boolean
  error: unknown
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
    queryKey: ['runner', org?.id, runnerId],
    queryFn: () => getRunner({ orgId: org.id, runnerId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!runnerId,
  })

  return (
    <RunnerContext.Provider
      value={{
        runner: runner ?? null,
        isLoading,
        error,
      }}
    >
      {children}
    </RunnerContext.Provider>
  )
}
