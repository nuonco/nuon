import { useMemo } from 'react'
import type { TWorkflow } from '@/types'

export const useWorkflowMetrics = (workflow: TWorkflow | undefined) => {
  return useMemo(() => {
    const workflowSteps = workflow?.steps?.filter((s) => s?.execution_type !== 'hidden') || []
    
    const hasApprovals = workflowSteps.some(
      (step) => step?.execution_type === 'approval'
    )
    
    const failedSteps = workflowSteps.filter(
      (step) => step?.status?.status === 'error'
    )
    
    const pendingApprovals = workflowSteps.filter(
      (s) =>
        s?.execution_type === 'approval' &&
        !s?.approval?.response &&
        s?.status?.status !== 'discarded'
    )
    
    const discardedSteps = workflowSteps.filter(
      (s) => s?.status?.status === 'discarded'
    )
    
    const completedSteps = workflowSteps.filter((s) => s?.finished)

    const stepsWithPolicyViolations = workflowSteps.filter(
      (s) =>
        ((s?.status?.metadata?.deny_violations as unknown[])?.length || 0) > 0 ||
        ((s?.status?.metadata?.warn_violations as unknown[])?.length || 0) > 0
    )

    return {
      workflowSteps,
      hasApprovals,
      failedSteps,
      pendingApprovals,
      discardedSteps,
      completedSteps,
      stepsWithPolicyViolations,
      totalSteps: workflowSteps.length,
      pendingApprovalsCount: pendingApprovals.length,
      discardedStepsCount: discardedSteps.length,
      completedStepsCount: completedSteps.length,
      failedStepsCount: failedSteps.length,
      policyViolationsCount: stepsWithPolicyViolations.length,
    }
  }, [workflow])
}