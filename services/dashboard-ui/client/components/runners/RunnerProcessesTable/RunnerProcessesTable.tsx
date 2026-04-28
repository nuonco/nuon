import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { Badge } from '@/components/common/Badge'
import { Duration } from '@/components/common/Duration'
import { ID } from '@/components/common/ID'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useAuth } from '@/hooks/use-auth'
import type { TRunnerProcess } from '@/types'

interface IRunnerProcessesTable {
  processes: TRunnerProcess[]
  isLoading: boolean
}

export const RunnerProcessesTable = ({
  processes,
  isLoading,
}: IRunnerProcessesTable) => {
  const { isAdmin } = useAuth()

  const columns = useMemo<ColumnDef<TRunnerProcess>[]>(() => {
    const cols: ColumnDef<TRunnerProcess>[] = [
      {
        accessorKey: 'id',
        header: 'ID',
        cell: (info) => <ID>{info.getValue() as string}</ID>,
        enableSorting: false,
      },
      {
        accessorKey: 'type',
        header: 'Type',
        cell: (info) => (
          <Text variant="subtext">{(info.getValue() as string) || '-'}</Text>
        ),
      },
      {
        id: 'status',
        header: 'Status',
        accessorFn: (row) => row.composite_status?.status,
        cell: (info) => (
          <Status status={info.getValue() as string} variant="badge" />
        ),
      },
      {
        id: 'version',
        header: 'Version',
        cell: (info) => {
          const process = info.row.original
          return (
            <div className="flex items-center gap-2">
              <Text variant="subtext">{process.version || '-'}</Text>
              {process.labels?.map((label) => (
                <Badge key={label} theme="neutral" variant="code" size="sm">{label}</Badge>
              ))}
            </div>
          )
        },
        enableSorting: false,
      },
      {
        accessorKey: 'started_at',
        header: 'Uptime',
        cell: (info) => {
          const startedAt = info.getValue() as string | undefined
          return startedAt ? (
            <Duration
              variant="subtext"
              beginTime={startedAt}
              durationUnits={['hours', 'minutes']}
              unitDisplay="short"
            />
          ) : (
            <Text variant="subtext">-</Text>
          )
        },
      },
      {
        accessorKey: 'created_at',
        header: 'Started',
        cell: (info) => {
          const createdAt = info.getValue() as string | undefined
          return createdAt ? (
            <Time variant="subtext" time={createdAt} format="relative" />
          ) : (
            <Text variant="subtext">-</Text>
          )
        },
      },
    ]

    if (isAdmin) {
      cols.push({
        id: 'admin',
        header: '',
        cell: (info) => {
          const process = info.row.original
          return (
            <AdminDashboardLink
              path={`/queues?owner_id=${process.runner_id}&search=runner-process-${process.id}&redirect=true`}
              label="View in admin panel"
            />
          )
        },
        enableSorting: false,
      })
    }

    return cols
  }, [isAdmin])

  return (
    <Table
      data={processes}
      columns={columns}
      isLoading={isLoading}
      enableSearch={false}
      enableSorting={false}
      emptyStateProps={{
        emptyTitle: 'No processes',
        emptyMessage: 'No runner processes have been registered yet.',
      }}
    />
  )
}
