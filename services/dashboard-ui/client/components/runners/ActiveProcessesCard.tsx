import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Card, type ICard } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { getRunnerProcesses, getProcessLatestHeartbeat } from '@/lib'
import type { TRunnerProcess } from '@/types'
import { isLessThan15SecondsOld } from '@/utils/time-utils'

interface IActiveProcessesCard extends Omit<ICard, 'children'> {
  shouldPoll?: boolean
  pollInterval?: number
}

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

function ProcessHeartbeatInfo({
  runnerId,
  process,
  shouldPoll,
  pollInterval,
}: {
  runnerId: string
  process: TRunnerProcess
  shouldPoll: boolean
  pollInterval: number
}) {
  const { org } = useOrg()

  const { data: heartbeat } = useQuery({
    queryKey: ['process-heartbeat', org?.id, runnerId, process.id],
    queryFn: () =>
      getProcessLatestHeartbeat({
        orgId: org.id,
        runnerId,
        processId: process.id,
      }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!runnerId && !!process.id,
  })

  return (
    <div className="flex flex-col gap-3 rounded-md border border-neutral-100 p-3">
      <div className="flex items-center justify-between">
        <Text variant="label" weight="strong" className="capitalize">
          {process.type || 'unknown'}
        </Text>
        <Badge theme={getStatusTheme(process.status)}>{process.status}</Badge>
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
  shouldPoll = false,
  pollInterval = 10000,
  ...props
}: IActiveProcessesCard) => {
  const { org } = useOrg()
  const { runner } = useRunner()

  const { data: result, isLoading } = useQuery({
    queryKey: ['runner-processes-active', org?.id, runner?.id],
    queryFn: () =>
      getRunnerProcesses({
        orgId: org.id,
        runnerId: runner.id,
        status: 'active,offline,pending-shutdown',
        limit: 10,
      }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!runner?.id,
  })

  if (isLoading) {
    return <ActiveProcessesCardSkeleton {...props} />
  }

  const processes = result?.data ?? []

  return (
    <Card {...props}>
      <Text variant="base" weight="strong">
        Active processes
      </Text>

      {processes.length === 0 ? (
        <EmptyState
          emptyTitle="No active processes"
          emptyMessage="No runner processes are currently active."
          variant="table"
        />
      ) : (
        <div className="flex flex-col gap-3">
          {processes.map((process) => (
            <ProcessHeartbeatInfo
              key={process.id}
              runnerId={runner.id}
              process={process}
              shouldPoll={shouldPoll}
              pollInterval={pollInterval}
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
