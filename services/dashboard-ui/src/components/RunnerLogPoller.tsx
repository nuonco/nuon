'use client'

import React, { type FC, useEffect, useState } from 'react'
import { RunnerLogs, parseOTELLog } from '@/components/RunnerLogs'
import type { TOTELLog, TRunnerJob } from '@/types'
import { LOG_POLL_DURATION, SHORT_POLL_DURATION, sentanceCase } from '@/utils'

export interface IRunnerLogsPoller {
  heading?: string
  initJob: TRunnerJob
  initLogs: Array<TOTELLog>
  jobId: string
  orgId: string
  runnerId: string
  shouldPoll?: boolean
}

export const RunnerLogsPoller: FC<IRunnerLogsPoller> = ({
  initJob,
  shouldPoll,
  ...props
}) => {
  const [job, updateJob] = useState(initJob)
  const [isPolling, setPolling] = useState(shouldPoll)

  useEffect(() => {
    const fetchRunnerJob = () => {
      fetch(`/api/${props.orgId}/runner-jobs/${props.jobId}`)
        .then((res) => res.json().then((j) => updateJob(j)))
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollJob = setInterval(fetchRunnerJob, SHORT_POLL_DURATION)

      if (
        (job?.status === 'finished' ||
          job?.status === 'failed' ||
          job?.status === 'timed-out' ||
          job?.status === 'not-attempted' ||
          job?.status === 'cancelled' ||
          job?.status === 'unknown') &&
        pollJob
      ) {
        setPolling(false)
        clearInterval(pollJob)
      } else {
        setPolling(true)
      }

      return () => {
        clearInterval(pollJob)
      }
    }
  }, [job, shouldPoll])

  return <LogPoller {...props} shouldPoll={isPolling} />
}

const LogPoller: FC<Omit<IRunnerLogsPoller, 'initJob'>> = ({
  heading = 'Logs',
  initLogs,
  jobId,
  orgId,
  runnerId,
  shouldPoll = false,
}) => {
  const [logs, updateLogs] = useState(parseOTELLog(initLogs))

  useEffect(() => {
    const fetchRunnerLogs = () => {
      fetch(`/api/${orgId}/runners/${runnerId}/logs?job_id=${jobId}`)
        .then((res) => res.json().then((l) => updateLogs(parseOTELLog(l))))
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollLogs = setInterval(fetchRunnerLogs, LOG_POLL_DURATION)

      return () => clearInterval(pollLogs)
    }
  }, [logs, shouldPoll])

  return <RunnerLogs heading={sentanceCase(heading)} logs={logs} />
}
