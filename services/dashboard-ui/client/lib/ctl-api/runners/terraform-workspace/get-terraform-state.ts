import { api } from '@/lib/api'
import type { TTerraformState } from '@/types'

export const getTerraformState = ({
  workspaceId,
  stateId,
  orgId,
}: {
  workspaceId: string
  stateId: string
  orgId: string
}) =>
  api<TTerraformState>({
    path: `runners/terraform-workspace/${workspaceId}/state-json/${stateId}`,
    orgId,
  })
