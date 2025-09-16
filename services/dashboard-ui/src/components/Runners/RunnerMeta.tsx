'use client'

import { DateTime } from 'luxon'
import { StatusBadge } from '@/components/Status'
import { Time } from '@/components/Time'
import { ID, Text } from '@/components/Typography'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TRunner, TRunnerMngHeartbeat, TRunnerSettings } from '@/types'

function isLessThan15SecondsOld(timestampStr: string) {
  const date = DateTime.fromISO(timestampStr)
  const now = DateTime.now()
  const diffInSeconds = now.diff(date, 'seconds').seconds

  return diffInSeconds >= 0 && diffInSeconds < 15
}

interface IRunnerMeta extends IPollingProps {
  initHeartbeat: TRunnerMngHeartbeat
  initRunner: TRunner
  initSettings: TRunnerSettings
  installId?: string
}

export const RunnerMeta = ({
  initHeartbeat,
  initRunner,
  initSettings,
  shouldPoll = false,
}: IRunnerMeta) => {
  const { org } = useOrg()
  const { data: runner } = usePolling<TRunner>({
    initData: initRunner,
    path: `/api/orgs/${org.id}/runners/${initRunner.id}`,
    pollInterval: 20000,
    shouldPoll,
  })

  const { data: settings } = usePolling<TRunnerSettings>({
    initData: initSettings,
    path: `/api/orgs/${org.id}/runners/${initRunner.id}/settings`,
    pollInterval: 20000,
    shouldPoll,
  })

  const { data: heartbeats } = usePolling<TRunnerMngHeartbeat>({
    initData: initHeartbeat,
    path: `/api/orgs/${org.id}/runners/${initRunner.id}/heartbeat`,
    pollInterval: 5000,
    shouldPoll,
  })

  const runnerHeartbeat =
    heartbeats.install ?? heartbeats?.org ?? heartbeats[''] ?? undefined

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
            isWithoutBorder
          />
        ) : (
          <StatusBadge
            status="not-connected"
            description="Not connected to runner"
            descriptionAlignment="left"
            isWithoutBorder
          />
        )}
        <StatusBadge
          status={runner?.status === 'active' ? 'healthy' : 'unhealthy'}
          description={runner?.status_description}
          descriptionAlignment="left"
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
