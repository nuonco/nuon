export default {
  title: 'Features / Installs / Alignment / Drift resource card',
}

import {
  causeResources,
  manualOrigin,
  openDeploymentResource,
  stateDriftResource,
  unappliedResource,
} from '@/components/__fixtures__/install-alignment'
import type { TInstallResourceSyncStatusOrigin } from '@/types'
import { DriftResourceCard } from './DriftResourceCard'

const links = {
  originHref: (origin?: TInstallResourceSyncStatusOrigin) =>
    origin?.run
      ? `#run-${origin.run.run_id}`
      : `#workflow-${origin?.workflow_id}`,
  driftCheckHref: '#drift-check',
  historyHref: '#deploy-history',
  deploymentHref: '#deployment',
  nameHref: '#component',
  onResolve: () => {},
  onViewPlan: () => {},
}

export const NoOpenDeployment = () => (
  <DriftResourceCard {...links} resource={unappliedResource()} />
)

export const Resolving = () => (
  <DriftResourceCard {...links} resource={unappliedResource()} resolving />
)

export const ResolveError = () => (
  <DriftResourceCard
    {...links}
    resource={unappliedResource()}
    resolveError={{
      status: 409,
      user_error: true,
      error: 'Target changed',
      description: 'Input values changed. Reload Resources and try again.',
    }}
  />
)

export const LiveFromManualDeploy = () => (
  <DriftResourceCard
    {...links}
    resource={unappliedResource({ applied_by: manualOrigin })}
  />
)

export const OpenDeployment = () => (
  <DriftResourceCard {...links} resource={openDeploymentResource()} />
)

export const FirstImageSync = () => (
  <DriftResourceCard
    {...links}
    onViewPlan={undefined}
    resource={openDeploymentResource({
      kind: 'image',
      component_type: 'external_image',
      applied_build_id: undefined,
      applied_by: undefined,
      applied_deploy_at: undefined,
      plan: undefined,
      open_deployment: { id: 'wf-first-sync', status: 'in-progress' },
    })}
  />
)

export const SandboxAwaitingApproval = () => (
  <DriftResourceCard
    {...links}
    resource={openDeploymentResource({
      kind: 'sandbox',
      name: 'sandbox',
      component_id: undefined,
      component_type: undefined,
      open_deployment: {
        id: 'wf-sandbox',
        status: 'approval-awaiting',
        type: 'reprovision',
      },
    })}
  />
)

export const StateDriftDetected = () => (
  <DriftResourceCard
    {...links}
    resource={stateDriftResource({
      id: 'sbx-1',
      kind: 'sandbox',
      name: 'sandbox',
      component_id: undefined,
      component_type: undefined,
    })}
  />
)

const causeCard = (view: 'list' | 'grid') => {
  const Causes = () => (
    <div
      className={
        view === 'grid' ? 'grid md:grid-cols-2 gap-3' : 'flex flex-col divide-y'
      }
    >
      {causeResources().map((resource) => (
        <DriftResourceCard
          key={resource.id}
          {...links}
          view={view}
          resource={resource}
          driftCheckHref={resource.plan?.plan_only ? '#drift-check' : undefined}
          onViewPlan={resource.plan ? () => {} : undefined}
        />
      ))}
    </div>
  )
  return Causes
}

export const Causes = causeCard('list')
export const CausesGrid = causeCard('grid')

export const Grid = () => (
  <div className="grid md:grid-cols-2 gap-3">
    <DriftResourceCard {...links} view="grid" resource={unappliedResource()} />
    <DriftResourceCard
      {...links}
      view="grid"
      resource={openDeploymentResource({ id: 'ic-worker', name: 'worker' })}
    />
  </div>
)
