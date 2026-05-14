import { createContext, useMemo, useEffect, useCallback, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useResourceSSE } from '@/hooks/use-resource-sse'
import { useStatusToast } from '@/hooks/use-status-toast'
import { useToast } from '@/hooks/use-toast'
import { getInstallSandboxRun } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TSandboxRun, TWorkflow } from '@/types'

type SandboxRunContextValue = {
  sandboxRun: TSandboxRun
}

export const SandboxRunContext = createContext<SandboxRunContextValue | undefined>(
  undefined
)

const FALLBACK_POLL_MS = 4000
const FINISHED_POLL_MS = 30_000

export function SandboxRunProvider({
  children,
  installId,
  runId,
  shouldPoll = true,
}: {
  children: ReactNode
  installId?: string
  runId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const queryKey = ['sandbox-run', org?.id, runId]

  const invalidateTabQueries = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ['runner-job-plan'] })
  }, [queryClient])

  const sseUrl = org?.id && installId && runId
    ? `/api/orgs/${org.id}/installs/${installId}/sandbox-runs/${runId}/sse`
    : undefined

  const listeners = useMemo(() => ({
    'sandbox-run': (event: MessageEvent) => {
      try {
        const data: TSandboxRun = JSON.parse(event.data)
        queryClient.setQueryData(queryKey, data)
        invalidateTabQueries()
      } catch {}
    },
    workflow: (event: MessageEvent) => {
      try {
        const data: TWorkflow = JSON.parse(event.data)
        queryClient.setQueryData(['workflow', org?.id, data?.id], data)
      } catch {}
    },
  }), [org?.id, installId, runId])

  const { connected: sseConnected } = useResourceSSE({
    url: sseUrl,
    enabled: shouldPoll,
    listeners,
  })

  const { data: sandboxRun, isLoading, error } = useQuery({
    queryKey,
    queryFn: () => getInstallSandboxRun({ orgId: org!.id, runId }),
    refetchInterval: (query) => {
      if (sseConnected) return false
      if (!shouldPoll) return false
      const status = query.state.data?.status_v2?.status
      if (status === 'success' || status === 'error' || status === 'cancelled' || status === 'not-attempted') {
        return FINISHED_POLL_MS
      }
      return FALLBACK_POLL_MS
    },
    enabled: !!org?.id && !!runId,
  })

  useStatusToast({
    status: sandboxRun?.status_v2?.status,
    resourceType: 'sandbox run',
  })

  useEffect(() => {
    if (error && sandboxRun) {
      addToast(
        <Toast heading="Failed to refresh data" theme="warn">
          {(error as TAPIError)?.error ?? 'Connection issue'}
        </Toast>
      )
    }
  }, [error])

  if (error && !sandboxRun) return <ProviderError error={error} />
  if (isLoading || !sandboxRun) return <ProviderLoading />

  return (
    <SandboxRunContext.Provider value={{ sandboxRun }}>
      {children}
    </SandboxRunContext.Provider>
  )
}
