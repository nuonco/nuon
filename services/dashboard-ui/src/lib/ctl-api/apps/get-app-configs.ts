import { api } from '@/lib/api'
import type { TAppConfig, TPaginationParams } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getAppConfigs = ({
  appId,
  limit,
  offset,
  orgId,
}: { orgId: string; appId: string } & TPaginationParams) =>
  api<TAppConfig[]>({
    path: `apps/${appId}/configs${buildQueryParams({ limit, offset })}`,
    orgId,
  })
