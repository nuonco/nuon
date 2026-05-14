import { createContext, useState, useRef, useEffect, useCallback, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getInstallActionRun } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TInstallActionRun, TWorkflow } from '@/types'

type InstallActionRunContextValue = {
  installActionRun: TInstallActionRun
  refresh: () => void
}

export const InstallActionRunContext = createContext<
  InstallActionRunContextValue | undefined
>(undefined)

const FALLBACK_POLL_MS = 4000
const FINISHED_POLL_MS = 30_000

export function InstallActionRunProvider({
  children,
  runId,
  shouldPoll = false,
}: {
  children: ReactNode
  runId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { install } = useInstall()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const [sseConnected, setSSEConnected] = useState(false)
  const [sseEnabled, setSseEnabled] = useState(shouldPoll)
  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const reconnectAttemptRef = useRef(0)

  const queryKey = ['install-action-run', org?.id, install?.id, runId]

  const { data: installActionRun, isLoading, error, refetch } = useQuery({
    queryKey,
    queryFn: () => getInstallActionRun({ orgId: org!.id, installId: install!.id, runId }),
    refetchInterval: (query) => {
      if (sseConnected) return false
      if (!shouldPoll) return false
      const status = query.state.data?.status_v2?.status
      if (status === 'success' || status === 'error' || status === 'failed' || status === 'cancelled' || status === 'not-attempted') {
        return FINISHED_POLL_MS
      }
      return FALLBACK_POLL_MS
    },
    enabled: !!org?.id && !!install?.id && !!runId,
  })

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
    if (!org?.id || !install?.id || !runId || eventSourceRef.current) return

    const url = `/api/orgs/${org.id}/installs/${install.id}/action-runs/${runId}/sse`
    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    eventSource.addEventListener('action-run', (event: MessageEvent) => {
      try {
        const data: TInstallActionRun = JSON.parse(event.data)
        queryClient.setQueryData(queryKey, data)
        setSSEConnected(true)
        reconnectAttemptRef.current = 0
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
      // action run is done — server will slow down and eventually close
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
  }, [org?.id, install?.id, runId])

  useEffect(() => {
    if (sseEnabled && org?.id && install?.id && runId) {
      connectSSE()
    }
    return () => disconnect()
  }, [sseEnabled, org?.id, install?.id, runId, connectSSE, disconnect])

  useEffect(() => {
    if (error && installActionRun) {
      addToast(
        <Toast heading="Failed to refresh data" theme="warn">
          {(error as TAPIError)?.error ?? 'Connection issue'}
        </Toast>
      )
    }
  }, [error])

  if (error && !installActionRun) return <ProviderError error={error} />

  if (isLoading || !installActionRun) return <ProviderLoading />

  return (
    <InstallActionRunContext.Provider value={{ installActionRun, refresh: refetch }}>
      {children}
    </InstallActionRunContext.Provider>
  )
}
