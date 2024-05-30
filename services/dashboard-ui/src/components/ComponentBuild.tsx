'use client'

import React, { type FC } from 'react'
import { GoCommit, GoClock, GoPackage } from 'react-icons/go'
import { Card, Duration, Heading, Link, Status, Text, Time } from '@/components'
import { useBuildContext } from '@/context'

export const ComponentBuild: FC = async () => {
  const { build } = useBuildContext()

  return (
    <>
      <div className="flex flex-col gap-0">
        <ComponentBuildStatus />
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
        <Text variant="caption">
          <Link href={`/dashboard/${build?.org_id}/builds/${build.id}`}>
            Details
          </Link>
        </Text>
      </div>

      {build?.vcs_connection_commit ? <BuildCommit /> : null}
    </>
  )
}

export const BuildCommit: FC = () => {
  const {
    build: { vcs_connection_commit },
  } = useBuildContext()

  return (
    <div className="flex flex-col gap-0">
      <Text variant="label">Commit details</Text>
      <Text className="flex justify-between" variant="caption">
        <span className="flex gap-2 items-center">
          <GoCommit />
          {vcs_connection_commit?.author_name ? (
            <b>{vcs_connection_commit?.author_name}</b>
          ) : null}
          <span className="truncate">{vcs_connection_commit?.message}</span> (#
          {vcs_connection_commit?.sha?.slice(0, 7)})
        </span>
      </Text>
    </div>
  )
}

export interface IComponentBuildStatus {
  isCompact?: boolean
  isStatusTextHidden?: boolean
  showDescription?: boolean
}

export const ComponentBuildStatus: FC<IComponentBuildStatus> = ({
  isCompact = false,
  isStatusTextHidden = false,
  showDescription = false,
}) => {
  const { build } = useBuildContext()

  return (
    <Status
      status={build?.status}
      description={showDescription && build?.status_description}
      label={isCompact && build?.component_name}
      isLabelStatusText={isCompact}
      isStatusTextHidden={isStatusTextHidden}
    />
  )
}

export const ComponentBuildCard: FC<{ heading?: string }> = ({
  heading = 'Component build',
}) => (
  <Card>
    <Heading>{heading}</Heading>
    <ComponentBuild />
  </Card>
)
