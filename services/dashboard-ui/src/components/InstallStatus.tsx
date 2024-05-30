'use client'

import React, { type FC } from 'react'
import { Status } from '@/components'
import { useInstallContext } from '@/context'
import { getFullInstallStatus } from '@/utils'

export interface IInstallStatus {
  isCompact?: boolean
  isCompositeStatus?: boolean
  isStatusTextHidden?: boolean
}

export const InstallStatus: FC<IInstallStatus> = ({
  isCompact = false,
  isCompositeStatus = false,
  isStatusTextHidden = false,
}) => {
  const { install } = useInstallContext()
  const status = getFullInstallStatus(install)

  return isCompositeStatus ? (
    <Status
      status={status?.installStatus?.status}
      description={!isCompact && status?.installStatus?.status_description}
      isStatusTextHidden={isStatusTextHidden}
    />
  ) : (
    <div className="flex flex-auto gap-6">
      <Status
        status={status?.sandboxStatus?.status}
        description={!isCompact && status?.sandboxStatus?.status_description}
        isLabelStatusText={isCompact}
        isStatusTextHidden={isStatusTextHidden}
        label={!isStatusTextHidden && 'Sandbox'}
      />
      <Status
        status={status?.componentStatus?.status}
        description={!isCompact && status?.componentStatus?.status_description}
        isLabelStatusText={isCompact}
        isStatusTextHidden={isStatusTextHidden}
        label={!isStatusTextHidden && 'Components'}
      />
    </div>
  )
}
