import type { ColumnDef } from '@tanstack/react-table'
import { Hash } from '@/components/common/Hash'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TAppRelease } from '@/types'

interface IReleasesTable {
  appId: string
  data: TAppRelease[]
  isLoading?: boolean
  orgId: string
  pagination?: { hasNext?: boolean; limit: number; offset: number }
}

export const ReleasesTable = ({
  appId,
  data,
  isLoading,
  orgId,
  pagination,
}: IReleasesTable) => {
  const columns: ColumnDef<TAppRelease>[] = [
    {
      accessorKey: 'id',
      header: 'Release',
      cell: ({ row }) => (
        <span className="flex flex-col gap-1">
          <Link
            href={`/${orgId}/apps/${appId}/releases/${row.original.id}`}
            variant="inline"
          >
            {row.original.id}
          </Link>
          <ID>{row.original.app_config_id}</ID>
        </span>
      ),
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ row }) => (
        <span className="flex flex-col gap-1">
          <Status status={row.original.status} />
          {row.original.status_description ? (
            <Text variant="subtext" theme="neutral">
              {row.original.status_description}
            </Text>
          ) : null}
        </span>
      ),
    },
    {
      accessorKey: 'semantic_digest',
      header: 'Digest',
      cell: ({ row }) => (
        <Hash hash={row.original.semantic_digest ?? ''} />
      ),
    },
    {
      accessorKey: 'created_at',
      header: 'Created',
      cell: ({ row }) =>
        row.original.created_at ? (
          <Time time={row.original.created_at} format="relative" />
        ) : null,
    },
  ]

  return (
    <Table
      columns={columns}
      data={data}
      isLoading={isLoading}
      pagination={pagination}
      searchPlaceholder="Search by release or config ID..."
      emptyStateProps={{
        variant: 'app',
        emptyTitle: 'No releases yet',
        emptyMessage:
          'Releases will appear here when an app config is prepared for customer-managed installs.',
      }}
    />
  )
}
