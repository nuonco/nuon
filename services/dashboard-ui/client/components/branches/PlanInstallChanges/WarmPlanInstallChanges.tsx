import { usePlanGroupInstalls } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep'
import type { TInstallWorkflowStep } from '@/types'

const awaitingPlan = (step: TInstallWorkflowStep) =>
  step.execution_type === 'approval' &&
  step.status?.status === 'approval-awaiting' &&
  !step.approval?.response &&
  !!step.approval?.id &&
  !!step.id &&
  !!step.install_workflow_id

const PlanInstallChangesWarmup = ({ step }: { step: TInstallWorkflowStep }) => {
  usePlanGroupInstalls(
    step,
    (step.status?.metadata ?? {}) as Record<string, any>
  )
  return null
}

export const WarmPlanInstallChanges = ({
  steps,
}: {
  steps?: TInstallWorkflowStep[]
}) => (
  <>
    {(steps ?? []).filter(awaitingPlan).map((step) => (
      <PlanInstallChangesWarmup key={step.id} step={step} />
    ))}
  </>
)
