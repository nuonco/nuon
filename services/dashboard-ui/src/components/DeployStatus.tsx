'use client'

import { usePathname } from 'next/navigation'
import React, { type FC, useEffect } from 'react'
import { StatusBadge } from '@/components/Status'
import { revalidateData } from '@/components/actions'
import type { TInstallDeploy } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

export interface IDeployStatus {
  descriptionAlignment?: 'center' | 'left' | 'right'
  descriptionPosition?: 'bottom' | 'top'
  initDeploy: TInstallDeploy
  shouldPoll?: boolean
}

export const DeployStatus: FC<IDeployStatus> = ({
  initDeploy: deploy,
  shouldPoll = false,
  ...props
}) => {
  const path = usePathname()
  const status = deploy?.status_v2 || {
    status: deploy?.status || 'Unknown',
    status_human_description: deploy?.status_description || undefined,
  }

  useEffect(() => {
    const fetchDeploy = () => {
      revalidateData({ path })
    }
    if (shouldPoll) {
      const pollDeploy = setInterval(fetchDeploy, SHORT_POLL_DURATION)

      if (
        status.status === 'active' ||
        status.status === 'error' ||
        status.status === 'cancelled' ||
        status.status === 'not-attempted' ||
        status.status === 'noop'
      ) {
        clearInterval(pollDeploy)
      }

      return () => clearInterval(pollDeploy)
    }
  }, [deploy, shouldPoll])

  return (
    <StatusBadge
      description={status.status_human_description}
      status={status.status}
      label="Status"
      {...props}
    />
  )
}
