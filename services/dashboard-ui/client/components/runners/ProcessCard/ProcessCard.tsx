import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import type {
  TRunnerProcess,
  TRunnerHealthCheck,
  TRunnerSettings,
} from '@/types'
import { cn } from '@/utils/classnames'
import { toSentenceCase } from '@/utils/string-utils'

function HealthCheckGraph({
  healthchecks,
}: {
  healthchecks: TRunnerHealthCheck[]
}) {
  if (!healthchecks?.length) {
    return (
      <div className="flex items-center justify-center h-10 rounded-md border border-white/5 bg-white/[0.02] dark:border-white/5 dark:bg-white/[0.02]">
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
              '[&:hover+.heartbeat-item-parent_.heartbeat-item]:scale-y-[1.15]'
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
                }
              )}
            />
          </Tooltip>
        ))}
      </div>
    </div>
  )
}

interface IProcessCard {
  process: TRunnerProcess
  settings?: TRunnerSettings
  isConnected: boolean
  heartbeatCreatedAt?: string
  configuredVersion: string
  reportedVersion: string
  healthchecks: TRunnerHealthCheck[]
  managementDropdown?: React.ReactNode
}

export const ProcessCard = ({
  process,
  settings,
  isConnected,
  heartbeatCreatedAt,
  configuredVersion,
  reportedVersion,
  healthchecks,
  managementDropdown,
}: IProcessCard) => {
  const warnings = process.warnings ?? []

  return (
    <Card className="min-w-0">
      <div className="flex items-start justify-between">
        <div className="flex flex-col gap-1">
          <div className="flex items-center gap-3">
            <Text variant="base" weight="strong">
              {toSentenceCase(process.type || 'unknown')} process
            </Text>
            <Status status={process.composite_status?.status} variant="badge" />
            {process.labels?.map((label) => (
              <Badge key={label} theme="neutral" variant="code" size="sm">
                {label}
              </Badge>
            ))}
          </div>
          <ID>{process.id}</ID>
          <AdminDashboardLink
            path={`/queues?owner_id=${process.runner_id}&search=runner-process-${process.id}&redirect=true`}
            label="Admin panel"
          />
        </div>
        <div className="flex items-center gap-2">
          {managementDropdown}
        </div>
      </div>

      {warnings.map((warning, i) => (
        <Banner key={i} theme="warn">
          {warning}
        </Banner>
      ))}

      <HealthCheckGraph healthchecks={healthchecks} />

      <div className="grid grid-cols-3 gap-x-6 gap-y-4">
        <LabeledValue label="Connectivity">
          <Status
            status={isConnected ? 'connected' : 'not-connected'}
            variant="badge"
          />
        </LabeledValue>

        <LabeledValue label="Uptime">
          <Duration
            variant="subtext"
            beginTime={process.started_at}
            durationUnits={['hours', 'minutes']}
            unitDisplay="short"
          />
        </LabeledValue>

        <LabeledValue label="Last heartbeat">
          {heartbeatCreatedAt ? (
            <Time
              variant="subtext"
              time={heartbeatCreatedAt}
              format="relative"
            />
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
      </div>
    </Card>
  )
}

export const ProcessCardSkeleton = () => (
  <Card className="min-w-0">
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
