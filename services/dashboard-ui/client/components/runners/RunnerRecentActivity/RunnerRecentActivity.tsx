import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Timeline, type ITimeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import type { TRunnerJob } from '@/types'
import {
  getJobExecutionStatus,
  getJobHref,
  getJobName,
} from '@/utils/runner-utils'

export const RECENT_ACTIVITY_SEARCH_PARAM = 'recent-activity'
export const RECENT_ACTIVITY_LIMIT = 10

interface IRunnerRecentActivity
  extends Omit<ITimeline<TRunnerJob>, 'events' | 'renderEvent' | 'pagination'> {
  jobs: TRunnerJob[]
  isLoading: boolean
  hasNext: boolean
  offset: number
  jobDetailBasePath?: string
}

export const RunnerRecentActivity = ({
  jobs,
  isLoading,
  hasNext,
  offset,
  jobDetailBasePath,
  ...props
}: IRunnerRecentActivity) => {
  if (isLoading) {
    return (
      <>
        <Skeleton height="24px" width="110px" />
        <TimelineSkeleton eventCount={10} />
      </>
    )
  }

  return (
    <Timeline<TRunnerJob>
      events={jobs}
      pagination={{
        hasNext,
        offset,
        limit: RECENT_ACTIVITY_LIMIT,
        param: RECENT_ACTIVITY_SEARCH_PARAM,
      }}
      renderEvent={(job) => {
        const jobHref = getJobHref(job)
        const resolvedHref =
          jobHref === '' && jobDetailBasePath
            ? `${jobDetailBasePath}/jobs/${job.id}`
            : jobHref
        const jobTitle =
          resolvedHref === '' ? (
            <>
              {getJobName(job)} {getJobExecutionStatus(job)}
            </>
          ) : (
            <>
              <Link href={resolvedHref}>{getJobName(job)}</Link>{' '}
              {getJobExecutionStatus(job)}
            </>
          )

        return (
          <TimelineEvent
            key={job.id}
            caption={<ID>{job?.id}</ID>}
            createdAt={job?.created_at}
            status={job?.status}
            title={jobTitle}
          />
        )
      }}
      {...props}
    />
  )
}
