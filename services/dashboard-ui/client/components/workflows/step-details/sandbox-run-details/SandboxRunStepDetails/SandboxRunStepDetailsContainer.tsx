import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getInstallSandboxRun } from '@/lib'
import type { TWorkflowStep, TSandboxRun } from '@/types'
import { SandboxRunStepDetails } from './SandboxRunStepDetails'

interface ISandboxRunStepDetailsContainer {
  step?: TWorkflowStep
}

export const SandboxRunStepDetailsContainer = ({ step }: ISandboxRunStepDetailsContainer) => {
  const { org } = useOrg()

  const { data: sandboxRun, isLoading } = useQuery<TSandboxRun>({
    queryKey: ['sandbox-run', org?.id, step?.step_target_id],
    queryFn: () =>
      getInstallSandboxRun({ orgId: org!.id, runId: step!.step_target_id }),
    enabled: !!org?.id && !!step?.step_target_id,
  })

  return (
    <SandboxRunStepDetails
      step={step}
      orgId={org?.id}
      sandboxRun={sandboxRun}
      isLoading={isLoading}
    />
  )
}
