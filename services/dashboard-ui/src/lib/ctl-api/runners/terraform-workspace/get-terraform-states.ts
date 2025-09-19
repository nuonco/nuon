import { api } from '@/lib/api'
import type { TTerraformWorkspaceState, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getTerraformStates = ({
  workspaceId,
  limit = 10,
  offset,
  page,
  orgId,
}: {
  workspaceId: string
  orgId: string
  page?: number
} & TPaginationParams) =>
  api<TTerraformWorkspaceState[]>({
    path: `runners/terraform-workspace/${workspaceId}/state-json${buildQueryParams({ limit, offset, page })}`,
    orgId,
  })
