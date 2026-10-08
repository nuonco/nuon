import { useMemo } from 'react'
import { Button } from '@/components/common/Button'
import { CheckboxFilterDropdown } from '@/components/common/CheckboxFilterDropdown'
import {
  COLLECTION_VIEW_MODES,
  CollectionViewToggle,
  type TCollectionView,
} from '@/components/common/CollectionViewToggle'
import { EmptyState } from '@/components/common/EmptyState'
import type { IPagination } from '@/components/common/Pagination'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TDeploymentRun } from '@/components/installs/DeploymentDetail/DeploymentProgress'
import {
  deploymentOutcomes,
  deploymentSteps,
  recoveredDeploymentResources,
  stepAffectedResources,
  summaryDeploymentEvidence,
  type TDeploymentOutcome,
} from '@/components/installs/DeploymentDetail/deployment-progress'
import { useStoredViewMode } from '@/hooks/use-stored-view-mode'
import { useSurfaces } from '@/hooks/use-surfaces'
import type {
  TAPIError,
  TInstallDeploymentRecordType,
  TInstallDeploymentSummary,
} from '@/types'
import { cn } from '@/utils/classnames'
import {
  WORKFLOW_DATE_LABELS,
  WORKFLOW_STATUS_LABELS,
  workflowStatusOptions,
  type TWorkflowDatePreset,
  type TWorkflowStatusOption,
} from '@/utils/workflow-filters'
import { DeploymentCard } from './DeploymentCard'
import { DeploymentDetailPanel } from './DeploymentDetailPanel'
import { DeploymentPolicySummary } from './DeploymentPolicySummary'
import { DeploymentRow } from './DeploymentRow'

export const DEPLOYMENT_TYPE_LABELS: Record<
  TInstallDeploymentRecordType,
  string
> = {
  provision: 'Provision',
  reprovision: 'Reprovision',
  sandbox_reprovision: 'Sandbox reprovision',
  app_branch_update: 'App branch update',
  component_deploy: 'Component deploy',
  image_update: 'Image update',
  stack_update: 'Stack update',
  install_config_update: 'Install config update',
}

export const DEPLOYMENT_FILTER_TYPES = (
  Object.keys(DEPLOYMENT_TYPE_LABELS) as TInstallDeploymentRecordType[]
).filter((type) => type !== 'image_update')

const deploymentRun = (deployment: TInstallDeploymentSummary) => {
  const evidence = summaryDeploymentEvidence(deployment)
  const resources = stepAffectedResources(evidence.steps)
  return {
    resources,
    run: {
      status: deployment.status,
      activity: deployment.activity ?? '',
      steps: deploymentSteps(evidence),
      outcomes: deploymentOutcomes({ affected_resources: resources }, evidence),
    },
  }
}

interface IDeploymentRecordRow {
  deployment: TInstallDeploymentSummary
  run: TDeploymentRun
  orgId: string
  appId: string
  installId: string
  repo?: string
  previousOutcomes: TDeploymentOutcome[][]
  active?: boolean
  view?: TCollectionView
}

const DeploymentRecordRow = ({
  deployment,
  run,
  orgId,
  appId,
  installId,
  repo,
  previousOutcomes,
  active,
  view = 'grid',
}: IDeploymentRecordRow) => {
  const { addPanel } = useSurfaces()
  const recovered =
    run.status === 'success'
      ? recoveredDeploymentResources(run.outcomes, previousOutcomes)
      : []
  const onViewDetails = () =>
    addPanel(
      <DeploymentDetailPanel
        deployment={deployment}
        orgId={orgId}
        appId={appId}
        installId={installId}
        repo={repo}
      />
    )
  if (active && view === 'grid') {
    return (
      <DeploymentCard
        run={run}
        title={deployment.title}
        typeLabel={DEPLOYMENT_TYPE_LABELS[deployment.type]}
        createdAt={deployment.created_at}
        onViewDetails={onViewDetails}
      >
        <DeploymentPolicySummary steps={run.steps} />
      </DeploymentCard>
    )
  }
  return (
    <DeploymentRow
      run={run}
      title={deployment.title}
      createdAt={deployment.created_at}
      onViewDetails={onViewDetails}
      history={!active}
    >
      <DeploymentPolicySummary steps={run.steps} history={!active} />
      {recovered.map((category) => (
        <Text key={category} variant="subtext" theme="neutral">
          Previous {category.toLowerCase()} update failed
        </Text>
      ))}
    </DeploymentRow>
  )
}

export interface IDeploymentFilter {
  search: string
  status: Set<TWorkflowStatusOption>
  type: Set<TInstallDeploymentRecordType>
  resource?: string
  date?: TWorkflowDatePreset
}

export interface IDeploymentsListPresenter {
  deployments: TInstallDeploymentSummary[]
  activeDeployments?: TInstallDeploymentSummary[]
  activeTotal?: number
  activeLoading?: boolean
  activeError?: TAPIError | null
  hasMoreActive?: boolean
  showActive?: boolean
  showHistory?: boolean
  onLoadMoreActive?: () => void
  isLoading: boolean
  error?: TAPIError | null
  pagination: Omit<IPagination, 'position'> & {
    onNext?: () => void
    onPrevious?: () => void
  }
  orgId: string
  appId: string
  installId: string
  repo?: string
  search: string
  filter: IDeploymentFilter
  onSearchChange: (value: string) => void
  onStatusChange: (value: Set<TWorkflowStatusOption>) => void
  onTypeChange: (value: Set<TInstallDeploymentRecordType>) => void
  onResourceChange: (value?: string) => void
  onDateChange: (value?: string) => void
  onClearFilters: () => void
}

const hasDeploymentFilters = (filter: IDeploymentFilter) =>
  filter.search !== '' ||
  filter.status.size > 0 ||
  filter.type.size > 0 ||
  !!filter.resource ||
  !!filter.date

type TDeploymentsListFilters = Pick<
  IDeploymentsListPresenter,
  | 'search'
  | 'filter'
  | 'onSearchChange'
  | 'onStatusChange'
  | 'onTypeChange'
  | 'onResourceChange'
  | 'onDateChange'
  | 'onClearFilters'
> & { resources: string[] }

export const DeploymentsListFilters = ({
  resources,
  search,
  filter,
  onSearchChange,
  onStatusChange,
  onTypeChange,
  onResourceChange,
  onDateChange,
  onClearFilters,
}: TDeploymentsListFilters) => (
  <div className="flex min-w-0 flex-wrap items-center gap-2">
    <SearchInput
      aria-label="Search deployments"
      placeholder="Search deployments"
      value={search}
      onChange={onSearchChange}
      onClear={() => onSearchChange('')}
    />
    <CheckboxFilterDropdown
      id="deployments-filter-status"
      label="Status"
      options={workflowStatusOptions().map((value) => ({
        value,
        label: WORKFLOW_STATUS_LABELS[value],
      }))}
      selected={filter.status}
      onChange={(value) => onStatusChange(value as Set<TWorkflowStatusOption>)}
    />
    <CheckboxFilterDropdown
      id="deployments-filter-type"
      label="Type"
      options={DEPLOYMENT_FILTER_TYPES.map((value) => ({
        value,
        label: DEPLOYMENT_TYPE_LABELS[value],
      }))}
      selected={filter.type}
      onChange={(value) =>
        onTypeChange(value as Set<TInstallDeploymentRecordType>)
      }
    />
    {resources.length > 0 ? (
      <RadioFilterDropdown
        id="deployments-filter-resource"
        label="Resource"
        options={resources.map((value) => ({ value, label: value }))}
        selected={filter.resource}
        onChange={onResourceChange}
      />
    ) : null}
    <RadioFilterDropdown
      id="deployments-filter-date"
      label="Date"
      options={(Object.keys(WORKFLOW_DATE_LABELS) as TWorkflowDatePreset[]).map(
        (value) => ({ value, label: WORKFLOW_DATE_LABELS[value] })
      )}
      selected={filter.date}
      onChange={onDateChange}
    />
    {hasDeploymentFilters(filter) ? (
      <Button variant="ghost" onClick={onClearFilters}>
        Clear filters
      </Button>
    ) : null}
  </div>
)

export const DeploymentsListPresenter = ({
  deployments,
  activeDeployments = [],
  activeTotal,
  activeLoading = false,
  activeError,
  hasMoreActive = false,
  showActive = true,
  showHistory = true,
  onLoadMoreActive,
  isLoading,
  error,
  pagination,
  orgId,
  appId,
  installId,
  repo,
  search,
  filter,
  onSearchChange,
  onStatusChange,
  onTypeChange,
  onResourceChange,
  onDateChange,
  onClearFilters,
}: IDeploymentsListPresenter) => {
  const [view, setView] = useStoredViewMode<TCollectionView>(
    'nuon:deployments-view',
    COLLECTION_VIEW_MODES,
    'grid'
  )
  const runs = useMemo(() => deployments.map(deploymentRun), [deployments])
  const activeRuns = useMemo(
    () => activeDeployments.map(deploymentRun),
    [activeDeployments]
  )
  const allResources = [
    ...new Set([
      'stack',
      'sandbox',
      ...(filter.resource ? [filter.resource] : []),
      ...[...activeRuns, ...runs].flatMap(
        ({ resources }) => resources.components
      ),
    ]),
  ]
  const hasActiveFilters = hasDeploymentFilters(filter)
  return (
    <div className="@container flex flex-col gap-6">
      <DeploymentsListFilters
        resources={allResources}
        search={search}
        filter={filter}
        onSearchChange={onSearchChange}
        onStatusChange={onStatusChange}
        onTypeChange={onTypeChange}
        onResourceChange={onResourceChange}
        onDateChange={onDateChange}
        onClearFilters={onClearFilters}
      />
      {showActive ? (
        <section aria-label="In progress" className="flex flex-col gap-3">
          <SectionHeader
            title={`In progress${activeTotal !== undefined ? ` (${activeTotal})` : ''}`}
            description="Follow active deployments and rollouts."
            actions={<CollectionViewToggle value={view} onChange={setView} />}
          />
          {activeLoading && activeDeployments.length === 0 ? (
            <TimelineSkeleton eventCount={4} />
          ) : activeError ? (
            <EmptyState
              variant="history"
              emptyTitle="Active deployments failed to load"
              emptyMessage="Unable to load active deployments. Try refreshing the page."
            />
          ) : activeDeployments.length === 0 ? (
            <Text variant="subtext" theme="neutral">
              {hasActiveFilters
                ? 'No in-progress deployments match these filters. Adjust or clear the filters to see more.'
                : 'No deployments in progress. Start a deployment to follow its rollout here.'}
            </Text>
          ) : (
            <>
              <div
                className={cn(
                  'grid grid-cols-1',
                  view === 'grid' ? 'gap-3' : 'pl-4',
                  view === 'grid' &&
                    activeDeployments.length > 1 &&
                    '@4xl:grid-cols-2'
                )}
              >
                {activeDeployments.map((deployment, index) => (
                  <DeploymentRecordRow
                    key={deployment.id}
                    deployment={deployment}
                    run={activeRuns[index].run}
                    orgId={orgId}
                    appId={appId}
                    installId={installId}
                    repo={repo}
                    previousOutcomes={[]}
                    active
                    view={view}
                  />
                ))}
              </div>
              {hasMoreActive ? (
                <Button
                  className="w-full justify-center"
                  disabled={activeLoading}
                  onClick={onLoadMoreActive}
                >
                  {activeLoading
                    ? 'Loading deployments...'
                    : `Show ${Math.max(0, (activeTotal ?? activeDeployments.length + 4) - activeDeployments.length)} more in progress`}
                </Button>
              ) : null}
            </>
          )}
        </section>
      ) : null}
      {showHistory ? (
        <section aria-label="History" className="flex flex-col gap-3">
          <SectionHeader
            title="History"
            description="Review completed deployments and rollouts."
          />
          {isLoading && deployments.length === 0 ? (
            <TimelineSkeleton eventCount={pagination.limit ?? 5} />
          ) : error ? (
            <EmptyState
              variant="history"
              emptyTitle="Deployments failed to load"
              emptyMessage="Unable to load deployments. Try refreshing the page."
              className="my-12"
            />
          ) : deployments.length === 0 ? (
            <EmptyState
              variant={hasActiveFilters ? 'search' : 'history'}
              emptyTitle={
                hasActiveFilters ? 'No deployments found' : 'No deployments yet'
              }
              emptyMessage={
                hasActiveFilters
                  ? 'No deployments match the current filters. Try adjusting or clearing them.'
                  : 'Deployments will appear here once a workflow deploys components to this install.'
              }
              className="my-12"
            />
          ) : (
            <Timeline
              className="w-full"
              events={deployments}
              groupByDate={false}
              pagination={{ hasNext: false, offset: 0 }}
              getEventKey={(deployment) => deployment.id}
              renderEvent={(deployment, index) => (
                <DeploymentRecordRow
                  deployment={deployment}
                  run={runs[index].run}
                  orgId={orgId}
                  appId={appId}
                  installId={installId}
                  repo={repo}
                  previousOutcomes={runs
                    .slice(index + 1)
                    .map((previous) => previous.run.outcomes)}
                />
              )}
            />
          )}
          {pagination.hasNext || (pagination.offset ?? 0) > 0 ? (
            <div className="flex items-center justify-end gap-3">
              <Button
                disabled={isLoading || !pagination.offset}
                onClick={pagination.onPrevious}
              >
                Previous
              </Button>
              <Text variant="subtext" theme="neutral">
                Page{' '}
                {Math.floor(
                  (pagination.offset ?? 0) / (pagination.limit ?? 20)
                ) + 1}
              </Text>
              <Button
                disabled={isLoading || !pagination.hasNext}
                onClick={pagination.onNext}
              >
                Next
              </Button>
            </div>
          ) : null}
        </section>
      ) : null}
    </div>
  )
}
