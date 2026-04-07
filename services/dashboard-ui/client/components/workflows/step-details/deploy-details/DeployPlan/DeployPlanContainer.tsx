import { useQueryApprovalPlan } from '@/hooks/use-query-approval-plan'
import type { TWorkflowStep } from '@/types'
import { DeployPlan } from './DeployPlan'

export const DeployPlanContainer = ({
  step,
  panelId,
}: {
  step: TWorkflowStep
  panelId?: string
}) => {
  const { plan, isLoading } = useQueryApprovalPlan({ step })

  return <DeployPlan step={step} plan={plan} isLoading={isLoading} panelId={panelId} />
}
