import { api } from '@/lib/api'
import type { TComponent, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export interface IGetComponents extends TPaginationParams {
  appId: string
  orgId: string
  q?: string
  types?: string
}

export async function getComponents({
  appId,
  orgId,
  limit,
  offset,
  q,
  types,
}: IGetComponents) {
  return api<TComponent[]>({
    orgId,
    path: `apps/${appId}/components${buildQueryParams({ limit, offset, q, types })}`,
  })
}
