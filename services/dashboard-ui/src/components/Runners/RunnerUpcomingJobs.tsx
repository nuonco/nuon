import React, { type FC } from 'react'
import {
  CancelRunnerJobButton,
  type TCancelJobType,
} from '@/components/CancelRunnerJobButton'
import { Config, ConfigContent } from '@/components/Config'
import { EmptyStateGraphic } from '@/components/EmptyStateGraphic'
import { Pagination } from '@/components/Pagination'
import { ToolTip } from '@/components/ToolTip'
import { Text, Truncate } from '@/components/Typography'
import { getRunnerJobs, type TRunnerJobGroup } from '@/lib'
import { jobName } from './helpers'

interface IRunnerUpcomingJobs {
  groups?: Array<TRunnerJobGroup>
  offset: string
  orgId: string
  runnerId: string
}

export const RunnerUpcomingJobs: FC<IRunnerUpcomingJobs> = async ({
  groups = ['build', 'deploy', 'sync', 'actions'],
  offset,
  orgId,
  runnerId,
}) => {
  const { runnerJobs, pageData } = await getRunnerJobs({
    orgId,
    runnerId,
    options: {
      groups,
      statuses: ['available', 'queued'],
      limit: '10',
      offset,
    },
  })

  return (
    <>
      {runnerJobs?.length ? (
        <div className="flex flex-col gap-6">
          <div className="divide-y flex-auto w-full">
            {runnerJobs?.map((job) => {
              const name = jobName(job)

              return (
                <div
                  className="flex items-center justify-between w-full py-3"
                  key={job.id}
                >
                  <Config>
                    <ConfigContent
                      label="Name"
                      value={
                        name ? (
                          name?.length >= 12 ? (
                            <ToolTip tipContent={name} alignment="right">
                              <Truncate variant="small">{name}</Truncate>
                            </ToolTip>
                          ) : (
                            name
                          )
                        ) : (
                          <span>Unknown</span>
                        )
                      }
                    />

                    <ConfigContent
                      label="Group"
                      value={runnerJobs?.[0]?.group}
                    />
                  </Config>
                  <div className="">
                    <CancelRunnerJobButton
                      runnerJobId={job?.id}
                      orgId={orgId}
                      jobType={job.group as TCancelJobType}
                    />
                  </div>
                </div>
              )
            })}
          </div>
          <Pagination
            param="upcoming-jobs"
            pageData={pageData}
            position="right"
          />
        </div>
      ) : (
        <div className="m-auto flex flex-col items-center max-w-[200px] my-6">
          <EmptyStateGraphic variant="table" />
          <Text className="mt-3 mb-1" variant="med-12">
            No jobs in queue yet!
          </Text>
          <Text variant="reg-12" className="text-center">
            Runner jobs will appear here as they become available and queued.
          </Text>
        </div>
      )}
    </>
  )
}
