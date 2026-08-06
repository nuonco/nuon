import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TVCSConnection, TVCSConnectionStatus } from '@/types'

export type TVCSConnectionRow = {
  connection: TVCSConnection
  href: string
  status?: TVCSConnectionStatus['status']
  checkedAt?: string
  isLoadingStatus?: boolean
}

const accountName = (connection: TVCSConnection) =>
  connection.github_account_name || connection.github_account_id || 'GitHub'

export const VCSConnectionsTable = ({
  data,
  isLoading,
}: {
  data: TVCSConnectionRow[]
  isLoading?: boolean
}) => {
  const columns: ColumnDef<TVCSConnectionRow>[] = useMemo(
    () => [
      {
        header: 'Account',
        id: 'account',
        accessorFn: (row) => accountName(row.connection),
        cell: ({ row }) => (
          <span className="flex items-center gap-2">
            <Icon variant="GitHub" size={16} />
            <Link href={row.original.href}>
              {accountName(row.original.connection)}
            </Link>
          </span>
        ),
      },
      {
        header: 'Status',
        id: 'status',
        enableSorting: false,
        cell: ({ row }) => {
          const { status, isLoadingStatus } = row.original
          if (isLoadingStatus) return <Skeleton height="20px" width="72px" />
          if (!status) return <Text theme="neutral">—</Text>
          return <Status status={status} variant="badge" />
        },
      },
      {
        header: 'Last checked',
        id: 'checked_at',
        enableSorting: false,
        cell: ({ row }) =>
          row.original.checkedAt ? (
            <Time time={row.original.checkedAt} format="relative" />
          ) : (
            <Text theme="neutral">—</Text>
          ),
      },
    ],
    []
  )

  return (
    <Table<TVCSConnectionRow>
      columns={columns}
      data={data}
      isLoading={isLoading}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No VCS connections yet',
        emptyMessage:
          'Connect a GitHub account to build components from your repositories.',
      }}
    />
  )
}
