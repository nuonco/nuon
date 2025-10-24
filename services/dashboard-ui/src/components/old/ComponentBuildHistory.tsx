'use client'

import { Empty } from '@/components/old/Empty'
import { Timeline } from '@/components/old/Timeline'
import { ToolTip } from '@/components/old/ToolTip'
import { Text, Truncate } from '@/components/old/Typography'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TBuild, TPaginationParams } from '@/types'

export interface IComponentBuildHistory
  extends IPollingProps,
    TPaginationParams {
  componentId: string
  initBuilds: Array<TBuild>
}

export const ComponentBuildHistory = ({
  componentId,
  initBuilds,
  pollInterval = 5000,
  shouldPoll = false,
  offset,
  limit,
}: IComponentBuildHistory) => {
  const { app } = useApp()
  const { org } = useOrg()
  const params = useQueryParams({ offset, limit })
  const { data: builds } = usePolling({
    dependencies: [params],
    initData: initBuilds,
    path: `/api/orgs/${org.id}/components/${componentId}/builds${params}`,
    pollInterval,
    shouldPoll,
  })

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
                  <ToolTip
                    tipContent={`SHA: ${b?.vcs_connection_commit?.sha}`}
                    alignment="right"
                  >
                    <Text
                      className="truncate !block w-20 !text-[11px]"
                      variant="mono-12"
                    >
                      # {b?.vcs_connection_commit?.sha}
                    </Text>
                  </ToolTip>
                  <Text
                    className="!text-[11px] font-normal pr-2 !block max-w-[250px] truncate"
                    isMuted
                  >
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
            ? `/${org.id}/apps/${app.id}/components/${b.component_id}/builds/${b.id}`
            : null,
        isMostRecent: i === 0,
      }))}
    />
  )
}
