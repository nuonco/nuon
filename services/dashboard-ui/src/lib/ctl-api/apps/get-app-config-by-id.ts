import { api } from '@/lib/api'
import type { TAppConfig } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export const getAppConfigById = ({
  appId,
  appConfigId,
  orgId,
  recurse,
}: {
  orgId: string
  appId: string
  appConfigId: string
  recurse?: boolean
}) =>
  api<TAppConfig>({
    path: `apps/${appId}/config/${appConfigId}${buildQueryParams({ recurse })}`,
    orgId,
  })
