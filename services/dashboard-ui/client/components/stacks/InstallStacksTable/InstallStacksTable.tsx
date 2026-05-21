import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TInstallStack } from '@/types'
import { StackVersionActions } from '../StackVersionActions'

export type TInstallStackRow = {
  versionId: string
  appConfigId: string
  appStackConfigHref: string
  status: ReactNode
  runs: string
  createdAt: string
  more: ReactNode
}

export function parseInstallStackSummaryToTableData(
  stack: TInstallStack,
  orgId: string,
  appId: string
): TInstallStackRow[] {
  return stack?.versions.map((version) => {
    return {
      versionId: version?.id,
      appConfigId: version?.app_config_id,
      appStackConfigHref: `/${orgId}/apps/${appId}`,
      status: (
        <Status variant="badge" status={version.composite_status?.status} />
      ),
      runs: version?.runs?.length?.toString() || '-',
      createdAt: version?.created_at,
      more: <StackVersionActions version={version} />,
    }
  })
}

const columns: ColumnDef<TInstallStackRow>[] = [
  {
    accessorKey: 'versionId',
    header: 'Version',
    cell: (info) => <ID>{info.getValue<string>()}</ID>,
    enableSorting: true,
  },
  {
    accessorKey: 'appConfigId',
    header: 'App version',
    cell: (info) => (
      <Text variant="subtext">
        <Link href={info?.row?.original?.appStackConfigHref}>
          {info.getValue<string>()}
        </Link>
      </Text>
    ),
  },
  {
    enableSorting: false,
    accessorKey: 'status',
    header: 'Status',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    enableSorting: true,
    accessorKey: 'runs',
    header: 'Runs',
    cell: (info) => info.getValue<string>(),
  },
  {
    accessorKey: 'createdAt',
    header: 'Created',
    cell: (info) => (
      <Time
        time={info.getValue() as string}
        variant="subtext"
        format="relative"
      />
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'more',
    header: '',
    id: 'more-options',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
]

interface IInstallStacksTable {
  data: TInstallStackRow[]
  isLoading: boolean
  pagination: { hasNext: boolean; offset: number; limit: number }
}

export const InstallStacksTable = ({
  data,
  isLoading,
  pagination,
}: IInstallStacksTable) => {
  if (!data && !isLoading) return null

  return (
    <Table<TInstallStackRow>
      columns={columns}
      data={data ?? []}
      isLoading={isLoading}
      emptyMessage="No stack found"
      pagination={pagination}
      searchPlaceholder="Search stack version..."
    />
  )
}

export const InstallStacksTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
