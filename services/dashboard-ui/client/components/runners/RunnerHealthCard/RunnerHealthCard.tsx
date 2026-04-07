import { Card, type ICard } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import { Text } from '@/components/common/Text'
import type { TRunnerHealthCheck } from '@/types'
import { cn } from '@/utils/classnames'

interface IRunnerHealthCard extends Omit<ICard, 'children'> {
  healthchecks?: TRunnerHealthCheck[]
  isLoading?: boolean
}

export const RunnerHealthCard = ({
  className,
  healthchecks,
  isLoading,
  ...props
}: IRunnerHealthCard) => {
  if (isLoading) {
    return <RunnerHealthCardSkeleton className={className} {...props} />
  }

  const checkLength = healthchecks?.length

  return checkLength ? (
    <Card className={cn('flex-auto justify-between', className)} {...props}>
      <Text variant="base" weight="strong">
        Health status
      </Text>
      <div className="flex flex-col gap-6 w-full">
        <div className="flex items-center gap-0.5">
          {healthchecks.map((healthcheck) => (
            <Tooltip
              key={healthcheck?.id}
              position={'top'}
              className={cn(
                'flex-auto transition-all duration-fastest ease-cubic group heartbeat-item-parent',
                '[&:has(+.heartbeat-item-parent:hover)]:scale-y-[1.15]',
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

export const RunnerHealthCardSkeleton = ({
  className,
  ...props
}: Omit<ICard, 'children'>) => {
  return (
    <Card className={cn('flex-auto justify-between', className)} {...props}>
      <Skeleton height="24px" width="98px" />

      <div className="flex flex-col gap-6 w-full">
        <Skeleton height="24px" width="180px" />

        <Skeleton height="56px" width="100%" />

        <Skeleton height="25px" width="100%" />
      </div>
    </Card>
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
    <Card className="min-w-1/2 flex-auto">
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
  result.push(healthchecks[0])

  const interval = (length - 2) / 4

  for (let i = 1; i <= 3; i++) {
    result.push(healthchecks[Math.round(interval * i)])
  }

  result.push(healthchecks[length - 1])

  return result
}
