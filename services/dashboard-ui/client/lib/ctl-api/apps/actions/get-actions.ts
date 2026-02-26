import { api } from '@/lib/api'
import type { TAction, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export interface IGetActions extends TPaginationParams {
  appId: string
  orgId: string
  q?: string
}

export async function getActions({
  appId,
  orgId,
  limit,
  offset,
  q,
}: IGetActions) {
  return api<TAction[]>({
    orgId,
    path: `apps/${appId}/action-workflows${buildQueryParams({ limit, offset, q })}`,
  })
}
