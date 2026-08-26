import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { ServiceAccountsTable } from '@/components/service-accounts/ServiceAccountsTable'
import { CreateServiceAccountButton } from '@/components/service-accounts/CreateServiceAccount'
import { ShowRunnerAccountsContainer as ShowRunnerAccounts } from '@/components/service-accounts/filters/ShowRunnerAccounts'

import { useOrg } from '@/hooks/use-org'

export const ServiceAccounts = () => {
  const { org } = useOrg()

  return (
    <>
      <PageTitle title="Service accounts" />
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
            path: `/${org.id}/settings/service-accounts`,
            text: 'Service accounts',
          },
        ]}
      />
      <PageHeader className="flex items-center justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Service accounts
          </Text>
          <Text theme="neutral">
            Manage machine users.
          </Text>
        </HeadingGroup>
        <CreateServiceAccountButton variant="primary" />
      </PageHeader>
      <PageContent>
        <PageSection>
          <div className="flex items-center justify-end">
            <ShowRunnerAccounts />
          </div>
          <ServiceAccountsTable shouldPoll />
        </PageSection>
      </PageContent>
    </>
  )
}
