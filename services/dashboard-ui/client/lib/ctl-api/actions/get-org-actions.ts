import { api } from '@/lib/api'
import type { TAction, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

// getOrgActions lists every action workflow visible to the calling org.
// Mirrors the slack subscribe modal's actions picker (org-isolated via
// JOIN apps → apps.org_id) so the dashboard MatchPicker shares the same
// backing query and search semantics.
export interface IGetOrgActions extends TPaginationParams {
  orgId: string
  q?: string
}

export const getOrgActions = ({ orgId, q, limit, offset }: IGetOrgActions) =>
  api<TAction[]>({
    orgId,
    path: `action-workflows${buildQueryParams({ limit, offset, q })}`,
    paginated: true,
  })
