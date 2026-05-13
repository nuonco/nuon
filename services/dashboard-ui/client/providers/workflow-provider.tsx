import { createContext, useState, useRef, useEffect, useCallback, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useWorkflowMetrics } from '@/hooks/use-workflow-metrics'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { getWorkflow } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TAPIError, TWorkflow } from '@/types'

interface WorkflowContextValue {
  workflow: TWorkflow
  stopPolling: () => void
  workflowSteps: any[]
  hasApprovals: boolean
  failedSteps: any[]
  pendingApprovals: any[]
  discardedSteps: any[]
  completedSteps: any[]
  stepsWithPolicyViolations: any[]
  totalSteps: number
  pendingApprovalsCount: number
  discardedStepsCount: number
  completedStepsCount: number
  failedStepsCount: number
  policyViolationsCount: number
}

export const WorkflowContext = createContext<WorkflowContextValue | undefined>(undefined)

const FALLBACK_POLL_MS = 4000
const FINISHED_POLL_MS = 30_000

export const WorkflowProvider = ({
  children,
  workflowId,
  shouldPoll = false,
}: {
  children: ReactNode
  workflowId: string
  shouldPoll?: boolean
}) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { addToast } = useToast()

  const [sseConnected, setSSEConnected] = useState(false)
  const [sseEnabled, setSseEnabled] = useState(shouldPoll)
  const eventSourceRef = useRef<EventSource | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const reconnectAttemptRef = useRef(0)

  const queryKey = ['workflow', org?.id, workflowId]

  const { data: workflow, isLoading, error } = useQuery({
    queryKey,
    queryFn: () => getWorkflow({ orgId: org!.id, workflowId }),
    refetchInterval: (query) => {
      if (sseConnected) return false
      if (!shouldPoll) return false
      if (query.state.data?.finished) return FINISHED_POLL_MS
      return FALLBACK_POLL_MS
    },
    enabled: !!org?.id && !!workflowId,
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
    if (!org?.id || !workflowId || eventSourceRef.current) return

    const url = `/api/orgs/${org.id}/workflows/${workflowId}/sse`
    const eventSource = new EventSource(url)
    eventSourceRef.current = eventSource

    eventSource.onmessage = (event) => {
      try {
        const data: TWorkflow = JSON.parse(event.data)
        queryClient.setQueryData(queryKey, data)
        setSSEConnected(true)
        reconnectAttemptRef.current = 0
      } catch {
        // ignore parse errors
      }
    }

    eventSource.addEventListener('finished', () => {
      // workflow is done — server will slow down and eventually close
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
  }, [org?.id, workflowId])

  useEffect(() => {
    if (sseEnabled && org?.id && workflowId) {
      connectSSE()
    }
    return () => disconnect()
  }, [sseEnabled, org?.id, workflowId, connectSSE, disconnect])

  const metrics = useWorkflowMetrics(workflow)

  useEffect(() => {
    if (error && workflow) {
      addToast(
        <Toast heading="Failed to refresh data" theme="warn">
          {(error as TAPIError)?.error ?? 'Connection issue'}
        </Toast>
      )
    }
  }, [error])

  if (error && !workflow) return <ProviderError error={error} />

  if (isLoading || !workflow) return <ProviderLoading />

  const value: WorkflowContextValue = {
    workflow,
    stopPolling: () => {
      setSseEnabled(false)
      disconnect()
    },
    ...metrics,
  }

  return (
    <WorkflowContext.Provider value={value}>
      {children}
    </WorkflowContext.Provider>
  )
}
