export default {
  title: 'Features / Installs / Alignment / Drifted resources',
}

import { useState } from 'react'
import {
  CollectionViewToggle,
  type TCollectionView,
} from '@/components/common/CollectionViewToggle'
import { SearchInput } from '@/components/common/SearchInput'
import {
  alignmentResource,
  causeResources,
  inputUpdateResource,
  inSyncResources,
  mixedAlignmentResources,
} from '@/components/__fixtures__/install-alignment'
import type { TInstallAlignmentResource } from '@/types'
import { explanationEvidence, hasOngoingDeployment } from '../alignment-utils'
import { DriftResourceCard } from '../DriftResourceCard'
import { DriftedResources, type IDriftedResources } from './DriftedResources'

const card =
  (view: TCollectionView) => (resource: TInstallAlignmentResource) => (
    <DriftResourceCard
      resource={resource}
      view={view}
      explanationHref={explanationEvidence(resource) ? '#evidence' : undefined}
      branchRunHref="#branch-run"
      lastAttemptHref="#last-attempt"
      deploymentHref="#deployment"
      onViewPlan={resource.plan ? () => {} : undefined}
    />
  )

const Interactive = ({
  initialView = 'grid',
  resources = [...mixedAlignmentResources(12), ...inSyncResources()],
  ...props
}: Partial<IDriftedResources> & { initialView?: TCollectionView }) => {
  const [view, setView] = useState<TCollectionView>(initialView)
  const [search, setSearch] = useState('')
  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <SearchInput
          aria-label="Search components"
          placeholder="Search by name or ID..."
          value={search}
          onChange={setSearch}
        />
        <CollectionViewToggle value={view} onChange={setView} />
      </div>
      <DriftedResources
        noun="components"
        resources={resources}
        search={search}
        view={view}
        renderCard={card(view)}
        {...props}
      />
    </div>
  )
}

const causes = causeResources()

export const Grid = () => (
  <Interactive
    resources={[...causes, ...inSyncResources()]}
    pageSize={causes.length}
  />
)
export const List = () => (
  <Interactive
    initialView="list"
    resources={[...causes, ...inSyncResources()]}
    pageSize={causes.length}
  />
)
export const ManyDrifted = () => (
  <Interactive resources={mixedAlignmentResources(25)} />
)
export const Collapsed = () => <Interactive initiallyExpanded={false} />
export const ChangesNotApplied = () => (
  <Interactive initialDriftFilter="changes-not-applied" />
)
export const OpenDeployment = () => (
  <Interactive resources={causes.filter(hasOngoingDeployment)} />
)
export const NoOpenDeployment = () => (
  <Interactive
    initialDriftFilter="changes-not-applied"
    resources={causes.filter((resource) => !hasOngoingDeployment(resource))}
  />
)
export const StateDriftDetected = () => (
  <Interactive initialDriftFilter="state-drift-detected" />
)
export const InputUpdates = () => (
  <Interactive
    initialView="list"
    initialDriftFilter="changes-not-applied"
    resources={[
      inputUpdateResource({
        id: 'ic-search',
        name: 'search',
        component_id: 'cmp-search',
        status: 'inactive',
        reason: 'This component is enabled, but hasn’t been deployed.',
        applied_build_id: undefined,
        applied_config_id: undefined,
      }),
      inputUpdateResource({
        id: 'ic-webhooks',
        name: 'webhooks',
        component_id: 'cmp-webhooks',
      }),
      inputUpdateResource({
        id: 'ic-worker',
        name: 'worker',
        component_id: 'cmp-worker',
        open_deployment: {
          id: 'wf-open-input-worker',
          type: 'input_update',
          status: 'approval-awaiting',
          created_at: new Date(Date.now() - 1_800_000).toISOString(),
        },
        explanation: {
          summary: 'These changes are waiting for approval.',
          status: 'approval-awaiting',
          workflow_id: 'wf-open-input-worker',
        },
      }),
    ]}
  />
)
export const AllInSync = () => <Interactive resources={inSyncResources(4)} />
export const RefreshFailed = () => (
  <Interactive
    resources={inSyncResources(4)}
    error={{
      error: 'Internal Server Error',
      description: 'Overview unavailable',
      user_error: false,
    }}
  />
)
export const SomeNotCompared = () => (
  <Interactive
    resources={[
      ...inSyncResources(2),
      alignmentResource({
        id: 'ic-unknown',
        name: 'reports',
        category: 'unknown',
      }),
      alignmentResource({
        id: 'ic-off',
        name: 'exporter',
        category: 'disabled',
      }),
    ]}
  />
)
export const Loading = () => <Interactive resources={[]} loading />
export const LoadFailed = () => (
  <Interactive
    resources={[]}
    error={{
      error: 'Internal Server Error',
      description: 'unable to get install overview',
      user_error: false,
    }}
    onRetry={() => {}}
  />
)
