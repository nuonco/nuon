import { createContext, useState, type ReactNode } from 'react'
import { useWorkflowMetrics } from '@/hooks/use-workflow-metrics'
import { useOrg } from '@/hooks/use-org'
import { useSSEResourceQuery } from '@/hooks/use-sse-resource-query'
import { getWorkflow } from '@/lib'
import { ProviderError } from '@/components/layout/ProviderError'
import { ProviderLoading } from '@/components/layout/ProviderLoading'
import type { TWorkflow } from '@/types'

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
  const [sseEnabled, setSseEnabled] = useState(shouldPoll)

  const { data: workflow, isLoading, error, disconnect } = useSSEResourceQuery<TWorkflow>({
    sseUrl: org?.id && workflowId
      ? `/api/orgs/${org.id}/workflows/${workflowId}/sse`
      : undefined,
    queryKey: ['workflow', org?.id, workflowId],
    queryFn: () => getWorkflow({ orgId: org!.id, workflowId }),
    enabled: !!org?.id && !!workflowId,
    shouldPoll,
    sseEnabled,
    eventName: 'workflow',
    isFinished: (data) => !!data?.finished,
  })

  const metrics = useWorkflowMetrics(workflow)

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
