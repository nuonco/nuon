import { api } from '@/lib/api'
import type { TAppPoliciesConfig } from '@/types'

export const getAppPoliciesConfigs = ({
  appId,
  orgId,
}: {
  orgId: string
  appId: string
}) =>
  api<TAppPoliciesConfig[]>({
    path: `apps/${appId}/policies-configs`,
    orgId,
  })
