'use client'

import { useState } from 'react'
import { Link } from '@/components/common/Link'
import { Timeline, type ITimeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TDeploy } from '@/types'

interface IDeployTimeline
  extends Omit<ITimeline<TDeploy>, 'events' | 'renderEvent'>,
    IPollingProps {
  componentName: string
  componentId: string
}

export const DeployTimeline = ({
  componentName,
  componentId,
  shouldPoll = false,
  pollInterval = 20000,
}: IDeployTimeline) => {
  const { install } = useInstall()
  const { org } = useOrg()
  const [pagination, setPagination] = useState({
    offset: 0,
    limit: 10,
  })

  const queryParams = useQueryParams({
    offset: pagination.offset,
    limit: pagination.limit,
  })
  const { data: deploys, error } = usePolling<TDeploy[]>({
    path: `/api/orgs/${org?.id}/installs/${install.id}/components/${componentId}/deploys${queryParams}`,
    shouldPoll,
    pollInterval,
  })

  return deploys ? (
    <Timeline<TDeploy>
      events={deploys}
      pagination={pagination}
      renderEvent={(deploy, idx) => {
        return (
          <TimelineEvent
            key={deploy.id}
            badge={
              idx === 0 ? { theme: 'info', children: 'Latest' } : undefined
            }
            caption={deploy?.id}
            createdAt={deploy?.created_at}
            status={deploy?.status}
            title={
              <Link
                href={`/${org.id}/installs/${install.id}/components/${componentId}/deploys/${deploy.id}`}
              >
                {componentName} deploy
              </Link>
            }
          />
        )
      }}
    />
  ) : (
    <>
      <TimelineSkeleton eventCount={10} />
    </>
  )
}
