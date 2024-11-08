'use client'

import React, { type FC, useEffect, useState } from 'react'
import { StatusBadge } from '@/components/Status'
import type { TSandboxRun } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

export interface ISandboxRunStatus {
  descriptionAlignment?: 'center' | 'left' | 'right'
  descriptionPosition?: 'bottom' | 'top'
  initSandboxRun: TSandboxRun
  shouldPoll?: boolean
}

export const SandboxRunStatus: FC<ISandboxRunStatus> = ({
  initSandboxRun,
  shouldPoll = false,
  ...props
}) => {
  const [run, updateSandboxRun] = useState<TSandboxRun>(initSandboxRun)

  useEffect(() => {
    const fetchSandboxRun = () => {
      fetch(`/api/${run.org_id}/installs/${run.install_id}/runs/${run.id}`)
        .then((res) =>
          res.json().then((r) => {
            updateSandboxRun(r)
          })
        )
        .catch(console.error)
    }
    if (shouldPoll) {
      const pollSandboxRun = setInterval(fetchSandboxRun, SHORT_POLL_DURATION)

      if (
        run?.status === 'active' ||
        run?.status === 'error' ||
        run?.status === 'failed' ||
        run?.status === 'noop'
      ) {
        clearInterval(pollSandboxRun)
      }

      return () => clearInterval(pollSandboxRun)
    }
  }, [run, shouldPoll])

  return (
    <StatusBadge
      description={run?.status_description}
      status={run?.status}
      label="Status"
      {...props}
    />
  )
}
