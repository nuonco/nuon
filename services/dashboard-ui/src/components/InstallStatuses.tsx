'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useState } from 'react'
import { StatusBadge } from '@/components/Status'
// import { revalidateInstallData } from '@/components/install-actions'
import type { TInstall } from '@/types'
import { POLL_DURATION } from '@/utils'

export interface IInstallStatus {
  initInstall: TInstall
  isCompact?: boolean
  isCompositeStatus?: boolean
  isStatusTextHidden?: boolean
  shouldPoll?: boolean
}

export const InstallStatuses: FC<IInstallStatus> = ({
  initInstall,
  isCompact = false,
  shouldPoll = false,
}) => {
  const [install, updateInstall] = useState<TInstall>(initInstall)

  useEffect(() => {
    const fetchInstall = () => {
      fetch(`/api/${install.org_id}/installs/${install.id}`)
        .then((res) =>
          res.json().then((o) => {
            updateInstall(o)
          })
        )
        .catch(console.error)
      // revalidateInstallData({ installId: install.id, orgId: install.org_id })
    }
    if (shouldPoll) {
      const pollInstall = setInterval(fetchInstall, POLL_DURATION)
      return () => clearInterval(pollInstall)
    }
  }, [install, shouldPoll])

  return (
    <div
      className={classNames('flex', {
        'gap-6 items-center': !isCompact,
        '': isCompact,
      })}
    >
      <StatusBadge label="Sandbox" status={install.sandbox_status} />
      <StatusBadge label="Runner" status={install.runner_status} />
      <StatusBadge
        label="Components"
        status={install.composite_component_status}
      />
    </div>
  )
}
