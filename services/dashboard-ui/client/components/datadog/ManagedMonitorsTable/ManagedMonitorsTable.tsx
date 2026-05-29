import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Code } from '@/components/common/Code'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DeleteManagedMonitorButton } from '@/components/datadog/DeleteManagedMonitor'
import type {
  TDatadogConnection,
  TDatadogManagedMonitor,
} from '@/types'

const connectionName = (
  m: TDatadogManagedMonitor,
  connections: TDatadogConnection[]
): string => {
  const conn = connections.find((c) => c.id === m.connection_id)
  return conn?.name || m.connection_id || '—'
}

export const ManagedMonitorsTable = ({
  data,
  connections,
  isLoading,
}: {
  data: TDatadogManagedMonitor[]
  connections: TDatadogConnection[]
  isLoading: boolean
}) => {
  const columns: ColumnDef<TDatadogManagedMonitor>[] = useMemo(
    () => [
      {
        header: 'Target',
        id: 'target',
        cell: (props) => {
          const m = props.row.original
          return (
            <div className="flex flex-col gap-1">
              <div className="flex items-center gap-2">
                <Badge theme="neutral">{m.target_type}</Badge>
                <Badge
                  theme={m.preset === 'failure' ? 'error' : 'warn'}
                >
                  {m.preset}
                </Badge>
                <Badge theme={m.mode === 'metric' ? 'info' : 'neutral'}>
                  {m.mode || 'event'}
                </Badge>
              </div>
              <Code variant="inline" className="!px-2 !py-0.5 w-fit">
                {m.target_id}
              </Code>
            </div>
          )
        },
      },
      {
        header: 'Connection',
        id: 'connection',
        cell: (props) => (
          <Text variant="subtext" theme="neutral">
            {connectionName(props.row.original, connections)}
          </Text>
        ),
      },
      {
        header: 'Notifies',
        accessorKey: 'notify_handles',
        cell: (props) => {
          const handles = props.getValue<string[] | undefined>() ?? []
          if (handles.length === 0) {
            return (
              <Text variant="subtext" theme="neutral">
                (none — alert visible in DD only)
              </Text>
            )
          }
          return (
            <div className="flex flex-wrap gap-1">
              {handles.map((h) => (
                <Badge key={h} theme="neutral">
                  {h}
                </Badge>
              ))}
            </div>
          )
        },
      },
      {
        header: 'Created',
        accessorKey: 'created_at',
        cell: (props) => {
          const time = props.getValue<string | undefined>()
          return time ? (
            <Time variant="subtext" time={time} format="relative" />
          ) : (
            <Text variant="subtext" theme="neutral">
              —
            </Text>
          )
        },
      },
      {
        id: 'action',
        header: '',
        cell: (props) => (
          <div className="flex justify-end gap-1">
            <DeleteManagedMonitorButton monitor={props.row.original} size="sm" />
          </div>
        ),
      },
    ],
    [connections]
  )

  if (isLoading) return <ManagedMonitorsTableSkeleton />

  return (
    <Table<TDatadogManagedMonitor>
      columns={columns}
      data={data}
      enableSearch={false}
      emptyStateProps={{
        emptyTitle: 'No managed monitors',
        emptyMessage:
          'Open an install, action, or component and click "Alert in Datadog" to create a one-click monitor here.',
      }}
    />
  )
}

const skeletonColumns: ColumnDef<TDatadogManagedMonitor>[] = [
  { header: 'Target', id: 'target' },
  { header: 'Connection', id: 'connection' },
  { header: 'Notifies', accessorKey: 'notify_handles' },
  { header: 'Created', accessorKey: 'created_at' },
  { header: '', id: 'action' },
]

export const ManagedMonitorsTableSkeleton = () => (
  <TableSkeleton<TDatadogManagedMonitor>
    columns={skeletonColumns}
    skeletonRows={3}
  />
)
