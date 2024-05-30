'use client'

import React, { type FC, useEffect, useState } from 'react'
import { Card, Heading, Logs } from '@/components'
import { useBuildContext } from '@/context'
import type { TComponentBuildLogs } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

interface IBuildLogs {
  initLogs?: TComponentBuildLogs
  shouldPoll?: boolean
}

export const BuildLogs: FC<IBuildLogs> = ({
  initLogs = [],
  shouldPoll = false,
}) => {
  const { build } = useBuildContext()
  const [logs, setLogs] = useState<TComponentBuildLogs>(initLogs)

  useEffect(() => {
    const fetchLogs = () => {
      fetch(
        `/api/${build.org_id}/components/${build.component_id}/builds/${build.id}/logs`
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
  }, [logs, build, shouldPoll])

  return <Logs logs={logs} />
}

export const BuildLogsCard: FC<IBuildLogs & { heading?: string }> = async ({
  heading = 'Build logs',
  ...props
}) => {
  return (
    <Card className="flex-initial">
      <Heading>{heading}</Heading>
      <BuildLogs {...props} />
    </Card>
  )
}
