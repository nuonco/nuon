// TODO(nnnnat):  rename to InstallComponentDeployHistory

'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useState } from 'react'
import { GoChevronRight } from 'react-icons/go'
import { Link, Status, Text, Time } from '@/components'
import type { TInstallDeploy } from '@/types'
import { SHORT_POLL_DURATION, sentanceCase } from '@/utils'

export interface IInstallComponentDeploys {
  installId: string
  installComponentId: string
  initDeploys: Array<TInstallDeploy>
  shouldPoll?: boolean
  orgId: string
}

export const InstallComponentDeploys: FC<IInstallComponentDeploys> = ({
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
        `/api/${orgId}/installs/${installId}/components/${installComponentId}/deploys`
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
  deploy: TInstallDeploy
  installId: string
  installComponentId: string
  isMostRecent?: boolean
  orgId: string
}

const InstallDeployEvent: FC<IInstallDeployEvent> = ({
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
          <span>{deploy.id}</span>
          <>
            / <span>{deploy.component_name}</span>
          </>
        </Text>
      </div>

      <div className="flex items-center gap-4">
        <Time time={deploy.updated_at} format="relative" variant="overline" />

        <Link
          key={deploy.id}
          href={`/beta/${orgId}/installs/${installId}/components/${installComponentId}/deploys/${deploy.id}`}
        >
          <GoChevronRight />
        </Link>
      </div>
    </div>
  )
}
