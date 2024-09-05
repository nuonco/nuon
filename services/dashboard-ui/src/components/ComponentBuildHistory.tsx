'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useState } from 'react'
import { GoChevronRight } from 'react-icons/go'
import { Link, Status, Text, Time } from '@/components'
import type { TBuild } from '@/types'
import { SHORT_POLL_DURATION, sentanceCase } from '@/utils'

export interface IComponentBuildHistory {
  appId: string
  componentId: string
  initBuilds: Array<TBuild>
  orgId: string
  shouldPoll?: boolean
}

export const ComponentBuildHistory: FC<IComponentBuildHistory> = ({
  componentId,
  initBuilds,
  shouldPoll = false,
  ...props
}) => {
  const [builds, setComponentBuilds] = useState(initBuilds)

  useEffect(() => {
    const fetchComponentBuilds = () => {
      fetch(`/api/${props.orgId}/components/${componentId}/builds`)
        .then((res) => res.json().then((b) => setComponentBuilds(b)))
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollBuilds = setInterval(
        fetchComponentBuilds,
        SHORT_POLL_DURATION
      )
      return () => clearInterval(pollBuilds)
    }
  }, [builds, componentId, props.orgId, shouldPoll])

  return (
    <div>
      {builds.map((build, i) => (
        <ComponentBuildEvent
          key={`${build.id}-${i}`}
          build={build}
          isMostRecent={i === 0}
          {...props}
        />
      ))}
    </div>
  )
}

interface IComponentBuildEvent {
  appId: string
  build: TBuild
  isMostRecent?: boolean
  orgId: string
}

const ComponentBuildEvent: FC<IComponentBuildEvent> = ({
  appId,
  build,
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
          <Status status={build.status} isStatusTextHidden />
          <Text variant="label">{sentanceCase(build.status)}</Text>
        </span>

        <Text className="flex items-center gap-4 ml-8" variant="overline">
          <span>{build.id}</span>
          <>
            / <span>{build.component_name}</span>
          </>
        </Text>
      </div>

      <div className="flex items-center gap-4">
        <Time time={build.updated_at} format="relative" variant="overline" />

        <Link
          href={`/beta/${orgId}/apps/${appId}/components/${build.component_id}/builds/${build.id}`}
        >
          <GoChevronRight />
        </Link>
      </div>
    </div>
  )
}
