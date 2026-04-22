import { useMemo } from 'react'
import type { TWorkflow } from '@/types'

export const useWorkflowMetrics = (workflow: TWorkflow | undefined) => {
  return useMemo(() => {
    const workflowSteps =
      workflow?.steps?.filter((s) => s?.execution_type !== 'hidden') || []

    const metricSteps = workflowSteps.filter((s) => !s?.retried)

    const hasApprovals = metricSteps.some(
      (step) => step?.execution_type === 'approval'
    )
    
    const failedSteps = metricSteps.filter(
      (step) => step?.status?.status === 'error'
    )

    const pendingApprovals = metricSteps.filter(
      (s) =>
        s?.execution_type === 'approval' &&
        !s?.approval?.response &&
        s?.status?.status !== 'discarded'
    )

    const discardedSteps = metricSteps.filter(
      (s) => s?.status?.status === 'discarded'
    )

    const completedSteps = metricSteps.filter((s) => s?.finished)

    const stepsWithPolicyViolations = metricSteps.filter(
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
      totalSteps: metricSteps.length,
      pendingApprovalsCount: pendingApprovals.length,
      discardedStepsCount: discardedSteps.length,
      completedStepsCount: completedSteps.length,
      failedStepsCount: failedSteps.length,
      policyViolationsCount: stepsWithPolicyViolations.length,
    }
  }, [workflow])
}