import { useEffect } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CheckboxFilterDropdown } from '@/components/common/CheckboxFilterDropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Pagination, type IPagination } from '@/components/common/Pagination'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
import { SearchInput } from '@/components/common/SearchInput'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { usePagination } from '@/hooks/use-pagination'
import { useSurfaces } from '@/hooks/use-surfaces'
import { PaginationProvider } from '@/providers/pagination-provider'
import type {
  TAPIError,
  TInstallDeploymentRecord,
  TInstallDeploymentRecordType,
} from '@/types'
import {
  WORKFLOW_DATE_LABELS,
  WORKFLOW_STATUS_LABELS,
  workflowStatusOptions,
  type TWorkflowDatePreset,
  type TWorkflowStatusOption,
} from '@/utils/workflow-filters'
import { DeploymentDetailPanel } from './DeploymentDetailPanel'

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

const DEPLOYMENT_TYPE_ICON: Record<TInstallDeploymentRecordType, TIconVariant> =
  {
    provision: 'CardsIcon',
    reprovision: 'CardsIcon',
    sandbox_reprovision: 'ShippingContainerIcon',
    app_branch_update: 'GitBranchIcon',
    component_deploy: 'CardsIcon',
    image_update: 'PackageIcon',
    stack_update: 'StackIcon',
    install_config_update: 'FileCodeIcon',
  }

interface IDeploymentCard {
  deployment: TInstallDeploymentRecord
  orgId: string
  appId: string
  installId: string
  onViewDetails: () => void
}

const DeploymentCard = ({
  deployment,
  orgId,
  appId,
  installId,
  onViewDetails,
}: IDeploymentCard) => {
  const branchHref = deployment.app_branch
    ? `/${orgId}/apps/${appId}/branches/${deployment.app_branch.id}`
    : undefined
  const workflowHref = deployment.workflow
    ? `/${orgId}/installs/${installId}/history/${deployment.workflow.id}`
    : undefined

  const affectedResources = [
    ...(deployment.affected_resources.stack ? ['stack'] : []),
    ...(deployment.affected_resources.sandbox ? ['sandbox'] : []),
    ...deployment.affected_resources.components,
    ...deployment.affected_resources.images,
  ]

  return (
    <Card className="!p-4 !gap-3 !shadow-none">
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-start gap-3 min-w-0">
          <span className="mt-0.5 text-cool-grey-400 shrink-0">
            <Icon variant={DEPLOYMENT_TYPE_ICON[deployment.type]} size={16} />
          </span>
          <div className="flex flex-col gap-1 min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <Text weight="strong">{deployment.title}</Text>
              <Status status={deployment.status} variant="badge" />
              <Badge size="sm" theme="neutral">
                {DEPLOYMENT_TYPE_LABELS[deployment.type]}
              </Badge>
            </div>
            {deployment.summary && (
              <Text variant="subtext" theme="neutral">
                {deployment.summary}
              </Text>
            )}
          </div>
        </div>
        <div className="flex items-center gap-3 shrink-0">
          <Time
            time={deployment.created_at}
            format="relative"
            variant="subtext"
            theme="neutral"
          />
          <Button variant="secondary" size="sm" onClick={onViewDetails}>
            View details
          </Button>
        </div>
      </div>

      <div className="flex items-center gap-x-6 gap-y-2 flex-wrap">
        {deployment.app_branch && branchHref && (
          <span className="flex items-center gap-2">
            <Text as="span" variant="subtext" theme="neutral">
              {deployment.type === 'provision'
                ? 'Originally configured with'
                : 'App branch'}
            </Text>
            <Link href={branchHref} textVariant="subtext">
              {deployment.app_branch.name}
            </Link>
            {deployment.app_branch.sha && (
              <Badge size="sm" variant="code" theme="neutral">
                {deployment.app_branch.sha.slice(0, 8)}
              </Badge>
            )}
          </span>
        )}
        {deployment.workflow && workflowHref && (
          <span className="flex items-center gap-2">
            <Text as="span" variant="subtext" theme="neutral">
              Workflow
            </Text>
            <Link href={workflowHref} textVariant="subtext">
              {deployment.workflow.name}
            </Link>
          </span>
        )}
      </div>

      {deployment.image && (
        <div className="flex items-center gap-2 flex-wrap">
          <Text variant="subtext" family="mono">
            {deployment.image.repository}
          </Text>
          {deployment.image.previous_tag && (
            <>
              <Badge size="sm" variant="code" theme="neutral">
                {deployment.image.previous_tag}
              </Badge>
              <Icon
                variant="ArrowRightIcon"
                size={12}
                className="text-cool-grey-400"
              />
            </>
          )}
          <Badge size="sm" variant="code" theme="neutral">
            {deployment.image.next_tag}
          </Badge>
        </div>
      )}

      {affectedResources.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {deployment.affected_resources.stack && (
            <Badge size="sm" theme="neutral">
              Stack
            </Badge>
          )}
          {deployment.affected_resources.sandbox && (
            <Badge size="sm" theme="neutral">
              Sandbox
            </Badge>
          )}
          {deployment.affected_resources.components.map((c) => (
            <Badge key={c} size="sm" variant="code" theme="neutral">
              {c}
            </Badge>
          ))}
          {deployment.affected_resources.images.map((img) => (
            <Badge key={img} size="sm" variant="code" theme="neutral">
              {img}
            </Badge>
          ))}
        </div>
      )}
    </Card>
  )
}

const DeploymentCardSkeleton = () => (
  <Card className="!p-4 !gap-3 !shadow-none">
    <div className="flex items-start justify-between gap-4">
      <div className="flex items-start gap-3 min-w-0">
        <Text variant="subtext" loading loadingWidth={2} className="mt-0.5" />
        <div className="flex flex-col gap-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <Text loading loadingWidth={22} />
            <Status loading variant="badge" loadingWidth={8} />
            <Badge loading size="sm" loadingWidth={12} />
          </div>
          <Text variant="subtext" loading loadingWidth={44} />
        </div>
      </div>
      <div className="flex items-center gap-3 shrink-0">
        <Text variant="subtext" loading loadingWidth={10} />
        <Badge loading size="lg" loadingWidth={11} className="!rounded-lg" />
      </div>
    </div>

    <div className="flex items-center gap-x-6 gap-y-2 flex-wrap">
      <Text variant="subtext" loading loadingWidth={16} />
      <Text variant="subtext" loading loadingWidth={20} />
    </div>

    <div className="flex flex-wrap gap-2">
      <Badge loading size="sm" variant="code" loadingWidth={9} />
      <Badge loading size="sm" variant="code" loadingWidth={7} />
      <Badge loading size="sm" variant="code" loadingWidth={11} />
    </div>
  </Card>
)

export interface IDeploymentFilter {
  search: string
  status: Set<TWorkflowStatusOption>
  type: Set<TInstallDeploymentRecordType>
  resource?: string
  date?: TWorkflowDatePreset
}

export interface IDeploymentsListPresenter {
  deployments: TInstallDeploymentRecord[]
  isLoading: boolean
  error?: TAPIError | null
  pagination: Omit<IPagination, 'position'>
  orgId: string
  appId: string
  installId: string
  search: string
  filter: IDeploymentFilter
  onSearchChange: (value: string) => void
  onStatusChange: (value: Set<TWorkflowStatusOption>) => void
  onTypeChange: (value: Set<TInstallDeploymentRecordType>) => void
  onResourceChange: (value?: string) => void
  onDateChange: (value?: string) => void
  onClearFilters: () => void
}

const DeploymentsListBase = ({
  deployments,
  isLoading,
  error,
  pagination,
  orgId,
  appId,
  installId,
  search,
  filter,
  onSearchChange,
  onStatusChange,
  onTypeChange,
  onResourceChange,
  onDateChange,
  onClearFilters,
}: IDeploymentsListPresenter) => {
  const { addPanel } = useSurfaces()
  const { isPaginating, setIsPaginating } = usePagination()

  useEffect(() => {
    setIsPaginating(false)
  }, [deployments])

  const allComponents = Array.from(
    new Set(
      deployments.flatMap((d) => [
        ...(d.affected_resources.stack ? ['stack'] : []),
        ...(d.affected_resources.sandbox ? ['sandbox'] : []),
        ...d.affected_resources.components,
        ...d.affected_resources.images,
      ])
    )
  )

  const hasActiveFilters =
    filter.search !== '' ||
    filter.status.size > 0 ||
    filter.type.size > 0 ||
    !!filter.resource ||
    !!filter.date

  return (
    <div className="flex flex-col gap-4">
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
          onChange={(value) =>
            onStatusChange(value as Set<TWorkflowStatusOption>)
          }
        />
        <CheckboxFilterDropdown
          id="deployments-filter-type"
          label="Type"
          options={(
            Object.keys(
              DEPLOYMENT_TYPE_LABELS
            ) as TInstallDeploymentRecordType[]
          ).map((value) => ({
            value,
            label: DEPLOYMENT_TYPE_LABELS[value],
          }))}
          selected={filter.type}
          onChange={(value) =>
            onTypeChange(value as Set<TInstallDeploymentRecordType>)
          }
        />
        {allComponents.length > 0 && (
          <RadioFilterDropdown
            id="deployments-filter-resource"
            label="Resource"
            options={allComponents.map((value) => ({ value, label: value }))}
            selected={filter.resource}
            onChange={onResourceChange}
          />
        )}
        <RadioFilterDropdown
          id="deployments-filter-date"
          label="Date"
          options={(
            Object.keys(WORKFLOW_DATE_LABELS) as TWorkflowDatePreset[]
          ).map((value) => ({
            value,
            label: WORKFLOW_DATE_LABELS[value],
          }))}
          selected={filter.date}
          onChange={onDateChange}
        />
        {hasActiveFilters && (
          <Button variant="ghost" onClick={onClearFilters}>
            Clear filters
          </Button>
        )}
      </div>

      {isPaginating || (isLoading && deployments.length === 0) ? (
        <div className="flex flex-col gap-3">
          {Array.from({ length: pagination.limit ?? 5 }).map((_, index) => (
            <DeploymentCardSkeleton key={index} />
          ))}
        </div>
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
        <div className="flex flex-col gap-3">
          {deployments.map((deployment) => (
            <DeploymentCard
              key={deployment.id}
              deployment={deployment}
              orgId={orgId}
              appId={appId}
              installId={installId}
              onViewDetails={() =>
                addPanel(
                  <DeploymentDetailPanel
                    deployment={deployment}
                    orgId={orgId}
                    appId={appId}
                    installId={installId}
                  />
                )
              }
            />
          ))}
        </div>
      )}

      {pagination.hasNext || pagination.offset !== 0 ? (
        <Pagination {...pagination} />
      ) : null}
    </div>
  )
}

export const DeploymentsListPresenter = (props: IDeploymentsListPresenter) => (
  <PaginationProvider>
    <DeploymentsListBase {...props} />
  </PaginationProvider>
)
