import { DateTime } from 'luxon'
import React, { type FC } from 'react'
import { StatusBadge } from '@/components/Status'
import { Time } from '@/components/Time'
import { ID, Text } from '@/components/Typography'
import type { TRunner, TRunnerHeartbeat, TRunnerGroupSettings } from '@/types'
import { nueQueryData } from '@/utils'

function isLessThan15SecondsOld(timestampStr: string) {
  const date = DateTime.fromISO(timestampStr)
  const now = DateTime.now()
  const diffInSeconds = now.diff(date, 'seconds').seconds

  return diffInSeconds >= 0 && diffInSeconds < 15
}

interface IRunnerMeta {
  orgId: string
  runner: TRunner
  installId?: string
}

export const RunnerMeta: FC<IRunnerMeta> = async ({ orgId, runner }) => {
  const [{ data: heartbeats }, { data: settings }] = await Promise.all([
    nueQueryData<{ install: TRunnerHeartbeat; mng: TRunnerHeartbeat }>({
      orgId,
      path: `runners/${runner?.id}/heart-beats/latest`,
    }),
    nueQueryData<TRunnerGroupSettings>({
      orgId,
      path: `runners/${runner?.id}/settings`,
    }),
  ])

  const runnerHeartbeat = heartbeats?.install || undefined

  return (
    <div className="grid md:grid-cols-3 gap-8 items-start justify-start">
      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Status
        </Text>

        {runnerHeartbeat &&
        isLessThan15SecondsOld(runnerHeartbeat?.created_at) ? (
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
      {settings ? (
        <>
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Tag
            </Text>
            <Text>{settings?.container_image_tag}</Text>
          </span>
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Platform
            </Text>
            <Text>{settings?.metadata?.['runner.platform'] || 'Unknown'}</Text>
          </span>
        </>
      ) : null}
      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">ID</Text>
        <ID className="break-all" id={runner?.id} />
      </span>
    </div>
  )
}
