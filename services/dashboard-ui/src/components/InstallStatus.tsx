'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { StatusBadge } from '@/components'
import type { TInstall } from '@/types'

export interface IInstallStatus {
  isCompact?: boolean
  isCompositeStatus?: boolean
  isStatusTextHidden?: boolean
}

// TODO(nnnnat): rename and remove the old install statues
// TODO(nnnnat): add polling for install statues
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
