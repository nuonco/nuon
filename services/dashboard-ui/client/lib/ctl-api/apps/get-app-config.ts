import { api } from '@/lib/api'
import type { TAppConfig } from '@/types'
import { buildQueryParams } from '@/utils/build-query-params'

export type TAppConfigWithIntermediate = TAppConfig & {
  intermediate_config_json?: string
}

export const getAppConfig = ({
  appId,
  appConfigId,
  orgId,
  recurse,
  includeIntermediate,
}: {
  orgId: string
  appId: string
  appConfigId: string
  recurse?: boolean
  includeIntermediate?: boolean
}) =>
  api<TAppConfigWithIntermediate>({
    path: `apps/${appId}/configs/${appConfigId}${buildQueryParams({
      recurse,
      include_intermediate: includeIntermediate,
    })}`,
    orgId,
  })
