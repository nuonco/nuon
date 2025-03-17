import React, { type FC } from 'react'
import { CaretRight } from '@phosphor-icons/react/dist/ssr'
import { Config, ConfigContent } from '@/components/Config'
import { Link } from '@/components/Link'
import { StatusBadge } from '@/components/Status'
import { Text } from '@/components/Typography'
import { getRunnerJobs, type TRunnerJobGroup } from '@/lib'
import { jobHrefPath, jobName } from './helpers'

interface IRunnerRecentJob {
  orgId: string
  runnerId: string
  groups?: Array<TRunnerJobGroup>
}

export const RunnerRecentJob: FC<IRunnerRecentJob> = async ({
  orgId,
  runnerId,
  groups = ['actions', 'build', 'deploy', 'operations', 'sandbox', 'sync'],
}) => {
  const { runnerJobs } = await getRunnerJobs({
    orgId,
    runnerId,
    options: {
      limit: '1',
      groups,
      statuses: ['finished', 'failed'],
    },
  })

  const job = runnerJobs?.[0]
  const name = jobName(job)
  const hrefPath = jobHrefPath(job)

  return runnerJobs?.length ? (
    <div className="flex items-start justify-between">
      <Config>
        <ConfigContent label="Name" value={name || 'Unknown'} />

        <ConfigContent label="Group" value={job?.group} />

        <ConfigContent
          label="Status"
          value={
            <span className="flex items-center gap-2">
              <StatusBadge
                status={job?.status}
                isWithoutBorder
                isStatusTextHidden
              />
              {job?.status}
            </span>
          }
        />
      </Config>
      {job?.metadata && hrefPath !== '' ? (
        <Link className="text-sm" href={`/${orgId}/${hrefPath}`}>
          Details <CaretRight />
        </Link>
      ) : null}
    </div>
  ) : (
    <Text>No job to show.</Text>
  )
}
