import { useQuery } from '@tanstack/react-query'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { RunnerDetailsCard, RunnerDetailsCardSkeleton } from '@/components/runners/RunnerDetailsCard'
import { RunnerHealthCard, RunnerHealthCardSkeleton } from '@/components/runners/RunnerHealthCard'
import { useOrg } from '@/hooks/use-org'
import { getRunner, getRunnerLatestHeartbeat, getRunnerRecentHealthChecks } from '@/lib'
import { RunnerProvider } from '@/providers/runner-provider'
import type { IStepDetails } from './types'

interface IRunnerStepDetails extends IStepDetails {}

export const RunnerStepDetails = ({ step }: IRunnerStepDetails) => {
  const { org } = useOrg()
  const runnerId = step.step_target_id

  const { data: runner, isLoading: isRunnerLoading } = useQuery({
    queryKey: ['runner', org?.id, runnerId],
    queryFn: () => getRunner({ orgId: org.id, runnerId }),
    enabled: !!org?.id && !!runnerId,
  })

  const { data: runnerHeartbeat, isLoading: isHeartbeatLoading } = useQuery({
    queryKey: ['runner-heartbeat', org?.id, runnerId],
    queryFn: () => getRunnerLatestHeartbeat({ orgId: org.id, runnerId }),
    enabled: !!org?.id && !!runnerId,
  })

  const { data: runnerHealthCheck, isLoading: isHealthCheckLoading } = useQuery({
    queryKey: ['runner-health-checks', org?.id, runnerId],
    queryFn: () => getRunnerRecentHealthChecks({ orgId: org.id, runnerId }),
    enabled: !!org?.id && !!runnerId,
  })

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-4">
        <Text variant="base" weight="strong">
          Install runner
        </Text>

        <Text variant="subtext">
          <Link href={`/${org.id}/installs/${step.owner_id}/runner`}>
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
