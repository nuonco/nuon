'use client'

import React, { type FC, useEffect, useState } from 'react'
import { StatusBadge } from '@/components/Status'
import type { TInstallDeploy } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

export interface IDeployStatus {
  descriptionAlignment?: 'center' | 'left' | 'right'
  descriptionPosition?: 'bottom' | 'top'
  initDeploy: TInstallDeploy
  shouldPoll?: boolean
}

export const DeployStatus: FC<IDeployStatus> = ({
  initDeploy,
  shouldPoll = false,
  ...props
}) => {
  const [deploy, updateDeploy] = useState<TInstallDeploy>(initDeploy)

  useEffect(() => {
    const fetchDeploy = () => {
      fetch(
        `/api/${deploy?.org_id}/installs/${deploy?.install_id}/deploys/${deploy?.id}`
      )
        .then((res) =>
          res.json().then((d) => {
            updateDeploy(d)
          })
        )
        .catch(console.error)
    }
    if (shouldPoll) {
      const pollDeploy = setInterval(fetchDeploy, SHORT_POLL_DURATION)

      if (
        deploy?.status === 'active' ||
        deploy?.status === 'error' ||
        deploy?.status === 'failed' ||
        deploy?.status === 'noop'
      ) {
        clearInterval(pollDeploy)
      }

      return () => clearInterval(pollDeploy)
    }
  }, [deploy, shouldPoll])

  return (
    <StatusBadge
      description={deploy?.status_description}
      status={deploy?.status}
      {...props}
    />
  )
}
