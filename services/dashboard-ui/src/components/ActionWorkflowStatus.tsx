'use client'

import { usePathname } from 'next/navigation'
import React, { type FC, useEffect } from 'react'
import { StatusBadge } from '@/components/Status'
import { revalidateData } from '@/components/actions'
import type { TInstallActionWorkflowRun } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

export interface IActionWorkflowStatus {
  descriptionAlignment?: 'center' | 'left' | 'right'
  descriptionPosition?: 'bottom' | 'top'
  actionWorkflowRun: TInstallActionWorkflowRun
  shouldPoll?: boolean
}

export const ActionWorkflowStatus: FC<IActionWorkflowStatus> = ({
  actionWorkflowRun,
  shouldPoll = false,
  ...props
}) => {
  const path = usePathname()
  const status = actionWorkflowRun?.status_v2 || {
    status: actionWorkflowRun?.status || 'Unknown',
    status_human_description:
      actionWorkflowRun?.status_description || undefined,
  }

  useEffect(() => {
    const refreshData = () => {
      revalidateData({ path })
    }
    if (shouldPoll) {
      const pollBuild = setInterval(refreshData, SHORT_POLL_DURATION)

      if (
        status.status === 'active' ||
        status.status === 'error' ||
        status.status === 'cancelled' ||
        status.status === 'not-attempted' ||
        status.status === 'noop'
      ) {
        clearInterval(pollBuild)
      }

      return () => clearInterval(pollBuild)
    }
  }, [actionWorkflowRun, shouldPoll])

  return (
    <StatusBadge
      description={status.status_human_description}
      status={status.status}
      {...props}
    />
  )
}
