'use client'

import {
  CancelRunnerJobButton,
  type TCancelJobType,
} from '@/components/CancelRunnerJobButton'
import { Config, ConfigContent } from '@/components/Config'
import { EmptyStateGraphic } from '@/components/EmptyStateGraphic'
import { ToolTip } from '@/components/ToolTip'
import { Text, Truncate } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TRunnerJob, TPaginationParams } from '@/types'
import { jobName } from './helpers'

interface IRunnerUpcomingJobs extends IPollingProps, TPaginationParams {
  initRunnerJobs: TRunnerJob[]
  runnerId: string
}

export const RunnerUpcomingJobs = ({
  initRunnerJobs,
  limit,
  offset,
  pollInterval = 5000,
  runnerId,
  shouldPoll = false,
}: IRunnerUpcomingJobs) => {
  const { org } = useOrg()
  const params = useQueryParams({
    ['upcoming-jobs']: offset,
    limit,
    groups: ['actions', 'build', 'deploy', 'operations', 'sandbox', 'sync'],
    statuses: ['available', 'queued'],
  })
  const { data: runnerJobs, error } = usePolling<TRunnerJob[]>({
    dependencies: [params],
    initData: initRunnerJobs,
    path: `/api/orgs/${org.id}/runners/${runnerId}/jobs${params}`,
    pollInterval,
    shouldPoll,
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
                      orgId={org.id}
                      jobType={job.group as TCancelJobType}
                    />
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      ) : (
        <div className="m-auto flex flex-col items-center max-w-[200px] my-6">
          <EmptyStateGraphic variant="table" isSmall />
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
