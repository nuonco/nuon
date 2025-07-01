'use client'

import React from 'react'
import { Link } from '@/stratus/components/common'
import {
  Timeline,
  TimelineEvent,
  type ITimeline,
} from '@/stratus/components/timeline'
import type { TRunnerJob } from '@/types'
import { getJobExecutionStatus, getJobHref, getJobName } from './helpers'

interface IRunnerRecentActivity
  extends Omit<ITimeline<TRunnerJob>, 'events' | 'renderEvent'> {
  jobs: Array<TRunnerJob>
}

export const RunnerRecentActivity = ({
  jobs,
  pagination,
}: IRunnerRecentActivity) => {
  return (
    <Timeline<TRunnerJob>
      events={jobs}
      pagination={pagination}
      renderEvent={(job) => {
        const jobHref = getJobHref(job)
        const jobTitle =
          jobHref === '' ? (
            <>
              {getJobName(job)} {getJobExecutionStatus(job)}
            </>
          ) : (
            <>
              <Link href={jobHref}>{getJobName(job)}</Link>{' '}
              {getJobExecutionStatus(job)}
            </>
          )

        return (
          <TimelineEvent
            key={job.id}
            caption={job?.id}
            createdAt={job?.created_at}
            status={job?.status}
            title={jobTitle}
          />
        )
      }}
    />
  )
}
