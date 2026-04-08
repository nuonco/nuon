import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getRunnerJobs } from '@/lib'
import { LoadRunnerJob } from './LoadRunnerJob'

interface LoadRunnerJobContainerProps {
  runnerId: string
  groups?: Array<'operations'>
  statuses?: Array<'finished' | 'failed' | 'timed-out' | 'cancelled' | 'not-attempted'>
  title: string
}

export const LoadRunnerJobContainer = ({
  runnerId,
  groups,
  statuses,
  title,
}: LoadRunnerJobContainerProps) => {
  const { org } = useOrg()
  const orgId = org.id

  const { data, error: queryError, isLoading } = useQuery({
    queryKey: ['runner-jobs', orgId, runnerId, groups, statuses],
    queryFn: () =>
      getRunnerJobs({
        orgId,
        runnerId,
        groups,
        statuses,
        limit: 1,
      }),
    enabled: !!orgId && !!runnerId,
  })

  return (
    <LoadRunnerJob
      job={data?.data?.[0]}
      error={queryError ? 'Unable to load runner job' : null}
      isLoading={isLoading}
      title={title}
    />
  )
}
