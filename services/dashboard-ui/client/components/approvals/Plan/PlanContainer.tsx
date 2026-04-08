import { useQueryApprovalPlan } from '@/hooks/use-query-approval-plan'
import type { TWorkflowStep } from '@/types'
import { Plan } from './Plan'

export const PlanContainer = ({ step }: { step: TWorkflowStep }) => {
  const { plan, isLoading, error } = useQueryApprovalPlan({ step })
  return <Plan step={step} plan={plan} isLoading={isLoading} error={error} />
}
