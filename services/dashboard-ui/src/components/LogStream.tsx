'use client'

import React, { type FC, useEffect, useState } from 'react'
import { Section } from '@/components/Card'
import { RunnerLogs, parseOTELLog } from '@/components/RunnerLogs'
import { Text } from '@/components/Typography'
import type { TOTELLog, TLogStream } from '@/types'
import { LOG_POLL_DURATION, SHORT_POLL_DURATION, sentanceCase } from '@/utils'

export interface ILogStreamPoller {
  heading?: string
  initLogStream: TLogStream
  initLogs: Array<TOTELLog>
  orgId: string
  logStreamId: string
  shouldPoll?: boolean
}

export const LogStreamPoller: FC<ILogStreamPoller> = ({
  initLogStream,
  shouldPoll,
  ...props
}) => {
  const [logStream, updateLogStream] = useState(initLogStream)
  const [isPolling, setPolling] = useState(shouldPoll)

  useEffect(() => {
    const fetchLogStream = () => {
      fetch(`/api/${props.orgId}/log-streams/${props.logStreamId}`)
        .then((res) => res.json().then((l) => updateLogStream(l)))
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollJob = setInterval(fetchLogStream, SHORT_POLL_DURATION)

      if (!logStream.open) {
        setPolling(false)
        clearInterval(pollJob)
      } else {
        setPolling(true)
      }

      return () => {
        clearInterval(pollJob)
      }
    }
  }, [logStream, shouldPoll])

  if (!logStream?.id) {
    return (
      <Section heading={sentanceCase(props.heading)} className="border-r">
        <Text>Waiting on log stream to start.</Text>
      </Section>
    )
  }

  return <LogPoller {...props} shouldPoll={isPolling} />
}

const LogPoller: FC<Omit<ILogStreamPoller, 'initLogStream'>> = ({
  heading = 'Logs',
  initLogs,
  orgId,
  logStreamId,
  shouldPoll = false,
}) => {
  const [logs, updateLogs] = useState(parseOTELLog(initLogs))

  useEffect(() => {
    const fetchLogs = () => {
      fetch(`/api/${orgId}/log-streams/${logStreamId}/logs`)
        .then((res) => res.json().then((l) => updateLogs(parseOTELLog(l))))
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollLogs = setInterval(fetchLogs, LOG_POLL_DURATION)

      return () => clearInterval(pollLogs)
    }
  }, [logs, shouldPoll])

  return <RunnerLogs heading={sentanceCase(heading)} logs={logs} />
}
