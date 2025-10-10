'use client'

import { Link } from '@/components/common/Link'
import { Timeline, type ITimeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TRunnerJob } from '@/types'
import {
  getJobExecutionStatus,
  getJobHref,
  getJobName,
  type TJobGroup,
} from '@/utils/runner-utils'

export const RECENT_ACTIVITY_SEARCH_PARAM = 'recent-activity'
export const RECENT_ACTIVITY_LIMIT = 10
export const RECENT_ACTIVITY_GROUPS: TJobGroup[] = [
  'actions',
  'build',
  'deploy',
  'operations',
  'sandbox',
  'sync',
]

interface IRunnerRecentActivity
  extends Omit<ITimeline<TRunnerJob>, 'events' | 'renderEvent'>,
    IPollingProps {
  initJobs: Array<TRunnerJob>
  runnerId: string
}

export const RunnerRecentActivity = ({
  initJobs,
  pagination,
  runnerId,
  shouldPoll = false,
  pollInterval = 20000,
}: IRunnerRecentActivity) => {
  const { org } = useOrg()
  const queryParams = useQueryParams({
    offset: pagination?.offset,
    limit: 10,
  })
  const { data: jobs } = usePolling<TRunnerJob[]>({
    path: `/api/orgs/${org?.id}/runners/${runnerId}/jobs${queryParams}`,
    shouldPoll,
    initData: initJobs,
    pollInterval,
  })

  return (
    <Timeline<TRunnerJob>
      events={jobs}
      pagination={pagination}
      renderEvent={(job) => {
        const jobHref = getJobHref(job)
        const jobTitle =
          jobHref === '' ? (
            <>
              {getJobName(job)} {getJobExecutionStatus(job)}
            </>
          ) : (
            <>
              <Link href={jobHref}>{getJobName(job)}</Link>{' '}
              {getJobExecutionStatus(job)}
            </>
          )

        return (
          <TimelineEvent
            key={job.id}
            caption={job?.id}
            createdAt={job?.created_at}
            status={job?.status}
            title={jobTitle}
          />
        )
      }}
    />
  )
}
