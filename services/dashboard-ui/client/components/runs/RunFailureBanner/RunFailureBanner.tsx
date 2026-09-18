import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import type { TRunnerJob } from '@/types'
import { getStatusTheme } from '@/utils/status-utils'
import { humanize } from '@/utils/string-utils'

export interface IRunFailureBanner {
  jobs?: TRunnerJob[]
  status?: { status?: string; status_human_description?: string }
  statusDescription?: string
}

const jobLabel = (job: TRunnerJob) =>
  humanize(job?.operation ?? job?.type ?? 'Job')

export const RunFailureBanner = ({
  jobs,
  status,
  statusDescription,
}: IRunFailureBanner) => {
  const hasFailed = getStatusTheme(status?.status ?? '') === 'error'
  const failedJobs = (jobs ?? []).filter(
    (job) =>
      getStatusTheme(job?.status_v2?.status ?? job?.status ?? '') === 'error'
  )
  const reason =
    status?.status_human_description ||
    statusDescription ||
    failedJobs.at(0)?.status_description

  if (!hasFailed || !reason) return null

  return (
    <Banner theme="error">
      <div className="flex flex-col gap-1 min-w-0">
        <Text weight="strong">Run failed</Text>
        <Text>{reason}</Text>
        {failedJobs.map((job) =>
          job.status_description && job.status_description !== reason ? (
            <Text key={job.id} variant="subtext">
              {jobLabel(job)}: {job.status_description}
            </Text>
          ) : null
        )}
      </div>
    </Banner>
  )
}
