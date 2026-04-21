import { createContext, useState, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useWorkflowMetrics } from '@/hooks/use-workflow-metrics'
import { useOrg } from '@/hooks/use-org'
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
  pollInterval = 4000,
  shouldPoll = false,
}: {
  children: ReactNode
  workflowId: string
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const { org } = useOrg()
  const [pollingEnabled, setPollingEnabled] = useState(shouldPoll)

  const { data: workflow, isLoading, error } = useQuery({
    queryKey: ['workflow', org.id!, workflowId],
    queryFn: () => getWorkflow({ orgId: org.id!, workflowId }),
    refetchInterval: pollingEnabled ? pollInterval : false,
    enabled: !!org.id && !!workflowId,
  })

  const metrics = useWorkflowMetrics(workflow)

  if (error) return <ProviderError error={error} />

  if (isLoading || !workflow) return <ProviderLoading />

  const value: WorkflowContextValue = {
    workflow,
    stopPolling: () => setPollingEnabled(false),
    ...metrics,
  }

  return (
    <WorkflowContext.Provider value={value}>
      {children}
    </WorkflowContext.Provider>
  )
}
