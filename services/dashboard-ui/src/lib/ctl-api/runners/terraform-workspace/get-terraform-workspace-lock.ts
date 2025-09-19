import { api } from '@/lib/api'
import type { TTerraformWorkspaceLock } from '@/types'

export const getTerraformWorkspaceLock = ({
  workspaceId,
  orgId,
}: {
  workspaceId: string
  orgId: string
}) =>
  api<TTerraformWorkspaceLock>({
    path: `terraform-workspaces/${workspaceId}/lock`,
    orgId,
  })
