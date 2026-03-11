import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getRunnerJob } from '@/lib'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TRunnerJob } from '@/types'

type RunnerJobContextValue = {
  job: TRunnerJob
}

export const RunnerJobContext = createContext<RunnerJobContextValue | undefined>(undefined)

const ACTIVE_STATUSES = ['in-progress', 'queued', 'available']

export function RunnerJobProvider({
  children,
  runnerJobId,
}: {
  children: ReactNode
  runnerJobId: string
}) {
  const { org } = useOrg()

  const { data: job, isLoading, error } = useQuery({
    queryKey: ['runner-job', org.id, runnerJobId],
    queryFn: () => getRunnerJob({ runnerJobId, orgId: org.id }),
    refetchInterval: (query) => {
      const data = query.state.data
      return data && ACTIVE_STATUSES.includes(data.status ?? '') ? 5000 : false
    },
    enabled: !!org.id && !!runnerJobId,
  })

  if (error) return <ProviderError error={error} />
  if (isLoading || !job) return <ProviderLoading />

  return (
    <RunnerJobContext.Provider value={{ job }}>
      {children}
    </RunnerJobContext.Provider>
  )
}
