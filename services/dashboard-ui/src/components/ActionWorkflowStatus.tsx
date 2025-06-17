'use client'

import { usePathname } from "next/navigation"
import React, { type FC, useEffect } from 'react'
import { StatusBadge } from '@/components/Status'
import { revalidateData } from "@/components/actions"
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

  useEffect(() => {
    const refreshData = () => {
      revalidateData({ path })
    }
    if (shouldPoll) {
      const pollBuild = setInterval(refreshData, SHORT_POLL_DURATION)

      if (
        actionWorkflowRun?.status_v2?.status === 'active' ||
        actionWorkflowRun?.status_v2?.status === 'error' ||
        actionWorkflowRun?.status_v2?.status === 'cancelled' ||
        actionWorkflowRun?.status_v2?.status === 'not-attempted' ||
        actionWorkflowRun?.status_v2?.status === 'noop'
      ) {
        clearInterval(pollBuild)
      }

      return () => clearInterval(pollBuild)
    }
  }, [actionWorkflowRun, shouldPoll])

  return (
    <StatusBadge
      description={actionWorkflowRun?.status_v2?.status_human_description}
      status={actionWorkflowRun?.status_v2?.status}
      {...props}
    />
  )
}
