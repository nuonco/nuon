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
        status: b?.status_v2?.status || b?.status,
        underline: (
          <div>
            <Text>
              <ToolTip tipContent={`Build ID: ${b.id}`}>
                <Text
                  className="truncate !block text-ellipsis w-16"
                  variant="mono-12"
                >
                  {b.id}
                </Text>
              </ToolTip>
              <>
                /{' '}
                {b.component_name.length >= 12 ? (
                  <ToolTip
                    tipContent={`Component: ${b.component_name}`}
                    alignment="right"
                  >
                    <Truncate variant="small">{b.component_name}</Truncate>
                  </ToolTip>
                ) : (
                  b.component_name
                )}
              </>
            </Text>
            <span className="flex flex-col gap-2 mt-2">
              {b?.vcs_connection_commit?.message &&
              b?.vcs_connection_commit?.sha ? (
                <span className="">
                  <ToolTip tipContent={`SHA: ${b?.vcs_connection_commit?.sha}`} alignment="right">
                    <Text
                      className="truncate !block w-20 !text-[11px]"
                      variant="mono-12"
                    >
                      # {b?.vcs_connection_commit?.sha}
                    </Text>
                  </ToolTip>
                  <Text className="!text-[11px] font-normal pr-2 !block max-w-[250px] truncate" isMuted>
                    <span className="truncate">
                      {b?.vcs_connection_commit?.message}
                    </span>
                  </Text>
                </span>
              ) : null}
              <Text className="!text-[10px]" isMuted>
                Build by: {b?.created_by?.email}
              </Text>
            </span>
          </div>
        ),
        time: b.updated_at,
        href:
          (b?.status_v2?.status &&
            (b?.status_v2?.status as string) !== 'queued') ||
          (b?.status && b?.status !== 'queued')
            ? `/${orgId}/apps/${appId}/components/${b.component_id}/builds/${b.id}`
            : null,
        isMostRecent: i === 0,
      }))}
    />
  )
}
