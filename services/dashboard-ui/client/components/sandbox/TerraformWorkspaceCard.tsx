import { useQuery } from '@tanstack/react-query'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { UnlockSandboxTerraformStateButton } from '@/components/sandbox/management/UnlockSandboxTerraformState'
import { TerraformState } from '@/components/terraform-workspace/TerraformState'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import {
  getTerraformState,
  getTerraformStates,
  getTerraformWorkspaceLock,
} from '@/lib'
import type { TTerraformState } from '@/types'

function getResourceAddresses(
  rootModule: TTerraformState['values']['root_module']
): string[] {
  const addresses: string[] = []

  if (rootModule?.resources) {
    for (const res of rootModule.resources) {
      if (res.address) addresses.push(res.address)
    }
  }

  if (rootModule?.child_modules) {
    for (const mod of rootModule.child_modules) {
      if (mod.resources) {
        for (const res of mod.resources) {
          if (res.address) addresses.push(res.address)
        }
      }
    }
  }

  return addresses
}

export const TerraformWorkspaceCard = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const workspaceId = install?.sandbox?.terraform_workspace?.id

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

  const { data: lock } = useQuery({
    queryKey: ['terraform-workspace-lock', org?.id, workspaceId],
    queryFn: () =>
      getTerraformWorkspaceLock({
        orgId: org.id,
        workspaceId: workspaceId!,
      }),
    enabled: !!org?.id && !!workspaceId,
  })

  if (!workspaceId) return null

  const resources = currentRevision
    ? getResourceAddresses(currentRevision?.values?.root_module)
    : []
  const outputs = currentRevision?.values?.outputs || {}
  const outputKeys = Object.keys(outputs)

  const hasData = resources.length > 0 || outputKeys.length > 0

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <Text variant="base" weight="strong">
          Terraform state
        </Text>
        {lock ? <UnlockSandboxTerraformStateButton /> : null}
      </div>

      {!hasData ? (
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
