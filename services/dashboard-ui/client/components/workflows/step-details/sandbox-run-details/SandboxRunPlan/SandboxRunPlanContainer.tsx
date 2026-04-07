import { useQueryApprovalPlan } from '@/hooks/use-query-approval-plan'
import type { TWorkflowStep } from '@/types'
import { SandboxRunPlan } from './SandboxRunPlan'

export const SandboxRunPlanContainer = ({ step }: { step: TWorkflowStep }) => {
  const { plan, isLoading } = useQueryApprovalPlan({ step })

  return <SandboxRunPlan plan={plan} isLoading={isLoading} />
}
