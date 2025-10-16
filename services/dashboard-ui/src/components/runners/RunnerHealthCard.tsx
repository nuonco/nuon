'use client'

import { Card, type ICard } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TRunnerHealthCheck } from '@/types'
import { cn } from '@/utils/classnames'

interface IRunnerHealthCard extends Omit<ICard, 'children'>, IPollingProps {
  runnerId: string
  initHealthchecks: TRunnerHealthCheck[]
}

export const RunnerHealthCard = ({
  className,
  initHealthchecks,
  runnerId,
  shouldPoll = false,
  pollInterval = 60000,
  ...props
}: IRunnerHealthCard) => {
  const { org } = useOrg()
  const { data: healthchecks, error } = usePolling<TRunnerHealthCheck[]>({
    path: `/api/orgs/${org?.id}/runners/${runnerId}/health-checks`,
    shouldPoll,
    initData: initHealthchecks,
    pollInterval,
  })

  const checkLength = healthchecks?.length
  const checkFirstThrid = Math.ceil(checkLength / 3)
  const checkSecondThrid = Math.ceil((checkLength * 2) / 3)

  return checkLength ? (
    <Card className={cn('flex-auto justify-between', className)} {...props}>
      <Text variant="base" weight="strong">
        Health status
      </Text>
      <div className="flex flex-col gap-6 w-full">
        <Text variant="base">
          Last hour: {getHealthyPercentage(healthchecks)}% healthy
        </Text>
        <div className="flex items-center gap-0.5">
          {healthchecks.map((healthcheck) => (
            <Tooltip
              key={healthcheck?.id}
              position={'top'}
              // Refactored parent className to use Tailwind utility classes and arbitrary variants
              className={cn(
                'flex-auto transition-all duration-fastest ease-cubic group heartbeat-item-parent',
                // scaleY(115%) if next sibling .heartbeat-item-parent is hovered
                '[&:has(+.heartbeat-item-parent:hover)]:scale-y-[1.15]',
                // scaleY(115%) for .heartbeat-item in next sibling if this one is hovered
                '[&:hover+.heartbeat-item-parent_.heartbeat-item]:scale-y-[1.15]'
              )}
              tipContent={
                <div className="flex flex-col w-36">
                  {healthcheck?.status_code === 0 ? (
                    <>
                      <Text variant="label" weight="strong">
                        Healthy
                      </Text>
                      <Time
                        variant="subtext"
                        time={healthcheck?.minute_bucket}
                      />
                    </>
                  ) : healthcheck?.status_code === 900 ? (
                    <>
                      <Text variant="label">Unknown</Text>
                      <Text variant="subtext">No healthcheck record</Text>
                    </>
                  ) : (
                    <>
                      <Text variant="label">Unhealthy</Text>
                      <Time
                        variant="subtext"
                        time={healthcheck?.minute_bucket}
                      />
                    </>
                  )}
                </div>
              }
            >
              <div
                className={cn(
                  'flex-auto w-full h-14 rounded-xs transition-all duration-fastest ease-cubic heartbeat-item',
                  // scaleY(130%) on hover of parent (group)
                  'group-hover:scale-y-[1.3]',
                  {
                    'bg-green-500': healthcheck?.status_code === 0,
                    'bg-red-500':
                      healthcheck?.status_code !== 0 &&
                      healthcheck?.status_code !== 900,
                    'bg-cool-grey-500': healthcheck?.status_code === 900,
                  }
                )}
              />
            </Tooltip>
          ))}
        </div>
        <div className="flex items-center justify-between bg-black/5 dark:bg-white/5 px-4 py-1">
          {buildTimelineFromHealthChecks(healthchecks)?.map((healthcheck) => (
            <Time
              key={`label-${healthcheck?.id}`}
              variant="subtext"
              time={healthcheck?.minute_bucket}
              format="time-only"
            />
          ))}
        </div>
      </div>
    </Card>
  ) : (
    <RunnerHealthEmptyCard />
  )
}

export const RunnerHealthEmptyCard = ({
  title = 'No health check data',
  caption = 'Runner health check wiil display here once available',
}: {
  title?: string
  caption?: string
}) => {
  return (
    <Card className="min-w-1/2">
      <Text variant="base" weight="strong">
        Health status
      </Text>
      <EmptyState emptyMessage={caption} emptyTitle={title} variant="diagram" />
    </Card>
  )
}

function buildTimelineFromHealthChecks(
  healthchecks: TRunnerHealthCheck[]
): TRunnerHealthCheck[] {
  const length = healthchecks.length

  if (length < 5) {
    return healthchecks
  }

  const result = []
  result.push(healthchecks[0]) // First item

  const interval = (length - 2) / 4

  for (let i = 1; i <= 3; i++) {
    result.push(healthchecks[Math.round(interval * i)])
  }

  result.push(healthchecks[length - 1]) // Last item

  return result
}

function getHealthyPercentage(healthchecks: TRunnerHealthCheck[]): number {
  if (!Array.isArray(healthchecks) || healthchecks.length === 0) return 0.0

  const known = healthchecks.filter((h) => h.status_code !== 900)
  if (known.length === 0) return 0.0

  const healthy = known.filter((h) => h.status_code === 0).length
  const percent = (healthy / known.length) * 100

  // Round to two decimal places
  return Math.round(percent * 100) / 100
}
