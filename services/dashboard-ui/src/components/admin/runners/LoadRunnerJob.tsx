'use client'

import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Status } from '@/components/common/Status'
import { Time } from '@/components/common/Time'
import { useQuery } from '@/hooks/use-query'
import { useOrg } from '@/hooks/use-org'
import type { TRunnerJob } from '@/types'

interface LoadRunnerJobProps {
  runnerId: string
  groups?: Array<'operations'>
  statuses?: Array<'finished' | 'error' | 'timed-out' | 'cancelled' | 'not-attempted'>
  title: string
}

export const LoadRunnerJob = ({ 
  runnerId, 
  groups, 
  statuses,
  title 
}: LoadRunnerJobProps) => {
  const { org } = useOrg()
  const orgId = org.id
  
  const queryParams = new URLSearchParams({
    limit: '1',
    ...(groups ? { groups: groups.join(',') } : {}),
    ...(statuses ? { statuses: statuses.join(',') } : {}),
  })

  const { data, error: queryError, isLoading } = useQuery<TRunnerJob[]>({
    path: `/api/orgs/${orgId}/runners/${runnerId}/jobs?${queryParams.toString()}`,
    dependencies: [runnerId, groups, statuses]
  })

  const job = data?.[0]
  const error = queryError ? 'Unable to load runner job' : null

  if (error) {
    return (
      <Text variant="subtext" className="text-red-600">
        {error}
      </Text>
    )
  }

  if (isLoading) {
    return (
      <div className="flex items-center gap-2">
        <Icon variant="Loading" className="animate-spin" size="16" />
        <Text variant="subtext">Loading job...</Text>
      </div>
    )
  }

  if (!job) {
    return <Text variant="subtext">No job to show.</Text>
  }

  return (
    <div className="flex items-start justify-between">
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <Status status={job.status} />
          <Text variant="base">{job.id || 'Unknown Job'}</Text>
        </div>
        <div>
          <Text variant="subtext">
            Updated: <Time time={job.updated_at} />
          </Text>
        </div>
      </div>
    </div>
  )
}