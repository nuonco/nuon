import { api } from '@/lib/api'
import type { TAction, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export interface IGetActions extends TPaginationParams {
  appId: string
  labels?: string
  orgId: string
  q?: string
}

export async function getActions({
  appId,
  labels,
  orgId,
  limit,
  offset,
  q,
}: IGetActions) {
  return api<TAction[]>({
    orgId,
    path: `apps/${appId}/action-workflows${buildQueryParams({ limit, offset, q, labels })}`,
    paginated: true,
  })
}
