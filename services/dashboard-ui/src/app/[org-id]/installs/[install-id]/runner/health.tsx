import { RunnerHealthChart } from '@/components'
import { getRunnerRecentHealthChecks } from '@/lib'

export const Health = async ({
  runnerId,
  orgId,
}: {
  orgId: string
  runnerId: string
}) => {
  const { data: healthchecks, error } = await getRunnerRecentHealthChecks({
    orgId,
    runnerId,
  })
  return healthchecks && !error ? (
    <RunnerHealthChart
      initRunnerHealthChecks={healthchecks}
      runnerId={runnerId}
      shouldPoll
    />
  ) : (
    <span>Unable to load runner health checks</span>
  )
}
