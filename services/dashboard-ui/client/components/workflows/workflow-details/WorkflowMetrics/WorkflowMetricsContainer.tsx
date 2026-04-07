import { useWorkflow } from '@/hooks/use-workflow'
import { WorkflowMetrics } from './WorkflowMetrics'

export const WorkflowMetricsContainer = () => {
  const {
    workflow,
    pendingApprovalsCount,
    policyViolationsCount,
    discardedStepsCount,
    completedStepsCount,
    totalSteps,
  } = useWorkflow()

  if (!workflow) return null

  return (
    <WorkflowMetrics
      workflow={workflow}
      pendingApprovalsCount={pendingApprovalsCount}
      policyViolationsCount={policyViolationsCount}
      discardedStepsCount={discardedStepsCount}
      completedStepsCount={completedStepsCount}
      totalSteps={totalSteps}
    />
  )
}
