import React, { type FC } from 'react'
import { StatusBadge } from '@/components/Status'
import { Time } from '@/components/Time'
import { ID, Text } from '@/components/Typography'
import {
  getRunnerLatestHeartbeat,
  getInstallRunnerGroup,
  getOrgRunnerGroup,
} from '@/lib'
import type { TRunner } from '@/types'

interface IRunnerMeta {
  orgId: string
  runner: TRunner
  installId?: string
}

export const RunnerMeta: FC<IRunnerMeta> = async ({
  orgId,
  runner,
  installId = '',
}) => {
  const getRunnerGroup = () =>
    installId === ''
      ? getOrgRunnerGroup({ orgId })
      : getInstallRunnerGroup({ orgId, installId })

  const [runnerHeartbeat, runnerGroup] = await Promise.all([
    getRunnerLatestHeartbeat({
      orgId,
      runnerId: runner.id,
    }).catch(console.error),
    getRunnerGroup().catch(console.error),
  ])

  return (
    <div className="flex gap-8 items-start justify-start flex-wrap">
      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Status
        </Text>

        {runnerHeartbeat && runnerHeartbeat?.alive_time > 15 * 1_000_000_000 ? (
          <StatusBadge
            status="connected"
            description="Connected to runner"
            descriptionAlignment="left"
            shouldPoll
            isWithoutBorder
          />
        ) : (
          <StatusBadge
            status="not-connected"
            description="Not connected to runner"
            descriptionAlignment="left"
            shouldPoll
            isWithoutBorder
          />
        )}
        <StatusBadge
          status={runner?.status === 'active' ? 'healthy' : 'unhealthy'}
          description={runner?.status_description}
          descriptionAlignment="left"
          shouldPoll
          isWithoutBorder
        />
      </span>
      {runnerHeartbeat ? (
        <>
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Started at
            </Text>
            <Text>
              <Time time={runnerHeartbeat?.started_at} format="default" />
            </Text>
          </span>
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Version
            </Text>
            <Text>{runnerHeartbeat?.version}</Text>
          </span>
        </>
      ) : null}
      {runnerGroup ? (
        <span className="flex flex-col gap-2">
          <Text className="text-cool-grey-600 dark:text-cool-grey-500">
            Platform
          </Text>
          <Text>{runnerGroup?.platform}</Text>
        </span>
      ) : null}
      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">ID</Text>
        <ID id={runner?.id} />
      </span>
    </div>
  )
}
