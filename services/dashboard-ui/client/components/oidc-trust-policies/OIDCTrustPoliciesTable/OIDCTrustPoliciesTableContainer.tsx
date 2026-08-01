import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getCurrentOrgOIDCTrustPolicies } from '@/lib'
import { OIDCTrustPoliciesTable } from './OIDCTrustPoliciesTable'

export const OIDCTrustPoliciesTableContainer = ({
  pollInterval = 20000,
  shouldPoll = true,
}: {
  pollInterval?: number
  shouldPoll?: boolean
} = {}) => {
  const { org } = useOrg()

  const { data, isLoading } = useQuery({
    queryKey: ['oidc-trust-policies', org.id],
    queryFn: () => getCurrentOrgOIDCTrustPolicies({ orgId: org.id }),
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  return <OIDCTrustPoliciesTable data={data ?? []} isLoading={isLoading} />
}
