import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getRunnerProcesses, getRunnerSettings } from '@/lib'
import { RunnerProvider } from '@/providers/runner-provider'
import type { IStepDetails } from '../types'
import { RunnerStepDetails } from './RunnerStepDetails'

interface IRunnerStepDetailsContainer extends IStepDetails {}

export const RunnerStepDetailsContainer = ({ step }: IRunnerStepDetailsContainer) => {
  const { org } = useOrg()
  const runnerId = step?.step_target_id

  const { data: settings } = useQuery({
    queryKey: ['runner-settings', org?.id, runnerId],
    queryFn: () => getRunnerSettings({ orgId: org!.id, runnerId: runnerId! }),
    enabled: !!org?.id && !!runnerId,
  })

  const { data: processResult, isLoading: processesLoading } = useQuery({
    queryKey: ['runner-processes-active', org?.id, runnerId],
    queryFn: () =>
      getRunnerProcesses({
        orgId: org!.id,
        runnerId: runnerId!,
        status: 'pending,active,offline,pending-shutdown',
        limit: 2,
      }),
    refetchInterval: 10000,
    enabled: !!org?.id && !!runnerId,
  })

  const processes = processResult?.data ?? []

  return (
    <RunnerProvider runnerId={runnerId}>
      <RunnerStepDetails
        step={step}
        orgId={org?.id}
        processes={processes}
        processesLoading={processesLoading}
        settings={settings}
      />
    </RunnerProvider>
  )
}
