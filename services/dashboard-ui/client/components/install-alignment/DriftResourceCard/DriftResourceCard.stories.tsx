export default {
  title: 'Features / Installs / Alignment / Drift resource card',
}

import {
  causeResources,
  openDeploymentResource,
  stateDriftResource,
  unappliedResource,
} from '@/components/__fixtures__/install-alignment'
import { explanationEvidence } from '../alignment-utils'
import { DriftResourceCard } from './DriftResourceCard'

const links = {
  branchRunHref: '#branch-run',
  explanationHref: '#evidence',
  lastAttemptHref: '#last-attempt',
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
      last_attempt: undefined,
      plan: undefined,
      newer_branch_runs: null,
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
