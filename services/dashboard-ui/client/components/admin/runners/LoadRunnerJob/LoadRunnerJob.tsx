import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Status } from '@/components/common/Status'
import { Time } from '@/components/common/Time'
import type { TRunnerJob } from '@/types'

interface ILoadRunnerJob {
  job?: TRunnerJob
  error: string | null
  isLoading: boolean
  title: string
}

export const LoadRunnerJob = ({
  job,
  error,
  isLoading,
  title,
}: ILoadRunnerJob) => {
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
