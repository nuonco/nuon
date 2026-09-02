import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { BranchConfigDetailsButton } from '@/components/branches/BranchConfigDetails'
import type { TAppConfig } from '@/types'

export interface IBranchConfigsTable {
  configs: TAppConfig[]
  isLoading?: boolean
  appId?: string
}

export const BranchConfigsTable = ({
  configs,
  isLoading,
  appId,
}: IBranchConfigsTable) => {
  const columns = useMemo<ColumnDef<TAppConfig, unknown>[]>(
    () => [
      {
        accessorKey: 'version',
        header: 'Version',
        cell: ({ row }) => (
          <BranchConfigDetailsButton config={row.original} appId={appId}>
            {row.original?.version ? `v${row.original.version}` : 'View config'}
          </BranchConfigDetailsButton>
        ),
      },
      {
        id: 'status',
        accessorFn: (config) =>
          config?.status_v2?.status || config?.status || '',
        header: 'Status',
        enableSorting: false,
        cell: ({ row }) => (
          <Status
            variant="badge"
            status={
              row.original?.status_v2?.status ||
              row.original?.status ||
              'unknown'
            }
          />
        ),
      },
      {
        id: 'source',
        accessorFn: (config) =>
          config?.vcs_connection_commit?.author_name ||
          config?.cli_version ||
          '',
        header: 'Source',
        cell: ({ row }) => {
          const commit = row.original?.vcs_connection_commit

          if (commit) {
            return (
              <span className="flex flex-wrap items-center gap-2">
                {commit.sha ? (
                  <Badge size="sm" variant="code">
                    {commit.sha.slice(0, 7)}
                  </Badge>
                ) : null}
                <Text variant="subtext" theme="neutral">
                  {commit.author_name || 'git'}
                </Text>
              </span>
            )
          }

          return (
            <Text variant="subtext" theme="neutral">
              {row.original?.cli_version
                ? `CLI ${row.original.cli_version}`
                : 'Unknown'}
            </Text>
          )
        },
      },
      {
        accessorKey: 'id',
        header: 'Config',
        enableSorting: false,
        cell: ({ row }) => <ID>{row.original?.id}</ID>,
      },
      {
        accessorKey: 'created_at',
        header: 'Created',
        cell: ({ row }) => (
          <Time
            variant="subtext"
            time={row.original?.created_at}
            format="relative"
          />
        ),
      },
    ],
    [appId]
  )

  return (
    <Table<TAppConfig>
      columns={columns}
      data={configs ?? []}
      enableSearch={false}
      isLoading={isLoading}
      initialSorting={[{ id: 'version', desc: true }]}
      emptyStateProps={{
        variant: 'table',
        emptyTitle: 'No configs yet',
        emptyMessage:
          'Config versions appear here once this branch syncs an app config.',
      }}
    />
  )
}
