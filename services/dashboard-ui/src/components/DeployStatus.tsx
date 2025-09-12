'use client'

import { StatusBadge } from '@/components/Status'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TDeploy } from '@/types'

export interface IDeployStatus extends IPollingProps {
  descriptionAlignment?: 'center' | 'left' | 'right'
  descriptionPosition?: 'bottom' | 'top'
  initDeploy: TDeploy
}

export const DeployStatus = ({
  initDeploy,
  pollInterval = 5000,
  shouldPoll = false,
  ...props
}: IDeployStatus) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { data: deploy } = usePolling<TDeploy>({
    initData: initDeploy,
    path: `/api/orgs/${org.id}/installs/${install.id}/deploys/${initDeploy.id}`,
    pollInterval,
    shouldPoll,
  })
  const status = deploy?.status_v2 || {
    status: deploy?.status || 'Unknown',
    status_human_description: deploy?.status_description || undefined,
  }

  return (
    <StatusBadge
      description={status.status_human_description}
      status={status.status}
      label="Status"
      {...props}
    />
  )
}
