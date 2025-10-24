'use client'

import { StatusBadge } from '@/components/old/Status'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TSandboxRun } from '@/types'

export interface ISandboxRunStatus extends IPollingProps {
  descriptionAlignment?: 'center' | 'left' | 'right'
  descriptionPosition?: 'bottom' | 'top'
  initSandboxRun: TSandboxRun
}

export const SandboxRunStatus = ({
  initSandboxRun,
  pollInterval = 5000,
  shouldPoll = false,
  ...props
}: ISandboxRunStatus) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { data: run } = usePolling<TSandboxRun>({
    initData: initSandboxRun,
    path: `/api/orgs/${org.id}/installs/${install.id}/sandbox/runs/${initSandboxRun.id}`,
    pollInterval,
    shouldPoll,
  })

  const status = run?.status_v2 || {
    status: run?.status || 'Unknown',
    status_human_description: run?.status_description || undefined,
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
