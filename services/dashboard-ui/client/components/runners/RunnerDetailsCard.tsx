import { useQuery } from '@tanstack/react-query'
import { Card, type ICard } from '@/components/common/Card'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { getRunnerLatestHeartbeat } from '@/lib'
import type { TRunnerGroup, TRunnerMngHeartbeat } from '@/types'
import { isLessThan15SecondsOld } from '@/utils/time-utils'

interface IRunnerDetailsCard extends Omit<ICard, 'children'> {
  initHeartbeat?: TRunnerMngHeartbeat
  runnerGroup: TRunnerGroup
  shouldPoll?: boolean
  pollInterval?: number
}

export const RunnerDetailsCard = ({
  initHeartbeat,
  pollInterval = 5000,
  runnerGroup,
  shouldPoll = false,
  ...props
}: IRunnerDetailsCard) => {
  const { org } = useOrg()
  const { runner } = useRunner()

  const { data: heartbeats } = useQuery({
    queryKey: ['runner-heartbeat', org?.id, runner?.id],
    queryFn: () => getRunnerLatestHeartbeat({ orgId: org.id, runnerId: runner.id }),
    initialData: initHeartbeat,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!runner?.id,
  })

  const runnerHeartbeat =
    heartbeats?.install ??
    heartbeats?.org ??
    heartbeats?.build ??
    heartbeats?.[''] ??
    undefined

  return (
    <Card {...props}>
      <Text variant="base" weight="strong">
        Runner details
      </Text>

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
            {runnerGroup?.platform || runnerGroup?.['metadata']?.['runner.platform'] || 'Unknown'}
          </Text>
        </LabeledValue>

        <LabeledValue label="Started at">
          <Time variant="subtext" time={runnerHeartbeat?.started_at} />
        </LabeledValue>

        <LabeledValue label="Runner ID">
          <ID theme="default">{runner?.id}</ID>
        </LabeledValue>
      </div>
    </Card>
  )
}

export const RunnerDetailsCardSkeleton = (props: Omit<ICard, 'children'>) => {
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
