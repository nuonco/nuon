import { EmptyState } from '@/components/common/EmptyState'
import { PageTitle } from '@/components/navigation/PageTitle'
import { TerraformOutputs } from '@/components/terraform-outputs/TerraformOutputs'
import { useInstall } from '@/hooks/use-install'
import { useSandboxRun } from '@/hooks/use-sandbox-run'

export const SandboxRunOutputsTab = () => {
  const { sandboxRun } = useSandboxRun()
  const { install } = useInstall()
  const outputs = sandboxRun?.outputs

  return (
    <>
      <PageTitle segments={['Sandbox run outputs', install?.name]} />
      {!outputs || Object.keys(outputs).length === 0 ? (
        <EmptyState
          variant="table"
          emptyTitle="No outputs"
          emptyMessage="No outputs available for this sandbox run."
        />
      ) : (
        <TerraformOutputs heading="Sandbox run outputs" outputs={outputs} />
      )}
    </>
  )
}
