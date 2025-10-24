'use client'

import { StatusBadge } from '@/components/old/Status'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TBuild } from '@/types'

export interface IBuildStatus extends IPollingProps {
  descriptionAlignment?: 'center' | 'left' | 'right'
  descriptionPosition?: 'bottom' | 'top'
  initBuild: TBuild
}

export const BuildStatus = ({
  initBuild,
  pollInterval = 5000,
  shouldPoll = false,
  ...props
}: IBuildStatus) => {
  const { org } = useOrg()
  const { data: build } = usePolling<TBuild>({
    initData: initBuild,
    path: `/api/orgs/${org.id}/components/${initBuild?.component_id}/builds/${initBuild?.id}`,
    pollInterval,
    shouldPoll,
  })

  const status = build?.status_v2 || {
    status: build?.status || 'Unknown',
    status_human_description: build?.status_description || undefined,
  }

  return (
    <StatusBadge
      description={status?.status_human_description}
      status={status?.status}
      {...props}
    />
  )
}
