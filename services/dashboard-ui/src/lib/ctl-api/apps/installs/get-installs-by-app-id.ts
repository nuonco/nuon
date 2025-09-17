import { api } from '@/lib/api'
import type { TInstall, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export interface IGetInstallsByAppId extends TPaginationParams {
  appId: string
  orgId: string
  q?: string
}

export async function getInstallsByAppId({
  appId,
  orgId,
  limit,
  offset,
  q,
}: IGetInstallsByAppId) {
  return api<TInstall[]>({
    orgId,
    path: `apps/${appId}/installs${buildQueryParams({ limit, offset, q })}`,
  })
}
