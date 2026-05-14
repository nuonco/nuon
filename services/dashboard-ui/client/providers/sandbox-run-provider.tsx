import { createContext, useState, useRef, useEffect, useCallback, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
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

  const [sseConnected, setSSEConnected] = useState(false)
  const [sseEnabled, setSseEnabled] = useState(shouldPoll)
  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const reconnectAttemptRef = useRef(0)

  const queryKey = ['sandbox-run', org?.id, runId]

  const { data: sandboxRun, isLoading, error } = useQuery({
    queryKey,
    queryFn: () => getInstallSandboxRun({ orgId: org!.id, runId }),
    refetchInterval: (query) => {
      if (sseConnected) return false
      if (!shouldPoll) return false
      const status = query.state.data?.status_v2?.status
      if (status === 'success' || status === 'error' || status === 'failed' || status === 'cancelled' || status === 'not-attempted') {
        return FINISHED_POLL_MS
      }
      return FALLBACK_POLL_MS
    },
    enabled: !!org?.id && !!runId,
  })

  const invalidateTabQueries = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ['runner-job-plan'] })
  }, [queryClient])

  const disconnect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
      reconnectTimeoutRef.current = null
    }
    setSSEConnected(false)
  }, [])

  const connectSSE = useCallback(() => {
    if (!org?.id || !installId || !runId || eventSourceRef.current) return

    const url = `/api/orgs/${org.id}/installs/${installId}/sandbox-runs/${runId}/sse`
    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    eventSource.addEventListener('sandbox-run', (event: MessageEvent) => {
      try {
        const data: TSandboxRun = JSON.parse(event.data)
        queryClient.setQueryData(queryKey, data)
        setSSEConnected(true)
        reconnectAttemptRef.current = 0
        invalidateTabQueries()
      } catch {
        // ignore parse errors
      }
    })

    eventSource.addEventListener('workflow', (event: MessageEvent) => {
      try {
        const data: TWorkflow = JSON.parse(event.data)
        queryClient.setQueryData(['workflow', org?.id, data?.id], data)
      } catch {
        // ignore parse errors
      }
    })

    eventSource.addEventListener('finished', () => {
      // sandbox run is done — server will slow down and eventually close
    })

    eventSource.addEventListener('fetch-error', (event: MessageEvent) => {
      try {
        const errorData = JSON.parse(event.data)
        addToast(
          <Toast heading="Failed to refresh data" theme="warn">
            {errorData?.error ?? 'Connection issue'}
          </Toast>
        )
      } catch {
        // non-JSON error event, handled by onerror
      }
    })

    eventSource.onerror = () => {
      eventSource.close()
      eventSourceRef.current = null
      setSSEConnected(false)

      const backoffDelay = Math.min(1000 * Math.pow(2, reconnectAttemptRef.current), 30000)
      reconnectAttemptRef.current += 1

      reconnectTimeoutRef.current = setTimeout(() => {
        connectSSE()
      }, backoffDelay)
    }

    eventSource.onopen = () => {
      setSSEConnected(true)
      reconnectAttemptRef.current = 0
    }
  }, [org?.id, installId, runId])

  useEffect(() => {
    if (sseEnabled && org?.id && installId && runId) {
      connectSSE()
    }
    return () => disconnect()
  }, [sseEnabled, org?.id, installId, runId, connectSSE, disconnect])

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
