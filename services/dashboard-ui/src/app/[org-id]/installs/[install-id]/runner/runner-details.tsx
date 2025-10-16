import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { RunnerDetailsCard } from '@/components/runners/RunnerDetailsCard'
import { getRunnerById, getRunnerLatestHeartbeat } from '@/lib'
import type { TRunnerGroup, TRunnerSettings } from '@/types'

export async function RunnerDetails({
  orgId,
  runnerId,
  settings,
}: {
  orgId: string
  runnerId: string
  settings: TRunnerSettings
}) {
  const [
    { data: runnerHeartbeat, error: runnerHeartbeatError },
    { data: runner, error: runnerError },
  ] = await Promise.all([
    getRunnerLatestHeartbeat({
      orgId,
      runnerId,
    }),
    getRunnerById({
      orgId,
      runnerId,
    }),
  ])

  const error = runnerError || runnerHeartbeatError || null

  return runner && !error ? (
    <RunnerDetailsCard
      className="md:flex-initial"
      initHeartbeat={runnerHeartbeat}
      runner={runner}
      runnerGroup={settings as TRunnerGroup}
      shouldPoll
    />
  ) : (
    <RunnerDetailsError />
  )
}

export const RunnerDetailsError = () => (
  <Card className="flex-auto">
    <EmptyState
      emptyMessage="Runner details will display here once available."
      emptyTitle="No runner details"
      variant="table"
    />
  </Card>
)
