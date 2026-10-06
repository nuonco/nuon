import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { PlanGroupStep } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep'
import type { TInstallWorkflowStep } from '@/types'
import { getGroupName } from './use-group-plan-href'

export const PlanDiffPanel = ({
  step,
  installFacts,
  ...props
}: IPanel & {
  step: TInstallWorkflowStep
  installFacts?: Record<
    string,
    {
      labels?: Record<string, string>
      region?: string
      status?: string
      detail?: string
      appliedConfigId?: string
    }
  >
}) => (
  <Panel
    {...props}
    size="3/4"
    heading={`Plan for ${getGroupName(step.name)}`}
  >
    <PlanGroupStep
      step={step}
      metadata={(step.status?.metadata ?? {}) as Record<string, any>}
      hideHeading
      diffOnly
      installFacts={installFacts}
    />
  </Panel>
)
