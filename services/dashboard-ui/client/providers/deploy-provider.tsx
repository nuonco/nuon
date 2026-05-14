import { createContext, useState, useRef, useEffect, useCallback, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { useStatusToast } from '@/hooks/use-status-toast'
import { useToast } from '@/hooks/use-toast'
import { getDeploy } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TComponent, TDeploy, TWorkflow } from '@/types'

type DeployContextValue = {
  deploy: TDeploy
}

export const DeployContext = createContext<DeployContextValue | undefined>(
  undefined
)

const FALLBACK_POLL_MS = 4000
const FINISHED_POLL_MS = 30_000

export function DeployProvider({
  children,
  deployId,
  installId,
  shouldPoll = true,
}: {
  children: ReactNode
  deployId: string
  installId: string
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

  const queryKey = ['deploy', org?.id, installId, deployId]

  const { data: deploy, isLoading, error } = useQuery({
    queryKey,
    queryFn: () => getDeploy({ orgId: org!.id, installId, deployId }),
    refetchInterval: (query) => {
      if (sseConnected) return false
      if (!shouldPoll) return false
      const status = query.state.data?.status_v2?.status
      if (status === 'success' || status === 'error' || status === 'failed' || status === 'cancelled' || status === 'not-attempted') {
        return FINISHED_POLL_MS
      }
      return FALLBACK_POLL_MS
    },
    enabled: !!org?.id && !!installId && !!deployId,
  })

  const invalidateTabQueries = useCallback(() => {
    queryClient.invalidateQueries({ queryKey: ['runner-job-plan'] })
    queryClient.invalidateQueries({ queryKey: ['install-component-outputs', org?.id, installId] })
    queryClient.invalidateQueries({ queryKey: ['install-component', org?.id, installId] })
  }, [queryClient, org?.id, installId])

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
    if (!org?.id || !installId || !deployId || eventSourceRef.current) return

    const url = `/api/orgs/${org.id}/installs/${installId}/deploys/${deployId}/sse`
    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    eventSource.addEventListener('deploy', (event: MessageEvent) => {
      try {
        const data: TDeploy = JSON.parse(event.data)
        queryClient.setQueryData(queryKey, data)
        setSSEConnected(true)
        reconnectAttemptRef.current = 0
        invalidateTabQueries()
      } catch {
        // ignore parse errors
      }
    })

    eventSource.addEventListener('component', (event: MessageEvent) => {
      try {
        const data: TComponent = JSON.parse(event.data)
        queryClient.setQueryData(['component', org?.id, data?.id], data)
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
      // deploy is done — server will slow down and eventually close
    })

    eventSource.addEventListener('error', (event: MessageEvent) => {
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
  }, [org?.id, installId, deployId])

  useEffect(() => {
    if (sseEnabled && org?.id && installId && deployId) {
      connectSSE()
    }
    return () => disconnect()
  }, [sseEnabled, org?.id, installId, deployId, connectSSE, disconnect])

  useStatusToast({
    status: deploy?.status_v2?.status,
    label: deploy?.component_name,
    resourceType: 'deploy',
  })

  useEffect(() => {
    if (error && deploy) {
      addToast(
        <Toast heading="Failed to refresh data" theme="warn">
          {(error as TAPIError)?.error ?? 'Connection issue'}
        </Toast>
      )
    }
  }, [error])

  if (error && !deploy) return <ProviderError error={error} />

  if (isLoading || !deploy) return <ProviderLoading />

  return (
    <DeployContext.Provider value={{ deploy }}>
      {children}
    </DeployContext.Provider>
  )
}
