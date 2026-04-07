import { useQuery } from '@tanstack/react-query'
import type { ICard } from '@/components/common/Card'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { getRunnerLatestHeartbeat } from '@/lib'
import type { TRunnerGroup, TRunnerMngHeartbeat } from '@/types'
import { RunnerDetailsCard as RunnerDetailsCardComponent } from './RunnerDetailsCard'

interface IRunnerDetailsCardContainer extends Omit<ICard, 'children'> {
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
}: IRunnerDetailsCardContainer) => {
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
    <RunnerDetailsCardComponent
      runner={runner}
      runnerGroup={runnerGroup}
      heartbeat={runnerHeartbeat}
      {...props}
    />
  )
}
