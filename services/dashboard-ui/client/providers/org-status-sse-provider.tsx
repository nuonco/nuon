import { createContext, useMemo, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useResourceSSE } from '@/hooks/use-resource-sse'
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
    org: (event: MessageEvent) => {
      try {
        const data: TOrg = JSON.parse(event.data)
        queryClient.setQueryData(['org', org?.id], data)
      } catch {}
    },
    'active-workflows': (event: MessageEvent) => {
      try {
        const data: TWorkflow[] = JSON.parse(event.data)
        queryClient.setQueryData<TPaginatedResult<TWorkflow[]>>(
          ['active-workflows', org?.id],
          { data, pagination: { hasNext: false, offset: 0, limit: 50 } },
        )
      } catch {}
    },
    'pending-approvals': (event: MessageEvent) => {
      try {
        const data: TWorkflowStepApproval[] = JSON.parse(event.data)
        queryClient.setQueryData(['workflow-approvals', org?.id], data)
      } catch {}
    },
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
