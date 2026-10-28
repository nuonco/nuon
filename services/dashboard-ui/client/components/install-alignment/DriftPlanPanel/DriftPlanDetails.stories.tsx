export default {
  title: 'Features / Installs / Alignment / Drift plan',
}

import { Button } from '@/components/common/Button'
import { PlanComponent } from '@/components/approvals/Plan'
import {
  openDeploymentResource,
  selectedCommit,
  stateDriftResource,
  unappliedResource,
} from '@/components/__fixtures__/install-alignment'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { securityGroupUpdatePlan } from '@/lib/fixtures/plan-diffs/terraform'
import type { TInstallAlignmentResource, TWorkflowStep } from '@/types'
import { DriftPlanDetails, type IDriftPlanDetails } from './DriftPlanDetails'

const step = {
  id: 'step-plan-ic-api',
  install_workflow_id: 'wf-plan-ic-api',
  execution_type: 'approval',
  finished: true,
  approval: { id: 'apr-ic-api', type: 'terraform_plan' },
} as TWorkflowStep

const details = (
  resource: TInstallAlignmentResource,
  props: Partial<IDriftPlanDetails> = {}
) => (
  <DriftPlanDetails
    resource={resource}
    selectedCommit={selectedCommit}
    selectedCommitHref="#selected-commit"
    planWorkflowHref="#planning-deployment"
    deploymentHref="#deployment"
    onApproveAndResolve={() => {}}
    plan={
      <PlanComponent
        step={step}
        plan={securityGroupUpdatePlan}
        isLoading={false}
        error={null}
      />
    }
    {...props}
  />
)

export const DeploymentPlan = () => (
  <div className="p-4">{details(unappliedResource())}</div>
)

export const Resolving = () => (
  <div className="p-4">{details(unappliedResource(), { resolving: true })}</div>
)

export const ResolveError = () => (
  <div className="p-4">
    {details(unappliedResource(), {
      resolveError: {
        status: 409,
        user_error: true,
        error: 'Target changed',
        description: 'Input values changed. Reload Resources and try again.',
      },
    })}
  </div>
)

export const OpenDeployment = () => (
  <div className="p-4">{details(openDeploymentResource())}</div>
)

export const StateDriftScan = () => (
  <div className="p-4">{details(stateDriftResource())}</div>
)

export const PlanLoading = () => (
  <div className="p-4">
    {details(unappliedResource(), {
      resolveDisabled: true,
      plan: (
        <PlanComponent step={step} plan={undefined} isLoading error={null} />
      ),
    })}
  </div>
)

export const InPanel = () => {
  const { addPanel } = useSurfaces()
  const open = () =>
    addPanel(
      <Panel
        heading="api · Deployment plan"
        size="half"
        className="!h-auto max-h-screen"
        data-testid="drift-plan-panel"
      >
        {details(unappliedResource())}
      </Panel>
    )
  return (
    <div className="p-4">
      <Button onClick={open}>View plan</Button>
    </div>
  )
}
