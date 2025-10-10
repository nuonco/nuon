import { Card } from '@/components/common/Card'
import { RunnerDetailsCard } from '@/components/runners/RunnerDetailsCard'
import { Text } from '@/components/common/Text'
import { getRunnerLatestHeartbeat } from '@/lib'
import type { TOrg } from '@/types'

export async function RunnerDetails({ org }: { org: TOrg }) {
  const runnerGroup = org?.runner_group
  const runner = runnerGroup?.runners?.at(0)
  const { data: runnerHeartbeat, error } = await getRunnerLatestHeartbeat({
    orgId: org.id,
    runnerId: runner.id,
  })

  return runnerGroup && runner && !error ? (
    <RunnerDetailsCard
      className="flex-initial"
      initHeartbeat={runnerHeartbeat}
      runner={runner}
      runnerGroup={runnerGroup}
      shouldPoll
    />
  ) : (
    <RunnerError />
  )
}

export const RunnerError = () => (
  <Card className="flex-initial">
    <Text>Unable to load build runner</Text>
  </Card>
)
