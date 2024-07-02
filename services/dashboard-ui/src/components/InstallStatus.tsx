'use client'

import React, { type FC } from 'react'
import { Status } from '@/components'
import { useInstallContext } from '@/context'

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

  return isCompositeStatus ? (
    <Status
      status={install.status}
      isStatusTextHidden={isStatusTextHidden}
    />
  ) : (
    <div className="flex flex-auto gap-6 flex-wrap">
      <Status
        status={install.sandbox_status}
        isLabelStatusText={isCompact}
        isStatusTextHidden={isStatusTextHidden}
        label={!isStatusTextHidden && 'Sandbox'}
      />
      <Status
        status={install.runner_status}
        isLabelStatusText={isCompact}
        isStatusTextHidden={isStatusTextHidden}
        label={!isStatusTextHidden && 'Runner'}
      />      
      <Status
        status={install.composite_component_status}
        isLabelStatusText={isCompact}
        isStatusTextHidden={isStatusTextHidden}
        label={!isStatusTextHidden && 'Components'}
      />
    </div>
  )
}
