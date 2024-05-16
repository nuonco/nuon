'use client'

import React, { type FC, useEffect, useState } from 'react'
import { Logs } from '@/components'
import type { TComponentBuildLogs } from '@/types'

export const BuildLogs: FC<{
  ssrLogs?: TComponentBuildLogs
  orgId: string
  buildId: string
  componentId: string
}> = ({ buildId, componentId, orgId, ssrLogs = [] }) => {
  const [logs, setLogs] = useState<TComponentBuildLogs>(ssrLogs)
  const fetchLogs = () => {
    fetch(`/api/${orgId}/components/${componentId}/builds/${buildId}/logs`)
      .then((r) => r.json().then((l) => setLogs(l)))
      .catch(console.error)
  }

  useEffect(() => {
    let pollLogs: NodeJS.Timeout
    if (
      logs?.[1]?.State?.current !== 'SUCCESS' &&
      logs?.[1]?.State?.current !== 'ERROR'
    ) {
      pollLogs = setInterval(fetchLogs, 2000)
    }

    return () => clearInterval(pollLogs)
  }, [logs])

  return <Logs logs={logs} />
}
