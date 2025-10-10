import { Pagination } from '@/components/Pagination'
import { RunnerPastJobs } from '@/components/OldRunners/RunnerPastJobs'
import { getRunnerJobs } from '@/lib'

export const Activity = async ({
  orgId,
  runnerId,
  offset,
  limit = 10,
}: {
  runnerId: string
  orgId: string
  offset?: string
  limit?: number
}) => {
  const {
    data: runnerJobs,
    error,
    headers,
  } = await getRunnerJobs({
    orgId,
    runnerId,
    groups: ['actions', 'build', 'deploy', 'operations', 'sandbox', 'sync'],
    limit: 10,
    offset,
  })

  const pageData = {
    hasNext: headers?.['x-nuon-page-next'] || 'false',
    offset: headers?.['x-nuon-page-offset'] || '0',
  }

  return runnerJobs && !error ? (
    <div className="flex flex-col gap-6">
      <RunnerPastJobs
        runnerId={runnerId}
        initRunnerJobs={runnerJobs}
        offset={offset}
        limit={limit}
        shouldPoll
      />
      <Pagination param="past-jobs" pageData={pageData} limit={limit} />
    </div>
  ) : (
    <span>Error loading runner jobs</span>
  )
}
