import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSandboxRun } from '@/hooks/use-sandbox-run'
import type { TWorkflow } from '@/types'
import { SandboxHeader } from './SandboxHeader'

interface ISandboxHeaderContainer {
  workflow: TWorkflow
  stepId: string
  flush?: boolean
}

export const SandboxHeaderContainer = ({
  workflow,
  stepId,
  flush,
}: ISandboxHeaderContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { sandboxRun } = useSandboxRun()

  return (
    <SandboxHeader
      workflow={workflow}
      stepId={stepId}
      sandboxRun={sandboxRun}
      install={install}
      orgId={org?.id}
      flush={flush}
    />
  )
}
