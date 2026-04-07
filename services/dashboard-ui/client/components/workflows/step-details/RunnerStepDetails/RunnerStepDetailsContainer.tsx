import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getRunner, getRunnerLatestHeartbeat, getRunnerRecentHealthChecks } from '@/lib'
import type { IStepDetails } from '../types'
import { RunnerStepDetails } from './RunnerStepDetails'

interface IRunnerStepDetailsContainer extends IStepDetails {}

export const RunnerStepDetailsContainer = ({ step }: IRunnerStepDetailsContainer) => {
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
    <RunnerStepDetails
      step={step}
      orgId={org?.id}
      runner={runner}
      runnerHeartbeat={runnerHeartbeat}
      runnerHealthCheck={runnerHealthCheck}
      isRunnerLoading={isRunnerLoading}
      isHeartbeatLoading={isHeartbeatLoading}
      isHealthCheckLoading={isHealthCheckLoading}
    />
  )
}
