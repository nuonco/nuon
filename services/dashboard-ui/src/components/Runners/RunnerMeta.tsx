import React, { type FC } from 'react'
import { Time } from '@/components/Time'
import { Text } from '@/components/Typography'
import {
  getRunnerLatestHeartbeat,
  getInstallRunnerGroup,
  getOrgRunnerGroup,
} from '@/lib'

interface IRunnerMeta {
  orgId: string
  runnerId: string
  installId?: string
}

export const RunnerMeta: FC<IRunnerMeta> = async ({
  orgId,
  runnerId,
  installId = '',
}) => {
  const [runnerHeartbeat] = await Promise.all([
    getRunnerLatestHeartbeat({
      orgId,
      runnerId,
    }),
  ])

  const runnerGroup =
    installId === ''
      ? await getInstallRunnerGroup({ orgId, installId })
      : await getOrgRunnerGroup({ orgId })

  return (
    <>
      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Started at
        </Text>
        <Text>
          <Time
            time={runnerHeartbeat?.started_at}
            format="default"
            variant="med-12"
          />
        </Text>
      </span>
      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Version
        </Text>
        <Text variant="med-12">{runnerHeartbeat?.version}</Text>
      </span>
      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Platform
        </Text>
        <Text variant="med-12">{runnerGroup?.platform}</Text>
      </span>
    </>
  )
}
