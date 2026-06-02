import { TerraformWorkspaceCard } from '@/components/terraform-workspace/TerraformWorkspaceCard'
import { EmptyState } from '@/components/common/EmptyState'
import { useInstall } from '@/hooks/use-install'
import { useSandboxRun } from '@/hooks/use-sandbox-run'

export const SandboxRunStateTab = () => {
  const { install } = useInstall()
  const { sandboxRun } = useSandboxRun()

  const isPulumi = sandboxRun?.app_sandbox_config?.type === 'pulumi'
  const workspaceId = install?.sandbox?.terraform_workspace?.id

  if (!workspaceId) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No state available"
        emptyMessage={
          isPulumi
            ? 'No Pulumi workspace found for this sandbox.'
            : 'No Terraform workspace found for this sandbox.'
        }
      />
    )
  }

  return (
    <TerraformWorkspaceCard
      workspaceId={workspaceId}
      componentType={isPulumi ? 'pulumi' : 'terraform_module'}
    />
  )
}
