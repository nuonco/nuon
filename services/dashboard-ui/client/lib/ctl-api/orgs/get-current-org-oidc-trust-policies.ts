import { api } from '@/lib/api'
import type { TOIDCTrustPolicy } from '@/types'

export const getCurrentOrgOIDCTrustPolicies = ({ orgId }: { orgId: string }) =>
  api<TOIDCTrustPolicy[]>({
    orgId,
    path: `oidc/trust-policies`,
  })
