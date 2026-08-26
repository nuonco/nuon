import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { ListPage } from '@/components/layout/ListPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import {
  CreateOIDCTrustPolicyButton,
  OIDCTrustPoliciesTable,
} from '@/components/oidc-trust-policies'
import { useOrg } from '@/hooks/use-org'
import { getCurrentOrgOIDCTrustPolicies } from '@/lib'

export const OIDCTrustPolicies = () => {
  const { org } = useOrg()

  const { data: policies } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['oidc-trust-policies', org.id],
    queryFn: () => getCurrentOrgOIDCTrustPolicies({ orgId: org.id }),
  })

  return (
    <>
      <PageTitle title="OIDC federation" />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org?.name },
          { path: `/${org.id}/settings`, text: 'Settings' },
          { path: `/${org.id}/settings/oidc`, text: 'OIDC federation' },
        ]}
      />
      <ListPage
        title="OIDC federation"
        description="Grant OIDC providers access to the Nuon control plane without storing long-lived static tokens."
        createAction={
          <CreateOIDCTrustPolicyButton
            reservedNames={(policies ?? []).map((policy) => policy.name ?? '')}
          />
        }
      >
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Trust policies
          </Text>
          <OIDCTrustPoliciesTable shouldPoll />
        </div>
      </ListPage>
    </>
  )
}
