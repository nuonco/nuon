import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import {
  CheckboxFilterDropdown,
  type IFilterDropdownOption,
} from '@/components/common/CheckboxFilterDropdown'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { CloudRegion } from '@/components/common/CloudRegion'
import { ContextTooltip } from '@/components/common/ContextTooltip'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { InstallStatuses } from '@/components/installs/InstallStatuses'
import { InstallStatusAxisCell } from '@/components/installs/InstallStatusSummary'
import { QuickManagementDropdown } from '@/components/installs/management/QuickManagementDropdown'
import { LabelBadge } from '@/components/common/LabelBadge'
import { installAppBranchGroup } from '@/components/branches/active-app-branch-connection'
import type { TCloudPlatform, TInstall } from '@/types'
import { humanize } from '@/utils/string-utils'

export type TInstallsTableScope = 'org' | 'app' | 'branch'

export const INSTALL_HEALTH_FILTER_VALUES = [
  'unhealthy',
  'degraded',
  'other',
  'unknown',
  'progressing',
  'healthy',
  'unset',
] as const

export const INSTALL_LIFECYCLE_PHASE_FILTER_VALUES = [
  'provisioning',
  'provisioned',
  'deprovisioning',
  'deprovisioned',
  'reprovisioning',
  'unset',
  'other',
] as const

const HEALTH_FILTER_OPTIONS: IFilterDropdownOption[] =
  INSTALL_HEALTH_FILTER_VALUES.map((value) => ({
    value,
    label: value === 'unset' ? 'No recorded verdict' : humanize(value),
  }))

const LIFECYCLE_PHASE_FILTER_OPTIONS: IFilterDropdownOption[] =
  INSTALL_LIFECYCLE_PHASE_FILTER_VALUES.map((value) => ({
    value,
    label: value === 'unset' ? 'Not set' : humanize(value),
  }))

interface IInstallHealthFilters {
  health: Set<string>
  lifecyclePhase: Set<string>
  onHealthChange: (selected: Set<string>) => void
  onLifecyclePhaseChange: (selected: Set<string>) => void
}

export const InstallHealthFilters = ({
  health,
  lifecyclePhase,
  onHealthChange,
  onLifecyclePhaseChange,
}: IInstallHealthFilters) => (
  <>
    <CheckboxFilterDropdown
      id="installs-filter-health"
      label="Health"
      options={HEALTH_FILTER_OPTIONS}
      selected={health}
      onChange={onHealthChange}
    />
    <CheckboxFilterDropdown
      id="installs-filter-lifecycle-phase"
      label="Lifecycle"
      options={LIFECYCLE_PHASE_FILTER_OPTIONS}
      selected={lifecyclePhase}
      onChange={onLifecyclePhaseChange}
    />
  </>
)

const COVERAGE_VERDICTS = [
  'unhealthy',
  'degraded',
  'other',
  'unknown',
  'progressing',
  'healthy',
] as const

type TCoverageBucket =
  | (typeof COVERAGE_VERDICTS)[number]
  | 'not-applicable'
  | 'unset'

const COVERAGE_BUCKET_LABELS: Record<TCoverageBucket, string> = {
  unhealthy: 'unhealthy',
  degraded: 'degraded',
  other: 'other',
  unknown: 'unknown',
  progressing: 'progressing',
  healthy: 'healthy',
  'not-applicable': 'not applicable',
  unset: 'no recorded verdict',
}

const toCoverageBucket = (status?: string | null): TCoverageBucket => {
  const value = status ?? ''
  if (value === '' || value === 'unset') return 'unset'
  if (value === 'not-applicable') return 'not-applicable'
  return (COVERAGE_VERDICTS as readonly string[]).includes(value)
    ? (value as TCoverageBucket)
    : 'other'
}

function getHealthCoverage(
  componentHealthStatuses?: Record<string, string | null> | null
) {
  const counts: Partial<Record<TCoverageBucket, number>> = {}
  const statuses = Object.values(componentHealthStatuses ?? {})
  statuses.forEach((status) => {
    const bucket = toCoverageBucket(status)
    counts[bucket] = (counts[bucket] ?? 0) + 1
  })
  const recorded = statuses.length - (counts.unset ?? 0)
  const breakdown = (Object.keys(COVERAGE_BUCKET_LABELS) as TCoverageBucket[])
    .filter((bucket) => counts[bucket])
    .map((bucket) => `${counts[bucket]} ${COVERAGE_BUCKET_LABELS[bucket]}`)
  return { total: statuses.length, recorded, breakdown }
}

export const HealthCoverage = ({
  componentHealthStatuses,
}: {
  componentHealthStatuses?: Record<string, string | null> | null
}) => {
  const { total, recorded, breakdown } = getHealthCoverage(
    componentHealthStatuses
  )
  if (total === 0) {
    return (
      <Text variant="subtext" theme="neutral">
        No components
      </Text>
    )
  }
  return (
    <span className="flex flex-col">
      <Text variant="subtext">
        {recorded} of {total} with recorded verdicts
      </Text>
      <Text variant="label" theme="neutral">
        {breakdown.join(' · ')}
      </Text>
    </span>
  )
}

export type InstallRow = {
  action: ReactNode
  activity: ReactNode
  branch: ReactNode
  group: ReactNode
  healthCoverage?: ReactNode
  updatedAt: string
  appHref: string
  appName: string
  installId: string
  labels: ReactNode
  name: string
  nameHref: string
  region?: ReactNode
  statuses: ReactNode
  platform: ReactNode
}

function getCreatedBySubtitle(
  install: TInstall
): { email: string; source: string } | undefined {
  const account = install?.created_by
  if (!account?.email) return undefined
  const source = account.account_type === 'service' ? 'API / CLI' : 'Dashboard'
  return { email: account.email, source }
}

function ActivityCell({ install }: { install: TInstall }) {
  const createdBy = getCreatedBySubtitle(install)

  return (
    <ContextTooltip
      position="top"
      width="w-64"
      items={[
        ...(createdBy
          ? [
              {
                id: 'created-by',
                title: createdBy.email,
                subtitle: `via ${createdBy.source}`,
                leftContent: <Icon variant="UserIcon" size={16} />,
              },
            ]
          : []),
        {
          id: 'created',
          title: 'Created',
          subtitle: (
            <Time
              variant="label"
              time={install?.created_at}
              format="long-datetime"
            />
          ),
          leftContent: <Icon variant="PlusCircleIcon" size={16} />,
        },
        {
          id: 'updated',
          title: 'Updated',
          subtitle: (
            <Time
              variant="label"
              time={install?.updated_at}
              format="long-datetime"
            />
          ),
          leftContent: <Icon variant="ClockCounterClockwiseIcon" size={16} />,
        },
      ]}
    >
      <span className="inline-flex items-center gap-1.5 cursor-default">
        <Time time={install?.updated_at} variant="subtext" format="relative" />
        <Icon variant="InfoIcon" size={12} theme="neutral" />
      </span>
    </ContextTooltip>
  )
}

export function parseInstallsToTableData(
  installs: TInstall[],
  orgId: string,
  labelColorsByApp?: Record<string, Record<string, string>>,
  lazyComponents = false,
  installHref?: (install: TInstall) => string
): InstallRow[] {
  return installs.map((install) => ({
    appHref: `/${install.org_id}/apps/${install.app_id}`,
    appName: install?.app?.name,
    name: install.name,
    nameHref: installHref?.(install) ?? `/${orgId}/installs/${install.id}`,
    installId: install.id,
    region: (
      <CloudRegion
        variant="subtext"
        platform={(install?.cloud_platform as TCloudPlatform) || 'unknown'}
        region={install.gcp_account?.region || install.aws_account?.region}
        location={install.azure_account?.location}
      />
    ),
    statuses: (
      <InstallStatuses
        install={install}
        isLabelHidden
        lazyComponents={lazyComponents}
        tooltipPosition="top"
      />
    ),
    platform: (
      <CloudPlatform
        platform={(install?.cloud_platform as TCloudPlatform) || 'unknown'}
        variant="subtext"
        colorVariant="color"
        displayVariant="icon-only"
        iconSize="20"
      />
    ),
    labels: (() => {
      const lbls = install.labels
      if (!lbls || Object.keys(lbls).length === 0)
        return <Icon variant="MinusIcon" />
      return (
        <span className="flex flex-wrap gap-1">
          {Object.keys(lbls)
            .sort()
            .map((k) => (
              <LabelBadge
                key={k}
                size="sm"
                labelKey={k}
                labelValue={lbls[k]}
                customColor={labelColorsByApp?.[install.app_id ?? '']?.[k]}
              />
            ))}
        </span>
      )
    })(),
    branch: install.app_branch?.id ? (
      <span className="flex items-center gap-1.5">
        <Icon variant="GitBranchIcon" size={14} />
        <Link
          href={`/${orgId}/apps/${install.app_id}/branches/${install.app_branch.id}`}
          variant="inline"
        >
          {install.app_branch.name}
        </Link>
      </span>
    ) : (
      <Text variant="subtext" theme="neutral">
        —
      </Text>
    ),
    group: (
      <Text variant="subtext" theme="neutral" family="mono">
        {installAppBranchGroup(install) || '—'}
      </Text>
    ),
    healthCoverage: (
      <HealthCoverage
        componentHealthStatuses={install?.component_health_statuses}
      />
    ),
    activity: <ActivityCell install={install} />,
    updatedAt: install?.updated_at ?? '',
    action: (
      <div className="hidden md:block">
        <QuickManagementDropdown install={install} />
      </div>
    ),
  }))
}

const columns: ColumnDef<InstallRow>[] = [
  {
    accessorKey: 'name',
    header: 'Install name',
    cell: (info) => (
      <span>
        <Text variant="body">
          <Link href={info.row.original.nameHref} variant="inline">
            {info.getValue() as string}
          </Link>
        </Text>
        <ID>{info.row.original.installId as string}</ID>
      </span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'appName',
    header: 'App',
    cell: (info) => (
      <Link href={info.row.original.appHref}>{info.getValue() as string}</Link>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'statuses',
    header: 'Statuses',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    id: 'statusAxes',
    enableSorting: false,
    header: 'Statuses',
    cell: (info) => (
      <InstallStatusAxisCell installId={info.row.original.installId} />
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'healthCoverage',
    header: 'Health coverage',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    accessorKey: 'platform',
    header: 'Platform',
    cell: (info) => (
      <Text className="flex items-center gap-1">
        {info.getValue() as string}
      </Text>
    ),
    enableSorting: true,
  },
  {
    enableSorting: true,
    accessorKey: 'region',
    header: 'Region',
    cell: (info) => <Text>{info.getValue() as string}</Text>,
  },
  {
    enableSorting: false,
    accessorKey: 'labels',
    header: 'Labels',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    enableSorting: false,
    accessorKey: 'branch',
    header: 'Branch',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    enableSorting: false,
    accessorKey: 'group',
    header: 'Group',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    id: 'activity',
    accessorKey: 'updatedAt',
    header: 'Activity',
    cell: (info) => info.row.original.activity,
    enableSorting: true,
  },
  {
    enableSorting: false,
    accessorKey: 'action',
    id: 'action',
    header: '',
    cell: (info) => info.getValue<ReactNode>(),
  },
]

const HIDDEN_COLUMNS: Record<TInstallsTableScope, string[]> = {
  org: [],
  app: ['appName'],
  branch: ['appName', 'branch'],
}

interface IInstallsTable {
  data: InstallRow[]
  isLoading: boolean
  emptyStateAction?: ReactNode
  emptyTitle?: string
  emptyMessage?: string
  errorMessage?: string
  filterActions?: ReactNode
  hasActiveFilters?: boolean
  pagination: { hasNext?: boolean; offset: number; limit: number }
  scope?: TInstallsTableScope
  showHealthCoverage?: boolean
  statusAxes?: boolean
}

function getEmptyStateProps({
  emptyStateAction,
  emptyTitle,
  emptyMessage,
  errorMessage,
  hasActiveFilters,
}: Pick<
  IInstallsTable,
  | 'emptyStateAction'
  | 'emptyTitle'
  | 'emptyMessage'
  | 'errorMessage'
  | 'hasActiveFilters'
>) {
  if (errorMessage) {
    return { emptyTitle: 'Installs failed to load', emptyMessage: errorMessage }
  }
  if (hasActiveFilters) {
    return {
      emptyTitle: 'No matching installs',
      emptyMessage:
        'No installs match the current filters. Try adjusting or clearing them.',
      variant: 'search' as const,
    }
  }
  return { emptyTitle, emptyMessage, action: emptyStateAction }
}

export const InstallsTable = ({
  data,
  isLoading,
  emptyStateAction,
  emptyTitle = 'No installs created',
  emptyMessage = 'An install is an instance of an application running in a customer cloud account.',
  errorMessage,
  filterActions,
  hasActiveFilters = false,
  pagination,
  scope = 'org',
  showHealthCoverage = false,
  statusAxes = false,
}: IInstallsTable) => {
  const hidden = [
    ...HIDDEN_COLUMNS[scope],
    ...(showHealthCoverage ? [] : ['healthCoverage']),
    ...(statusAxes ? ['statuses', 'platform'] : ['statusAxes']),
  ]
  const scopedColumns = columns
    .filter(
      (c) =>
        !hidden.includes((c as { accessorKey?: string }).accessorKey ?? '') &&
        !hidden.includes(c.id ?? '')
    )
    .map((column) => {
      if (!statusAxes) return column
      const key = (column as { accessorKey?: string }).accessorKey ?? column.id
      if (key === 'region') {
        return {
          ...column,
          meta: { className: 'hidden @min-[80rem]:table-cell' },
        }
      }
      if (key === 'updatedAt' || key === 'activity') {
        return {
          ...column,
          meta: { className: 'hidden @min-[64rem]:table-cell' },
        }
      }
      return column
    })

  return (
    <Table<InstallRow>
      columns={scopedColumns}
      data={errorMessage ? [] : data}
      isLoading={isLoading && !errorMessage}
      emptyStateProps={getEmptyStateProps({
        emptyStateAction,
        emptyTitle,
        emptyMessage,
        errorMessage,
        hasActiveFilters,
      })}
      filterActions={filterActions}
      pagination={
        errorMessage ? { ...pagination, hasNext: false, offset: 0 } : pagination
      }
      className={statusAxes ? '@container' : ''}
      searchPlaceholder={
        scope === 'org'
          ? 'Search by name, branch or ID...'
          : 'Search by name or ID...'
      }
    />
  )
}
