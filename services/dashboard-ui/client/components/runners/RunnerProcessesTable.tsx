import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { getRunnerProcesses } from '@/lib'
import type { TRunnerProcess } from '@/types'

const LIMIT = 20

function getStatusTheme(status: string) {
  switch (status) {
    case 'active':
      return 'success'
    case 'shutting-down':
      return 'warning'
    case 'shut-down':
      return 'neutral'
    case 'error':
      return 'danger'
    default:
      return 'neutral'
  }
}

function formatUptime(startedAt: string | undefined): string {
  if (!startedAt) return '-'
  const start = new Date(startedAt)
  const now = new Date()
  const diffMs = now.getTime() - start.getTime()
  const hours = Math.floor(diffMs / (1000 * 60 * 60))
  const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60))
  if (hours > 0) return `${hours}h ${minutes}m`
  if (minutes < 1) return 'less than a minute'
  return `${minutes}m`
}

function ProcessRow({ process }: { process: TRunnerProcess }) {
  return (
    <tr className="border-b border-neutral-100 last:border-0">
      <td className="px-4 py-3">
        <ID id={process.id} />
      </td>
      <td className="px-4 py-3">
        <Text variant="small">{process.type || '-'}</Text>
      </td>
      <td className="px-4 py-3">
        <Badge theme={getStatusTheme(process.status)}>
          {process.status}
        </Badge>
      </td>
      <td className="px-4 py-3">
        <Text variant="small">{process.version || '-'}</Text>
      </td>
      <td className="px-4 py-3">
        <Text variant="small">{formatUptime(process.started_at)}</Text>
      </td>
      <td className="px-4 py-3">
        <Text variant="small">
          {process.created_at
            ? new Date(process.created_at).toLocaleString()
            : '-'}
        </Text>
      </td>
    </tr>
  )
}

export const RunnerProcessesTable = ({
  shouldPoll = true,
  pollInterval = 20000,
  filterStatus,
}: {
  shouldPoll?: boolean
  pollInterval?: number
  filterStatus?: string
}) => {
  const { org } = useOrg()
  const { runner } = useRunner()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: ['runner-processes', org?.id, runner?.id, offset, filterStatus],
    queryFn: () =>
      getRunnerProcesses({
        orgId: org.id,
        runnerId: runner.id,
        limit: LIMIT,
        offset,
        status: filterStatus,
      }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!runner?.id,
  })

  if (isLoading) {
    return <RunnerProcessesTableSkeleton />
  }

  const processes = result?.data ?? []

  if (processes.length === 0) {
    return (
      <Card>
        <EmptyState
          emptyTitle="No processes"
          emptyMessage="No runner processes have been registered yet."
          variant="table"
        />
      </Card>
    )
  }

  return (
    <Card className="overflow-hidden">
      <table className="w-full text-left">
        <thead className="border-b border-neutral-200 bg-neutral-50">
          <tr>
            <th className="px-4 py-2">
              <Text variant="small" weight="strong">ID</Text>
            </th>
            <th className="px-4 py-2">
              <Text variant="small" weight="strong">Type</Text>
            </th>
            <th className="px-4 py-2">
              <Text variant="small" weight="strong">Status</Text>
            </th>
            <th className="px-4 py-2">
              <Text variant="small" weight="strong">Version</Text>
            </th>
            <th className="px-4 py-2">
              <Text variant="small" weight="strong">Uptime</Text>
            </th>
            <th className="px-4 py-2">
              <Text variant="small" weight="strong">Started</Text>
            </th>
          </tr>
        </thead>
        <tbody>
          {processes.map((process) => (
            <ProcessRow key={process.id} process={process} />
          ))}
        </tbody>
      </table>
    </Card>
  )
}

export const RunnerProcessesTableSkeleton = () => (
  <Card>
    <div className="flex flex-col gap-3 p-4">
      {Array.from({ length: 5 }).map((_, i) => (
        <Skeleton key={i} height="32px" />
      ))}
    </div>
  </Card>
)
