import { Button } from '@/components/common/Button'
import { CheckboxFilterDropdown } from '@/components/common/CheckboxFilterDropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { CommitLink } from '@/components/common/GitReferenceLink'
import { Link } from '@/components/common/Link'
import type { IPagination } from '@/components/common/Pagination'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import {
  deploymentOutcomes,
  deploymentSteps,
  recoveredDeploymentResources,
  type TDeploymentOutcome,
} from '@/components/installs/DeploymentDetail/deployment-progress'
import { useSurfaces } from '@/hooks/use-surfaces'
import type {
  TAPIError,
  TInstallDeploymentRecord,
  TInstallDeploymentRecordType,
  TWorkflow,
} from '@/types'
import {
  WORKFLOW_DATE_LABELS,
  WORKFLOW_STATUS_LABELS,
  workflowStatusOptions,
  type TWorkflowDatePreset,
  type TWorkflowStatusOption,
} from '@/utils/workflow-filters'
import { DeploymentDetailPanel } from './DeploymentDetailPanel'
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

interface IDeploymentRecordRow {
  deployment: TInstallDeploymentRecord
  orgId: string
  appId: string
  installId: string
  repo?: string
  workflow?: TWorkflow
  previousOutcomes: TDeploymentOutcome[][]
}

const DeploymentRecordRow = ({
  deployment,
  orgId,
  appId,
  installId,
  repo,
  workflow,
  previousOutcomes,
}: IDeploymentRecordRow) => {
  const { addPanel } = useSurfaces()
  const branchHref = deployment.app_branch
    ? `/${orgId}/apps/${appId}/branches/${deployment.app_branch.id}`
    : undefined
  const run = {
    status: workflow?.status?.status ?? deployment.status,
    activity: workflow?.status?.status_human_description ?? deployment.summary,
    steps: deploymentSteps(workflow),
    outcomes: deploymentOutcomes(deployment, workflow),
  }
  const recovered =
    run.status === 'success'
      ? recoveredDeploymentResources(run.outcomes, previousOutcomes)
      : []
  return (
    <DeploymentRow
      run={run}
      title={deployment.title}
      createdAt={deployment.created_at}
      onViewDetails={() =>
        addPanel(
          <DeploymentDetailPanel
            deployment={deployment}
            orgId={orgId}
            appId={appId}
            installId={installId}
            repo={repo}
          />
        )
      }
    >
      {recovered.map((category) => (
        <Text key={category} variant="subtext" theme="neutral">
          Previous {category.toLowerCase()} update failed
        </Text>
      ))}
      {deployment.app_branch && branchHref ? (
        <span className="flex flex-wrap items-center gap-2">
          <Text as="span" variant="subtext" theme="neutral">
            {deployment.type === 'provision'
              ? 'Originally configured with'
              : 'App branch'}
          </Text>
          <Link href={branchHref} textVariant="subtext">
            {deployment.app_branch.name}
          </Link>
          {deployment.app_branch.sha ? (
            <CommitLink sha={deployment.app_branch.sha} repo={repo} />
          ) : null}
        </span>
      ) : null}
      {deployment.image ? (
        <Text variant="subtext" family="mono">
          {deployment.image.repository}:{' '}
          {deployment.image.previous_tag
            ? `${deployment.image.previous_tag} → `
            : ''}
          {deployment.image.next_tag}
        </Text>
      ) : null}
      {!workflow && deployment.workflow ? (
        <Text variant="subtext" theme="neutral">
          Resource outcomes unavailable
        </Text>
      ) : null}
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
  deployments: TInstallDeploymentRecord[]
  workflowsById?: Record<string, TWorkflow>
  isLoading: boolean
  error?: TAPIError | null
  pagination: Omit<IPagination, 'position'>
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
      options={(
        Object.keys(DEPLOYMENT_TYPE_LABELS) as TInstallDeploymentRecordType[]
      ).map((value) => ({ value, label: DEPLOYMENT_TYPE_LABELS[value] }))}
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
  workflowsById = {},
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
  const allResources = [
    ...new Set(
      deployments.flatMap((deployment) => [
        ...(deployment.affected_resources.stack ? ['stack'] : []),
        ...(deployment.affected_resources.sandbox ? ['sandbox'] : []),
        ...(deployment.affected_resources.components ?? []),
        ...(deployment.affected_resources.images ?? []),
      ])
    ),
  ]
  const hasActiveFilters = hasDeploymentFilters(filter)
  return (
    <div className="flex flex-col gap-4">
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
      {isLoading && deployments.length === 0 ? (
        <TimelineSkeleton eventCount={pagination.limit ?? 5} />
      ) : error ? (
        <EmptyState
          emptyTitle="Deployments failed to load"
          emptyMessage="Unable to load deployments. Try refreshing the page."
          className="my-12"
        />
      ) : deployments.length === 0 ? (
        <EmptyState
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
          pagination={pagination}
          getEventKey={(deployment) => deployment.id}
          renderEvent={(deployment, index) => (
            <DeploymentRecordRow
              deployment={deployment}
              workflow={
                deployment.workflow
                  ? workflowsById[deployment.workflow.id]
                  : undefined
              }
              orgId={orgId}
              appId={appId}
              installId={installId}
              repo={repo}
              previousOutcomes={
                deployments
                  .slice(index + 1)
                  .every(
                    (previous) =>
                      !previous.workflow || workflowsById[previous.workflow.id]
                  )
                  ? deployments
                      .slice(index + 1)
                      .map((previous) =>
                        deploymentOutcomes(
                          previous,
                          previous.workflow
                            ? workflowsById[previous.workflow.id]
                            : undefined
                        )
                      )
                  : []
              }
            />
          )}
        />
      )}
    </div>
  )
}
