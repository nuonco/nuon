'use client'

import { Badge } from '@/components/old/Badge'
import { Empty } from '@/components/old/Empty'
import { Timeline } from '@/components/old/Timeline'
import { ToolTip } from '@/components/old/ToolTip'
import { Truncate } from '@/components/old/Typography'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TRunnerJob, TPaginationParams } from '@/types'
import { jobHrefPath, jobName, jobOperation } from './helpers'

interface IRunnerPastJobs extends IPollingProps, TPaginationParams {
  groups?: TRunnerJob['group'][]
  initRunnerJobs: TRunnerJob[]
  runnerId: string
}

export const RunnerPastJobs = ({
  groups = ['actions', 'build', 'deploy', 'operations', 'sandbox', 'sync'],
  initRunnerJobs,
  limit,
  offset,
  pollInterval = 10000,
  shouldPoll = false,
  runnerId,
}: IRunnerPastJobs) => {
  const { org } = useOrg()
  const params = useQueryParams({ ['past-jobs']: offset, limit, groups })
  const { data: runnerJobs } = usePolling<TRunnerJob[]>({
    dependencies: [params],
    initData: initRunnerJobs,
    path: `/api/orgs/${org.id}/runners/${runnerId}/jobs${params}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <Timeline
      emptyContent={
        <Empty
          emptyMessage="Waiting on runner to pick up jobs."
          emptyTitle="No runner jobs yet"
          variant="history"
        />
      }
      events={runnerJobs.map((job, i) => {
        const hrefPath = jobHrefPath(job)
        const name = jobName(job)

        return {
          id: job?.id,
          status: job?.status,
          underline: (
            <>
              {name ? (
                name?.length >= 12 ? (
                  <ToolTip tipContent={name} alignment="right">
                    <Truncate variant="small">{name}</Truncate>
                  </ToolTip>
                ) : (
                  name
                )
              ) : (
                <span>Unknown</span>
              )}{' '}
              /
              <span className="!inline truncate max-w-[100px]">
                {job?.group}
              </span>
              <Badge className="text-[11px]">{job?.operation}</Badge>
            </>
          ),
          time: job?.updated_at,
          href: hrefPath !== '' ? `/${org.id}/${hrefPath}` : null,
          isMostRecent: i === 0,
        }
      })}
    />
  )
}
