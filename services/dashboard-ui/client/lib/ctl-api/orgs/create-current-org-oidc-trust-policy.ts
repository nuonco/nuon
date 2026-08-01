import { api } from '@/lib/api'
import type { TCreateOIDCTrustPolicyBody, TOIDCTrustPolicy } from '@/types'

export const createCurrentOrgOIDCTrustPolicy = ({
  body,
  orgId,
}: {
  body: TCreateOIDCTrustPolicyBody
  orgId: string
}) =>
  api<TOIDCTrustPolicy>({
    body,
    method: 'POST',
    orgId,
    path: `oidc/trust-policies`,
  })
