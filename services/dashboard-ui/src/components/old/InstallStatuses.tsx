'use client'

import classNames from 'classnames'
import { StatusBadge } from '@/components/old/Status'
import { useInstall } from '@/hooks/use-install'

export interface IInstallStatus {
  isCompact?: boolean
  isCompositeStatus?: boolean
  isStatusTextHidden?: boolean
}

export const InstallStatuses = ({ isCompact = false }: IInstallStatus) => {
  const { install } = useInstall()

  return (
    <div
      className={classNames('flex', {
        'gap-6 items-center': !isCompact,
        '': isCompact,
      })}
    >
      <StatusBadge
        label="Runner"
        status={install.runner_status}
        description={install?.runner_status_description}
        descriptionAlignment="right"
        descriptionPosition="bottom"
      />
      <StatusBadge
        label="Sandbox"
        status={install.sandbox_status}
        description={install?.sandbox_status_description}
        descriptionAlignment="right"
        descriptionPosition="bottom"
      />
      <StatusBadge
        label="Components"
        status={install.composite_component_status}
        description={install?.composite_component_status_description}
        descriptionAlignment="right"
        descriptionPosition="bottom"
      />
    </div>
  )
}
