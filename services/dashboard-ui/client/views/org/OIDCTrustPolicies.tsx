import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
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
      <PageTitle title={`OIDC federation | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org?.name },
          { path: `/${org.id}/settings`, text: 'Settings' },
          { path: `/${org.id}/settings/oidc`, text: 'OIDC federation' },
        ]}
      />
      <PageHeader className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            OIDC federation
          </Text>
          <Text theme="neutral">
            Grant OIDC providers access to the Nuon control plane without
            storing long-lived static tokens.
          </Text>
        </HeadingGroup>
        <CreateOIDCTrustPolicyButton
          reservedNames={(policies ?? []).map((policy) => policy.name ?? '')}
        />
      </PageHeader>
      <PageContent>
        <PageSection>
          <div className="flex flex-col gap-4">
            <Text variant="base" weight="strong">
              Trust policies
            </Text>
            <OIDCTrustPoliciesTable shouldPoll />
          </div>
        </PageSection>
      </PageContent>
    </>
  )
}
