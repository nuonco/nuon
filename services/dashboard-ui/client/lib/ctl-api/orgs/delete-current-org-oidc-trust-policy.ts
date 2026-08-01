import { api } from '@/lib/api'

export const deleteCurrentOrgOIDCTrustPolicy = ({
  orgId,
  policyId,
}: {
  orgId: string
  policyId: string
}) =>
  api({
    method: 'DELETE',
    orgId,
    path: `oidc/trust-policies/${policyId}`,
  })
