import { EmptyState } from '@/components/common/EmptyState'
import { TerraformOutputs } from '@/components/terraform-outputs/TerraformOutputs'
import { useSandboxRun } from '@/hooks/use-sandbox-run'

export const SandboxRunOutputsTab = () => {
  const { sandboxRun } = useSandboxRun()
  const outputs = sandboxRun?.outputs

  if (!outputs || Object.keys(outputs).length === 0) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No outputs"
        emptyMessage="No outputs available for this sandbox run."
      />
    )
  }

  return <TerraformOutputs heading="Sandbox run outputs" outputs={outputs} />
}
