'use client'

import React, { type FC, useEffect, useState } from 'react'
import { GoClock, GoPackage } from 'react-icons/go'
import {
  BuildCommit,
  Duration,
  Heading,
  PageHeader,
  Status,
  Text,
  Time,
} from '@/components'
import { type IGetBuild } from '@/lib'
import type { TComponent, TBuild } from '@/types'

export const BuildHeader: FC<
  { ssrBuild: TBuild; component: TComponent } & IGetBuild
> = ({ ssrBuild, component, orgId, buildId }) => {
  const [build, setBuild] = useState(ssrBuild)

  const fetchBuild = () => {
    fetch(`/api/${orgId}/components/${component?.id}/builds/${buildId}`)
      .then((r) => r.json().then((b) => setBuild(b)))
      .catch(console.error)
  }

  useEffect(() => {
    let pollBuild: NodeJS.Timeout
    if (build?.status !== 'active' && build?.status !== 'error') {
      pollBuild = setInterval(fetchBuild, 10000)
    }

    return () => clearInterval(pollBuild)
  }, [build])

  return (
    <PageHeader
      info={
        <>
          <Status status={build?.status} />
          <div className="flex flex-col flex-auto gap-1">
            <Text variant="caption">
              <b>Build ID:</b> {build.id}
            </Text>
            <Text variant="caption">
              <b>Component ID:</b> {component?.id}
            </Text>
            <BuildCommit {...build} />
          </div>
        </>
      }
      title={
        <span className="flex flex-col gap-2">
          <Text variant="overline">{build?.id}</Text>
          <Heading
            level={1}
            variant="title"
            className="flex gap-1 items-center"
          >
            {component?.name} build
          </Heading>
        </span>
      }
      summary={
        <div className="flex gap-6">
          <Text variant="caption">
            <GoPackage />
            <Time time={build?.updated_at} />
          </Text>
          <Text variant="caption">
            <GoClock />
            <Duration
              unitDisplay="short"
              listStyle="long"
              variant="caption"
              beginTime={build?.created_at}
              endTime={build?.updated_at}
            />
          </Text>
        </div>
      }
    />
  )
}
