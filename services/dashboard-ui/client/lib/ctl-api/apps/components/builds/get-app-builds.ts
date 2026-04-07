import type { TBuild } from '@/types'
import { api } from '@/lib/api'
import { buildQueryParams } from '@/utils/build-query-params'

export async function getAppBuilds({
  appId,
  orgId,
  limit = 50,
}: {
  appId: string
  orgId: string
  limit?: number
}) {
  return api<TBuild[]>({
    orgId,
    path: `builds${buildQueryParams({ limit, app_id: appId })}`,
  })
}
