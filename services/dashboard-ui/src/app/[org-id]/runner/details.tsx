import { RunnerMeta } from '@/components/Runners/RunnerMeta'
import { getRunnerLatestHeartbeat } from '@/lib'
import type { TRunner, TRunnerSettings } from '@/types'

export const Details = async ({
  orgId,
  runner,
  settings,
}: {
  orgId: string
  runner: TRunner
  settings: TRunnerSettings
}) => {
  const { data: heartbeat, error } = await getRunnerLatestHeartbeat({
    orgId,
    runnerId: runner.id,
  })

  return heartbeat && !error ? (
    <RunnerMeta
      initHeartbeat={heartbeat}
      initRunner={runner}
      initSettings={settings}
      shouldPoll
    />
  ) : (
    <span>Unable to load runner details</span>
  )
}
