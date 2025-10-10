'use client'

import { Card, type ICard } from '@/components/common/Card'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TRunner, TRunnerGroup, TRunnerMngHeartbeat } from '@/types'
import { isLessThan15SecondsOld } from '@/utils/time-utils'

interface IRunnerDetailsCard extends Omit<ICard, 'children'>, IPollingProps {
  initHeartbeat: TRunnerMngHeartbeat
  runner: TRunner
  runnerGroup: TRunnerGroup
}

export const RunnerDetailsCard = ({
  initHeartbeat,
  pollInterval = 5000,
  runner,
  runnerGroup,
  shouldPoll = false,
  ...props
}: IRunnerDetailsCard) => {
  const { org } = useOrg()
  const { data: heartbeats, error } = usePolling<TRunnerMngHeartbeat>({
    path: `/api/orgs/${org?.id}/runners/${runner?.id}/heartbeat`,
    shouldPoll,
    initData: initHeartbeat,
    pollInterval,
  })

  const runnerHeartbeat =
    heartbeats.install ??
    heartbeats?.org ??
    heartbeats?.build ??
    heartbeats[''] ??
    undefined

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
            <Text variant="subtext">
              {runnerHeartbeat?.version || 'Waiting on version'}
            </Text>
          </LabeledValue>

          <LabeledValue label="Platform">
            <Text variant="subtext" className="uppercase">
              {runnerGroup?.platform}
            </Text>
          </LabeledValue>

          <LabeledValue label="Started at">
            <Time variant="subtext" time={runnerHeartbeat?.started_at} />
          </LabeledValue>

          <LabeledValue label="Runner ID">
            <ID theme="default">{runner?.id}</ID>
          </LabeledValue>
        </div>
      )}
    </Card>
  )
}
