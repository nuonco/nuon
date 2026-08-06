import { api } from '@/lib/api'
import { buildQueryParams } from '@/utils/build-query-params'
import type { TAirgapBundle, TPaginationParams } from '@/types'

export interface IGetAirgapBundles extends TPaginationParams {
  appId: string
  orgId: string
}

export async function getAirgapBundles({
  appId,
  orgId,
  limit,
  offset,
}: IGetAirgapBundles) {
  return api<TAirgapBundle[]>({
    orgId,
    path: `apps/${appId}/airgap-bundles${buildQueryParams({ limit, offset })}`,
    paginated: true,
  })
}
