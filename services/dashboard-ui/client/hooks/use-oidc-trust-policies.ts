import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getCurrentOrgOIDCTrustPolicies } from '@/lib'

// Same query key the trust policies table uses, so creating a policy from
// anywhere refreshes every view of the list.
export const useOIDCTrustPolicies = ({ enabled = true } = {}) => {
  const { org } = useOrg()

  return useQuery({
    queryKey: ['oidc-trust-policies', org?.id],
    queryFn: () => getCurrentOrgOIDCTrustPolicies({ orgId: org!.id }),
    enabled: enabled && !!org?.id,
    staleTime: 60 * 1000,
  })
}
