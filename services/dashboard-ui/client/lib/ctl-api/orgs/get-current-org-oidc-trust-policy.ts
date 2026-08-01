import { api } from '@/lib/api'
import type { TOIDCTrustPolicy } from '@/types'

export const getCurrentOrgOIDCTrustPolicy = ({
  orgId,
  policyId,
}: {
  orgId: string
  policyId: string
}) =>
  api<TOIDCTrustPolicy>({
    orgId,
    path: `oidc/trust-policies/${policyId}`,
  })
