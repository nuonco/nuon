import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { RunnerHealthCard } from '@/components/runners/RunnerHealthCard'
import { getRunnerRecentHealthChecks } from '@/lib'
import type { TOrg } from '@/types'

export async function RunnerHealth({ org }: { org: TOrg }) {
  const runnerGroup = org?.runner_group
  const runner = runnerGroup?.runners?.at(0)
  const { data: healthchecks, error } = await getRunnerRecentHealthChecks({
    orgId: org?.id,
    runnerId: runner?.id,
  })

  return runnerGroup && runner && !error ? (
    <RunnerHealthCard
      className="flex-auto"
      initHealthchecks={healthchecks}
      runnerId={runner.id}
      shouldPoll
    />
  ) : (
    <RunnerHealthError />
  )
}

export const RunnerHealthError = () => (
  <Card className="flex-auto">
    <EmptyState
      emptyMessage="Runner health checks will display here once available."
      emptyTitle="No health check data"
      variant="diagram"
    />
  </Card>
)
