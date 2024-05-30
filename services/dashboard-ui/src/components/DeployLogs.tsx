'use client'

import React, { useEffect, useState, type FC } from 'react'
import { Card, Heading, Logs } from '@/components'
import { useInstallDeployContext } from '@/context'
import type { TInstallDeployLogs } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

interface IDeployLogs {
  initLogs?: TInstallDeployLogs
  shouldPoll?: boolean
}

export const DeployLogs: FC<IDeployLogs> = ({
  initLogs = [],
  shouldPoll = false,
}) => {
  const { deploy } = useInstallDeployContext()
  const [logs, setLogs] = useState<TInstallDeployLogs>(initLogs)

  useEffect(() => {
    const fetchLogs = () => {
      fetch(
        `/api/${deploy.org_id}/installs/${deploy.install_id}/deploys/${deploy.id}/logs`
      )
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
  }, [logs, deploy, shouldPoll])

  return <Logs logs={logs} />
}

export const DeployLogsCard: FC<IDeployLogs & { heading?: string }> = async ({
  heading = 'Deploy logs',
  ...props
}) => {
  return (
    <Card className="flex-initial">
      <Heading>{heading}</Heading>
      <DeployLogs {...props} />
    </Card>
  )
}
