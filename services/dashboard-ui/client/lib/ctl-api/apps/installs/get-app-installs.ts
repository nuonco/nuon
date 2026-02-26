import { api } from '@/lib/api'
import type { TInstall, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export interface IGetAppInstalls extends TPaginationParams {
  appId: string
  orgId: string
  q?: string
}

export async function getAppInstalls({
  appId,
  orgId,
  limit,
  offset,
  q,
}: IGetAppInstalls) {
  return api<TInstall[]>({
    orgId,
    path: `apps/${appId}/installs${buildQueryParams({ limit, offset, q })}`,
  })
}
