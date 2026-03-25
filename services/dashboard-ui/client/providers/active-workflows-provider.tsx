import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getOrgWorkflows } from '@/lib'
import { useOrg } from '@/hooks/use-org'
import type { TWorkflow } from '@/types'

type ActiveWorkflowsContextValue = {
  activeWorkflows: TWorkflow[]
  isLoading: boolean
  refresh: () => void
}

export const ActiveWorkflowsContext = createContext<
  ActiveWorkflowsContextValue | undefined
>(undefined)

export function ActiveWorkflowsProvider({
  children,
}: {
  children: ReactNode
}) {
  const { org } = useOrg()

  const {
    data,
    isLoading,
    refetch,
  } = useQuery({
    queryKey: ['active-workflows', org.id],
    queryFn: () =>
      getOrgWorkflows({
        orgId: org.id,
        finished: false,
        limit: 50,
        offset: 0,
        planonly: false,
      }),
    // refetchInterval: 20_000,
  })

  const activeWorkflows = (data?.data ?? []).filter(
    (w) =>
      w.status?.status &&
      w.status.status !== 'pending' &&
      w.status.status !== 'queued' &&
      w.status.status !== 'cancelled' &&
      w.status.status !== 'error' &&
      w.status.status !== 'success'
  )

  return (
    <ActiveWorkflowsContext.Provider
      value={{ activeWorkflows, isLoading, refresh: refetch }}
    >
      {children}
    </ActiveWorkflowsContext.Provider>
  )
}
