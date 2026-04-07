import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import type { ITimeline } from '@/components/common/Timeline'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { getRunnerJobs } from '@/lib'
import type { TRunnerJob } from '@/types'
import type { TJobGroup } from '@/utils/runner-utils'
import {
  RunnerRecentActivity,
  RECENT_ACTIVITY_SEARCH_PARAM,
  RECENT_ACTIVITY_LIMIT,
} from './RunnerRecentActivity'

export const RECENT_ACTIVITY_GROUPS: TJobGroup[] = [
  'actions',
  'build',
  'deploy',
  'operations',
  'sandbox',
  'sync',
]
const HIDDEN_JOB_TYPES = ['fetch-image-metadata']

interface IRunnerRecentActivityContainer
  extends Omit<ITimeline<TRunnerJob>, 'events' | 'renderEvent' | 'pagination'> {
  shouldPoll?: boolean
  pollInterval?: number
  jobDetailBasePath?: string
}

export const RunnerRecentActivityContainer = ({
  shouldPoll = false,
  pollInterval = 20000,
  jobDetailBasePath,
  ...props
}: IRunnerRecentActivityContainer) => {
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

  const visibleJobs = (result?.data ?? [])
    .filter((job) => !HIDDEN_JOB_TYPES.includes(job.type))

  return (
    <RunnerRecentActivity
      jobs={visibleJobs}
      isLoading={isLoading}
      hasNext={result?.pagination?.hasNext ?? false}
      offset={offset}
      jobDetailBasePath={jobDetailBasePath}
      {...props}
    />
  )
}
