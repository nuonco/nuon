import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { HealthBars } from '@/components/common/HealthBars'
import type {
  TRunnerProcess,
  TRunnerHealthCheck,
  TRunnerSettings,
} from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

function healthCheckColorClass(hc: TRunnerHealthCheck): string {
  if (hc?.status_code === 0) return 'bg-green-500'
  if (hc?.status_code === 900) return 'bg-cool-grey-500'
  return 'bg-red-500'
}

function HealthCheckTooltip({ hc }: { hc: TRunnerHealthCheck }) {
  return (
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
  )
}

function HealthCheckGraph({
  healthchecks,
}: {
  healthchecks: TRunnerHealthCheck[]
}) {
  return (
    <HealthBars
      animated
      grow
      barClassName="h-8 rounded-xs"
      emptyMessage="No health data"
      bars={(healthchecks ?? []).map((hc) => ({
        key: hc?.id ?? '',
        colorClass: healthCheckColorClass(hc),
        tooltip: <HealthCheckTooltip hc={hc} />,
      }))}
    />
  )
}

interface IProcessCard {
  process?: TRunnerProcess
  settings?: TRunnerSettings
  isConnected?: boolean
  heartbeatCreatedAt?: string
  configuredVersion?: string
  reportedVersion?: string
  healthchecks?: TRunnerHealthCheck[]
  managementDropdown?: React.ReactNode
  loading?: boolean
}

export const ProcessCard = ({
  process,
  isConnected,
  heartbeatCreatedAt,
  configuredVersion,
  reportedVersion,
  healthchecks,
  managementDropdown,
  loading,
}: IProcessCard) => {
  if (loading || !process) {
    return (
      <Card className="min-w-0">
        <div className="flex items-start justify-between">
          <div className="flex flex-col gap-1">
            <div className="flex items-center gap-3">
              <Text variant="base" weight="strong" loading loadingWidth={16} />
              <Status loading variant="badge" />
            </div>
            <ID loading loadingWidth={26} />
          </div>
        </div>
        <div className="grid grid-cols-3 gap-x-6 gap-y-4">
          <LabeledValue label="Connectivity" loading />
          <LabeledValue label="Uptime" loading />
          <LabeledValue label="Last heartbeat" loading />
          <LabeledValue label="Configured version" loading />
          <LabeledValue label="Reported version" loading />
        </div>
      </Card>
    )
  }

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

      <HealthCheckGraph healthchecks={healthchecks ?? []} />

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
