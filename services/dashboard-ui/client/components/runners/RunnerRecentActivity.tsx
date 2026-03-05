import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Timeline, type ITimeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { getRunnerJobs } from '@/lib'
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
const HIDDEN_JOB_TYPES = ['fetch-image-metadata']

interface IRunnerRecentActivity
  extends Omit<ITimeline<TRunnerJob>, 'events' | 'renderEvent' | 'pagination'> {
  shouldPoll?: boolean
  pollInterval?: number
}

export const RunnerRecentActivity = ({
  shouldPoll = false,
  pollInterval = 20000,
  ...props
}: IRunnerRecentActivity) => {
  const { org } = useOrg()
  const { runner } = useRunner()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get(RECENT_ACTIVITY_SEARCH_PARAM) ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: ['runner-jobs', org?.id, runner?.id, offset, RECENT_ACTIVITY_GROUPS],
    queryFn: () =>
      getRunnerJobs({
        orgId: org.id,
        runnerId: runner.id,
        groups: RECENT_ACTIVITY_GROUPS,
        limit: RECENT_ACTIVITY_LIMIT,
        offset,
      }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!runner?.id,
  })

  if (isLoading) {
    return (
      <>
        <Skeleton height="24px" width="110px" />
        <TimelineSkeleton eventCount={10} />
      </>
    )
  }

  const visibleJobs = (result?.data ?? [])
    .filter((job) => !HIDDEN_JOB_TYPES.includes(job.type))

  return (
    <Timeline<TRunnerJob>
      events={visibleJobs}
      pagination={{
        hasNext: result?.pagination?.hasNext ?? false,
        offset,
        limit: RECENT_ACTIVITY_LIMIT,
        param: RECENT_ACTIVITY_SEARCH_PARAM,
      }}
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
