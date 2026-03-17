import { api } from '@/lib/api'
import { buildQueryParams } from '@/utils/build-query-params'
import type { TAppSandboxBuild, TPaginationParams } from '@/types'

export interface IGetSandboxBuilds extends TPaginationParams {
  appId: string
  orgId: string
}

export async function getSandboxBuilds({
  appId,
  orgId,
  limit,
  offset,
}: IGetSandboxBuilds) {
  return api<TAppSandboxBuild[]>({
    orgId,
    path: `apps/${appId}/sandbox/builds${buildQueryParams({ limit, offset })}`,
    paginated: true,
  })
}
