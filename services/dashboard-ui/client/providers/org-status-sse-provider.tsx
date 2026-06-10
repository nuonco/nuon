import { createContext, useMemo, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useResourceSSE } from '@/hooks/use-resource-sse'
import { createSSEQueryListener } from '@/lib/sse-listeners'
import type { TPaginatedResult } from '@/lib/api'
import type {
  TOrg,
  TWorkflow,
  TWorkflowStepApproval,
} from '@/types'

type OrgStatusSSEContextValue = {
  sseConnected: boolean
}

export const OrgStatusSSEContext = createContext<
  OrgStatusSSEContextValue | undefined
>(undefined)

export function OrgStatusSSEProvider({
  children,
}: {
  children: ReactNode
}) {
  const { org } = useOrg()
  const queryClient = useQueryClient()

  const sseUrl = org?.id
    ? `/api/orgs/${org.id}/status/sse`
    : undefined

  const listeners = useMemo(() => ({
    org: createSSEQueryListener<TOrg>(queryClient, ['org', org?.id]),
    'active-workflows': createSSEQueryListener<TWorkflow[]>(
      queryClient,
      ['active-workflows', org?.id],
      {
        transform: (data): TPaginatedResult<TWorkflow[]> => ({
          data,
          pagination: { hasNext: false, offset: 0, limit: 50 },
        }),
      }
    ),
    'pending-approvals': createSSEQueryListener<TWorkflowStepApproval[]>(
      queryClient,
      ['workflow-approvals', org?.id]
    ),
    'runner-heartbeat': (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data)
        const runnerId = org?.runner_group?.runners?.[0]?.id
        if (runnerId) {
          queryClient.setQueryData(['runner-heartbeat', org?.id, runnerId], data)
        }
      } catch {}
    },
  }), [org?.id, queryClient])

  const { connected } = useResourceSSE({
    url: sseUrl,
    enabled: true,
    listeners,
  })

  return (
    <OrgStatusSSEContext.Provider value={{ sseConnected: connected }}>
      {children}
    </OrgStatusSSEContext.Provider>
  )
}
