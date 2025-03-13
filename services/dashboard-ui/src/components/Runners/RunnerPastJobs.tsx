import React, { type FC } from 'react'
import { Pagination } from '@/components/Pagination'
import { Timeline } from '@/components/Timeline'
import { ToolTip } from '@/components/ToolTip'
import { Truncate } from '@/components/Typography'
import { getRunnerJobs, type TRunnerJobGroup } from '@/lib'
import { jobHrefPath, jobName } from './helpers'

interface IRunnerPastJobs {
  orgId: string
  runnerId: string
  offset: string
  groups?: Array<TRunnerJobGroup>
}

export const RunnerPastJobs: FC<IRunnerPastJobs> = async ({
  groups = ['build', 'deploy', 'sync', 'actions', 'operations'],
  offset,
  orgId,
  runnerId,
}) => {
  const { runnerJobs, pageData } = await getRunnerJobs({
    orgId,
    runnerId,
    options: {
      groups,
      limit: '10',
      offset,
    },
  })

  return (
    <div className="flex flex-col gap-6">
      <Timeline
        emptyTitle="No runner jobs yet"
        emptyMessage="Waiting on install runner jobs."
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
              </>
            ),
            time: job?.updated_at,
            href: hrefPath !== '' ? `/${orgId}/${hrefPath}` : null,
            isMostRecent: i === 0,
          }
        })}
      />
      <Pagination param="past-jobs" pageData={pageData} />
    </div>
  )
}
