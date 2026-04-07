import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { RunnerDetailsCard, RunnerDetailsCardSkeleton } from '@/components/runners/RunnerDetailsCard'
import { RunnerHealthCard, RunnerHealthCardSkeleton } from '@/components/runners/RunnerHealthCard'
import { RunnerProvider } from '@/providers/runner-provider'
import type { TRunner, TRunnerMngHeartbeat, TRunnerHealthCheck, TWorkflowStep } from '@/types'

export interface IRunnerStepDetails {
  step?: TWorkflowStep
  orgId: string
  runner?: TRunner
  runnerHeartbeat?: TRunnerMngHeartbeat
  runnerHealthCheck?: TRunnerHealthCheck[]
  isRunnerLoading: boolean
  isHeartbeatLoading: boolean
  isHealthCheckLoading: boolean
}

export const RunnerStepDetails = ({
  step,
  orgId,
  runner,
  runnerHeartbeat,
  runnerHealthCheck,
  isRunnerLoading,
  isHeartbeatLoading,
  isHealthCheckLoading,
}: IRunnerStepDetails) => {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-4">
        <Text variant="base" weight="strong">
          Install runner
        </Text>

        <Text variant="subtext">
          <Link href={`/${orgId}/installs/${step.owner_id}/runner`}>
            View runner <Icon variant="CaretRight" />
          </Link>
        </Text>
      </div>
      <div className="flex flex-col @min-4xl:flex-row gap-6">
        {(isRunnerLoading || isHeartbeatLoading) && !runner ? (
          <RunnerDetailsCardSkeleton />
        ) : (
          <RunnerProvider runnerId={runner?.id}>
            <RunnerDetailsCard
              initHeartbeat={runnerHeartbeat}
              runnerGroup={{ platform: 'local' }}
              shouldPoll
            />
          </RunnerProvider>
        )}

        {(isHealthCheckLoading || !runnerHealthCheck) && !runner ? (
          <RunnerHealthCardSkeleton />
        ) : (
          <RunnerProvider runnerId={runner.id}>
            <RunnerHealthCard
              initHealthchecks={runnerHealthCheck}
              shouldPoll
            />
          </RunnerProvider>
        )}
      </div>
    </div>
  )
}
