import { api } from '@/lib/api'
import type { TAction, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export interface IGetActions extends TPaginationParams {
  appId: string
  branch_id?: string
  labels?: string
  orgId: string
  q?: string
  trigger_types?: string
}

export async function getActions({
  appId,
  branch_id,
  labels,
  orgId,
  limit,
  offset,
  q,
  trigger_types,
}: IGetActions) {
  return api<TAction[]>({
    orgId,
    path: `apps/${appId}/action-workflows${buildQueryParams({ limit, offset, q, labels, trigger_types, branch_id })}`,
    paginated: true,
  })
}
