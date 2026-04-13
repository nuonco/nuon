import { useQuery } from '@tanstack/react-query'
import { TerraformBackendConfigButton } from '@/components/terraform-workspace/TerraformBackendConfig'
import { UnlockTerraformWorkspaceButton } from '@/components/terraform-workspace/UnlockTerraformWorkspace'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import {
  getTerraformState,
  getTerraformStates,
  getWorkspaceStateRaw,
} from '@/lib'
import type { TComponentType } from '@/types'
import { TerraformWorkspaceCard } from './TerraformWorkspaceCard'

export const TerraformWorkspaceCardContainer = ({
  workspaceId: workspaceIdProp,
  description,
  componentType,
}: {
  workspaceId?: string
  description?: string
  componentType?: TComponentType
} = {}) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const workspaceId =
    workspaceIdProp ?? install?.sandbox?.terraform_workspace?.id

  const isPulumi = componentType === 'pulumi'

  const { data: states } = useQuery({
    queryKey: ['workspace-states', org?.id, workspaceId],
    queryFn: () =>
      getTerraformStates({
        orgId: org.id,
        workspaceId: workspaceId!,
      }),
    enabled: !!org?.id && !!workspaceId,
  })

  const latestStateId = states?.[0]?.id

  // For terraform, use the parsed state endpoint.
  // For pulumi, use the raw endpoint (pulumi state isn't terraform JSON).
  const { data: currentRevision } = useQuery({
    queryKey: ['workspace-state', org?.id, workspaceId, latestStateId, isPulumi],
    queryFn: () =>
      isPulumi
        ? getWorkspaceStateRaw({
            orgId: org.id,
            workspaceId: workspaceId!,
            stateId: latestStateId!,
          })
        : getTerraformState({
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
      componentType={componentType}
      actions={
        isPulumi ? undefined : (
          <>
            <TerraformBackendConfigButton
              workspaceId={workspaceId}
              variant="secondary"
              size="sm"
            >
              Use Terraform CLI
            </TerraformBackendConfigButton>
            <UnlockTerraformWorkspaceButton
              workspaceId={workspaceId}
              description={description}
              size="sm"
            />
          </>
        )
      }
    />
  )
}
