'use client'

import React, { type FC } from 'react'
import { StatusBadge } from '@/components/old/Status'
import type { TOrg } from '@/types'

export interface IOrgStatus {
  descriptionAlignment?: 'center' | 'left' | 'right'
  descriptionPosition?: 'bottom' | 'top'
  initOrg: TOrg
  shouldPoll?: boolean
}

export const OrgStatus: FC<IOrgStatus> = ({
  initOrg: org,
  shouldPoll = false,
  ...props
}) => {
  return (
    <StatusBadge
      description={org?.status_description}
      status={org?.status}
      {...props}
      isWithoutBorder
    />
  )
}
