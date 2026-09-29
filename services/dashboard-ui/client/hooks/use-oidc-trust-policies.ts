import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getCurrentOrgOIDCTrustPolicies } from '@/lib'

export const useOIDCTrustPolicies = ({ enabled = true } = {}) => {
  const { org } = useOrg()

  return useQuery({
    queryKey: ['oidc-trust-policies', org?.id],
    queryFn: () => getCurrentOrgOIDCTrustPolicies({ orgId: org!.id }),
    enabled: enabled && !!org?.id,
    staleTime: 60 * 1000,
  })
}
