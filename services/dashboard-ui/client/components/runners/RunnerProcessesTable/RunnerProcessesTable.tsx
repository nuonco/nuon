import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import type { TRunnerProcess } from '@/types'

function getStatusTheme(status: string) {
  switch (status) {
    case 'active':
      return 'success'
    case 'shutting-down':
      return 'warn'
    case 'shut-down':
      return 'neutral'
    case 'error':
      return 'error'
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

function ProcessRow({ process, adminDashboardUrl }: { process: TRunnerProcess; adminDashboardUrl?: string }) {
  return (
    <tr className="border-b border-neutral-100 last:border-0">
      <td className="px-4 py-3">
        <ID id={process.id} />
      </td>
      <td className="px-4 py-3">
        <Text variant="subtext">{process.type || '-'}</Text>
      </td>
      <td className="px-4 py-3">
        <Badge theme={getStatusTheme(process.composite_status?.status)}>
          {process.composite_status?.status}
        </Badge>
      </td>
      <td className="px-4 py-3">
        <div className="flex items-center gap-2">
          <Text variant="subtext">{process.version || '-'}</Text>
          {process.labels?.map((label) => (
            <Badge key={label} theme="neutral">{label}</Badge>
          ))}
        </div>
      </td>
      <td className="px-4 py-3">
        <Text variant="subtext">{formatUptime(process.started_at)}</Text>
      </td>
      <td className="px-4 py-3">
        <Text variant="subtext">
          {process.created_at
            ? new Date(process.created_at).toLocaleString()
            : '-'}
        </Text>
      </td>
      {adminDashboardUrl && (
        <td className="px-4 py-3">
          <Button
            size="sm"
            href={`${adminDashboardUrl}/queues?owner_id=${process.runner_id}&search=runner-process-${process.id}&redirect=true`}
            target="_blank"
          >
            View in admin panel <Icon variant="ArrowSquareOutIcon" />
          </Button>
        </td>
      )}
    </tr>
  )
}

interface IRunnerProcessesTable {
  processes: TRunnerProcess[]
  isLoading: boolean
  adminDashboardUrl?: string
}

export const RunnerProcessesTable = ({
  processes,
  isLoading,
  adminDashboardUrl,
}: IRunnerProcessesTable) => {
  if (isLoading) {
    return <RunnerProcessesTableSkeleton />
  }

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
              <Text variant="subtext" weight="strong">ID</Text>
            </th>
            <th className="px-4 py-2">
              <Text variant="subtext" weight="strong">Type</Text>
            </th>
            <th className="px-4 py-2">
              <Text variant="subtext" weight="strong">Status</Text>
            </th>
            <th className="px-4 py-2">
              <Text variant="subtext" weight="strong">Version</Text>
            </th>
            <th className="px-4 py-2">
              <Text variant="subtext" weight="strong">Uptime</Text>
            </th>
            <th className="px-4 py-2">
              <Text variant="subtext" weight="strong">Started</Text>
            </th>
            {adminDashboardUrl && (
              <th className="px-4 py-2">
                <Text variant="subtext" weight="strong">Admin</Text>
              </th>
            )}
          </tr>
        </thead>
        <tbody>
          {processes.map((process) => (
            <ProcessRow key={process.id} process={process} adminDashboardUrl={adminDashboardUrl} />
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
