import React, { type FC } from 'react'
import { Heartbeat, Timer } from '@phosphor-icons/react/dist/ssr'
import { Duration, Time } from '@/components/Time'
import { Text } from '@/components/Typography'
import { getRunnerLatestHeartbeat } from '@/lib'

interface IRunnerHeartbeat {
  orgId: string
  runnerId: string
  runnerType: string
}

export const RunnerHeartbeat: FC<IRunnerHeartbeat> = async ({
  orgId,
  runnerId,
  runnerType,
}) => {
  const runnerHeartbeat = await getRunnerLatestHeartbeat({
    orgId,
    runnerId,
  }).catch(console.error)

  return runnerHeartbeat ? (
    <>
      <div>
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Last heartbeat
        </Text>
        <Text>
          <Heartbeat size={14} />
          <Time
            time={runnerHeartbeat.created_at}
            format="relative"
            variant="med-12"
          />
        </Text>
      </div>
      <div>
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Alive time
        </Text>
        <Text>
          <Timer size={14} />
          <Duration nanoseconds={runnerHeartbeat.alive_time} variant="med-12" />
        </Text>
      </div>
      <div>
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Version
        </Text>
        <Text variant="med-12">{runnerHeartbeat?.version}</Text>
      </div>
      <div>
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">Type</Text>
        <Text>{runnerType}</Text>
      </div>
    </>
  ) : (
    <Text>No runner heartbeat found</Text>
  )
}
