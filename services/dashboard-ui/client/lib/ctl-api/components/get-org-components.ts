import { api } from '@/lib/api'
import type { TComponent, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

// getOrgComponents lists every component visible to the calling org. Mirrors
// the slack subscribe modal's components picker (org-isolated via JOIN apps
// → apps.org_id) so the dashboard MatchPicker shares the same backing
// query and search semantics.
export interface IGetOrgComponents extends TPaginationParams {
  orgId: string
  q?: string
  component_ids?: string
}

export const getOrgComponents = ({
  orgId,
  q,
  component_ids,
  limit,
  offset,
}: IGetOrgComponents) =>
  api<TComponent[]>({
    orgId,
    path: `components${buildQueryParams({ limit, offset, q, component_ids })}`,
    paginated: true,
  })
