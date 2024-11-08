// TODO(nnnnat):  rename to InstallComponentDeployHistory

'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useState } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { Link } from '@/components/Link'
import { StatusBadge } from '@/components/Status'
import { Time } from '@/components/Time'
import { ToolTip } from '@/components/ToolTip'
import { Text, Truncate } from '@/components/Typography'
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
  initDeploys,
  shouldPoll = false,
  orgId,
}) => {
  const [deploys, setInstallComponentDeploys] = useState(initDeploys)

  useEffect(() => {
    const fetchInstallComponentDeploys = () => {
      fetch(
        `/api/${orgId}/installs/${installId}/components/${component.id}/deploys`
      )
        .then((res) => res.json().then((b) => setInstallComponentDeploys(b)))
        .catch(console.error)
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
    <div className="flex flex-col gap-2">
      {deploys.map((deploy, i) => (
        <InstallDeployEvent
          key={`${deploy.id}-${i}`}
          component={component}
          deploy={deploy}
          installId={installId}
          installComponentId={installComponentId}
          isMostRecent={i === 0}
          orgId={orgId}
        />
      ))}
    </div>
  )
}

interface IInstallDeployEvent {
  component: TComponent
  deploy: TInstallDeploy
  installId: string
  installComponentId: string
  isMostRecent?: boolean
  orgId: string
}

const InstallDeployEvent: FC<IInstallDeployEvent> = ({
  component,
  deploy,
  installId,
  installComponentId,
  isMostRecent = false,
  orgId,
}) => {
  return (
    <Link
      className="!block w-full !p-0"
      href={`/${orgId}/installs/${installId}/components/${installComponentId}/deploys/${deploy.id}`}
      variant="ghost"
    >
      <div
        className={classNames('flex items-center justify-between p-4', {
          'border rounded-md shadow-sm': isMostRecent,
        })}
      >
        <div className="flex flex-col">
          <span className="flex items-center gap-2">
            <StatusBadge
              status={deploy.status}
              isStatusTextHidden
              isWithoutBorder
            />
          </span>

          <Text className="flex items-center gap-2 ml-4 text-sm">
            <ToolTip tipContent={deploy.id}>
              <span className="truncate text-ellipsis w-16">{deploy.id}</span>
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
        </div>

        <div className="flex items-center gap-2">
          <Time time={deploy.updated_at} format="relative" />

          <Link
            href={`/${orgId}/installs/${installId}/components/${installComponentId}/deploys/${deploy.id}`}
            variant="ghost"
          >
            <CaretRight />
          </Link>
        </div>
      </div>
    </Link>
  )
}
