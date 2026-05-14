import { createContext, useState, useMemo, useEffect, type ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useWorkflowMetrics } from '@/hooks/use-workflow-metrics'
import { useOrg } from '@/hooks/use-org'
import { useResourceSSE } from '@/hooks/use-resource-sse'
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
  const [sseEnabled, setSseEnabled] = useState(shouldPoll)
  const queryKey = ['workflow', org?.id, workflowId]

  const sseUrl = org?.id && workflowId
    ? `/api/orgs/${org.id}/workflows/${workflowId}/sse`
    : undefined

  const onMessage = useMemo(() => (event: MessageEvent) => {
    try {
      const data: TWorkflow = JSON.parse(event.data)
      queryClient.setQueryData(queryKey, data)
    } catch {}
  }, [org?.id, workflowId])

  const { connected: sseConnected, disconnect } = useResourceSSE({
    url: sseUrl,
    enabled: sseEnabled,
    onMessage,
  })

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
