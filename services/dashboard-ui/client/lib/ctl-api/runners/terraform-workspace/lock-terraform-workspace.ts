import { api } from '@/lib/api'
import type { TTerraformState } from '@/types'

export async function lockTerraformWorkspace({
  orgId,
  terraformWorkspaceId,
}: {
  terraformWorkspaceId: string
  orgId: string
}) {
  return api<TTerraformState>({
    body: {},
    method: 'POST',
    orgId,
    path: `terraform-workspaces/${terraformWorkspaceId}/lock`,
  })
}
