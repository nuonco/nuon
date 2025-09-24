import { Pagination } from '@/components/Pagination'
import { RunnerUpcomingJobs } from '@/components/Runners/RunnerUpcomingJobs'
import { getRunnerJobs } from '@/lib'

export const UpcomingJobs = async ({
  runnerId,
  orgId,
  offset,
  limit = 10,
}: {
  orgId: string
  runnerId: string
  offset?: string
  limit?: number
}) => {
  const {
    data: runnerJobs,
    error,
    headers,
  } = await getRunnerJobs({
    runnerId,
    orgId,
    offset,
    limit,
    groups: ['actions', 'build', 'deploy', 'operations', 'sandbox', 'sync'],
    statuses: ['available', 'queued'],
  })

  const pageData = {
    hasNext: headers?.['x-nuon-page-next'] || 'false',
    offset: headers?.['x-nuon-page-offset'] || '0',
  }

  return runnerJobs && !error ? (
    <div>
      <RunnerUpcomingJobs
        initRunnerJobs={runnerJobs}
        runnerId={runnerId}
        offset={offset}
        limit={limit}
        shouldPoll
      />
      <Pagination
        param="upcoming-jobs"
        pageData={pageData}
        limit={limit}
        position="right"
      />
    </div>
  ) : (
    <span>Unable to load upcoming runner jobs</span>
  )
}
