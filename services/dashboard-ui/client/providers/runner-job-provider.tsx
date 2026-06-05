import { createContext, useEffect, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getRunnerJob } from '@/lib'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TRunnerJob } from '@/types'

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
  const { addToast } = useToast()

  const { data: job, isLoading, error } = useQuery({
    queryKey: ['runner-job', org.id, runnerJobId],
    queryFn: () => getRunnerJob({ runnerJobId, orgId: org.id }),
    refetchInterval: (query) => {
      const data = query.state.data
      return data && ACTIVE_STATUSES.includes(data.status ?? '') ? 5000 : false
    },
    enabled: !!org.id && !!runnerJobId,
  })

  useEffect(() => {
    if (error && job) {
      addToast(
        <Toast heading="Refresh failed" theme="warn">
          <Text>{(error as TAPIError)?.error ?? 'Connection issue'}</Text>
        </Toast>
      )
    }
  }, [error])

  if (error && !job) return <ProviderError error={error} />
  if (isLoading || !job) return <ProviderLoading />

  return (
    <RunnerJobContext.Provider value={{ job }}>
      {children}
    </RunnerJobContext.Provider>
  )
}
