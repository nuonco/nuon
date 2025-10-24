'use client'

import classNames from 'classnames'
import { EmptyStateGraphic } from '@/components/old/EmptyStateGraphic'
import { Time } from '@/components/old/Time'
import { ToolTip } from '@/components/old/ToolTip'
import { Text } from '@/components/old/Typography'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TRunnerHealthCheck } from '@/types'

interface IRunnerHealthChart extends IPollingProps {
  runnerId: string
  initRunnerHealthChecks: TRunnerHealthCheck[]
}

export const RunnerHealthChart = ({
  initRunnerHealthChecks,
  runnerId,
  pollInterval = 60000,
  shouldPoll = false,
}: IRunnerHealthChart) => {
  const { org } = useOrg()
  const { data: healthchecks } = usePolling<TRunnerHealthCheck[]>({
    initData: initRunnerHealthChecks,
    path: `/api/orgs/${org.id}/runners/${runnerId}/health-checks`,
    pollInterval,
    shouldPoll,
  })

  const checkLength = healthchecks?.length
  const checkFirstThrid = Math.ceil(checkLength / 3)
  const checkSecondThrid = Math.ceil((checkLength * 2) / 3)

  return checkLength ? (
    <div className="flex flex-col gap-6 w-full">
      <div className="flex items-center gap-0.5">
        {healthchecks.map((healthcheck, i) => (
          <ToolTip
            key={healthcheck?.id}
            alignment={
              i < checkFirstThrid
                ? 'left'
                : i > checkSecondThrid
                  ? 'right'
                  : 'center'
            }
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
      <EmptyStateGraphic variant="diagram" />
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
