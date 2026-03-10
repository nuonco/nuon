import { useQuery } from '@tanstack/react-query'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { TerraformBackendConfigButton } from '@/components/terraform-workspace/TerraformBackendConfig'
import { TerraformState } from '@/components/terraform-workspace/TerraformState'
import { UnlockTerraformWorkspaceButton } from '@/components/terraform-workspace/UnlockTerraformWorkspace'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getTerraformState, getTerraformStates } from '@/lib'

export const TerraformWorkspaceCard = ({
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
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <Text variant="base" weight="strong">
          Terraform state
        </Text>
        <div className="flex items-center gap-2">
          <TerraformBackendConfigButton workspaceId={workspaceId} variant="secondary" size="sm">
            Use Terraform CLI
          </TerraformBackendConfigButton>
          <UnlockTerraformWorkspaceButton
            workspaceId={workspaceId}
            description={description}
            size="sm"
          />
        </div>
      </div>

      {!currentRevision ? (
        <EmptyState
          variant="diagram"
          emptyTitle="No revisions yet"
          emptyMessage="The workspace has been created, but no state has been written."
        />
      ) : (
        <TerraformState terraformState={currentRevision} />
      )}
    </div>
  )
}
