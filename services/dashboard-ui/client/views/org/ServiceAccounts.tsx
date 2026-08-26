import { ListPage } from '@/components/layout/ListPage'
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
      <ListPage
        title="Service accounts"
        description="Manage machine users."
        createAction={<CreateServiceAccountButton variant="primary" />}
      >
        <div className="flex items-center justify-end">
          <ShowRunnerAccounts />
        </div>
        <ServiceAccountsTable shouldPoll />
      </ListPage>
    </>
  )
}
