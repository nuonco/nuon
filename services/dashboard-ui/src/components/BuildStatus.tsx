'use client'

import React, { type FC, useEffect, useState } from 'react'
import { StatusBadge } from '@/components/Status'
import type { TBuild } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

export interface IBuildStatus {
  descriptionAlignment?: 'center' | 'left' | 'right'
  descriptionPosition?: 'bottom' | 'top'
  initBuild: TBuild
  shouldPoll?: boolean
}

export const BuildStatus: FC<IBuildStatus> = ({
  initBuild,
  shouldPoll = false,
  ...props
}) => {
  const [build, updateBuild] = useState<TBuild>(initBuild)

  useEffect(() => {
    const fetchBuild = () => {
      fetch(
        `/api/${build?.org_id}/components/${build.component_id}/builds/${build.id}`
      )
        .then((res) =>
          res.json().then((b) => {
            updateBuild(b)
          })
        )
        .catch(console.error)
    }
    if (shouldPoll) {
      const pollBuild = setInterval(fetchBuild, SHORT_POLL_DURATION)

      if (
        build?.status === 'active' ||
        build?.status === 'error' ||
        build?.status === 'failed' ||
        build?.status === 'noop'
      ) {
        clearInterval(pollBuild)
      }

      return () => clearInterval(pollBuild)
    }
  }, [build, shouldPoll])

  return (
    <StatusBadge
      description={build?.status_description}
      status={build?.status}
      {...props}
    />
  )
}
