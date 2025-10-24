'use client'

import { StatusBadge } from '@/components/old/Status'
import { useInstallActionRun } from '@/hooks/use-install-action-run'

export interface IActionWorkflowStatus {
  descriptionAlignment?: 'center' | 'left' | 'right'
  descriptionPosition?: 'bottom' | 'top'
  shouldPoll?: boolean
}

export const ActionWorkflowStatus = ({
  shouldPoll = false,
  ...props
}: IActionWorkflowStatus) => {
  const { installActionRun } = useInstallActionRun()
  const status = installActionRun?.status_v2 || {
    status: installActionRun?.status || 'Unknown',
    status_human_description: installActionRun?.status_description || undefined,
  }

  return (
    <StatusBadge
      description={status.status_human_description}
      status={status.status}
      {...props}
    />
  )
}
