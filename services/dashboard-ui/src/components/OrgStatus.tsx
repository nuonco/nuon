'use client'

import React, { type FC, useEffect, useState } from 'react'
import { StatusBadge } from '@/components/Status'
import type { TOrg } from '@/types'
import { POLL_DURATION } from '@/utils'

export interface IOrgStatus {
  descriptionAlignment?: 'center' | 'left' | 'right'
  descriptionPosition?: 'bottom' | 'top'
  initOrg: TOrg
  shouldPoll?: boolean
}

export const OrgStatus: FC<IOrgStatus> = ({
  initOrg,
  shouldPoll = false,
  ...props
}) => {
  const [org, updateOrg] = useState<TOrg>(initOrg)

  useEffect(() => {
    const fetchOrg = () => {
      fetch(`/api/${initOrg.id}`)
        .then((res) =>
          res.json().then((o) => {
            updateOrg(o)
          })
        )
        .catch(console.error)
    }
    if (shouldPoll) {
      const pollOrg = setInterval(fetchOrg, POLL_DURATION)

      /* if (
       *   org?.status === 'active' ||
       *   org?.status === 'error' ||
       *   org?.status === 'failed' ||
       *   org?.status === 'noop'
       * ) {
       *   clearInterval(pollOrg)
       * } */

      return () => clearInterval(pollOrg)
    }
  }, [org, shouldPoll])

  return (
    <StatusBadge
      description={org?.status_description}
      status={org?.status}
      {...props}
      isWithoutBorder
    />
  )
}
