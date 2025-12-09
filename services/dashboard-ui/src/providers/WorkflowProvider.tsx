'use client'

import { createContext, type ReactNode } from 'react'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useWorkflowMetrics } from '@/hooks/use-workflow-metrics'
import { useOrg } from '@/hooks/use-org'
import type { TWorkflow } from '@/types'

interface WorkflowContextValue {
  workflow: TWorkflow
  isLoading: boolean
  error?: any
  stopPolling: () => void
  // Workflow metrics
  workflowSteps: any[]
  hasApprovals: boolean
  failedSteps: any[]
  pendingApprovals: any[]
  discardedSteps: any[]
  completedSteps: any[]
  totalSteps: number
  pendingApprovalsCount: number
  discardedStepsCount: number
  completedStepsCount: number
  failedStepsCount: number
}

interface WorkflowProviderProps extends Partial<IPollingProps> {
  children: ReactNode
  initWorkflow: TWorkflow
}

export const WorkflowContext = createContext<WorkflowContextValue | undefined>(undefined)

export const WorkflowProvider = ({
  children,
  initWorkflow,
  pollInterval = 10000,
  shouldPoll = false,
}: WorkflowProviderProps) => {
  const { org } = useOrg()
  
  const { data: workflow, isLoading, error, stopPolling } = usePolling<TWorkflow>({
    initData: initWorkflow,
    path: `/api/orgs/${org.id}/workflows/${initWorkflow.id}`,
    pollInterval,
    shouldPoll,
  })

  const metrics = useWorkflowMetrics(workflow)

  const value: WorkflowContextValue = {
    workflow,
    isLoading,
    error,
    stopPolling,
    ...metrics,
  }

  return (
    <WorkflowContext.Provider value={value}>
      {children}
    </WorkflowContext.Provider>
  )
}