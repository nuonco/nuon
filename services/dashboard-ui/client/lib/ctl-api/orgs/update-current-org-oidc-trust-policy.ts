import { api } from '@/lib/api'
import type { TOIDCTrustPolicy, TUpdateOIDCTrustPolicyBody } from '@/types'

export const updateCurrentOrgOIDCTrustPolicy = ({
  body,
  orgId,
  policyId,
}: {
  body: TUpdateOIDCTrustPolicyBody
  orgId: string
  policyId: string
}) =>
  api<TOIDCTrustPolicy>({
    body,
    method: 'PATCH',
    orgId,
    path: `oidc/trust-policies/${policyId}`,
  })
