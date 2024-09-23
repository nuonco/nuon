'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { Status, StatusBadge } from '@/components'
import { useInstallContext } from '@/context'
import type { TInstall } from '@/types'

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
    <Status status={install.status} isStatusTextHidden={isStatusTextHidden} />
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

// TODO(nnnnat): rename and remove the old install statues
export const InstallStatuesV2: FC<IInstallStatus & { install: TInstall }> = ({
  install,
  isCompact = false,
}) => {
  return (
    <div className={classNames("flex", {
      'gap-6 items-center': !isCompact,
      '': isCompact,
    })}>
      <StatusBadge label="Sandbox" status={install.sandbox_status} />
      <StatusBadge label="Runner" status={install.runner_status} />
      <StatusBadge label="Components" status={install.composite_component_status} />
    </div>
  )
}
