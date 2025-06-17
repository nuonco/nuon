'use client'

import { usePathname } from 'next/navigation'
import React, { type FC, useEffect } from 'react'
import { StatusBadge } from '@/components/Status'
import { revalidateData } from "@/components/actions"
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
  
  useEffect(() => {
    const fetchDeploy = () => {
      revalidateData({ path })
    }
    if (shouldPoll) {
      const pollDeploy = setInterval(fetchDeploy, SHORT_POLL_DURATION)

      if (
        deploy?.status_v2.status === 'active' ||
        deploy?.status_v2.status === 'error' ||
        deploy?.status_v2.status === 'cancelled' ||
        deploy?.status_v2.status === 'not-attempted' ||
        deploy?.status_v2.status === 'noop'
      ) {
        clearInterval(pollDeploy)
      }

      return () => clearInterval(pollDeploy)
    }
  }, [deploy, shouldPoll])

  return (
    <StatusBadge
      description={deploy?.status_v2.status_human_description}
      status={deploy?.status_v2.status}
      label="Status"
      {...props}
    />
  )
}
