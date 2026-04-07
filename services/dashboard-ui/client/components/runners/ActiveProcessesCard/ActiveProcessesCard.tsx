import { Badge } from '@/components/common/Badge'
import { Card, type ICard } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TRunnerProcess, TRunnerHeartbeat } from '@/types'
import { isLessThan15SecondsOld } from '@/utils/time-utils'

function getStatusTheme(status: string) {
  switch (status) {
    case 'active':
      return 'success' as const
    case 'offline':
      return 'warn' as const
    case 'shutting-down':
      return 'warn' as const
    case 'shut-down':
      return 'neutral' as const
    case 'error':
      return 'error' as const
    default:
      return 'neutral' as const
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

export type TProcessWithHeartbeat = {
  process: TRunnerProcess
  heartbeat?: TRunnerHeartbeat
}

interface IActiveProcessesCard extends Omit<ICard, 'children'> {
  processes?: TProcessWithHeartbeat[]
  isLoading?: boolean
}

function ProcessHeartbeatInfo({
  process,
  heartbeat,
}: TProcessWithHeartbeat) {
  return (
    <div className="flex flex-col gap-3 rounded-md border border-neutral-100 p-3">
      <div className="flex items-center justify-between">
        <Text variant="label" weight="strong" className="capitalize">
          {process.type || 'unknown'}
        </Text>
        <Badge theme={getStatusTheme(process.composite_status?.status)}>{process.composite_status?.status}</Badge>
        {process.labels?.map((label) => (
          <Badge key={label} theme="neutral">{label}</Badge>
        ))}
      </div>

      <div className="grid grid-cols-2 gap-3">
        <LabeledValue label="Uptime">
          <Text variant="subtext">{formatUptime(process.started_at)}</Text>
        </LabeledValue>

        <LabeledValue label="Version">
          <Text variant="subtext">{process.version || heartbeat?.version || '-'}</Text>
        </LabeledValue>

        <LabeledValue label="Heartbeat">
          <Text variant="subtext">
            {heartbeat?.created_at ? (
              isLessThan15SecondsOld(heartbeat.created_at) ? (
                <Badge theme="success">connected</Badge>
              ) : (
                <Time variant="subtext" time={heartbeat.created_at} />
              )
            ) : (
              '-'
            )}
          </Text>
        </LabeledValue>
      </div>
    </div>
  )
}

export const ActiveProcessesCard = ({
  processes,
  isLoading,
  ...props
}: IActiveProcessesCard) => {
  if (isLoading) {
    return <ActiveProcessesCardSkeleton {...props} />
  }

  const items = processes ?? []

  return (
    <Card {...props}>
      <Text variant="base" weight="strong">
        Active processes
      </Text>

      {items.length === 0 ? (
        <EmptyState
          emptyTitle="No active processes"
          emptyMessage="No runner processes are currently active."
          variant="table"
        />
      ) : (
        <div className="flex flex-col gap-3">
          {items.map(({ process, heartbeat }) => (
            <ProcessHeartbeatInfo
              key={process.id}
              process={process}
              heartbeat={heartbeat}
            />
          ))}
        </div>
      )}
    </Card>
  )
}

const ActiveProcessesCardSkeleton = (props: Omit<ICard, 'children'>) => (
  <Card {...props}>
    <Skeleton height="24px" width="130px" />
    <div className="flex flex-col gap-3">
      {Array.from({ length: 2 }).map((_, i) => (
        <Skeleton key={i} height="100px" />
      ))}
    </div>
  </Card>
)
