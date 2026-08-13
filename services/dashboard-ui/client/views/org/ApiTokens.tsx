import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { ApiTokensTable } from '@/components/api-tokens/ApiTokensTable'
import { CreateApiTokenButton } from '@/components/api-tokens/CreateApiToken'

import { useOrg } from '@/hooks/use-org'

export const ApiTokens = () => {
  const { org } = useOrg()

  return (
    <>
      <PageTitle title={`API tokens | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${org.id}`,
            text: org?.name,
          },
          {
            path: `/${org.id}/settings`,
            text: 'Settings',
          },
          {
            path: `/${org.id}/settings/api-tokens`,
            text: 'API tokens',
          },
        ]}
      />
      <PageHeader className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            API tokens
          </Text>
          <Text theme="neutral">
            Manage static tokens for accessing the Nuon API in this org.
          </Text>
        </HeadingGroup>
        <CreateApiTokenButton variant="primary" />
      </PageHeader>
      <PageContent>
        <PageSection>
          <ApiTokensTable shouldPoll />
        </PageSection>
      </PageContent>
    </>
  )
}
