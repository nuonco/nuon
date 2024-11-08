'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useState } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { Link } from '@/components/Link'
import { StatusBadge } from '@/components/Status'
import { Time } from '@/components/Time'
import { ToolTip } from '@/components/ToolTip'
import { Text, Truncate } from '@/components/Typography'
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
      const pollBuilds = setInterval(fetchComponentBuilds, SHORT_POLL_DURATION)
      return () => clearInterval(pollBuilds)
    }
  }, [builds, componentId, props.orgId, shouldPoll])

  return (
    <div className="flex flex-col gap-2">
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
    <Link
      className="!block w-full !p-0"
      href={`/${orgId}/apps/${appId}/components/${build.component_id}/builds/${build.id}`}
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
              status={build.status}
              isStatusTextHidden
              isWithoutBorder
            />
          </span>

          <Text className="flex items-center gap-2 ml-4">
            <ToolTip tipContent={build.id}>
              <Text className="truncate text-ellipsis w-16" variant="mono-12">
                {build.id}
              </Text>
            </ToolTip>
            <>
              /{' '}
              {build.component_name.length >= 12 ? (
                <ToolTip tipContent={build.component_name} alignment="right">
                  <Truncate variant="small">{build.component_name}</Truncate>
                </ToolTip>
              ) : (
                build.component_name
              )}
            </>
          </Text>
        </div>

        <div className="flex items-center gap-2">
          <Time time={build.updated_at} format="relative" />
          <Link
            href={`/${orgId}/apps/${appId}/components/${build.component_id}/builds/${build.id}`}
            variant="ghost"
          >
            <CaretRight />
          </Link>
        </div>
      </div>
    </Link>
  )
}
