'use client'

import { usePathname } from 'next/navigation'
import React, { type FC, useEffect } from 'react'
import { StatusBadge } from '@/components/Status'
import { revalidateData } from '@/components/actions'
import type { TBuild } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

export interface IBuildStatus {
  descriptionAlignment?: 'center' | 'left' | 'right'
  descriptionPosition?: 'bottom' | 'top'
  initBuild: TBuild
  shouldPoll?: boolean
}

export const BuildStatus: FC<IBuildStatus> = ({
  initBuild: build,
  shouldPoll = false,
  ...props
}) => {
  const path = usePathname()
  const status = build?.status_v2 || {
    status: build?.status || 'Unknown',
    status_human_description: build?.status_description || undefined,
  }

  useEffect(() => {
    const fetchBuild = () => {
      revalidateData({ path })
    }
    if (shouldPoll) {
      const pollBuild = setInterval(fetchBuild, SHORT_POLL_DURATION)

      if (
        status?.status === 'active' ||
        status?.status === 'error' ||
        status?.status === 'cancelled' ||
        status?.status === 'not-attempted' ||
        status?.status === 'noop'
      ) {
        clearInterval(pollBuild)
      }

      return () => clearInterval(pollBuild)
    }
  }, [build, shouldPoll])

  return (
    <StatusBadge
      description={status?.status_human_description}
      status={status?.status}
      {...props}
    />
  )
}
