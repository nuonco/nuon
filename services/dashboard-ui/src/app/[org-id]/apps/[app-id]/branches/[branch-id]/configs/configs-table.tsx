import type { ColumnDef } from '@tanstack/react-table'
import { Link } from '@/components/common/Link'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Badge } from '@/components/common/Badge'
import { Table } from '@/components/common/Table'
import { getBranchConfigs } from '@/lib'
import type { TAppBranchConfig } from '@/types'

type TConfigRow = {
  id: string
  version: number
  createdAt: string
  installGroupsCount: number
  hasVcsConfig: boolean
  href: string
}

function parseConfigsToTableData(
  configs: TAppBranchConfig[],
  orgId: string,
  appId: string,
  branchId: string
): TConfigRow[] {
  return configs.map((config) => ({
    id: config.id || '',
    version: config.config_number || 0,
    createdAt: config.created_at || '',
    installGroupsCount: config.install_groups?.length || 0,
    hasVcsConfig: !!config.connected_github_vcs_config,
    href: `/${orgId}/apps/${appId}/branches/${branchId}/configs/${config.id}`,
  }))
}

const columns: ColumnDef<TConfigRow>[] = [
  {
    accessorKey: 'version',
    header: 'Version',
    cell: (info) => (
      <Link href={info.row.original.href}>
        <Badge theme="info" size="sm">
          v{info.getValue() as number}
        </Badge>
      </Link>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'createdAt',
    header: 'Created',
    cell: (info) => <Time time={info.getValue() as string} format="relative" />,
    enableSorting: true,
  },
  {
    accessorKey: 'installGroupsCount',
    header: 'Install Groups',
    cell: (info) => {
      const count = info.getValue() as number
      return (
        <Text variant="body">
          {count} group{count !== 1 ? 's' : ''}
        </Text>
      )
    },
  },
  {
    accessorKey: 'hasVcsConfig',
    header: 'VCS',
    cell: (info) => {
      const hasVcs = info.getValue() as boolean
      return hasVcs ? (
        <Badge theme="success" size="sm">
          Configured
        </Badge>
      ) : (
        <Text variant="subtext" theme="neutral">
          Not configured
        </Text>
      )
    },
  },
  {
    enableSorting: false,
    accessorKey: 'href',
    id: 'action',
    header: '',
    cell: (info) => (
      <Text>
        <Link className="text-left" href={info.getValue() as string}>
          View <Icon variant="CaretRightIcon" />
        </Link>
      </Text>
    ),
  },
]

interface IConfigsTable {
  appId: string
  branchId: string
  orgId: string
  offset: string
}

export const ConfigsTable = async ({
  appId,
  branchId,
  orgId,
  offset,
}: IConfigsTable) => {
  const { data: configs, error } = await getBranchConfigs({
    appId,
    branchId,
    orgId,
    limit: 50,
    offset: Number(offset),
  })

  if (error) {
    return (
      <div className="rounded-lg border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/10 p-4">
        <Text variant="body" theme="error">
          Failed to load configurations
        </Text>
      </div>
    )
  }

  // Sort by config_number descending (newest first)
  const sortedConfigs = (configs || []).sort(
    (a, b) => (b.config_number || 0) - (a.config_number || 0)
  )

  return (
    <Table<TConfigRow>
      columns={columns}
      data={parseConfigsToTableData(sortedConfigs, orgId, appId, branchId)}
      emptyStateProps={{
        emptyMessage:
          'No configurations yet. Create your first configuration to get started.',
        emptyTitle: 'No configurations',
      }}
      searchPlaceholder="Search configurations..."
      pagination={{
        hasNext: false,
        offset: 0,
      }}
    />
  )
}

export const ConfigsTableSkeleton = () => {
  return (
    <div className="rounded-lg border border-gray-200 dark:border-gray-700">
      <div className="h-12 bg-gray-100 dark:bg-gray-800 rounded-t-lg animate-pulse" />
      {Array.from({ length: 5 }).map((_, i) => (
        <div
          key={i}
          className="h-16 bg-gray-50 dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700 animate-pulse"
        />
      ))}
    </div>
  )
}