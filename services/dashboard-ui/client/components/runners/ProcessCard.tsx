import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { ProcessManagementDropdown } from '@/components/runners/ProcessManagementDropdown'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import {
  getProcessLatestHeartbeat,
  getRunnerRecentHealthChecks,
} from '@/lib'
import type { TRunnerProcess, TRunnerHealthCheck, TRunnerSettings } from '@/types'
import { isLessThan15SecondsOld } from '@/utils/time-utils'
import { cn } from '@/utils/classnames'

function formatRelativeTime(dateStr: string | undefined): string {
  if (!dateStr) return ''
  const diffMs = Date.now() - new Date(dateStr).getTime()
  const minutes = Math.floor(diffMs / (1000 * 60))
  if (minutes < 1) return '(less than a minute ago)'
  if (minutes === 1) return '(1 minute ago)'
  return `(${minutes} minutes ago)`
}

function getProcessWarnings(warnings?: string[]): string[] {
  return [...(warnings || [])]
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

function getStatusTheme(status: string) {
  switch (status) {
    case 'active':
      return 'success' as const
    case 'offline':
      return 'warn' as const
    case 'pending-shutdown':
      return 'warn' as const
    case 'shutting-down':
      return 'warn' as const
    case 'shut-down':
      return 'neutral' as const
    case 'inactive':
      return 'neutral' as const
    case 'error':
      return 'error' as const
    default:
      return 'neutral' as const
  }
}

function HealthCheckGraph({
  healthchecks,
}: {
  healthchecks: TRunnerHealthCheck[]
}) {
  if (!healthchecks?.length) {
    return (
      <div className="flex items-center justify-center h-10 text-cool-grey-500">
        <Text variant="subtext" theme="neutral">
          No health data
        </Text>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-2 w-full">
      <div className="flex items-center gap-0.5">
        {healthchecks.map((hc) => (
          <Tooltip
            key={hc?.id}
            position="top"
            className={cn(
              'flex-auto transition-all duration-fastest ease-cubic group heartbeat-item-parent',
              '[&:has(+.heartbeat-item-parent:hover)]:scale-y-[1.15]',
              '[&:hover+.heartbeat-item-parent_.heartbeat-item]:scale-y-[1.15]',
            )}
            tipContent={
              <div className="flex flex-col w-36">
                {hc?.status_code === 0 ? (
                  <>
                    <Text variant="label" weight="strong">
                      Healthy
                    </Text>
                    <Time variant="subtext" time={hc?.minute_bucket} />
                  </>
                ) : hc?.status_code === 900 ? (
                  <>
                    <Text variant="label">Unknown</Text>
                    <Text variant="subtext">No healthcheck record</Text>
                  </>
                ) : (
                  <>
                    <Text variant="label">Unhealthy</Text>
                    <Time variant="subtext" time={hc?.minute_bucket} />
                  </>
                )}
              </div>
            }
          >
            <div
              className={cn(
                'flex-auto w-full h-8 rounded-xs transition-all duration-fastest ease-cubic heartbeat-item',
                'group-hover:scale-y-[1.3]',
                {
                  'bg-green-500': hc?.status_code === 0,
                  'bg-red-500':
                    hc?.status_code !== 0 && hc?.status_code !== 900,
                  'bg-cool-grey-500': hc?.status_code === 900,
                },
              )}
            />
          </Tooltip>
        ))}
      </div>
    </div>
  )
}

export const ProcessCard = ({
  process,
  settings,
  shouldPoll = false,
  pollInterval = 10000,
}: {
  process: TRunnerProcess
  settings?: TRunnerSettings
  shouldPoll?: boolean
  pollInterval?: number
}) => {
  const { org } = useOrg()
  const { runner } = useRunner()

  const { data: heartbeat } = useQuery({
    queryKey: ['process-heartbeat', org?.id, runner?.id, process.id],
    queryFn: () =>
      getProcessLatestHeartbeat({
        orgId: org.id,
        runnerId: runner.id,
        processId: process.id,
      }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!runner?.id && !!process.id,
  })

  const { data: healthchecks } = useQuery({
    queryKey: ['process-health-checks', org?.id, runner?.id, process.id],
    queryFn: () =>
      getRunnerRecentHealthChecks({
        orgId: org.id,
        runnerId: runner.id,
        processId: process.id,
      }),
    refetchInterval: shouldPoll ? 60000 : false,
    enabled: !!org?.id && !!runner?.id && !!process.id,
  })

  const isConnected = isLessThan15SecondsOld(heartbeat?.created_at)
  const configuredVersion = settings?.container_image_tag || '-'
  const reportedVersion = heartbeat?.version || process.version || '-'
  const warnings = getProcessWarnings(process.warnings)

  return (
    <Card className="flex-1 min-w-[320px]">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-1">
          <div className="flex items-center gap-3">
            <Text variant="base" weight="strong" className="capitalize">
              {process.type || 'unknown'} process
            </Text>
            <Badge theme={getStatusTheme(process.status)}>{process.status}</Badge>
          </div>
          <Text variant="subtext" theme="neutral" className="italic">
            {process.id}
          </Text>
        </div>
        <ProcessManagementDropdown
          process={process}
          settings={settings}
        />
      </div>

      {warnings.map((warning, i) => (
        <Banner key={i} theme="warn">
          {warning}
        </Banner>
      ))}

      <HealthCheckGraph healthchecks={healthchecks || []} />

      <div className="grid grid-cols-2 gap-4 lg:grid-cols-3">
        <LabeledValue label="Connectivity">
          <Status
            status={isConnected ? 'connected' : 'not-connected'}
            variant="badge"
          />
        </LabeledValue>

        <LabeledValue label="Uptime">
          <Text variant="subtext">{formatUptime(process.started_at)}</Text>
        </LabeledValue>

        <LabeledValue label="Last heartbeat">
          {heartbeat?.created_at ? (
            <div className="flex flex-col">
              <Time variant="subtext" time={heartbeat.created_at} />
              <Text variant="subtext" theme="neutral">
                {formatRelativeTime(heartbeat.created_at)}
              </Text>
            </div>
          ) : (
            <Text variant="subtext">-</Text>
          )}
        </LabeledValue>

        <LabeledValue label="Configured version">
          <Text variant="subtext">{configuredVersion}</Text>
        </LabeledValue>

        <LabeledValue label="Reported version">
          <Text variant="subtext">{reportedVersion}</Text>
        </LabeledValue>

        <LabeledValue label="Runner ID">
          <ID theme="default">{runner?.id}</ID>
        </LabeledValue>
      </div>
    </Card>
  )
}

export const ProcessCardSkeleton = () => (
  <Card className="flex-1 min-w-[320px]">
    <div className="flex items-center justify-between">
      <Skeleton height="24px" width="160px" />
      <Skeleton height="32px" width="80px" />
    </div>
    <Skeleton height="32px" width="100%" />
    <div className="grid grid-cols-2 gap-4 lg:grid-cols-3">
      {Array.from({ length: 6 }).map((_, i) => (
        <div key={i} className="flex flex-col gap-1">
          <Skeleton height="14px" width="60px" />
          <Skeleton height="20px" width="80px" />
        </div>
      ))}
    </div>
  </Card>
)
