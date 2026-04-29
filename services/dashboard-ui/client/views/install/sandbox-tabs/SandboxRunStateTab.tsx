import { TerraformWorkspaceCard } from '@/components/terraform-workspace/TerraformWorkspaceCard'
import { EmptyState } from '@/components/common/EmptyState'
import { useInstall } from '@/hooks/use-install'

export const SandboxRunStateTab = () => {
  const { install } = useInstall()
  const workspaceId = install?.sandbox?.terraform_workspace?.id

  if (!workspaceId) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No state available"
        emptyMessage="No Terraform workspace found for this sandbox."
      />
    )
  }

  return (
    <TerraformWorkspaceCard
      workspaceId={workspaceId}
      componentType="terraform_module"
    />
  )
}
