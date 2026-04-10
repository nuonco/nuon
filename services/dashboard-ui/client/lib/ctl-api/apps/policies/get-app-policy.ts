import { api } from '@/lib/api'
import type { TAppPolicyConfig } from '@/types'

export const getAppPolicy = ({
  appId,
  orgId,
  policyId,
}: {
  appId: string
  orgId: string
  policyId: string
}) => {
  return api<TAppPolicyConfig>({
    path: `apps/${appId}/policy-config/${policyId}`,
    orgId,
  })
}
