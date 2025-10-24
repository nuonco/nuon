import classNames from 'classnames'
import { type FC } from 'react'
import { EmptyStateGraphic } from '@/components/old/EmptyStateGraphic'
import { Time } from '@/components/old/Time'
import { ToolTip } from '@/components/old/ToolTip'
import { Text } from '@/components/old/Typography'
import type { TRunnerHealthCheck } from '@/types'

export const RunnerHeartbeatChart: FC<{
  healthchecks?: Array<TRunnerHealthCheck>
}> = ({ healthchecks = [] }) => {
  return healthchecks?.length ? (
    <div className="flex flex-col gap-6 w-full">
      <div className="flex items-center gap-0.5">
        {healthchecks.map((healthcheck, i) => (
          <ToolTip
            key={healthcheck?.id}
            alignment={i <= 9 ? 'left' : i >= 49 ? 'right' : 'center'}
            parentClassName="flex-auto heartbeat-item-parent"
            tipContent={
              healthcheck?.status_code === 0 ? (
                <>
                  <Text variant="med-12">Healthy</Text>
                  <Time time={healthcheck?.minute_bucket} />
                </>
              ) : healthcheck?.status_code === 900 ? (
                <>
                  <Text variant="med-12">Unknown</Text>
                  <Text>No healthcheck record</Text>
                </>
              ) : (
                <>
                  <Text variant="med-12">Unhealthy</Text>
                  <Time time={healthcheck?.minute_bucket} />
                </>
              )
            }
            isIconHidden
          >
            <div
              className={classNames(
                'flex-auto w-full h-[46px] rounded-sm heartbeat-item',
                {
                  'bg-green-500': healthcheck?.status_code === 0,
                  'bg-red-500':
                    healthcheck?.status_code !== 0 &&
                    healthcheck?.status_code !== 900,
                  'bg-cool-grey-500': healthcheck?.status_code === 900,
                }
              )}
            />
          </ToolTip>
        ))}
      </div>
      <div className="flex items-center justify-between bg-black/5 dark:bg-white/5 px-4 py-1">
        {buildTimelineFromHealthChecks(healthchecks)?.map((healthcheck) => (
          <Time
            key={`label-${healthcheck?.id}`}
            variant="med-12"
            time={healthcheck?.minute_bucket}
            format="time-only"
          />
        ))}
      </div>
    </div>
  ) : (
    <div className="m-auto flex flex-col items-center max-w-[200px] my-6">
      <EmptyStateGraphic />
      <Text className="mt-6" variant="med-14">
        No health check data
      </Text>
      <Text variant="reg-12" className="text-center">
        Runner health checks will display here once available.
      </Text>
    </div>
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
