import { useContext } from 'react'
import { WorkflowContext } from '@/providers/workflow-provider'
import type { TWorkflow } from '@/types'

export const useWorkflow = (): { workflow: TWorkflow; stopPolling: () => void; workflowSteps: any[]; hasApprovals: boolean; failedSteps: any[]; pendingApprovals: any[]; discardedSteps: any[]; completedSteps: any[]; stepsWithPolicyViolations: any[]; totalSteps: number; pendingApprovalsCount: number; discardedStepsCount: number; completedStepsCount: number; failedStepsCount: number; policyViolationsCount: number } => {
  const context = useContext(WorkflowContext)
  if (context === undefined) {
    throw new Error('useWorkflow must be used within a WorkflowProvider')
  }
  return context
}
