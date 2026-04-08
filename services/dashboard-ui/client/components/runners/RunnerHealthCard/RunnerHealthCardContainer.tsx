import { useQuery } from '@tanstack/react-query'
import type { ICard } from '@/components/common/Card'
import { useOrg } from '@/hooks/use-org'
import { useRunner } from '@/hooks/use-runner'
import { getRunnerRecentHealthChecks } from '@/lib'
import type { TRunnerHealthCheck } from '@/types'
import { RunnerHealthCard as RunnerHealthCardComponent } from './RunnerHealthCard'

interface IRunnerHealthCardContainer extends Omit<ICard, 'children'> {
  initHealthchecks?: TRunnerHealthCheck[]
  shouldPoll?: boolean
  pollInterval?: number
}

export const RunnerHealthCard = ({
  initHealthchecks,
  shouldPoll = false,
  pollInterval = 60000,
  ...props
}: IRunnerHealthCardContainer) => {
  const { org } = useOrg()
  const { runner } = useRunner()

  const { data: healthchecks, isLoading } = useQuery({
    queryKey: ['runner-health-checks', org?.id, runner?.id],
    queryFn: () => getRunnerRecentHealthChecks({ orgId: org.id, runnerId: runner.id }),
    initialData: initHealthchecks,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!runner?.id,
  })

  return (
    <RunnerHealthCardComponent
      healthchecks={healthchecks}
      isLoading={isLoading && !initHealthchecks}
      {...props}
    />
  )
}
