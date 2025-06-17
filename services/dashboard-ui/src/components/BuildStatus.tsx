'use client'

import { usePathname } from "next/navigation"
import React, { type FC, useEffect } from 'react'
import { StatusBadge } from '@/components/Status'
import { revalidateData } from "@/components/actions"
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

  useEffect(() => {
    const fetchBuild = () => {
      revalidateData({ path })
    }
    if (shouldPoll) {
      const pollBuild = setInterval(fetchBuild, SHORT_POLL_DURATION)

      if (
        build?.status_v2?.status === 'active' ||
        build?.status_v2?.status === 'error' ||
        build?.status_v2?.status === 'cancelled' ||
        build?.status_v2?.status === 'not-attempted' ||
        build?.status_v2?.status === 'noop'
      ) {
        clearInterval(pollBuild)
      }

      return () => clearInterval(pollBuild)
    }
  }, [build, shouldPoll])

  return (
    <StatusBadge
      description={build?.status_v2?.status_human_description}
      status={build?.status_v2?.status}
      {...props}
    />
  )
}
