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
}): Promise<{
  data: TAppPolicyConfig | undefined
  error: { error: string } | undefined
  status: number
}> => {
  const {
    data: policiesConfigs,
    error,
    status,
  } = await api<TAppPoliciesConfig[]>({
    path: `apps/${appId}/policies-configs`,
    orgId,
  })

  if (error) {
    return { data: undefined, error, status }
  }

  const latestConfig = policiesConfigs
    ?.slice()
    .sort((a, b) => {
      const dateA = a.created_at ? new Date(a.created_at).getTime() : 0
      const dateB = b.created_at ? new Date(b.created_at).getTime() : 0
      return dateB - dateA
    })
    .at(0)

  const policy = latestConfig?.policies?.find((p) => p.id === policyId)

  if (!policy) {
    return {
      data: undefined,
      error: { error: 'Policy not found' },
      status: 404,
    }
  }

  return { data: policy, error: undefined, status: 200 }
}
