import { createContext, useState, useRef, useEffect, useCallback, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useStatusToast } from '@/hooks/use-status-toast'
import { useToast } from '@/hooks/use-toast'
import { getSandboxBuild } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TAppSandboxBuild } from '@/types'

type SandboxBuildContextValue = {
  build: TAppSandboxBuild
}

export const SandboxBuildContext = createContext<
  SandboxBuildContextValue | undefined
>(undefined)

const FALLBACK_POLL_MS = 4000
const FINISHED_POLL_MS = 30_000

export function SandboxBuildProvider({
  children,
  buildId,
  shouldPoll = true,
}: {
  children: ReactNode
  buildId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { app } = useApp()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const [sseConnected, setSSEConnected] = useState(false)
  const [sseEnabled, setSseEnabled] = useState(shouldPoll)
  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const reconnectAttemptRef = useRef(0)

  const queryKey = ['sandbox-build', org?.id, app?.id, buildId]

  const { data: build, isLoading, error } = useQuery({
    queryKey,
    queryFn: () => getSandboxBuild({ orgId: org!.id, appId: app!.id, buildId }),
    refetchInterval: (query) => {
      if (sseConnected) return false
      if (!shouldPoll) return false
      const status = query.state.data?.status_v2?.status
      if (status === 'success' || status === 'error' || status === 'failed' || status === 'cancelled' || status === 'not-attempted') {
        return FINISHED_POLL_MS
      }
      return FALLBACK_POLL_MS
    },
    enabled: !!org?.id && !!app?.id && !!buildId,
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
    if (!org?.id || !app?.id || !buildId || eventSourceRef.current) return

    const url = `/api/orgs/${org.id}/apps/${app.id}/sandbox-builds/${buildId}/sse`
    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    eventSource.addEventListener('sandbox-build', (event: MessageEvent) => {
      try {
        const data: TAppSandboxBuild = JSON.parse(event.data)
        queryClient.setQueryData(queryKey, data)
        setSSEConnected(true)
        reconnectAttemptRef.current = 0
      } catch {
        // ignore parse errors
      }
    })

    eventSource.addEventListener('finished', () => {
      // sandbox build is done — server will slow down and eventually close
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
  }, [org?.id, app?.id, buildId])

  useEffect(() => {
    if (sseEnabled && org?.id && app?.id && buildId) {
      connectSSE()
    }
    return () => disconnect()
  }, [sseEnabled, org?.id, app?.id, buildId, connectSSE, disconnect])

  useStatusToast({
    status: build?.status_v2?.status,
    resourceType: 'sandbox build',
  })

  useEffect(() => {
    if (error && build) {
      addToast(
        <Toast heading="Failed to refresh data" theme="warn">
          {(error as TAPIError)?.error ?? 'Connection issue'}
        </Toast>
      )
    }
  }, [error])

  if (error && !build) return <ProviderError error={error} />

  if (isLoading || !build) return <ProviderLoading />

  return (
    <SandboxBuildContext.Provider value={{ build }}>
      {children}
    </SandboxBuildContext.Provider>
  )
}
