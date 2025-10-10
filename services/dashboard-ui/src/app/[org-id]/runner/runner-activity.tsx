import { RunnerRecentActivity } from '@/components/runners/RunnerRecentActivity'
import { Text } from '@/components/common/Text'
import { getRunnerJobs } from '@/lib'
import type { TOrg } from '@/types'

interface ILoadRunnerRecentActivity {
  offset: string
  org: TOrg
}

export async function RunnerActivity({
  org,
  offset,
}: ILoadRunnerRecentActivity) {
  const runnerGroup = org?.runner_group
  const runner = runnerGroup?.runners?.at(0)
  const {
    data: jobs,
    error,
    headers,
  } = await getRunnerJobs({
    orgId: org.id,
    runnerId: runner.id,
    groups: ['actions', 'build', 'deploy', 'operations', 'sandbox', 'sync'],
    limit: 10,
    offset,
  })

  const pagination = {
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  return runnerGroup && runner && jobs && !error ? (
    <>
      <Text variant="base" weight="strong">
        Recent activity
      </Text>
      <RunnerRecentActivity
        initJobs={jobs}
        pagination={pagination}
        runnerId={runner?.id}
        shouldPoll
      />
    </>
  ) : (
    <RunnerActivityError />
  )
}

export const RunnerActivityError = () => (
  <div className="w-full">
    <Text>Error fetching recenty runner activity </Text>
  </div>
)
