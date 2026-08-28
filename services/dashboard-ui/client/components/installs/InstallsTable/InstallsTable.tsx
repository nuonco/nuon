import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
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
import { QuickManagementDropdown } from '@/components/installs/management/QuickManagementDropdown'
import { LabelBadge } from '@/components/common/LabelBadge'
import type { TCloudPlatform, TInstall } from '@/types'
import { isCustomerManagedInstall } from '@/utils/install-utils'

export type TInstallsTableScope = 'org' | 'app' | 'branch'

export type InstallRow = {
  action: ReactNode
  activity: ReactNode
  branch: ReactNode
  updatedAt: string
  appHref: string
  appName: string
  installId: string
  labels: ReactNode
  management: ReactNode
  name: string
  nameHref: string
  region?: ReactNode
  statuses: ReactNode
  platform: ReactNode
}

function ManagementBadge({ install }: { install: TInstall }) {
  const policy = install.management_policy
  if (!policy || policy.command_authority === 'nuon') {
    return (
      <Badge size="sm" theme="neutral">
        Nuon-managed
      </Badge>
    )
  }

  if (policy.connectivity === 'disconnected') {
    return (
      <Badge size="sm" theme="neutral">
        <Icon variant="CloudSlashIcon" size={12} />
        Customer-managed · offline
      </Badge>
    )
  }

  return (
    <Badge size="sm" theme="info">
      Customer-managed · connected
    </Badge>
  )
}

function isSupportedCloudPlatform(
  platform: string | undefined
): platform is Exclude<TCloudPlatform, 'unknown'> {
  return platform === 'aws' || platform === 'azure' || platform === 'gcp'
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
  labelColorsByApp?: Record<string, Record<string, string>>
): InstallRow[] {
  return installs.map((install) => {
    const platform = install.cloud_platform
    return {
      appHref: `/${install.org_id}/apps/${install.app_id}`,
      appName: install?.app?.name,
      name: install.name,
      nameHref: `/${orgId}/installs/${install.id}`,
      installId: install.id,
      management: <ManagementBadge install={install} />,
      region: (
        <CloudRegion
          variant="subtext"
          platform={(install?.cloud_platform as TCloudPlatform) || 'unknown'}
          region={install.gcp_account?.region || install.aws_account?.region}
          location={install.azure_account?.location}
        />
      ),
      statuses: (
        <InstallStatuses install={install} isLabelHidden tooltipPosition="top" />
      ),
      platform: isSupportedCloudPlatform(platform) ? (
        <CloudPlatform
          platform={platform}
          variant="subtext"
          colorVariant="color"
          displayVariant="icon-only"
          iconSize="20"
        />
      ) : (
        <CloudPlatform
          platform="unknown"
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
      activity: <ActivityCell install={install} />,
      updatedAt: install?.updated_at ?? '',
      action: (
        <div className="hidden md:block">
          <QuickManagementDropdown install={install} />
        </div>
      ),
    }
  })
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
    accessorKey: 'management',
    header: 'Management',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    enableSorting: false,
    accessorKey: 'statuses',
    header: 'Statuses',
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
  filterActions?: ReactNode
  pagination: { hasNext?: boolean; offset: number; limit: number }
  scope?: TInstallsTableScope
}

export const InstallsTable = ({
  data,
  isLoading,
  emptyStateAction,
  emptyTitle = 'No installs created',
  emptyMessage = 'An install is an instance of an application running in a customer cloud account.',
  filterActions,
  pagination,
  scope = 'org',
}: IInstallsTable) => {
  const hidden = HIDDEN_COLUMNS[scope]
  const scopedColumns = columns.filter(
    (c) => !hidden.includes((c as { accessorKey?: string }).accessorKey ?? '')
  )

  return (
    <Table<InstallRow>
      columns={scopedColumns}
      data={data}
      isLoading={isLoading}
      emptyStateProps={{
        emptyMessage,
        emptyTitle,
        action: emptyStateAction,
      }}
      filterActions={filterActions}
      pagination={pagination}
      searchPlaceholder={
        scope === 'org'
          ? 'Search by name, branch or ID...'
          : 'Search by name or ID...'
      }
    />
  )
}
