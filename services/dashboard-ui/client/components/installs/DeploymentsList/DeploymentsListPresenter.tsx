import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Dropdown } from '@/components/common/Dropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Menu } from '@/components/common/Menu'
import { SearchInput } from '@/components/common/SearchInput'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useSurfaces } from '@/hooks/use-surfaces'
import type {
  TInstallDeploymentRecord,
  TInstallDeploymentRecordType,
} from '@/types'
import { humanize } from '@/utils/string-utils'
import { DeploymentDetailPanel } from './DeploymentDetailPanel'

// ─── Type labels ─────────────────────────────────────────────────────────────

const DEPLOYMENT_TYPE_LABELS: Record<TInstallDeploymentRecordType, string> = {
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

const DATE_FILTER_LABELS: Record<string, string> = {
  '24h': 'Last 24 hours',
  '7d': 'Last 7 days',
  '30d': 'Last 30 days',
}

// ─── Individual card ─────────────────────────────────────────────────────────

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
              <Button
                variant="ghost"
                size="sm"
                className="!px-0 font-medium"
                onClick={onViewDetails}
              >
                {deployment.title}
              </Button>
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

// ─── Filter bar ───────────────────────────────────────────────────────────────

export interface IDeploymentFilter {
  search: string
  status: string
  type: string
  component: string
  date: string
}

export const DEFAULT_DEPLOYMENTS_FILTER: IDeploymentFilter = {
  search: '',
  status: 'all',
  type: 'all',
  component: 'all',
  date: 'all',
}

// ─── Presenter ───────────────────────────────────────────────────────────────

export interface IDeploymentsListPresenter {
  deployments: TInstallDeploymentRecord[]
  isLoading: boolean
  error?: Error | null
  page: number
  hasMore: boolean
  orgId: string
  appId: string
  installId: string
  filter: IDeploymentFilter
  onFilterChange: (patch: Partial<IDeploymentFilter>) => void
  onClearFilters: () => void
  onPageChange: (page: number) => void
}

export const DeploymentsListPresenter = ({
  deployments,
  isLoading,
  error,
  page,
  hasMore,
  orgId,
  appId,
  installId,
  filter,
  onFilterChange,
  onClearFilters,
  onPageChange,
}: IDeploymentsListPresenter) => {
  const { addPanel } = useSurfaces()

  const allStatuses = Array.from(new Set(deployments.map((d) => d.status)))

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
    filter.status !== 'all' ||
    filter.type !== 'all' ||
    filter.component !== 'all' ||
    filter.date !== 'all'

  return (
    <div className="flex flex-col gap-0">
      {/* Filter bar */}
      <div className="flex items-center flex-wrap gap-2 py-3 shrink-0">
        <SearchInput
          aria-label="Search deployments"
          value={filter.search}
          onChange={(v) => onFilterChange({ search: v })}
          placeholder="Search deployments…"
          labelClassName="flex-1 min-w-44"
        />

        <Dropdown
          id="deployments-filter-status"
          variant="secondary"
          size="sm"
          buttonText={
            filter.status === 'all' ? 'Status' : humanize(filter.status)
          }
          isActive={filter.status !== 'all'}
        >
          <Menu>
            <Button
              isMenuButton
              onClick={() => onFilterChange({ status: 'all' })}
            >
              All statuses
            </Button>
            {allStatuses.map((s) => (
              <Button
                isMenuButton
                key={s}
                onClick={() => onFilterChange({ status: s })}
              >
                {humanize(s)}
              </Button>
            ))}
          </Menu>
        </Dropdown>

        <Dropdown
          id="deployments-filter-type"
          variant="secondary"
          size="sm"
          buttonText={
            filter.type === 'all'
              ? 'Type'
              : (DEPLOYMENT_TYPE_LABELS[
                  filter.type as TInstallDeploymentRecordType
                ] ?? humanize(filter.type))
          }
          isActive={filter.type !== 'all'}
        >
          <Menu>
            <Button
              isMenuButton
              onClick={() => onFilterChange({ type: 'all' })}
            >
              All types
            </Button>
            {(
              Object.keys(
                DEPLOYMENT_TYPE_LABELS
              ) as TInstallDeploymentRecordType[]
            ).map((type) => (
              <Button
                isMenuButton
                key={type}
                onClick={() => onFilterChange({ type })}
              >
                {DEPLOYMENT_TYPE_LABELS[type]}
              </Button>
            ))}
          </Menu>
        </Dropdown>

        {allComponents.length > 0 && (
          <Dropdown
            id="deployments-filter-component"
            variant="secondary"
            size="sm"
            buttonText={
              filter.component === 'all' ? 'Resource' : filter.component
            }
            isActive={filter.component !== 'all'}
          >
            <Menu>
              <Button
                isMenuButton
                onClick={() => onFilterChange({ component: 'all' })}
              >
                All resources
              </Button>
              {allComponents.map((c) => (
                <Button
                  isMenuButton
                  key={c}
                  onClick={() => onFilterChange({ component: c })}
                >
                  {c}
                </Button>
              ))}
            </Menu>
          </Dropdown>
        )}

        <Dropdown
          id="deployments-filter-date"
          variant="secondary"
          size="sm"
          buttonText={
            filter.date === 'all'
              ? 'Date'
              : (DATE_FILTER_LABELS[filter.date] ?? filter.date)
          }
          isActive={filter.date !== 'all'}
        >
          <Menu>
            <Button
              isMenuButton
              onClick={() => onFilterChange({ date: 'all' })}
            >
              All time
            </Button>
            {Object.entries(DATE_FILTER_LABELS).map(([key, label]) => (
              <Button
                isMenuButton
                key={key}
                onClick={() => onFilterChange({ date: key })}
              >
                {label}
              </Button>
            ))}
          </Menu>
        </Dropdown>

        {hasActiveFilters && (
          <Button variant="ghost" size="sm" onClick={onClearFilters}>
            Clear filters
          </Button>
        )}
      </div>

      {/* Body */}
      {isLoading && deployments.length === 0 ? (
        <Loading variant="large" className="my-12" />
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

          {/* Pagination */}
          <div className="flex items-center justify-between gap-4 pt-2">
            <Button
              variant="secondary"
              size="sm"
              disabled={page === 0}
              onClick={() => onPageChange(page - 1)}
            >
              Previous
            </Button>
            <Text variant="subtext" theme="neutral">
              Page {page + 1}
            </Text>
            <Button
              variant="secondary"
              size="sm"
              disabled={!hasMore}
              onClick={() => onPageChange(page + 1)}
            >
              Next
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
