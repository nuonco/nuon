// TODO(nnnnat):  rename to InstallComponentDeployHistory

'use client'

import { Empty } from '@/components/old/Empty'
import { Timeline } from '@/components/old/Timeline'
import { ToolTip } from '@/components/old/ToolTip'
import { Truncate, Text } from '@/components/old/Typography'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TComponent, TDeploy, TPaginationParams } from '@/types'

export interface IInstallComponentDeploys
  extends TPaginationParams,
    IPollingProps {
  component: TComponent
  initDeploys: TDeploy[]
}

export const InstallComponentDeploys = ({
  component,

  initDeploys,
  limit,
  offset,
  pollInterval = 5000,
  shouldPoll = false,
}: IInstallComponentDeploys) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const params = useQueryParams({ limit, offset })
  const { data: deploys } = usePolling<TDeploy[]>({
    dependencies: [params],
    initData: initDeploys,
    path: `/api/orgs/${org.id}/installs/${install.id}/components/${component?.id}/deploys${params}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <Timeline
      emptyContent={
        <Empty
          emptyMessage="Waiting on component deployments."
          emptyTitle="No deployments yet"
          variant="history"
          isSmall
        />
      }
      events={deploys.map((d, i) => {
        return {
          id: d.id,
          status: d?.status_v2?.status || d?.status,
          underline: (
            <div>
              <Text>
                <ToolTip tipContent={d.id}>
                  <span className="truncate text-ellipsis w-16">{d.id}</span>
                </ToolTip>
                <>
                  /{' '}
                  {component.name.length >= 12 ? (
                    <ToolTip tipContent={component.name} alignment="right">
                      <Truncate variant="small">{component.name}</Truncate>
                    </ToolTip>
                  ) : (
                    component.name
                  )}
                </>
              </Text>
              {d?.created_by ? (
                <Text className="text-cool-grey-600 dark:text-white/70 !text-[10px]">
                  Deployed by: {d?.created_by?.email}
                </Text>
              ) : null}
            </div>
          ),
          time: d.updated_at,
          href:
            (d?.status_v2?.status &&
              (d?.status_v2?.status as string) !== 'queued') ||
            (d?.status && (d?.status as string) !== 'queued')
              ? `/${org.id}/installs/${install.id}/components/${component.id}/deploys/${d.id}`
              : null,
          isMostRecent: i === 0,
        }
      })}
    />
  )
}
