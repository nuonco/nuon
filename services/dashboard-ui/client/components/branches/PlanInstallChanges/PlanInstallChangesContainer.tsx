import { useApp } from '@/hooks/use-app'
import { usePlanGroupInstalls } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep'
import type { TInstallWorkflowStep } from '@/types'
import {
  PlanInstallChanges,
  type TPlanInstallFacts,
} from './PlanInstallChanges'

interface IPlanInstallChangesContainer {
  step: TInstallWorkflowStep
  installFacts?: TPlanInstallFacts
}

export const PlanInstallChangesContainer = ({
  step,
  installFacts,
}: IPlanInstallChangesContainer) => {
  const { labelColors } = useApp()
  const { installs, isLoading } = usePlanGroupInstalls(
    step,
    (step.status?.metadata ?? {}) as Record<string, any>
  )

  return (
    <PlanInstallChanges
      installs={installs}
      installFacts={installFacts}
      labelColors={labelColors}
      isLoading={isLoading}
    />
  )
}
