'use client'

import React, { type FC, useEffect } from 'react'
import { Empty } from '@/components/Empty'
import { Timeline } from '@/components/Timeline'
import { ToolTip } from '@/components/ToolTip'
import { Text, Truncate } from '@/components/Typography'
import { revalidateAppData } from '@/components/app-actions'
import type { TBuild } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

export interface IComponentBuildHistory {
  appId: string
  componentId: string
  initBuilds: Array<TBuild>
  orgId: string
  shouldPoll?: boolean
}

export const ComponentBuildHistory: FC<IComponentBuildHistory> = ({
  appId,
  componentId,
  initBuilds: builds,
  shouldPoll = false,
  orgId,
}) => {
  //const [builds, setComponentBuilds] = useState(initBuilds)

  useEffect(() => {
    const fetchComponentBuilds = () => {
      /* fetch(`/api/${orgId}/components/${componentId}/builds`)
       *   .then((res) => res.json().then((b) => setComponentBuilds(b)))
       *   .catch(console.error) */
      revalidateAppData({ appId, orgId })
    }

    if (shouldPoll) {
      const pollBuilds = setInterval(fetchComponentBuilds, SHORT_POLL_DURATION)
      return () => clearInterval(pollBuilds)
    }
  }, [builds, componentId, orgId, shouldPoll])

  return (
    <Timeline
      emptyContent={
        <Empty
          emptyTitle="No builds yet"
          emptyMessage="Waiting on component builds."
          variant="history"
          isSmall
        />
      }
      events={builds.map((b, i) => ({
        id: b.id,
        status: b.status,
        underline: (
          <div>
            <Text>
              <ToolTip tipContent={b.id}>
                <Text className="truncate text-ellipsis w-16" variant="mono-12">
                  {b.id}
                </Text>
              </ToolTip>
              <>
                /{' '}
                {b.component_name.length >= 12 ? (
                  <ToolTip tipContent={b.component_name} alignment="right">
                    <Truncate variant="small">{b.component_name}</Truncate>
                  </ToolTip>
                ) : (
                  b.component_name
                )}
              </>
            </Text>
            {b?.created_by ? (
              <Text className="text-cool-grey-600 dark:text-white/70 !text-[10px]">
                Build by: {b?.created_by?.email}
              </Text>
            ) : null}
          </div>
        ),
        time: b.updated_at,
        href: `/${orgId}/apps/${appId}/components/${b.component_id}/builds/${b.id}`,
        isMostRecent: i === 0,
      }))}
    />
  )
}
