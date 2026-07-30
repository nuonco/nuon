import { useMemo, type ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Link } from '@/components/common/Link'
import { Button } from '@/components/common/Button'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TTrigger } from '@/types'

export const TriggersTable = ({
  data,
  isLoading,
  error,
  filterActions,
  onRetry,
  orgId,
}: {
  data: TTrigger[]
  isLoading: boolean
  error?: boolean
  filterActions?: ReactNode
  onRetry?: () => void
  orgId: string
}) => {
  const columns = useMemo<ColumnDef<TTrigger>[]>(
    () => [
      {
        header: 'Name',
        accessorKey: 'name',
        enableSorting: false,
        cell: ({ row }) => (
          <div className="flex flex-col gap-1">
            <Link href={`/${orgId}/triggers/${row.original?.id}`}>
              {row.original?.name || row.original?.id || 'Unnamed trigger'}
            </Link>
            <Text variant="subtext" theme="neutral" family="mono">
              {row.original?.id || '—'}
            </Text>
          </div>
        ),
      },
      {
        header: 'Status',
        accessorKey: 'status',
        enableSorting: false,
        cell: ({ getValue }) => (
          <Status
            status={getValue<string>() === 'active' ? 'success' : 'warn'}
            variant="badge"
          >
            {getValue<string>() || 'Unknown'}
          </Status>
        ),
      },
      {
        header: 'Auth type',
        id: 'auth_type',
        accessorFn: (trigger) => trigger?.auth_type,
        enableSorting: false,
        cell: ({ getValue }) => (
          <Text variant="subtext">{getValue<string>() || '—'}</Text>
        ),
      },
      {
        header: 'Envelope',
        id: 'envelope',
        accessorFn: (trigger) => trigger?.envelope,
        enableSorting: false,
        cell: ({ getValue }) => (
          <Text variant="subtext">{getValue<string>() || '—'}</Text>
        ),
      },
      {
        header: 'Last event',
        id: 'last_event_at',
        accessorFn: (trigger) => trigger?.last_event_at || undefined,
        sortUndefined: 'last',
        sortDescFirst: true,
        cell: ({ getValue }) =>
          getValue<string>() ? (
            <Time
              time={getValue<string>()}
              format="relative"
              shouldTick
              variant="subtext"
            />
          ) : (
            <Text variant="subtext" theme="neutral">
              No events yet
            </Text>
          ),
      },
    ],
    [orgId]
  )
  if (error)
    return (
      <div className="flex flex-col items-start gap-3">
        <Text theme="error">Trigger loading failed.</Text>
        <Button variant="secondary" onClick={onRetry}>
          Retry loading triggers
        </Button>
      </div>
    )
  return (
    <Table
      columns={columns}
      data={data}
      enableSearch={false}
      filterActions={filterActions}
      initialSorting={[{ id: 'last_event_at', desc: true }]}
      isLoading={isLoading}
      emptyStateProps={{
        emptyTitle: 'No triggers yet',
        emptyMessage: 'Create a trigger to start receiving trigger events.',
      }}
    />
  )
}
