'use client'

import { usePathname } from 'next/navigation'
import React, { type FC, useEffect } from 'react'
import { StatusBadge } from '@/components/Status'
import { revalidateData } from '@/components/actions'
import type { TSandboxRun } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

export interface ISandboxRunStatus {
  descriptionAlignment?: 'center' | 'left' | 'right'
  descriptionPosition?: 'bottom' | 'top'
  initSandboxRun: TSandboxRun
  shouldPoll?: boolean
}

export const SandboxRunStatus: FC<ISandboxRunStatus> = ({
  initSandboxRun: run,
  shouldPoll = false,
  ...props
}) => {
  const path = usePathname()
  const status = run?.status_v2 || {
    status: run?.status || 'Unknown',
    status_human_description: run?.status_description || undefined,
  }

  useEffect(() => {
    const fetchSandboxRun = () => {
      revalidateData({ path })
    }
    if (shouldPoll) {
      const pollSandboxRun = setInterval(fetchSandboxRun, SHORT_POLL_DURATION)

      if (
        status.status === 'active' ||
        status.status === 'error' ||
        status.status === 'cancelled' ||
        status.status === 'not-attempted' ||
        status.status === 'noop'
      ) {
        clearInterval(pollSandboxRun)
      }

      return () => clearInterval(pollSandboxRun)
    }
  }, [run, shouldPoll])

  return (
    <StatusBadge
      description={status.status_human_description}
      status={status.status}
      label="Status"
      {...props}
    />
  )
}
