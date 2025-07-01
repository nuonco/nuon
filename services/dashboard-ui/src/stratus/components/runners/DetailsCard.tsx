'use client'

import React from 'react'
import {
  Card,
  LabeledValue,
  Skeleton,
  Text,
  Time,
  type ICard,
} from '@/stratus/components/common'
import { Status } from '@/stratus/components/statuses'
import { useOrg } from '@/stratus/context'
import type { TRunner, TRunnerGroup, TRunnerHeartbeat } from '@/types'
import { isLessThan15SecondsOld, usePolling } from '@/utils'

interface IDetailsCard extends Omit<ICard, 'children'> {}

interface IRunnerDetailsCard extends IDetailsCard {
  runner: TRunner
  runnerGroup: TRunnerGroup
  runnerHeartbeat: TRunnerHeartbeat
}

export const RunnerDetailsCard = ({
  runner,
  runnerGroup,
  runnerHeartbeat: initRunnerHeartbeat,
  ...props
}: IRunnerDetailsCard) => {
  const { org } = useOrg()
  const { data: runnerHeartbeat, error } = usePolling<TRunnerHeartbeat>({
    path: `/api/${org?.id}/runners/${runner?.id}/latest-heart-beat`,
    shouldPoll: true,
    initData: initRunnerHeartbeat,
  })

  return (
    <Card {...props}>
      <Text variant="base" weight="strong">
        Runner details
      </Text>

      {error ? (
        <Text>Unable to refresh runner heartbeat: {error?.error}</Text>
      ) : (
        <div className="grid gap-6 md:grid-cols-2">
          <LabeledValue label="Status">
            <Status
              status={runner?.status === 'active' ? 'healthy' : 'unhealthy'}
              variant="badge"
            />
          </LabeledValue>

          <LabeledValue label="Connectivity">
            <Status
              status={
                isLessThan15SecondsOld(runnerHeartbeat?.created_at)
                  ? 'connected'
                  : 'not-connected'
              }
              variant="badge"
            />
          </LabeledValue>

          <LabeledValue label="Version">
            <Text>{runnerHeartbeat?.version}</Text>
          </LabeledValue>

          <LabeledValue label="Platform">
            <Text className="uppercase">{runnerGroup?.platform}</Text>
          </LabeledValue>

          <LabeledValue label="Started at">
            <Text>
              <Time time={runnerHeartbeat?.started_at} />
            </Text>
          </LabeledValue>

          <LabeledValue label="Runner ID">
            <Text family="mono" theme="muted">
              {runner?.id}
            </Text>
          </LabeledValue>
        </div>
      )}
    </Card>
  )
}

export const RunnerDetailsCardSkeleton = (props: IDetailsCard) => {
  return (
    <Card {...props}>
      <Skeleton height="24px" width="106px" />

      <div className="grid gap-6 md:grid-cols-2">
        <LabeledValue label={<Skeleton height="17px" width="34px" />}>
          <Skeleton height="23px" width="75px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="68px" />}>
          <Skeleton height="23px" width="110px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="41px" />}>
          <Skeleton height="23px" width="50px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="45px" />}>
          <Skeleton height="23px" width="54px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="53px" />}>
          <Skeleton height="23px" width="148px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="53px" />}>
          <Skeleton height="23px" width="215px" />
        </LabeledValue>
      </div>
    </Card>
  )
}
