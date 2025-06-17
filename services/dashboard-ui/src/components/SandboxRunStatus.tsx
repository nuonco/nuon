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

  useEffect(() => {
    const fetchSandboxRun = () => {
      revalidateData({ path })
    }
    if (shouldPoll) {
      const pollSandboxRun = setInterval(fetchSandboxRun, SHORT_POLL_DURATION)

      if (
        run?.status_v2.status === 'active' ||
        run?.status_v2.status === 'error' ||
        run?.status_v2.status === 'cancelled' ||
        run?.status_v2.status === 'not-attempted' ||
        run?.status_v2.status === 'noop'
      ) {
        clearInterval(pollSandboxRun)
      }

      return () => clearInterval(pollSandboxRun)
    }
  }, [run, shouldPoll])

  return (
    <StatusBadge
      description={run?.status_v2.status_human_description}
      status={run?.status_v2.status}
      label="Status"
      {...props}
    />
  )
}
