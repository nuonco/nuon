import { Card, type ICard } from '@/components/common/Card'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TRunner, TRunnerGroup, TRunnerMngHeartbeat } from '@/types'
import { isLessThan30SecondsOld } from '@/utils/time-utils'

type TRunnerHeartbeatEntry = TRunnerMngHeartbeat[keyof TRunnerMngHeartbeat]

interface IRunnerDetailsCard extends Omit<ICard, 'children'> {
  runner?: TRunner
  runnerGroup: TRunnerGroup
  heartbeat?: TRunnerHeartbeatEntry
}

export const RunnerDetailsCard = ({
  runner,
  runnerGroup,
  heartbeat,
  ...props
}: IRunnerDetailsCard) => {
  return (
    <Card {...props}>
      <Text variant="base" weight="strong">
        Runner details
      </Text>

      <LabeledValue label="Runner ID">
        <ID theme="default">{runner?.id}</ID>
      </LabeledValue>

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
              isLessThan30SecondsOld(heartbeat?.created_at)
                ? 'connected'
                : 'not-connected'
            }
            variant="badge"
          />
        </LabeledValue>

        <LabeledValue label="Version">
          <Text variant="subtext">
            {heartbeat?.version || 'Waiting on version'}
          </Text>
        </LabeledValue>

        <LabeledValue label="Platform">
          <Text variant="subtext" className="uppercase">
            {runnerGroup?.platform || runnerGroup?.['metadata']?.['runner.platform'] || 'Unknown'}
          </Text>
        </LabeledValue>

        <LabeledValue label="Started at">
          <Time variant="subtext" time={heartbeat?.started_at} />
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
      </div>
    </Card>
  )
}
