import React, { type FC } from 'react'
import { GoCommit, GoClock, GoPackage } from 'react-icons/go'
import { Card, Duration, Heading, Link, Status, Text, Time } from '@/components'
import { getBuild, type IGetBuild } from '@/lib'
import type { TBuild } from '@/types'

export const Build: FC<IGetBuild> = async (props) => {
  let build: TBuild
  try {
    build = await getBuild(props)
  } catch (error) {
    return <>No build to show</>
  }

  return (
    <Card>
      <Heading>Build details</Heading>
      <div className="flex flex-col gap-0">
        <Status
          status={build?.status}
          description={build?.status_description}
        />
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

      {build?.vcs_connection_commit ? (
        <div className="flex flex-col gap-0">
          <Text variant="label">Commit details</Text>
          <Text className="flex justify-between" variant="caption">
            <span className="flex gap-2 items-center">
              <GoCommit />
              {build?.vcs_connection_commit?.author_name ? (
                <b>{build?.vcs_connection_commit?.author_name}</b>
              ) : null}

              <span className="truncate">
                {build?.vcs_connection_commit?.message}
              </span>
            </span>
            <Link href="#">
              {build?.vcs_connection_commit?.sha?.slice(0, 7)}
            </Link>
          </Text>
        </div>
      ) : null}
    </Card>
  )
}
