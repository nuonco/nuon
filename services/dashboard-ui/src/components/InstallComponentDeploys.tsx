// TODO(nnnnat):  rename to InstallComponentDeployHistory

'use client'

import React, { type FC, useEffect } from 'react'
import { Empty } from '@/components/Empty'
import { Timeline } from '@/components/Timeline'
import { ToolTip } from '@/components/ToolTip'
import { Truncate, Text } from '@/components/Typography'
import { revalidateInstallData } from '@/components/install-actions'
import type { TComponent, TInstallDeploy } from '@/types'
import { SHORT_POLL_DURATION } from '@/utils'

export interface IInstallComponentDeploys {
  component: TComponent
  installId: string
  installComponentId: string
  initDeploys: Array<TInstallDeploy>
  shouldPoll?: boolean
  orgId: string
}

export const InstallComponentDeploys: FC<IInstallComponentDeploys> = ({
  component,
  installId,
  installComponentId,
  initDeploys: deploys,
  shouldPoll = false,
  orgId,
}) => {
  //  const [deploys, setInstallComponentDeploys] = useState(initDeploys)

  useEffect(() => {
    const fetchInstallComponentDeploys = () => {
      /* fetch(
       *   `/api/${orgId}/installs/${installId}/components/${component.id}/deploys`
       * )
       *   .then((res) => res.json().then((b) => setInstallComponentDeploys(b)))
       *   .catch(console.error) */

      revalidateInstallData({ installId, orgId })
    }

    if (shouldPoll) {
      const pollDeploys = setInterval(
        fetchInstallComponentDeploys,
        SHORT_POLL_DURATION
      )
      return () => clearInterval(pollDeploys)
    }
  }, [deploys, shouldPoll])

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
      events={deploys.map((d, i) => ({
        id: d.id,
        status: d.status,
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
        href: `/${orgId}/installs/${installId}/components/${installComponentId}/deploys/${d.id}`,
        isMostRecent: i === 0,
      }))}
    />
  )
}
