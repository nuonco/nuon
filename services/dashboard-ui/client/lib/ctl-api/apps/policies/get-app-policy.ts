import { api } from '@/lib/api'
import type { TAppPoliciesConfig, TAppPolicyConfig } from '@/types'

export const getAppPolicy = async ({
  appId,
  orgId,
  policyId,
}: {
  appId: string
  orgId: string
  policyId: string
}): Promise<TAppPolicyConfig | undefined> => {
  const policiesConfigs = await api<TAppPoliciesConfig[]>({
    path: `apps/${appId}/policies-configs`,
    orgId,
  })

  const latestConfig = policiesConfigs
    ?.slice()
    .sort((a, b) => {
      const dateA = a.created_at ? new Date(a.created_at).getTime() : 0
      const dateB = b.created_at ? new Date(b.created_at).getTime() : 0
      return dateB - dateA
    })
    .at(0)

  return latestConfig?.policies?.find((p) => p.id === policyId)
}
