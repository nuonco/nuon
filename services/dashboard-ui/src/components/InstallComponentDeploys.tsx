// TODO(nnnnat):  rename to InstallComponentDeployHistory

'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useState } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { Link, Status, Text, Time, ToolTip } from '@/components'
import type { TComponent, TInstallDeploy } from '@/types'
import { SHORT_POLL_DURATION, sentanceCase } from '@/utils'

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
    <div>
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
    <div
      className={classNames('flex items-center justify-between p-4', {
        'border rounded-md shadow-sm': isMostRecent,
      })}
    >
      <div className="flex flex-col gap-2">
        <span className="flex items-center gap-4">
          <Status status={deploy.status} isStatusTextHidden />
          <Text variant="label">{sentanceCase(deploy.status)}</Text>
        </span>

        <Text className="flex items-center gap-4 ml-8" variant="overline">
          <ToolTip tipContent={deploy.id}>
            <span className="truncate text-ellipsis w-16">{deploy.id}</span>
          </ToolTip>
          <>
            / <span>{component.name}</span>
          </>
        </Text>
      </div>

      <div className="flex items-center gap-4">
        <Time time={deploy.updated_at} format="relative" variant="overline" />

        <Link
          href={`/${orgId}/installs/${installId}/components/${installComponentId}/deploys/${deploy.id}`}
          variant="ghost"
        >
          <CaretRight />
        </Link>
      </div>
    </div>
  )
}
