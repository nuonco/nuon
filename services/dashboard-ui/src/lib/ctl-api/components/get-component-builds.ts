import type { TBuild, TPaginationParams } from '@/types'
import { api } from '@/lib/api'
import { buildQueryParams } from '@/utils/build-query-params'

export interface IGetComponentBuilds extends TPaginationParams {
  componentId: string
  orgId: string
}

export async function getComponentBuilds({
  componentId,
  orgId,
  limit,
  offset,
}: IGetComponentBuilds) {
  return api<TBuild[]>({
    orgId,
    path: `builds${buildQueryParams({ limit, offset, component_id: componentId })}`,
  })
}
