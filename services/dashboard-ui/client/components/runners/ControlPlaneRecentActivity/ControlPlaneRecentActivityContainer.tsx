import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import type { ITimeline } from '@/components/common/Timeline'
import { useOrg } from '@/hooks/use-org'
import { getOrgRunnerJobs } from '@/lib'
import type { TRunnerJob } from '@/types'
import {
  RunnerRecentActivityComponent,
  RECENT_ACTIVITY_SEARCH_PARAM,
  RECENT_ACTIVITY_LIMIT,
} from '../RunnerRecentActivity'

const HIDDEN_JOB_TYPES = ['fetch-image-metadata']

interface IControlPlaneRecentActivityContainer
  extends Omit<ITimeline<TRunnerJob>, 'events' | 'renderEvent' | 'pagination'> {
  shouldPoll?: boolean
  pollInterval?: number
  jobDetailBasePath?: string
}

export const ControlPlaneRecentActivityContainer = ({
  shouldPoll = false,
  pollInterval = 20000,
  jobDetailBasePath,
  ...props
}: IControlPlaneRecentActivityContainer) => {
  const { org } = useOrg()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get(RECENT_ACTIVITY_SEARCH_PARAM) ?? 0)

  const { data: result, isLoading } = useQuery({
    queryKey: ['control-plane-runner-jobs', org?.id, offset],
    queryFn: () =>
      getOrgRunnerJobs({
        orgId: org!.id,
        executor: 'control-plane',
        limit: RECENT_ACTIVITY_LIMIT,
        offset,
      }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id,
  })

  const visibleJobs = (result?.data ?? []).filter(
    (job) => !HIDDEN_JOB_TYPES.includes(job.type)
  )

  return (
    <RunnerRecentActivityComponent
      jobs={visibleJobs}
      isLoading={isLoading}
      hasNext={result?.pagination?.hasNext ?? false}
      offset={offset}
      jobDetailBasePath={jobDetailBasePath}
      {...props}
    />
  )
}
