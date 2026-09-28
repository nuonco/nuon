import type { ColumnDef } from '@tanstack/react-table'
import { Badge, type TBadgeTheme } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TCloudConnection } from '@/types'

const statusTheme: Record<string, TBadgeTheme> = {
  verified: 'success',
  error: 'error',
  pending: 'warn',
}

interface ICloudConnectionsTable {
  connections: TCloudConnection[]
  isLoading: boolean
  pendingActionId?: string
  onOpen: (connection: TCloudConnection) => void
  onVerify: (connection: TCloudConnection) => void
  onDelete: (connection: TCloudConnection) => void
}

export const CloudConnectionsTable = ({
  connections,
  isLoading,
  pendingActionId,
  onOpen,
  onVerify,
  onDelete,
}: ICloudConnectionsTable) => {
  const columns: ColumnDef<TCloudConnection>[] = [
    {
      accessorKey: 'name',
      header: 'Connection',
      cell: ({ row }) => (
        <button
          className="text-left text-primary-600 dark:text-primary-400 hover:underline"
          onClick={() => onOpen(row.original)}
        >
          <Text weight="strong">{row.original.name}</Text>
        </button>
      ),
    },
    {
      accessorKey: 'platform',
      header: 'Cloud',
      cell: ({ getValue }) => String(getValue()).toUpperCase(),
    },
    { accessorKey: 'target_id', header: 'Target' },
    {
      accessorKey: 'principal',
      header: 'Principal',
      cell: ({ getValue }) => (
        <Text
          family="mono"
          className="max-w-64 truncate block"
          title={String(getValue())}
        >
          {String(getValue()).replace(/^.*\//, '')}
        </Text>
      ),
    },
    {
      accessorKey: 'status',
      header: 'Status',
      cell: ({ getValue }) => {
        const status = String(getValue() || 'pending')
        return (
          <Badge theme={statusTheme[status] ?? 'neutral'} size="sm">
            {status}
          </Badge>
        )
      },
    },
    {
      accessorKey: 'preset',
      header: 'Access preset',
    },
    {
      accessorKey: 'last_verified_at',
      header: 'Last verified',
      cell: ({ row }) =>
        row.original.last_verified_at ? (
          <Time
            time={row.original.last_verified_at}
            format="relative"
            className="whitespace-nowrap"
          />
        ) : (
          'Never'
        ),
    },
    {
      id: 'usedBy',
      header: 'Used by',
      cell: ({ row }) => {
        const installs = row.original.used_by?.installs ?? 0
        return `${installs} install${installs === 1 ? '' : 's'}`
      },
    },
    {
      id: 'actions',
      enableSorting: false,
      cell: ({ row }) => {
        const connection = row.original
        const inUse = (connection.used_by?.installs ?? 0) > 0
        const isPending = pendingActionId === connection.id
        return (
          <div className="flex items-center justify-end gap-1">
            <Button size="sm" onClick={() => onOpen(connection)}>
              Setup
            </Button>
            <Button
              size="sm"
              onClick={() => onVerify(connection)}
              disabled={isPending}
            >
              Verify
            </Button>
            <Button
              size="sm"
              variant="icon"
              aria-label={`Delete ${connection.name}`}
              disabled={inUse || isPending}
              tooltipProps={
                inUse ? { tipContent: 'This connection is in use' } : undefined
              }
              onClick={() => onDelete(connection)}
            >
              <Icon variant="TrashIcon" />
            </Button>
          </div>
        )
      },
    },
  ]

  return (
    <Table
      columns={columns}
      data={connections}
      isLoading={isLoading}
      searchPlaceholder="Search connections"
      emptyStateProps={{
        emptyTitle: 'No cloud connections yet',
        emptyMessage: 'Create an AWS connection to deploy install stacks.',
      }}
    />
  )
}
