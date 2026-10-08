import { Panel, type IPanel } from '@/components/surfaces/Panel'
import {
  PlanInstallChanges,
  type TPlanInstallFacts,
} from '@/components/branches/PlanInstallChanges'
import type { TInstallWorkflowStep } from '@/types'
import { getGroupName } from './use-group-plan-href'

export const PlanDiffPanel = ({
  step,
  installFacts,
  ...props
}: IPanel & {
  step: TInstallWorkflowStep
  installFacts?: TPlanInstallFacts
}) => (
  <Panel {...props} size="3/4" heading={`Plan for ${getGroupName(step.name)}`}>
    <PlanInstallChanges step={step} installFacts={installFacts} />
  </Panel>
)
