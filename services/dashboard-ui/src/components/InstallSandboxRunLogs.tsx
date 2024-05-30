'use client'

import React, { useEffect, useState, type FC } from 'react'
import { Card, Heading, Logs } from '@/components'
import { useSandboxRunContext } from '@/context'
import type { TSandboxRunLogs } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

interface ISandboxRunLogs {
  initLogs?: TSandboxRunLogs
  shouldPoll?: boolean
}

export const InstallSandboxRunLogs: FC<ISandboxRunLogs> = ({
  initLogs = [],
  shouldPoll = false,
}) => {
  const { run } = useSandboxRunContext()
  const [logs, setLogs] = useState<TSandboxRunLogs>(initLogs)

  useEffect(() => {
    const fetchLogs = () => {
      fetch(`/api/${run.org_id}/installs/${run.install_id}/runs/${run.id}/logs`)
        .then((r) => r.json().then((l) => setLogs(l)))
        .catch(console.error)
    }

    if (shouldPoll) {
      if (
        logs?.[1]?.State?.current !== 'SUCCESS' &&
        logs?.[1]?.State?.current !== 'ERROR'
      ) {
        const pollLogs = setInterval(fetchLogs, SHORT_POLL_DURATION)
        return () => clearInterval(pollLogs)
      }
    }
  }, [logs, run, shouldPoll])

  return <Logs logs={logs} />
}

export const InstallSandboxRunLogsCard: FC<
  ISandboxRunLogs & { heading?: string }
> = async ({ heading = 'Sandbox run logs', ...props }) => {
  return (
    <Card className="flex-initial">
      <Heading>{heading}</Heading>
      <InstallSandboxRunLogs {...props} />
    </Card>
  )
}
