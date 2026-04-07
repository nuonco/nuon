import { useQuery } from '@tanstack/react-query'
import { TerraformBackendConfigButton } from '@/components/terraform-workspace/TerraformBackendConfig'
import { UnlockTerraformWorkspaceButton } from '@/components/terraform-workspace/UnlockTerraformWorkspace'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getTerraformState, getTerraformStates } from '@/lib'
import { TerraformWorkspaceCard } from './TerraformWorkspaceCard'

export const TerraformWorkspaceCardContainer = ({
  workspaceId: workspaceIdProp,
  description,
}: { workspaceId?: string; description?: string } = {}) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const workspaceId = workspaceIdProp ?? install?.sandbox?.terraform_workspace?.id

  const { data: states } = useQuery({
    queryKey: ['terraform-states', org?.id, workspaceId],
    queryFn: () =>
      getTerraformStates({
        orgId: org.id,
        workspaceId: workspaceId!,
      }),
    enabled: !!org?.id && !!workspaceId,
  })

  const latestStateId = states?.[0]?.id

  const { data: currentRevision } = useQuery({
    queryKey: ['terraform-state', org?.id, workspaceId, latestStateId],
    queryFn: () =>
      getTerraformState({
        orgId: org.id,
        workspaceId: workspaceId!,
        stateId: latestStateId!,
      }),
    enabled: !!org?.id && !!workspaceId && !!latestStateId,
  })

  if (!workspaceId) return null

  return (
    <TerraformWorkspaceCard
      currentRevision={currentRevision}
      actions={
        <>
          <TerraformBackendConfigButton workspaceId={workspaceId} variant="secondary" size="sm">
            Use Terraform CLI
          </TerraformBackendConfigButton>
          <UnlockTerraformWorkspaceButton
            workspaceId={workspaceId}
            description={description}
            size="sm"
          />
        </>
      }
    />
  )
}
