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
