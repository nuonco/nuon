export default {
  title: 'Features / Installs / Alignment / Drift resource card',
}

import {
  causeResources,
  inputUpdateResource,
  openDeploymentResource,
  stateDriftResource,
  unappliedResource,
} from '@/components/__fixtures__/install-alignment'
import { explanationEvidence } from '../alignment-utils'
import { DriftResourceCard } from './DriftResourceCard'

const links = {
  buildHref: (buildId: string) => `#build-${buildId}`,
  explanationHref: '#evidence',
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

export const PolicyDenied = () => (
  <DriftResourceCard
    {...links}
    onViewPlan={undefined}
    resource={unappliedResource({
      blocking_attempt: {
        id: 'dep-ic-api',
        status: 'policy-failed',
        created_at: new Date(Date.now() - 2 * 3_600_000).toISOString(),
        workflow_id: 'wf-attempt-ic-api',
        build_id: 'bld-selected',
      },
      earlier_attempts: 1,
      plan: undefined,
      explanation: {
        summary: 'A policy denied the latest deployment.',
        status: 'policy-failed',
        workflow_id: 'wf-attempt-ic-api',
      },
    })}
  />
)

export const InputsChanged = () => (
  <DriftResourceCard {...links} resource={inputUpdateResource()} />
)

export const OpenDeployment = () => (
  <DriftResourceCard {...links} resource={openDeploymentResource()} />
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

export const NeverAttempted = () => (
  <DriftResourceCard
    {...links}
    explanationHref={undefined}
    onViewPlan={undefined}
    resource={unappliedResource({
      blocking_attempt: undefined,
      earlier_attempts: 0,
      plan: undefined,
      explanation: { summary: 'No deployment has started for these changes.' },
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
          explanationHref={
            explanationEvidence(resource) ? '#evidence' : undefined
          }
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
