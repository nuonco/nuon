import { createContext, useState, useRef, useEffect, useCallback, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useStatusToast } from '@/hooks/use-status-toast'
import { useToast } from '@/hooks/use-toast'
import { getComponentBuild } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TBuild } from '@/types'

type BuildContextValue = {
  build: TBuild
}

export const BuildContext = createContext<BuildContextValue | undefined>(
  undefined
)

const FALLBACK_POLL_MS = 4000
const FINISHED_POLL_MS = 30_000

export function BuildProvider({
  children,
  buildId,
  componentId,
  componentName,
  shouldPoll = true,
}: {
  children: ReactNode
  buildId: string
  componentId: string
  componentName?: string
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

  const queryKey = ['build', org?.id, componentId, buildId]

  const { data: build, isLoading, error } = useQuery({
    queryKey,
    queryFn: () => getComponentBuild({ orgId: org!.id, componentId, buildId }),
    refetchInterval: (query) => {
      if (sseConnected) return false
      if (!shouldPoll) return false
      const status = query.state.data?.status_v2?.status
      if (status === 'success' || status === 'error' || status === 'failed' || status === 'cancelled' || status === 'not-attempted') {
        return FINISHED_POLL_MS
      }
      return FALLBACK_POLL_MS
    },
    enabled: !!org?.id && !!componentId && !!buildId,
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
    if (!org?.id || !componentId || !buildId || eventSourceRef.current) return

    const url = `/api/orgs/${org.id}/components/${componentId}/builds/${buildId}/sse`
    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    eventSource.addEventListener('build', (event: MessageEvent) => {
      try {
        const data: TBuild = JSON.parse(event.data)
        queryClient.setQueryData(queryKey, data)
        setSSEConnected(true)
        reconnectAttemptRef.current = 0
      } catch {
        // ignore parse errors
      }
    })

    eventSource.addEventListener('finished', () => {
      // build is done — server will slow down and eventually close
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
  }, [org?.id, componentId, buildId])

  useEffect(() => {
    if (sseEnabled && org?.id && componentId && buildId) {
      connectSSE()
    }
    return () => disconnect()
  }, [sseEnabled, org?.id, componentId, buildId, connectSSE, disconnect])

  useStatusToast({
    status: build?.status_v2?.status,
    label: componentName ?? build?.component_name,
    resourceType: 'build',
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
    <BuildContext.Provider value={{ build }}>
      {children}
    </BuildContext.Provider>
  )
}
